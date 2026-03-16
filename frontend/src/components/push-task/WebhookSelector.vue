<template>
  <div class="form-group">
    <label>选择要发送的WebHook(多选):</label>
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
      <div v-else class="no-webhooks">
        <p>暂无可用的 WebHook，请先在 WebHook 管理中创建</p>
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
.webhook-selection { max-height: 200px; overflow-y: auto; border: 1px solid var(--color-border, #ddd); border-radius: 4px; padding: 8px; margin-bottom: 8px; }
.webhook-list { display: flex; flex-direction: column; gap: 8px; }
.webhook-item { display: flex; align-items: center; gap: 8px; }
.no-webhooks { padding: 12px; background-color: var(--color-bg-light, #f8f9fa); border-radius: 4px; text-align: center; color: var(--color-text-muted, #6c757d); }
.form-hint { font-size: 0.85em; color: var(--color-text-secondary, #666); margin-top: 5px; }
</style>
