<template>
  <div class="notification-container">
    <transition-group name="notification">
      <div
        v-for="notification in notifications"
        :key="notification.id"
        :class="['notification', `notification-${notification.type}`]"
      >
        <span class="notification-icon">
          <IconCheck v-if="notification.type === 'success'" :size="18" />
          <IconX v-if="notification.type === 'error'" :size="18" />
          <IconAlertTriangle v-if="notification.type === 'warning'" :size="18" />
          <IconInfo v-if="notification.type === 'info'" :size="18" />
        </span>
        <span class="notification-message">{{ notification.message }}</span>
        <button class="notification-close" @click="removeNotification(notification.id)">
          <IconX :size="14" />
        </button>
      </div>
    </transition-group>
  </div>
</template>

<script setup lang="ts">
import { useNotification } from '../composables/useNotification'
import { IconCheck, IconX, IconAlertTriangle, IconInfo } from './icons'

const { notifications, removeNotification } = useNotification()
</script>

<style scoped>
.notification-container {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-width: 400px;
}

.notification {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  font-size: 14px;
  line-height: 1.4;
  animation: slideIn 0.3s ease;
  background: var(--color-bg-white);
  border: 1px solid var(--color-border);
}

.notification-success {
  border-left: 4px solid var(--color-success);
}

.notification-success .notification-icon {
  color: var(--color-success);
}

.notification-error {
  border-left: 4px solid var(--color-danger);
}

.notification-error .notification-icon {
  color: var(--color-danger);
}

.notification-warning {
  border-left: 4px solid var(--color-warning);
}

.notification-warning .notification-icon {
  color: var(--color-warning);
}

.notification-info {
  border-left: 4px solid var(--color-info);
}

.notification-info .notification-icon {
  color: var(--color-info);
}

.notification-icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.notification-message {
  flex: 1;
  word-break: break-word;
  color: var(--color-text);
}

.notification-close {
  flex-shrink: 0;
  background: none;
  border: none;
  cursor: pointer;
  padding: 2px;
  color: var(--color-text-muted);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  transition: color var(--transition-fast);
}

.notification-close:hover {
  color: var(--color-text);
}

.notification-enter-active,
.notification-leave-active {
  transition: all 0.3s ease;
}

.notification-enter-from {
  opacity: 0;
  transform: translateX(100%);
}

.notification-leave-to {
  opacity: 0;
  transform: translateX(100%);
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateX(100%);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}
</style>
