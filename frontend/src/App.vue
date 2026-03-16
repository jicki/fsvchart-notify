<template>
  <div class="app-container">
    <header class="app-header">
      <h1>fsvchart-notify Manager</h1>

      <div v-if="authStore.isLoggedIn" class="user-info">
        <span>{{ authStore.userDisplayName }}</span>
        <div class="dropdown">
          <button class="dropdown-toggle">
            <svg class="user-icon" viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
              <circle cx="12" cy="7" r="4" />
            </svg>
          </button>
          <div class="dropdown-menu">
            <RouterLink to="/profile">个人信息</RouterLink>
            <a href="#" @click.prevent="handleLogout">退出登录</a>
          </div>
        </div>
      </div>
    </header>

    <nav v-if="authStore.isLoggedIn">
      <RouterLink to="/">首页</RouterLink>
      <RouterLink to="/send-records">发送记录</RouterLink>
    </nav>

    <div class="main-content">
      <RouterView />
    </div>

    <AppNotification />
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { onMounted, onUnmounted } from 'vue'
import { useAuthStore } from './stores/auth'
import AppNotification from './components/AppNotification.vue'

const router = useRouter()
const authStore = useAuthStore()

function handleLogout() {
  authStore.logout()
  router.push('/login')
}

function onStorageChange() {
  authStore.checkAuth()
}

onMounted(() => {
  window.addEventListener('storage', onStorageChange)
})

onUnmounted(() => {
  window.removeEventListener('storage', onStorageChange)
})
</script>

<style>
.app-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.app-header {
  padding: 0 var(--spacing-lg);
  background: var(--color-bg-light);
  border-bottom: 1px solid var(--color-border-light);
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 60px;
  position: relative;
  z-index: 100;
}

.app-header h1 {
  margin: 0;
  padding: 15px 0;
  font-size: 1.5rem;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-icon {
  color: var(--color-text-secondary);
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
  display: flex;
  align-items: center;
}

.dropdown-menu {
  display: none;
  position: absolute;
  right: 0;
  background-color: var(--color-bg-white);
  min-width: 120px;
  box-shadow: var(--shadow-dropdown);
  z-index: 1;
  border-radius: var(--radius-md);
}

.dropdown-menu a {
  color: var(--color-text);
  padding: 10px 15px;
  text-decoration: none;
  display: block;
  text-align: left;
}

.dropdown-menu a:hover {
  background-color: var(--color-bg-hover);
}

.dropdown:hover .dropdown-menu {
  display: block;
}

nav {
  padding: 10px var(--spacing-lg);
  background: var(--color-bg-white);
  border-bottom: 1px solid var(--color-border-light);
  display: flex;
  align-items: center;
}

nav a {
  margin-right: 15px;
  text-decoration: none;
  color: var(--color-text-secondary);
  padding: 5px 10px;
  border-radius: var(--radius-md);
}

nav a:hover {
  background: var(--color-bg-hover);
}

nav a.router-link-active {
  color: var(--color-primary);
  font-weight: bold;
}

.main-content {
  flex: 1;
  padding: var(--spacing-lg);
  background: var(--color-bg-white);
  margin: var(--spacing-lg);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
}
</style>
