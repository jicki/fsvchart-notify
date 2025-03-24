<template>
  <div class="profile-container">
    <h2>用户信息</h2>
    
    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error-message">{{ error }}</div>
    
    <div v-if="user" class="profile-content">
      <div class="profile-section">
        <h3>基本信息</h3>
        <form @submit.prevent="updateProfile">
          <div class="form-group">
            <label for="username">用户名</label>
            <input type="text" id="username" v-model="user.username" disabled />
          </div>
          <div class="form-group">
            <label for="displayName">显示名称</label>
            <input type="text" id="displayName" v-model="user.display_name" />
          </div>
          <div class="form-group">
            <label for="email">邮箱</label>
            <input type="email" id="email" v-model="user.email" />
          </div>
          <div class="form-group">
            <label for="role">角色</label>
            <input type="text" id="role" v-model="user.role" disabled />
          </div>
          <button type="submit" :disabled="profileLoading">
            {{ profileLoading ? '更新中...' : '更新信息' }}
          </button>
          <div v-if="profileSuccess" class="success-message">信息更新成功！</div>
          <div v-if="profileError" class="error-message">{{ profileError }}</div>
        </form>
      </div>
      
      <div class="profile-section">
        <h3>修改密码</h3>
        <form @submit.prevent="changePassword">
          <div class="form-group">
            <label for="oldPassword">当前密码</label>
            <input 
              type="password" 
              id="oldPassword" 
              v-model="passwordForm.oldPassword" 
              required 
            />
          </div>
          <div class="form-group">
            <label for="newPassword">新密码</label>
            <input 
              type="password" 
              id="newPassword" 
              v-model="passwordForm.newPassword" 
              required 
              minlength="6"
            />
          </div>
          <div class="form-group">
            <label for="confirmPassword">确认新密码</label>
            <input 
              type="password" 
              id="confirmPassword" 
              v-model="passwordForm.confirmPassword" 
              required 
            />
          </div>
          <button type="submit" :disabled="passwordLoading">
            {{ passwordLoading ? '更新中...' : '修改密码' }}
          </button>
          <div v-if="passwordSuccess" class="success-message">密码修改成功！</div>
          <div v-if="passwordError" class="error-message">{{ passwordError }}</div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { get, put } from '../utils/api'

const user = ref(null)
const loading = ref(true)
const error = ref('')

// 个人信息更新状态
const profileLoading = ref(false)
const profileSuccess = ref(false)
const profileError = ref('')

// 密码修改表单和状态
const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})
const passwordLoading = ref(false)
const passwordSuccess = ref(false)
const passwordError = ref('')

// 获取用户信息
const fetchUserInfo = async () => {
  loading.value = true
  error.value = ''
  
  try {
    user.value = await get('/api/user/current')
  } catch (err) {
    error.value = err.message || '获取用户信息失败'
  } finally {
    loading.value = false
  }
}

// 更新个人信息
const updateProfile = async () => {
  profileLoading.value = true
  profileSuccess.value = false
  profileError.value = ''
  
  try {
    await put('/api/user/info', {
      display_name: user.value.display_name,
      email: user.value.email
    })
    
    profileSuccess.value = true
    
    // 更新本地存储的用户信息
    const storedUser = JSON.parse(localStorage.getItem('user') || '{}')
    storedUser.displayName = user.value.display_name
    localStorage.setItem('user', JSON.stringify(storedUser))
    
  } catch (err) {
    profileError.value = err.message || '更新信息失败'
  } finally {
    profileLoading.value = false
  }
}

// 修改密码
const changePassword = async () => {
  // 验证两次密码输入是否一致
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
    
    // 清空表单
    passwordForm.oldPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
    
  } catch (err) {
    passwordError.value = err.message || '修改密码失败'
  } finally {
    passwordLoading.value = false
  }
}

// 组件挂载时获取用户信息
onMounted(fetchUserInfo)
</script>

<style scoped>
.profile-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

h2 {
  margin-bottom: 24px;
  color: #333;
}

.loading {
  text-align: center;
  padding: 20px;
  color: #666;
}

.profile-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 30px;
}

.profile-section {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

h3 {
  margin-top: 0;
  margin-bottom: 20px;
  color: #333;
  border-bottom: 1px solid #eee;
  padding-bottom: 10px;
}

.form-group {
  margin-bottom: 15px;
}

label {
  display: block;
  margin-bottom: 5px;
  font-weight: 500;
}

input {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

input:disabled {
  background-color: #f5f5f5;
  cursor: not-allowed;
}

button {
  padding: 10px 15px;
  background: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.3s;
  margin-top: 10px;
}

button:hover {
  background: #0056b3;
}

button:disabled {
  background: #cccccc;
  cursor: not-allowed;
}

.success-message {
  margin-top: 10px;
  padding: 8px;
  background: #d4edda;
  color: #155724;
  border-radius: 4px;
}

.error-message {
  margin-top: 10px;
  padding: 8px;
  background: #f8d7da;
  color: #721c24;
  border-radius: 4px;
}

@media (max-width: 768px) {
  .profile-content {
    grid-template-columns: 1fr;
  }
}
</style> 