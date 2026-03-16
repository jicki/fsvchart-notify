<template>
  <div class="card" style="margin-top: var(--spacing-lg)">
    <h4>任务列表</h4>
    <div v-if="tasks.length === 0" class="empty">
      暂无任务，请创建新任务
    </div>
    <table v-else class="data-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>名称</th>
          <th>数据源</th>
          <th>时间范围</th>
          <th>发送时间</th>
          <th>查询/图表标签</th>
          <th>绑定WebHook</th>
          <th>状态</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="task in tasks" :key="task.id" :class="{ 'editing-row': editingTaskId === task.id }">
          <td>{{ task.id }}</td>
          <td>{{ task.name }}</td>
          <td>{{ getSourceName(task.source_id) }}</td>
          <td>{{ formatTimeRange(task.time_range) }}</td>
          <td>
            <div class="send-times-list">
              <template v-if="task.send_times && task.send_times.length > 0">
                <div v-for="(time, index) in task.send_times" :key="index" class="send-time">
                  {{ getWeekdayText(time.weekday) }} {{ time.send_time }}
                </div>
              </template>
              <span v-else class="text-muted">未设置</span>
            </div>
          </td>
          <td>
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
          </td>
          <td>
            <div v-if="task.bound_webhooks && task.bound_webhooks.length > 0" class="webhook-tags">
              <span v-for="webhook in task.bound_webhooks" :key="webhook.id" class="webhook-tag">
                {{ webhook.name }}
              </span>
            </div>
            <span v-else class="text-muted">未绑定</span>
          </td>
          <td>
            <span :class="['badge', task.enabled ? 'badge-success' : 'badge-warning']">
              {{ task.enabled ? '启用' : '禁用' }}
            </span>
          </td>
          <td>
            <div class="action-group">
              <button class="btn-icon" @click.prevent="$emit('edit', task)" title="编辑">
                <IconEdit :size="16" />
              </button>
              <button class="btn-icon" @click.prevent="$emit('copy', task)" title="复制" style="color: var(--color-purple)">
                <IconCopy :size="16" />
              </button>
              <button class="btn-icon" @click.prevent="$emit('run', task.id)" title="执行" style="color: var(--color-success)">
                <IconPlay :size="16" />
              </button>
              <button
                class="btn-icon"
                @click.prevent="$emit('toggle', task.id, !task.enabled)"
                :title="task.enabled ? '禁用' : '启用'"
                :style="{ color: task.enabled ? 'var(--color-warning)' : 'var(--color-text-muted)' }"
              >
                <IconToggleRight v-if="task.enabled" :size="16" />
                <IconToggleLeft v-else :size="16" />
              </button>
              <button class="btn-icon" @click.prevent="$emit('delete', task.id)" title="删除" style="color: var(--color-danger)">
                <IconTrash :size="16" />
              </button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { formatTimeRange, getWeekdayText } from '../../utils/formatters'
import { IconEdit, IconCopy, IconPlay, IconToggleLeft, IconToggleRight, IconTrash } from '../icons'
import type { PushTask } from '../../types'

defineProps<{
  tasks: PushTask[]
  editingTaskId: number | null
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
h4 {
  margin: 0 0 var(--spacing-md);
  font-weight: 600;
}

.editing-row {
  background-color: var(--color-accent-light) !important;
}

.text-muted {
  color: var(--color-text-muted);
  font-style: italic;
  font-size: 13px;
}

.send-times-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.send-time {
  font-size: 13px;
  color: var(--color-text-secondary);
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
  font-size: 13px;
}

.config-tag {
  font-size: 11px;
  padding: 1px 6px;
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
  font-size: 13px;
  color: var(--color-accent);
}

.action-group {
  display: flex;
  gap: 2px;
  align-items: center;
}
</style>
