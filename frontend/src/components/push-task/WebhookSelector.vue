<template>
  <div class="form-group">
    <label>选择要发送的WebHook(多选)</label>
    <div class="webhook-selection">
      <div v-if="webhooks.length > 0" class="webhook-list">
        <div v-for="webhook in webhooks" :key="webhook.id" class="webhook-item">
          <input
            type="checkbox"
            :id="`${idPrefix}-webhook-${webhook.id}`"
            :value="webhook.id"
            v-model="selectedIds"
          >
          <label :for="`${idPrefix}-webhook-${webhook.id}`">{{ webhook.name }}</label>
        </div>
      </div>
      <div v-else class="empty">
        暂无可用的 WebHook，请先在 WebHook 管理中创建
      </div>
    </div>
    <div class="form-hint">请选择要发送的 WebHook，可多选</div>
  </div>
</template>

<script setup lang="ts">
import type { FeishuWebhook } from '../../types'

defineProps<{
  webhooks: FeishuWebhook[]
  idPrefix: string
}>()

const selectedIds = defineModel<number[]>('selectedIds', { default: () => [] })
</script>

<style scoped>
.webhook-selection {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--spacing-sm) var(--spacing-md);
  margin-bottom: var(--spacing-sm);
  background: var(--color-bg-white);
}

.webhook-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.webhook-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}
</style>
