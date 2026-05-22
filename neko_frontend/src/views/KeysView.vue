<template>
  <div>
    <!-- Filter bar -->
    <el-card shadow="never" class="mb-4 border border-gray-200">
      <div class="flex flex-wrap gap-3 items-center">
        <el-select
          v-model="filters.batch_id"
          placeholder="选择批次"
          clearable
          class="w-44"
          @change="handleSearch"
        >
          <el-option
            v-for="b in batches"
            :key="b.id"
            :label="`[${b.id}] ${b.note}`"
            :value="b.id"
          />
        </el-select>

        <el-select
          v-model="filters.status"
          placeholder="状态筛选"
          clearable
          class="w-36"
          @change="handleSearch"
        >
          <el-option label="未激活" value="inactive" />
          <el-option label="活跃中" value="active" />
          <el-option label="已过期" value="expired" />
          <el-option label="已禁用" value="disabled" />
        </el-select>

        <el-input
          v-model="filters.keyword"
          placeholder="搜索密钥/备注"
          clearable
          class="w-56"
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />

        <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
    </el-card>

    <!-- Table -->
    <el-card shadow="never" class="border border-gray-200">
      <el-table :data="keys" v-loading="loading" stripe class="w-full">
        <el-table-column label="密钥" min-width="220">
          <template #default="{ row }">
            <span class="font-mono text-xs text-gray-800 select-all">{{ row.key }}</span>
          </template>
        </el-table-column>

        <el-table-column label="类型" width="90">
          <template #default="{ row }">
            <el-tag :type="row.type === 'one_time' ? 'warning' : 'success'" size="small">
              {{ row.type === 'one_time' ? '一次性' : '多次' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="批次备注" width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.batch?.note || '-' }}</template>
        </el-table-column>

        <el-table-column label="到期时间" width="150">
          <template #default="{ row }">{{ formatDate(row.expires_at) }}</template>
        </el-table-column>

        <el-table-column label="使用次数" width="100" align="center">
          <template #default="{ row }">
            <span v-if="row.type === 'one_time'">{{ row.use_count }} 次</span>
            <span v-else>{{ row.use_count }} / {{ row.max_uses }}</span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="160" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" text size="small" @click="viewDetail(row.id)">
              详情
            </el-button>
            <el-button
              v-if="row.status !== 'expired'"
              :type="row.status === 'disabled' ? 'success' : 'danger'"
              text
              size="small"
              @click="toggleStatus(row)"
            >
              {{ row.status === 'disabled' ? '启用' : '禁用' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <div class="flex justify-end mt-4">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :total="total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @current-change="loadKeys"
          @size-change="handleSearch"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import { getKeys, getBatches, updateKey } from '../api'
import type { Key, KeyBatch } from '../types'

const route = useRoute()
const router = useRouter()

const keys = ref<Key[]>([])
const batches = ref<KeyBatch[]>([])
const loading = ref(false)
const total = ref(0)

const filters = reactive({
  batch_id: undefined as number | undefined,
  status: '',
  keyword: '',
})

const pagination = reactive({
  page: 1,
  page_size: 20,
})

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
  return dayjs(str).format('YYYY-MM-DD HH:mm')
}

const loadKeys = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.page_size,
    }
    if (filters.batch_id) params.batch_id = filters.batch_id
    if (filters.status) params.status = filters.status
    if (filters.keyword) params.keyword = filters.keyword

    const res = await getKeys(params)
    keys.value = res.list
    total.value = res.total
  } catch {
    ElMessage.error('加载密钥失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadKeys()
}

const handleReset = () => {
  filters.batch_id = undefined
  filters.status = ''
  filters.keyword = ''
  handleSearch()
}

const viewDetail = (id: number) => {
  router.push(`/keys/${id}`)
}

const toggleStatus = async (row: Key) => {
  const isDisabling = row.status !== 'disabled'
  const action = isDisabling ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(`确定要${action}该密钥吗？`, '确认', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消',
    })
    await updateKey(row.id, { status: isDisabling ? 'disabled' : 'inactive' })
    ElMessage.success(`${action}成功`)
    loadKeys()
  } catch (err: any) {
    if (err === 'cancel') return
    const msg = err.response?.data?.message || err.response?.data?.error || `${action}失败`
    ElMessage.error(msg)
  }
}

onMounted(async () => {
  // Check URL query for batch_id
  if (route.query.batch_id) {
    filters.batch_id = Number(route.query.batch_id)
  }

  try {
    batches.value = await getBatches()
  } catch {
    // ignore
  }

  loadKeys()
})
</script>
