# Neko Keys — 密钥管理系统设计文档

## 背景

外贸插件 shopee-super-assistant 使用密钥验证授权，原后端源码丢失。本项目重建密钥管理后台，需完整兼容插件现有 API 调用方式，并新增管理后台以支持批量发放、监控和告警。

## 技术栈

| 层 | 技术 |
|----|------|
| 前端 | Vue3 + TypeScript + Tailwind CSS + Pinia + Element Plus |
| 后端 | Go + Gin + GORM |
| 数据库 | MySQL |
| 缓存 | 进程内内存缓存（`sync.RWMutex` + map），首次查询 MySQL 后写入，状态变更时主动失效 |
| 部署 | Docker + docker-compose |

## 项目结构

```
neko_keys/
├── neko_frontend/     # Vue3 前端
├── neko_backend/      # Go 后端
└── docker-compose.yml
```

## 整体架构

```
neko_frontend (Vue3 SPA)
       │  HTTP/JSON
       ▼
neko_backend (Go + Gin)
  ├── 内存缓存层 (sync.RWMutex + map)
  │         ↑ 首次查询加载，写操作后失效
  └── GORM → MySQL
              ├── key_batches
              ├── keys
              ├── key_logs
              └── admin_users

插件 (shopee-super-assistant)
  └── POST /api/v1/activate  ← Authorization: Bearer nk_xxx
```

---

## 数据模型

### key_batches（批次表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| note | varchar(255) | 批次备注 |
| count | int | 本批次生成数量 |
| type | enum | `one_time` / `multi_use` |
| expires_at | datetime | 必填，1天～180天内 |
| max_uses | int | 仅 `multi_use` 填写 |
| created_at | datetime | 创建时间 |

### keys（密钥表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| batch_id | uint | 关联批次 |
| key | varchar(64) | `nk_` + 随机16位大写字母数字，唯一索引 |
| type | enum | `one_time` / `multi_use` |
| status | enum | `inactive` / `active` / `expired` / `disabled` |
| expires_at | datetime | 必填，继承自批次可单独覆盖 |
| max_uses | int | 仅 `multi_use` 有值 |
| use_count | int | 当前已使用次数，默认0 |
| note | varchar(255) | 单个密钥备注 |
| activated_at | datetime | 首次激活时间 |
| activated_ip | varchar(64) | 首次激活IP |
| created_at | datetime | 创建时间 |

**失效逻辑（满足任一即失效）：**
- `now > expires_at`
- `one_time`：`use_count >= 1`
- `multi_use`：`use_count >= max_uses`

### key_logs（使用日志表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| key_id | uint | 关联 keys.id |
| ip | varchar(64) | 调用者IP |
| action | enum | `activate` / `validate` |
| success | bool | 是否成功 |
| created_at | datetime | 时间 |

### admin_users（管理员表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| username | varchar(64) | 账号 |
| password_hash | varchar(255) | bcrypt |
| created_at | datetime | 创建时间 |

---

## API 接口

### 插件端（公开）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/activate` | 激活密钥，写入 activated_at/activated_ip，递增 use_count |
| POST | `/api/v1/validate` | 仅校验有效性，**不改变任何状态**（插件心跳用） |

**activate 响应格式（兼容插件现有解析）：**
```json
// 成功
{ "success": true, "key_id": 1, "expires_at": "2025-12-31T00:00:00Z" }
// 失败
{ "success": false, "error": "密钥已过期" }
```

### 管理端（需 JWT）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/admin/login` | 管理员登录，返回 JWT |
| GET | `/api/admin/dashboard` | 统计概览 |
| GET | `/api/admin/batches` | 批次列表 |
| POST | `/api/admin/batches` | 新建批次并批量生成密钥 |
| GET | `/api/admin/keys` | 密钥列表（支持按批次/状态/关键词过滤） |
| GET | `/api/admin/keys/:id` | 密钥详情 + 使用日志 |
| PUT | `/api/admin/keys/:id` | 修改备注、手动禁用/启用 |
| GET | `/api/admin/keys/:id/logs` | 该密钥的调用记录 |
| GET | `/api/admin/alerts` | 共享密钥告警列表 |

---

## 前端页面

```
登录页
└── 管理后台（侧边栏布局）
    ├── Dashboard     — 统计卡片 + 告警提示
    ├── 批次管理      — 新建批次、批量生成、列表
    ├── 密钥管理      — 表格：搜索/过滤/状态标签/完整密钥展示/操作
    ├── 密钥详情      — 基本信息 + 使用日志时间线 + IP列表
    └── 共享告警      — 触发共享检测的密钥列表
```

**Dashboard 卡片：** 总密钥数 / 活跃中 / 已过期 / 已禁用 / 今日激活数 / 今日验证数 / 告警数

**新建批次字段：** 批次备注、类型（一次性/多次）、生成数量、有效期（1天～180天）、最大使用次数（多次时显示）

**密钥表格列：** 密钥（完整展示）、类型、状态、批次备注、到期时间、使用次数、操作

---

## 共享检测

每次 activate/validate 时，查询该密钥最近1小时内的不同IP数，超过2个IP则在 `alerts` 接口中标记告警，管理员后台显示警告。

---

## 部署

- `docker-compose.yml` 包含三个服务：`mysql`、`neko_backend`、`neko_frontend`（nginx 静态托管）
- 后端环境变量：DB_DSN、JWT_SECRET、ADMIN_USERNAME、ADMIN_PASSWORD
- 后端监听 `:9002`，保持与插件现有配置一致

---

## 验证方式

1. 启动 `docker-compose up`
2. 访问前端管理页面，登录创建批次并生成密钥
3. 用 curl 模拟插件调用 `POST /api/v1/activate`，验证响应格式与插件兼容
4. 在插件中输入生成的 `nk_` 密钥，确认激活成功
5. 检查密钥日志和共享告警功能
