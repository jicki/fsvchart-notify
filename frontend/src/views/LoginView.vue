<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <div class="login-brand">
          <IconBarChart :size="32" />
        </div>
        <h2>FSVChart Notify</h2>
        <p class="login-desc">监控图表推送管理平台</p>
      </div>

      <div v-if="error" class="error-message">{{ error }}</div>

      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <label for="username">用户名</label>
          <input
            class="form-input"
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
            class="form-input"
            type="password"
            id="password"
            v-model="password"
            required
            placeholder="请输入密码"
          />
        </div>
        <button type="submit" class="btn btn-primary login-btn" :disabled="loading">
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
import { IconBarChart } from '../components/icons'

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
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: var(--color-bg-page);
}

.login-card {
  width: 100%;
  max-width: 400px;
  padding: 40px;
  background: var(--color-bg-white);
  border-radius: var(--radius-xl);
  border: 1px solid var(--color-border);
  box-shadow: var(--shadow-lg);
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-brand {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  background: var(--color-accent-light);
  color: var(--color-accent);
  border-radius: var(--radius-lg);
  margin-bottom: 16px;
}

.login-header h2 {
  margin: 0 0 8px;
  font-size: 24px;
  font-weight: 700;
  color: var(--color-text);
}

.login-desc {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 14px;
}

.login-card .form-input {
  padding: 12px 14px;
  font-size: 15px;
}

.login-btn {
  width: 100%;
  padding: 12px;
  font-size: 15px;
  font-weight: 600;
  margin-top: 8px;
}
</style>
