<template>
  <div>
    <!-- Breadcrumb -->
    <el-breadcrumb separator="/" class="mb-4">
      <el-breadcrumb-item :to="{ path: '/keys' }">密钥管理</el-breadcrumb-item>
      <el-breadcrumb-item>密钥详情</el-breadcrumb-item>
    </el-breadcrumb>

    <div v-if="loading" class="flex justify-center items-center h-64">
      <el-icon class="is-loading text-4xl text-blue-500"><Loading /></el-icon>
    </div>

    <template v-else-if="keyData">
      <!-- Basic info -->
      <el-card shadow="never" class="mb-4 border border-gray-200">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="font-semibold">基本信息</span>
            <el-tag :type="statusType(keyData.status)" size="large">
              {{ statusLabel(keyData.status) }}
            </el-tag>
          </div>
        </template>

        <el-descriptions :column="2" border>
          <el-descriptions-item label="密钥 ID">{{ keyData.id }}</el-descriptions-item>
          <el-descriptions-item label="所属批次">
            {{ keyData.batch?.note || keyData.batch_id }}
          </el-descriptions-item>

          <el-descriptions-item label="密钥内容" :span="2">
            <span class="font-mono text-sm select-all break-all text-blue-700">{{ keyData.key }}</span>
          </el-descriptions-item>

          <el-descriptions-item label="类型">
            <el-tag :type="keyData.type === 'one_time' ? 'warning' : 'success'" size="small">
              {{ keyData.type === 'one_time' ? '一次性' : '多次使用' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="使用次数">
            <span v-if="keyData.type === 'one_time'">{{ keyData.use_count }} 次</span>
            <span v-else>{{ keyData.use_count }} / {{ keyData.max_uses }}</span>
          </el-descriptions-item>

          <el-descriptions-item label="到期时间">{{ formatDate(keyData.expires_at) }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDate(keyData.created_at) }}</el-descriptions-item>

          <el-descriptions-item label="激活时间">
            {{ keyData.activated_at ? formatDate(keyData.activated_at) : '未激活' }}
          </el-descriptions-item>
          <el-descriptions-item label="激活 IP">{{ keyData.activated_ip || '-' }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- Edit note -->
      <el-card shadow="never" class="mb-4 border border-gray-200">
        <template #header><span class="font-semibold">备注</span></template>
        <div class="flex gap-3">
          <el-input
            v-model="noteEdit"
            placeholder="输入备注"
            class="flex-1"
            maxlength="200"
            show-word-limit
          />
          <el-button type="primary" :loading="savingNote" @click="handleSaveNote">保存</el-button>
        </div>
      </el-card>

      <!-- Status actions -->
      <el-card shadow="never" class="mb-4 border border-gray-200">
        <template #header><span class="font-semibold">操作</span></template>
        <div class="flex gap-3">
          <el-button
            v-if="keyData.status !== 'expired'"
            :type="keyData.status === 'disabled' ? 'success' : 'danger'"
            :loading="togglingStatus"
            @click="toggleStatus"
          >
            {{ keyData.status === 'disabled' ? '启用密钥' : '禁用密钥' }}
          </el-button>
          <el-tag v-else type="warning">该密钥已过期，无法操作</el-tag>
        </div>
      </el-card>

      <!-- Logs table -->
      <el-card shadow="never" class="border border-gray-200">
        <template #header><span class="font-semibold">使用日志</span></template>
        <el-table :data="logs" stripe class="w-full">
          <el-table-column label="时间" width="160">
            <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
          </el-table-column>
          <el-table-column prop="ip" label="IP 地址" width="150" />
          <el-table-column label="操作类型" width="100">
            <template #default="{ row }">
              <el-tag :type="row.action === 'activate' ? 'primary' : 'info'" size="small">
                {{ row.action === 'activate' ? '激活' : '验证' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="是否成功" width="100">
            <template #default="{ row }">
              <el-tag :type="row.success ? 'success' : 'danger'" size="small">
                {{ row.success ? '成功' : '失败' }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>

        <div v-if="logs.length === 0" class="text-center text-gray-400 py-8 text-sm">暂无使用日志</div>
      </el-card>
    </template>

    <div v-else class="text-center text-gray-400 py-20">密钥不存在或加载失败</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import { getKey, updateKey } from '../api'
import type { Key, KeyLog } from '../types'

const route = useRoute()

const keyData = ref<Key | null>(null)
const logs = ref<KeyLog[]>([])
const loading = ref(true)
const noteEdit = ref('')
const savingNote = ref(false)
const togglingStatus = ref(false)

const statusType = (s: string) => {
  const map: Record<string, string> = {
    active: 'success',
    inactive: 'info',
    expired: 'warning',
    disabled: 'danger',
  }
  return map[s] || 'info'
}

const statusLabel = (s: string) => {
  const map: Record<string, string> = {
    active: '活跃',
    inactive: '未激活',
    expired: '已过期',
    disabled: '已禁用',
  }
  return map[s] || s
}

const formatDate = (str: string) => {
  if (!str) return '-'
  return dayjs(str).format('YYYY-MM-DD HH:mm:ss')
}

const loadKey = async () => {
  const id = Number(route.params.id)
  loading.value = true
  try {
    const res = await getKey(id)
    keyData.value = res
    logs.value = res.logs || []
    noteEdit.value = res.note || ''
  } catch {
    keyData.value = null
  } finally {
    loading.value = false
  }
}

const handleSaveNote = async () => {
  if (!keyData.value) return
  savingNote.value = true
  try {
    const updated = await updateKey(keyData.value.id, { note: noteEdit.value })
    keyData.value = { ...keyData.value, ...updated }
    ElMessage.success('备注已保存')
  } catch (err: any) {
    const msg = err.response?.data?.message || err.response?.data?.error || '保存失败'
    ElMessage.error(msg)
  } finally {
    savingNote.value = false
  }
}

const toggleStatus = async () => {
  if (!keyData.value) return
  const isDisabling = keyData.value.status !== 'disabled'
  const action = isDisabling ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(`确定要${action}该密钥吗？`, '确认', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消',
    })
    togglingStatus.value = true
    const updated = await updateKey(keyData.value.id, {
      status: isDisabling ? 'disabled' : 'inactive',
    })
    keyData.value = { ...keyData.value, ...updated }
    ElMessage.success(`${action}成功`)
  } catch (err: any) {
    if (err === 'cancel') return
    const msg = err.response?.data?.message || err.response?.data?.error || `${action}失败`
    ElMessage.error(msg)
  } finally {
    togglingStatus.value = false
  }
}

onMounted(loadKey)
</script>
