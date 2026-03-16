<template>
  <RouterView />
  <AppNotification />
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'
import { onMounted, onUnmounted } from 'vue'
import { useAuthStore } from './stores/auth'
import AppNotification from './components/AppNotification.vue'

const authStore = useAuthStore()

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
