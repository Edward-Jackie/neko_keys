<template>
  <div>
    <!-- Description -->
    <el-alert
      type="warning"
      :closable="false"
      show-icon
      class="mb-4"
    >
      <template #default>
        以下密钥在最近 1 小时内被超过 2 个不同 IP 地址使用，可能存在<strong>密钥共享行为</strong>
      </template>
    </el-alert>

    <el-card shadow="never" class="border border-gray-200">
      <el-table :data="alerts" v-loading="loading" stripe class="w-full">
        <el-table-column label="密钥" min-width="200">
          <template #default="{ row }">
            <span class="font-mono text-xs text-gray-800 select-all">{{ row.key }}</span>
          </template>
        </el-table-column>

        <el-table-column label="IP 数量" width="90" align="center">
          <template #default="{ row }">
            <el-tag type="danger" size="small">{{ row.ip_count }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="IP 列表" min-width="300">
          <template #default="{ row }">
            <div class="flex flex-wrap gap-1">
              <el-tag
                v-for="ip in row.ips"
                :key="ip"
                type="info"
                size="small"
              >
                {{ ip }}
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="100" align="center">
          <template #default="{ row }">
            <el-button type="primary" text size="small" @click="viewDetail(row.key_id)">
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="!loading && alerts.length === 0" class="text-center text-gray-400 py-12 text-sm">
        暂无告警，密钥使用情况正常
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getAlerts } from '../api'
import type { AlertKey } from '../types'

const router = useRouter()
const alerts = ref<AlertKey[]>([])
const loading = ref(true)

const viewDetail = (id: number) => {
  router.push(`/keys/${id}`)
}

onMounted(async () => {
  try {
    alerts.value = await getAlerts()
  } catch {
    ElMessage.error('加载告警失败')
  } finally {
    loading.value = false
  }
})
</script>
