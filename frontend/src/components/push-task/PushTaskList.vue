<template>
  <div class="mt-lg">
    <h4 class="card-title">任务列表</h4>
    <div v-if="tasks.length === 0" class="empty">
      暂无任务，请创建新任务
    </div>
    <div v-else class="task-cards">
      <div
        v-for="task in tasks"
        :key="task.id"
        class="task-card"
        :class="{ 'task-card--editing': editingTaskId === task.id }"
      >
        <!-- 卡片头部：ID、名称、状态、操作 -->
        <div class="task-card__header">
          <div class="task-card__title">
            <span class="task-card__id">#{{ task.id }}</span>
            <span class="task-card__name">{{ task.name }}</span>
          </div>
          <div class="task-card__actions">
            <span :class="['badge', task.enabled ? 'badge-success' : 'badge-warning']">
              {{ task.enabled ? '启用' : '禁用' }}
            </span>
            <div v-if="isAdmin" class="action-group">
              <button class="btn-icon" @click.prevent="$emit('edit', task)" title="编辑">
                <IconEdit :size="16" />
              </button>
              <button class="btn-icon btn-icon-purple" @click.prevent="$emit('copy', task)" title="复制">
                <IconCopy :size="16" />
              </button>
              <button class="btn-icon btn-icon-success" @click.prevent="$emit('run', task.id)" title="执行">
                <IconPlay :size="16" />
              </button>
              <button
                class="btn-icon"
                :class="task.enabled ? 'btn-icon-warning' : ''"
                @click.prevent="$emit('toggle', task.id, !task.enabled)"
                :title="task.enabled ? '禁用' : '启用'"
              >
                <IconToggleRight v-if="task.enabled" :size="16" />
                <IconToggleLeft v-else :size="16" />
              </button>
              <button class="btn-icon btn-icon-danger" @click.prevent="$emit('delete', task.id)" title="删除">
                <IconTrash :size="16" />
              </button>
            </div>
          </div>
        </div>

        <!-- 元数据区域 -->
        <div class="task-card__meta">
          <div class="task-card__meta-item">
            <span class="task-card__meta-label">数据源</span>
            <span>{{ getSourceName(task.source_id) }}</span>
          </div>
          <div class="task-card__meta-item">
            <span class="task-card__meta-label">时间范围</span>
            <span>{{ formatTimeRange(task.time_range) }}</span>
          </div>
          <div class="task-card__meta-item">
            <span class="task-card__meta-label">发送时间</span>
            <span v-if="task.send_times && task.send_times.length > 0">
              <span v-for="(time, index) in task.send_times" :key="index">
                {{ getWeekdayText(time.weekday) }} {{ time.send_time }}{{ index < task.send_times.length - 1 ? '、' : '' }}
              </span>
            </span>
            <span v-else class="text-muted">未设置</span>
          </div>
        </div>

        <!-- 详情区域 -->
        <div class="task-card__details">
          <div class="task-card__detail-row">
            <span class="task-card__meta-label">查询</span>
            <div v-if="task.promql_configs && task.promql_configs.length > 0" class="promql-configs">
              <div v-for="(config, index) in task.promql_configs" :key="index" class="promql-config-item">
                <span class="promql-name">{{ config.promql_name }}</span>
                <span v-if="config.display_order !== undefined && config.display_order !== 0" class="config-tag">#{{ config.display_order }}</span>
                <span v-if="config.display_mode" class="config-tag" :class="`mode-${config.display_mode}`">
                  {{ config.display_mode === 'chart' ? '图表' : config.display_mode === 'text' ? '文本' : '混合' }}
                </span>
                <span v-if="config.initial_unit" class="config-tag">{{ config.initial_unit }}</span>
                <span v-if="config.unit" class="config-tag">{{ config.unit }}</span>
                <span v-if="config.custom_metric_label || config.metric_label" class="config-tag">
                  {{ config.custom_metric_label || config.metric_label }}
                </span>
              </div>
            </div>
            <div v-else-if="task.promql_ids && task.promql_ids.length > 0" class="promql-tags">
              <span v-for="promqlId in task.promql_ids" :key="promqlId" class="badge badge-success">
                {{ getPromqlName(promqlId) }}
              </span>
            </div>
            <span v-else class="text-muted">未选择</span>
          </div>
          <div class="task-card__detail-row">
            <span class="task-card__meta-label">WebHook</span>
            <div v-if="task.bound_webhooks && task.bound_webhooks.length > 0" class="webhook-tags">
              <span v-for="webhook in task.bound_webhooks" :key="webhook.id" class="webhook-tag">
                {{ webhook.name }}
              </span>
            </div>
            <span v-else class="text-muted">未绑定</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatTimeRange, getWeekdayText } from '../../utils/formatters'
import { IconEdit, IconCopy, IconPlay, IconToggleLeft, IconToggleRight, IconTrash } from '../icons'
import type { PushTask } from '../../types'

defineProps<{
  tasks: PushTask[]
  editingTaskId: number | null
  isAdmin: boolean
  getSourceName: (id: number) => string
  getPromqlName: (id: number) => string
}>()

defineEmits<{
  edit: [task: PushTask]
  copy: [task: PushTask]
  run: [taskId: number]
  toggle: [taskId: number, enabled: boolean]
  delete: [taskId: number]
}>()
</script>

<style scoped>
.task-cards {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.task-card {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-bg-white);
  box-shadow: var(--shadow-sm);
  padding: var(--spacing-lg);
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.task-card:hover {
  border-color: var(--color-border-focus);
  box-shadow: var(--shadow-md);
}

.task-card--editing {
  background-color: var(--color-accent-light);
  border-color: var(--color-accent);
}

.task-card__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-md);
}

.task-card__title {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.task-card__id {
  color: var(--color-text-muted);
  font-size: 13px;
  font-weight: 500;
}

.task-card__name {
  font-weight: 600;
  font-size: 15px;
  color: var(--color-text);
}

.task-card__actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.task-card__meta {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--spacing-md);
  padding: var(--spacing-md) 0;
  border-top: 1px solid var(--color-border);
  border-bottom: 1px solid var(--color-border);
  font-size: 14px;
}

.task-card__meta-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.task-card__meta-label {
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-secondary);
}

.task-card__details {
  padding-top: var(--spacing-md);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  font-size: 14px;
}

.task-card__detail-row {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-md);
}

.task-card__detail-row > .task-card__meta-label {
  min-width: 60px;
  padding-top: 2px;
}

.promql-configs {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.promql-config-item {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
}

.promql-name {
  font-weight: 500;
}

.config-tag {
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 4px;
  background: var(--color-bg-light);
  color: var(--color-text-secondary);
}

.config-tag.mode-chart { color: var(--color-accent); }
.config-tag.mode-text { color: var(--color-warning); }
.config-tag.mode-both { color: var(--color-purple); }

.promql-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.webhook-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.webhook-tag {
  font-size: 14px;
  color: var(--color-accent);
}

@media (max-width: 768px) {
  .task-card__meta {
    grid-template-columns: 1fr;
  }

  .task-card__header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-sm);
  }
}
</style>
