import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

import LoginView from '../views/LoginView.vue'
import DashboardView from '../views/DashboardView.vue'
import BatchesView from '../views/BatchesView.vue'
import KeysView from '../views/KeysView.vue'
import KeyDetailView from '../views/KeyDetailView.vue'
import AlertsView from '../views/AlertsView.vue'
import AppLayout from '../components/AppLayout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: LoginView,
      meta: { requiresAuth: false },
    },
    {
      path: '/',
      component: AppLayout,
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          redirect: '/dashboard',
        },
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: DashboardView,
          meta: { requiresAuth: true, title: 'Dashboard' },
        },
        {
          path: 'batches',
          name: 'Batches',
          component: BatchesView,
          meta: { requiresAuth: true, title: '批次管理' },
        },
        {
          path: 'keys',
          name: 'Keys',
          component: KeysView,
          meta: { requiresAuth: true, title: '密钥管理' },
        },
        {
          path: 'keys/:id',
          name: 'KeyDetail',
          component: KeyDetailView,
          meta: { requiresAuth: true, title: '密钥详情' },
        },
        {
          path: 'alerts',
          name: 'Alerts',
          component: AlertsView,
          meta: { requiresAuth: true, title: '共享告警' },
        },
      ],
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()
  if (to.meta.requiresAuth && !authStore.isLoggedIn) {
    next('/login')
  } else if (to.path === '/login' && authStore.isLoggedIn) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
