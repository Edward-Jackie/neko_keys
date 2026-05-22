<template>
  <div>
    <div v-if="loading" class="flex justify-center items-center h-64">
      <el-icon class="is-loading text-4xl text-blue-500"><Loading /></el-icon>
    </div>

    <template v-else-if="stats">
      <!-- Stats grid -->
      <div class="grid grid-cols-4 gap-4 mb-6">
        <StatCard
          label="总密钥数"
          :value="stats.total"
          color="blue"
          icon="Key"
        />
        <StatCard
          label="活跃中"
          :value="stats.active"
          color="green"
          icon="CircleCheck"
        />
        <StatCard
          label="已过期"
          :value="stats.expired"
          color="orange"
          icon="Clock"
        />
        <StatCard
          label="已禁用"
          :value="stats.disabled"
          color="red"
          icon="CircleClose"
        />
        <StatCard
          label="未激活"
          :value="stats.inactive"
          color="gray"
          icon="Remove"
        />
        <StatCard
          label="今日激活"
          :value="stats.today_activations"
          color="purple"
          icon="Promotion"
        />
        <StatCard
          label="今日验证"
          :value="stats.today_validations"
          color="cyan"
          icon="View"
        />
        <StatCard
          label="告警数"
          :value="stats.alert_count"
          :color="stats.alert_count > 0 ? 'danger' : 'gray'"
          icon="Bell"
          :highlight="stats.alert_count > 0"
        />
      </div>

      <!-- Alert link -->
      <div v-if="stats.alert_count > 0" class="mt-2">
        <el-alert
          type="warning"
          :closable="false"
          show-icon
        >
          <template #default>
            检测到 <strong>{{ stats.alert_count }}</strong> 个密钥共享告警
            <el-button
              type="warning"
              text
              class="ml-2"
              @click="$router.push('/alerts')"
            >
              去查看告警 →
            </el-button>
          </template>
        </el-alert>
      </div>
    </template>

    <div v-else class="text-center text-gray-400 py-20">加载失败，请刷新重试</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, defineComponent, h } from 'vue'
import { ElIcon } from 'element-plus'
import {
  Loading,
  Key,
  CircleCheck,
  Clock,
  CircleClose,
  Remove,
  Promotion,
  View,
  Bell,
} from '@element-plus/icons-vue'
import { getDashboard } from '../api'
import type { DashboardStats } from '../types'

const stats = ref<DashboardStats | null>(null)
const loading = ref(true)

const colorMap: Record<string, string> = {
  blue: 'bg-blue-50 border-blue-200 text-blue-600',
  green: 'bg-green-50 border-green-200 text-green-600',
  orange: 'bg-orange-50 border-orange-200 text-orange-600',
  red: 'bg-red-50 border-red-200 text-red-600',
  gray: 'bg-gray-50 border-gray-200 text-gray-500',
  purple: 'bg-purple-50 border-purple-200 text-purple-600',
  cyan: 'bg-cyan-50 border-cyan-200 text-cyan-600',
  danger: 'bg-red-50 border-red-400 text-red-600',
}

const iconMap: Record<string, any> = {
  Key,
  CircleCheck,
  Clock,
  CircleClose,
  Remove,
  Promotion,
  View,
  Bell,
}

// Inline StatCard component
const StatCard = defineComponent({
  props: {
    label: String,
    value: Number,
    color: String,
    icon: String,
    highlight: Boolean,
  },
  setup(props) {
    return () =>
      h(
        'div',
        {
          class: `rounded-xl border p-5 flex items-center gap-4 ${colorMap[props.color || 'gray']} ${props.highlight ? 'ring-2 ring-red-400' : ''}`,
        },
        [
          h(
            ElIcon,
            { size: 32 },
            { default: () => h(iconMap[props.icon || 'Key']) }
          ),
          h('div', {}, [
            h('p', { class: 'text-2xl font-bold' }, String(props.value ?? 0)),
            h('p', { class: 'text-xs mt-0.5 opacity-70' }, props.label),
          ]),
        ]
      )
  },
})

onMounted(async () => {
  try {
    stats.value = await getDashboard()
  } catch {
    stats.value = null
  } finally {
    loading.value = false
  }
})
</script>
