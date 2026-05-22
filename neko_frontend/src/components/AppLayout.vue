<template>
  <div class="flex h-screen bg-gray-100">
    <!-- Sidebar -->
    <aside class="w-[220px] flex-shrink-0 flex flex-col" style="background-color: #1a1a2e;">
      <!-- Logo -->
      <div class="px-6 py-5 border-b border-white/10">
        <h1 class="text-white text-lg font-bold tracking-wide">NekoKeys Admin</h1>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 py-4">
        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="flex items-center gap-3 px-6 py-3 text-sm transition-all duration-150"
          :class="isActive(item.path)
            ? 'bg-white/15 text-white font-medium border-r-2 border-blue-400'
            : 'text-gray-400 hover:text-white hover:bg-white/8'"
        >
          <el-icon :size="16"><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </router-link>
      </nav>

      <!-- Bottom user info -->
      <div class="px-6 py-4 border-t border-white/10">
        <p class="text-gray-500 text-xs truncate">{{ authStore.username }}</p>
      </div>
    </aside>

    <!-- Main content area -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- Top bar -->
      <header class="h-14 bg-white border-b border-gray-200 flex items-center justify-between px-6 flex-shrink-0">
        <h2 class="text-gray-800 font-semibold text-base">{{ currentTitle }}</h2>
        <el-button type="danger" plain size="small" @click="handleLogout">
          <el-icon class="mr-1"><SwitchButton /></el-icon>
          退出登录
        </el-button>
      </header>

      <!-- Page content -->
      <main class="flex-1 overflow-auto p-6">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import {
  Odometer,
  Collection,
  Key,
  Bell,
  SwitchButton,
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const navItems = [
  { path: '/dashboard', label: 'Dashboard', icon: Odometer },
  { path: '/batches', label: '批次管理', icon: Collection },
  { path: '/keys', label: '密钥管理', icon: Key },
  { path: '/alerts', label: '共享告警', icon: Bell },
]

const isActive = (path: string) => {
  if (path === '/dashboard') return route.path === '/dashboard'
  return route.path.startsWith(path)
}

const currentTitle = computed(() => {
  return (route.meta.title as string) || 'NekoKeys Admin'
})

const handleLogout = () => {
  authStore.clearAuth()
  router.push('/login')
}
</script>
