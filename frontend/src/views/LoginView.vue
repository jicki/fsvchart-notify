<template>
  <div class="login-container">
    <div class="login-box">
      <h2>fsvchart-notify 登录</h2>
      <div v-if="error" class="error-message">{{ error }}</div>
      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <label for="username">用户名</label>
          <input
            type="text"
            id="username"
            v-model="username"
            required
            placeholder="请输入用户名"
          />
        </div>
        <div class="form-group">
          <label for="password">密码</label>
          <input
            type="password"
            id="password"
            v-model="password"
            required
            placeholder="请输入密码"
          />
        </div>
        <button type="submit" :disabled="loading">
          {{ loading ? '登录中...' : '登录' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

const handleLogin = async () => {
  error.value = ''
  loading.value = true

  try {
    const response = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: username.value,
        password: password.value
      })
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.message || data.error || '登录失败')
    }

    authStore.login(data.token, {
      username: data.username,
      displayName: data.display_name,
      role: data.role
    })

    window.dispatchEvent(new Event('storage'))
    router.push('/')
  } catch (err: unknown) {
    console.error('Login error:', err)
    error.value = err instanceof Error ? err.message : '登录失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container { display: flex; justify-content: center; align-items: center; min-height: 100vh; background-color: var(--color-bg-page, #f5f5f5); }
.login-box { width: 400px; padding: 30px; background: white; border-radius: var(--radius-lg, 8px); box-shadow: var(--shadow-md, 0 2px 10px rgba(0, 0, 0, 0.1)); }
h2 { text-align: center; margin-bottom: 24px; color: var(--color-text, #333); }
.form-group { margin-bottom: 20px; }
label { display: block; margin-bottom: 8px; font-weight: 500; }
input { width: 100%; padding: 10px; border: 1px solid var(--color-border, #ddd); border-radius: var(--radius-md, 4px); font-size: 16px; }
button { width: 100%; padding: 12px; background: var(--color-primary, #007bff); color: white; border: none; border-radius: var(--radius-md, 4px); font-size: 16px; cursor: pointer; transition: background 0.3s; }
button:hover { background: var(--color-primary-hover, #0056b3); }
button:disabled { background: var(--color-bg-disabled, #cccccc); cursor: not-allowed; }
.error-message { background: var(--color-error-bg, #f8d7da); color: var(--color-error-text, #721c24); padding: 10px; border-radius: var(--radius-md, 4px); margin-bottom: 20px; }
</style>
