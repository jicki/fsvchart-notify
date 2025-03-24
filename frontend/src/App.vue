<template>
  <div class="app-container">
    <header class="app-header">
      <h1>fsvchart-notify Manager</h1>
      
      <div v-if="isLoggedIn" class="user-info">
        <span>{{ userDisplayName }}</span>
        <div class="dropdown">
          <button class="dropdown-toggle">
            <span class="user-icon">👤</span>
          </button>
          <div class="dropdown-menu">
            <RouterLink to="/profile">个人信息</RouterLink>
            <a href="#" @click.prevent="logout">退出登录</a>
          </div>
        </div>
      </div>
    </header>
    
    <nav v-if="isLoggedIn">
      <RouterLink to="/">首页</RouterLink>
      <RouterLink to="/send-records">发送记录</RouterLink>
    </nav>
    
    <div class="main-content">
      <RouterView />
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView, useRouter, useRoute } from 'vue-router'
import { ref, onMounted, computed, watch } from 'vue'

const router = useRouter()
const route = useRoute()
const isLoggedIn = ref(false)
const userInfo = ref<any>(null)
const showDebug = ref(false) // Set to false in production

// 计算属性：用户显示名称
const userDisplayName = computed(() => {
  if (!userInfo.value) return ''
  return userInfo.value.displayName || userInfo.value.username || ''
})

// 检查登录状态
const checkLoginStatus = () => {
  const token = localStorage.getItem('token')
  const user = localStorage.getItem('user')
  
  if (token && user) {
    isLoggedIn.value = true
    try {
      userInfo.value = JSON.parse(user)
    } catch (e) {
      console.error('Failed to parse user info:', e)
      userInfo.value = null
    }
  } else {
    isLoggedIn.value = false
    userInfo.value = null
  }
}

// 退出登录
const logout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  isLoggedIn.value = false
  userInfo.value = null
  router.push('/login')
}

// 清除存储（用于调试）
const clearStorage = () => {
  localStorage.clear()
  checkLoginStatus()
  router.push('/login')
}

// 监听路由变化，每次路由变化时检查登录状态
watch(
  () => route.path,
  () => {
    checkLoginStatus()
  }
)

// 组件挂载时检查登录状态
onMounted(() => {
  checkLoginStatus()
  
  // 监听 storage 事件，用于在登录/登出时更新状态
  window.addEventListener('storage', checkLoginStatus)
  
  // 组件卸载时移除事件监听
  return () => {
    window.removeEventListener('storage', checkLoginStatus)
  }
})
</script>

<style>
body {
  margin: 0;
  font-family: sans-serif;
  background: #f5f5f5;
}

.app-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.debug-info {
  background: #ffe;
  padding: 10px;
  border: 1px solid #ddd;
  margin: 10px;
  font-family: monospace;
  font-size: 12px;
}

.debug-info button {
  margin-right: 10px;
  padding: 5px;
}

.app-header {
  padding: 0 20px;
  background: #f8f9fa;
  border-bottom: 1px solid #e9ecef;
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 60px;
  position: relative;
  z-index: 100;
}

h1 {
  margin: 0;
  padding: 15px 0;
  font-size: 1.5rem;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.dropdown {
  position: relative;
  display: inline-block;
}

.dropdown-toggle {
  background: none;
  border: none;
  cursor: pointer;
  padding: 5px;
}

.user-icon {
  font-size: 1.5rem;
}

.dropdown-menu {
  display: none;
  position: absolute;
  right: 0;
  background-color: white;
  min-width: 120px;
  box-shadow: 0 2px 5px rgba(0,0,0,0.2);
  z-index: 1;
  border-radius: 4px;
}

.dropdown-menu a {
  color: #333;
  padding: 10px 15px;
  text-decoration: none;
  display: block;
  text-align: left;
}

.dropdown-menu a:hover {
  background-color: #f5f5f5;
}

.dropdown:hover .dropdown-menu {
  display: block;
}

nav {
  padding: 10px 20px;
  background: #fff;
  border-bottom: 1px solid #e9ecef;
  display: flex;
  align-items: center;
}

nav a {
  margin-right: 15px;
  text-decoration: none;
  color: #666;
  padding: 5px 10px;
  border-radius: 4px;
}

nav a:hover {
  background: #f8f9fa;
}

nav a.router-link-active {
  color: #007bff;
  font-weight: bold;
}

.main-content {
  flex: 1;
  padding: 20px;
  background: #fff;
  margin: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}
</style>
