<template>
  <div class="login-bg min-h-screen flex items-center justify-center p-4 relative overflow-hidden">
    <!-- 背景装饰光球 -->
    <div class="orb orb-1" />
    <div class="orb orb-2" />
    <div class="orb orb-3" />

    <!-- 登录卡片 -->
    <div class="glass-card w-full max-w-[420px] relative z-10">
      <!-- 顶部品牌区 -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-violet-500 to-indigo-600 shadow-lg shadow-violet-500/30 mb-4">
          <span class="text-3xl select-none">🐱</span>
        </div>
        <h1 class="text-2xl font-bold text-white tracking-tight">Neko Keys</h1>
        <p class="text-white/40 text-sm mt-1">密钥管理后台</p>
      </div>

      <!-- 表单 -->
      <div class="space-y-4">
        <!-- 用户名 -->
        <div class="input-group">
          <label class="input-label">用户名</label>
          <div class="input-wrapper" :class="{ focused: focusedField === 'username' }">
            <el-icon class="input-icon"><User /></el-icon>
            <input
              v-model="form.username"
              type="text"
              placeholder="请输入用户名"
              class="custom-input"
              @focus="focusedField = 'username'"
              @blur="focusedField = ''"
              @keyup.enter="handleLogin"
              autocomplete="username"
            />
          </div>
          <span v-if="errors.username" class="error-msg">{{ errors.username }}</span>
        </div>

        <!-- 密码 -->
        <div class="input-group">
          <label class="input-label">密码</label>
          <div class="input-wrapper" :class="{ focused: focusedField === 'password' }">
            <el-icon class="input-icon"><Lock /></el-icon>
            <input
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="请输入密码"
              class="custom-input"
              @focus="focusedField = 'password'"
              @blur="focusedField = ''"
              @keyup.enter="handleLogin"
              autocomplete="current-password"
            />
            <button
              class="eye-btn"
              type="button"
              @mousedown.prevent
              @click="showPassword = !showPassword"
            >
              <el-icon><component :is="showPassword ? View : Hide" /></el-icon>
            </button>
          </div>
          <span v-if="errors.password" class="error-msg">{{ errors.password }}</span>
        </div>

        <!-- 登录按钮 -->
        <button
          class="login-btn"
          :class="{ loading }"
          :disabled="loading"
          @click="handleLogin"
        >
          <span v-if="!loading">登 录</span>
          <span v-else class="flex items-center justify-center gap-2">
            <svg class="animate-spin h-4 w-4" viewBox="0 0 24 24" fill="none">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z"/>
            </svg>
            登录中...
          </span>
        </button>
      </div>

      <!-- 底部版权 -->
      <p class="text-center text-white/20 text-xs mt-8">© 2026 Neko Keys · 密钥管理系统</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, View, Hide } from '@element-plus/icons-vue'
import { login } from '../api'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const showPassword = ref(false)
const focusedField = ref('')

const form = reactive({ username: '', password: '' })
const errors = reactive({ username: '', password: '' })

const validate = () => {
  errors.username = form.username ? '' : '请输入用户名'
  errors.password = form.password ? '' : '请输入密码'
  return !errors.username && !errors.password
}

const handleLogin = async () => {
  if (!validate()) return
  loading.value = true
  try {
    const res = await login(form.username, form.password)
    authStore.setAuth(res.token, res.username)
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } catch (err: any) {
    const msg = err.response?.data?.error || err.response?.data?.message || '用户名或密码错误'
    ElMessage.error(msg)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
/* 背景 */
.login-bg {
  background: #0f0f1a;
}

/* 光球 */
.orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  pointer-events: none;
}
.orb-1 {
  width: 400px; height: 400px;
  background: rgba(124, 58, 237, 0.25);
  top: -100px; left: -100px;
  animation: float 8s ease-in-out infinite;
}
.orb-2 {
  width: 350px; height: 350px;
  background: rgba(99, 102, 241, 0.2);
  bottom: -80px; right: -80px;
  animation: float 10s ease-in-out infinite reverse;
}
.orb-3 {
  width: 250px; height: 250px;
  background: rgba(236, 72, 153, 0.12);
  top: 50%; left: 55%;
  animation: float 12s ease-in-out infinite 2s;
}
@keyframes float {
  0%, 100% { transform: translateY(0) scale(1); }
  50% { transform: translateY(-30px) scale(1.05); }
}

/* 玻璃卡片 */
.glass-card {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 24px;
  padding: 40px;
  box-shadow:
    0 25px 50px rgba(0, 0, 0, 0.4),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

/* 输入框组 */
.input-group { display: flex; flex-direction: column; gap: 6px; }

.input-label {
  font-size: 13px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.5);
  letter-spacing: 0.02em;
}

.input-wrapper {
  display: flex;
  align-items: center;
  gap: 10px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 0 14px;
  transition: border-color 0.2s, background 0.2s, box-shadow 0.2s;
}
.input-wrapper.focused {
  border-color: rgba(139, 92, 246, 0.7);
  background: rgba(139, 92, 246, 0.08);
  box-shadow: 0 0 0 3px rgba(139, 92, 246, 0.15);
}

.input-icon {
  color: rgba(255, 255, 255, 0.3);
  font-size: 16px;
  flex-shrink: 0;
}

.custom-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: #fff;
  font-size: 14px;
  padding: 13px 0;
  min-width: 0;
}
.custom-input::placeholder { color: rgba(255, 255, 255, 0.2); }

.eye-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  color: rgba(255, 255, 255, 0.3);
  padding: 0;
  display: flex;
  align-items: center;
  transition: color 0.2s;
  flex-shrink: 0;
}
.eye-btn:hover { color: rgba(255, 255, 255, 0.7); }

.error-msg {
  font-size: 12px;
  color: #f87171;
  margin-top: 2px;
}

/* 登录按钮 */
.login-btn {
  width: 100%;
  padding: 13px;
  border: none;
  border-radius: 12px;
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.1em;
  color: #fff;
  cursor: pointer;
  background: linear-gradient(135deg, #7c3aed, #6366f1);
  box-shadow: 0 8px 24px rgba(124, 58, 237, 0.35);
  transition: opacity 0.2s, transform 0.15s, box-shadow 0.2s;
  margin-top: 8px;
}
.login-btn:hover:not(:disabled) {
  opacity: 0.92;
  transform: translateY(-1px);
  box-shadow: 0 12px 28px rgba(124, 58, 237, 0.45);
}
.login-btn:active:not(:disabled) {
  transform: translateY(0);
}
.login-btn:disabled {
  cursor: not-allowed;
  opacity: 0.7;
}
</style>
