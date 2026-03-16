import { ref } from 'vue'
import type { Notification, NotificationType } from '../types'

const notifications = ref<Notification[]>([])
let nextId = 0

function addNotification(type: NotificationType, message: string, duration?: number) {
  const defaultDuration = type === 'error' ? 5000 : 3000
  const id = nextId++
  const notification: Notification = {
    id,
    type,
    message,
    duration: duration ?? defaultDuration
  }
  notifications.value.push(notification)

  setTimeout(() => {
    removeNotification(id)
  }, notification.duration)
}

function removeNotification(id: number) {
  const index = notifications.value.findIndex(n => n.id === id)
  if (index !== -1) {
    notifications.value.splice(index, 1)
  }
}

export function useNotification() {
  return {
    notifications,
    showSuccess: (msg: string) => addNotification('success', msg),
    showError: (msg: string) => addNotification('error', msg),
    showWarning: (msg: string) => addNotification('warning', msg),
    showInfo: (msg: string) => addNotification('info', msg),
    removeNotification
  }
}
