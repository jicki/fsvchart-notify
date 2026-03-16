<template>
  <div class="profile-container">
    <div class="page-header">
      <div>
        <h3>用户信息</h3>
        <p>管理个人信息和安全设置</p>
      </div>
    </div>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error-message">{{ error }}</div>

    <div v-if="user" class="profile-grid">
      <div class="card">
        <h4>基本信息</h4>
        <form @submit.prevent="updateProfile">
          <div class="form-group">
            <label for="username">用户名</label>
            <input class="form-input" type="text" id="username" v-model="user.username" disabled />
          </div>
          <div class="form-group">
            <label for="displayName">显示名称</label>
            <input class="form-input" type="text" id="displayName" v-model="user.display_name" />
          </div>
          <div class="form-group">
            <label for="email">邮箱</label>
            <input class="form-input" type="email" id="email" v-model="user.email" />
          </div>
          <div class="form-group">
            <label for="role">角色</label>
            <input class="form-input" type="text" id="role" v-model="user.role" disabled />
          </div>
          <button type="submit" class="btn btn-primary" :disabled="profileLoading">
            {{ profileLoading ? '更新中...' : '更新信息' }}
          </button>
          <div v-if="profileSuccess" class="success-message">信息更新成功</div>
          <div v-if="profileError" class="error-message">{{ profileError }}</div>
        </form>
      </div>

      <div class="card">
        <h4>修改密码</h4>
        <form @submit.prevent="changePassword">
          <div class="form-group">
            <label for="oldPassword">当前密码</label>
            <input class="form-input" type="password" id="oldPassword" v-model="passwordForm.oldPassword" required />
          </div>
          <div class="form-group">
            <label for="newPassword">新密码</label>
            <input class="form-input" type="password" id="newPassword" v-model="passwordForm.newPassword" required minlength="6" />
          </div>
          <div class="form-group">
            <label for="confirmPassword">确认新密码</label>
            <input class="form-input" type="password" id="confirmPassword" v-model="passwordForm.confirmPassword" required />
          </div>
          <button type="submit" class="btn btn-primary" :disabled="passwordLoading">
            {{ passwordLoading ? '更新中...' : '修改密码' }}
          </button>
          <div v-if="passwordSuccess" class="success-message">密码修改成功</div>
          <div v-if="passwordError" class="error-message">{{ passwordError }}</div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { get, put } from '../utils/api'
import { useAuthStore } from '../stores/auth'

interface UserProfile {
  username: string
  display_name: string
  email: string
  role: string
}

const authStore = useAuthStore()
const user = ref<UserProfile | null>(null)
const loading = ref(true)
const error = ref('')

const profileLoading = ref(false)
const profileSuccess = ref(false)
const profileError = ref('')

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})
const passwordLoading = ref(false)
const passwordSuccess = ref(false)
const passwordError = ref('')

const fetchUserInfo = async () => {
  loading.value = true
  error.value = ''
  try {
    user.value = await get<UserProfile>('/api/user/current')
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : '获取用户信息失败'
  } finally {
    loading.value = false
  }
}

const updateProfile = async () => {
  if (!user.value) return
  profileLoading.value = true
  profileSuccess.value = false
  profileError.value = ''

  try {
    await put('/api/user/info', {
      display_name: user.value.display_name,
      email: user.value.email
    })
    profileSuccess.value = true
    authStore.updateUser({ displayName: user.value.display_name })
  } catch (err: unknown) {
    profileError.value = err instanceof Error ? err.message : '更新信息失败'
  } finally {
    profileLoading.value = false
  }
}

const changePassword = async () => {
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    passwordError.value = '两次输入的密码不一致'
    return
  }

  passwordLoading.value = true
  passwordSuccess.value = false
  passwordError.value = ''

  try {
    await put('/api/user/password', {
      old_password: passwordForm.oldPassword,
      new_password: passwordForm.newPassword
    })
    passwordSuccess.value = true
    passwordForm.oldPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
  } catch (err: unknown) {
    passwordError.value = err instanceof Error ? err.message : '修改密码失败'
  } finally {
    passwordLoading.value = false
  }
}

onMounted(fetchUserInfo)
</script>

<style scoped>
.profile-container {
  max-width: 900px;
}

.profile-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-lg);
}

.card h4 {
  margin: 0 0 var(--spacing-lg);
  padding-bottom: var(--spacing-sm);
  border-bottom: 1px solid var(--color-border);
  font-weight: 600;
  color: var(--color-text);
}

@media (max-width: 768px) {
  .profile-grid {
    grid-template-columns: 1fr;
  }
}
</style>
