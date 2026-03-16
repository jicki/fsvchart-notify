<template>
  <div class="notification-container">
    <transition-group name="notification">
      <div
        v-for="notification in notifications"
        :key="notification.id"
        :class="['notification', `notification-${notification.type}`]"
        @click="removeNotification(notification.id)"
      >
        <span class="notification-icon">{{ iconMap[notification.type] }}</span>
        <span class="notification-message">{{ notification.message }}</span>
      </div>
    </transition-group>
  </div>
</template>

<script setup lang="ts">
import { useNotification } from '../composables/useNotification'

const { notifications, removeNotification } = useNotification()

const iconMap: Record<string, string> = {
  success: '\u2713',
  error: '\u2717',
  warning: '\u26A0',
  info: '\u2139'
}
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
  gap: 8px;
  padding: 12px 16px;
  border-radius: var(--radius-md, 4px);
  box-shadow: var(--shadow-md, 0 2px 10px rgba(0, 0, 0, 0.1));
  cursor: pointer;
  font-size: 14px;
  line-height: 1.4;
  animation: slideIn 0.3s ease;
}

.notification-success {
  background-color: #d4edda;
  color: #155724;
  border: 1px solid #c3e6cb;
}

.notification-error {
  background-color: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
}

.notification-warning {
  background-color: #fff3cd;
  color: #856404;
  border: 1px solid #ffeeba;
}

.notification-info {
  background-color: #d1ecf1;
  color: #0c5460;
  border: 1px solid #bee5eb;
}

.notification-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.notification-message {
  flex: 1;
  word-break: break-word;
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
