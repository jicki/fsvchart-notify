<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="visible" class="modal-overlay" @click.self="handleOverlayClick">
        <div class="modal-container" :style="{ maxWidth: maxWidth }">
          <div class="modal-header">
            <h3 class="modal-title">{{ title }}</h3>
            <button class="btn-icon modal-close" @click="$emit('close')" title="关闭">
              <IconX :size="18" />
            </button>
          </div>
          <div class="modal-body">
            <slot />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { IconX } from './icons'

const props = defineProps<{
  visible: boolean
  title: string
  maxWidth?: string
  closeOnOverlay?: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

function handleOverlayClick() {
  if (props.closeOnOverlay !== false) {
    emit('close')
  }
}

// 打开时禁止背景滚动
watch(() => props.visible, (val) => {
  document.body.style.overflow = val ? 'hidden' : ''
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 40px 16px;
  overflow-y: auto;
}

.modal-container {
  background: var(--color-bg-white);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  width: 100%;
  max-width: 900px;
  animation: modal-in 0.2s ease-out;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-md) var(--spacing-lg);
  border-bottom: 1px solid var(--color-border);
}

.modal-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
}

.modal-close {
  color: var(--color-text-muted);
}

.modal-close:hover {
  color: var(--color-text);
}

.modal-body {
  padding: var(--spacing-lg);
  max-height: calc(100vh - 160px);
  overflow-y: auto;
}

/* 过渡动画 */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

@keyframes modal-in {
  from {
    opacity: 0;
    transform: translateY(-10px) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}
</style>
