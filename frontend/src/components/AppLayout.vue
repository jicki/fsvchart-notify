<template>
  <div class="app-layout" :class="{ collapsed: sidebarCollapsed }">
    <aside class="sidebar">
      <div class="sidebar-header">
        <div class="sidebar-brand">
          <span class="brand-icon">
            <IconBarChart :size="24" />
          </span>
          <span v-show="!sidebarCollapsed" class="brand-text">FSVChart Notify</span>
        </div>
        <button class="sidebar-toggle" @click="toggleSidebar" :title="sidebarCollapsed ? '展开侧边栏' : '折叠侧边栏'">
          <IconChevronLeft v-if="!sidebarCollapsed" :size="18" />
          <IconChevronRight v-else :size="18" />
        </button>
      </div>

      <nav class="sidebar-nav">
        <RouterLink
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="nav-item"
          :title="sidebarCollapsed ? item.label : ''"
        >
          <component :is="item.icon" :size="20" />
          <span v-show="!sidebarCollapsed" class="nav-label">{{ item.label }}</span>
        </RouterLink>
      </nav>

      <div class="sidebar-footer">
        <div class="user-info" :title="sidebarCollapsed ? authStore.userDisplayName : ''">
          <IconUser :size="18" />
          <span v-show="!sidebarCollapsed" class="user-name">{{ authStore.userDisplayName }}</span>
        </div>
        <button class="nav-item logout-btn" @click="handleLogout" :title="sidebarCollapsed ? '退出登录' : ''">
          <IconLogOut :size="18" />
          <span v-show="!sidebarCollapsed" class="nav-label">退出登录</span>
        </button>
      </div>
    </aside>

    <main class="main-area">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, type Component } from 'vue'
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import {
  IconBarChart,
  IconChevronLeft,
  IconChevronRight,
  IconUser,
  IconLogOut,
  IconDatabase,
  IconWebhook,
  IconCode,
  IconSend,
  IconFileText
} from './icons'

const router = useRouter()
const authStore = useAuthStore()

const SIDEBAR_KEY = 'sidebar-collapsed'
const sidebarCollapsed = ref(false)

interface NavItem {
  path: string
  label: string
  icon: Component
}

const navItems: NavItem[] = [
  { path: '/datasources', label: '数据源', icon: IconDatabase },
  { path: '/webhooks', label: 'WebHook', icon: IconWebhook },
  { path: '/promql', label: 'PromQL', icon: IconCode },
  { path: '/push-tasks', label: '推送任务', icon: IconSend },
  { path: '/send-records', label: '发送记录', icon: IconFileText },
]

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
  localStorage.setItem(SIDEBAR_KEY, String(sidebarCollapsed.value))
}

function handleLogout() {
  authStore.logout()
  router.push('/login')
}

onMounted(() => {
  const saved = localStorage.getItem(SIDEBAR_KEY)
  if (saved === 'true') {
    sidebarCollapsed.value = true
  }
})
</script>

<style>
.app-layout {
  display: flex;
  min-height: 100vh;
}

/* 侧边栏 */
.sidebar {
  width: var(--sidebar-width);
  background: var(--color-bg-sidebar);
  color: var(--color-text-sidebar);
  display: flex;
  flex-direction: column;
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  z-index: 100;
  transition: width var(--transition-normal);
  overflow: hidden;
}

.collapsed .sidebar {
  width: var(--sidebar-collapsed-width);
}

/* 侧边栏头部 */
.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  min-height: 60px;
}

.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  overflow: hidden;
}

.brand-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--color-accent);
}

.brand-text {
  font-size: 16px;
  font-weight: 600;
  white-space: nowrap;
  color: var(--color-text-white);
}

.sidebar-toggle {
  background: none;
  border: none;
  color: var(--color-text-sidebar);
  cursor: pointer;
  padding: 4px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background var(--transition-fast);
  flex-shrink: 0;
}

.sidebar-toggle:hover {
  background: var(--color-bg-sidebar-hover);
}

.collapsed .sidebar-toggle {
  display: none;
}

/* 导航 */
.sidebar-nav {
  flex: 1;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: var(--radius-md);
  color: var(--color-text-sidebar);
  text-decoration: none;
  font-size: 14px;
  transition: all var(--transition-fast);
  white-space: nowrap;
  overflow: hidden;
  border: none;
  background: none;
  cursor: pointer;
  width: 100%;
  text-align: left;
}

.nav-item:hover {
  background: var(--color-bg-sidebar-hover);
  color: var(--color-text-white);
}

.nav-item.router-link-active {
  background: var(--color-bg-sidebar-active);
  color: var(--color-text-white);
  font-weight: 500;
}

.nav-label {
  overflow: hidden;
  text-overflow: ellipsis;
}

.collapsed .nav-item {
  justify-content: center;
  padding: 10px;
}

/* 侧边栏底部 */
.sidebar-footer {
  padding: 8px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  font-size: 14px;
  color: var(--color-text-sidebar);
  overflow: hidden;
}

.user-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collapsed .user-info {
  justify-content: center;
}

.logout-btn {
  color: var(--color-text-sidebar) !important;
  font-size: 14px;
}

.logout-btn:hover {
  color: var(--color-danger) !important;
}

/* 主内容区 */
.main-area {
  flex: 1;
  margin-left: var(--sidebar-width);
  padding: var(--spacing-lg);
  background: var(--color-bg-page);
  transition: margin-left var(--transition-normal);
  min-height: 100vh;
}

.collapsed .main-area {
  margin-left: var(--sidebar-collapsed-width);
}
</style>
