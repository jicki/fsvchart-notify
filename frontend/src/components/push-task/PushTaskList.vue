<template>
  <div class="task-list">
    <h3>任务列表</h3>
    <div v-if="tasks.length === 0" class="no-tasks">
      <p>暂无任务，请创建新任务</p>
    </div>
    <table v-else>
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
        <tr v-for="task in tasks" :key="task.id" :class="{ 'editing-task': editingTaskId === task.id }">
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
              <div v-else class="no-times">未设置发送时间</div>
            </div>
          </td>
          <td>
            <div v-if="task.promql_configs && task.promql_configs.length > 0" class="promql-configs">
              <div v-for="(config, index) in task.promql_configs" :key="index" class="promql-config-item">
                <span class="promql-name">{{ config.promql_name }}</span>
                <span v-if="config.display_order !== undefined && config.display_order !== 0" class="config-detail" style="color: #409eff;">(顺序: {{ config.display_order }})</span>
                <span v-if="config.display_mode" class="config-detail" :style="{ color: config.display_mode === 'both' ? '#9c27b0' : (config.display_mode === 'text' ? '#f57c00' : '#1976d2') }">
                  ({{ config.display_mode === 'chart' ? '图表' : config.display_mode === 'text' ? '文本' : '混合' }})
                </span>
                <span v-if="config.initial_unit" class="config-detail">(初始单位: {{ config.initial_unit }})</span>
                <span v-if="config.unit" class="config-detail">(单位: {{ config.unit }})</span>
                <span v-if="config.custom_metric_label || config.metric_label" class="config-detail">
                  (标签: {{ config.custom_metric_label || config.metric_label }})
                </span>
              </div>
            </div>
            <div v-else-if="task.promql_ids && task.promql_ids.length > 0" class="promql-names">
              <span v-for="(promqlId, index) in task.promql_ids" :key="promqlId" class="promql-tag">
                {{ getPromqlName(promqlId) }}{{ index < task.promql_ids.length - 1 ? ', ' : '' }}
              </span>
            </div>
            <div v-else class="no-promql-selected"><span>未选择</span></div>
          </td>
          <td>
            <div v-if="task.bound_webhooks && task.bound_webhooks.length > 0" class="bound-webhooks">
              <span v-for="(webhook, index) in task.bound_webhooks" :key="webhook.id" class="webhook-tag">
                {{ webhook.name }}{{ index < task.bound_webhooks.length - 1 ? ', ' : '' }}
              </span>
            </div>
            <div v-else class="no-webhooks-bound"><span>未绑定</span></div>
          </td>
          <td>{{ task.enabled ? '启用' : '禁用' }}</td>
          <td>
            <div class="task-actions">
              <button class="edit-btn" @click.prevent="$emit('edit', task)">编辑</button>
              <button class="copy-btn" @click.prevent="$emit('copy', task)">复制</button>
              <button class="run-btn" @click.prevent="$emit('run', task.id)">执行</button>
              <button
                :class="task.enabled ? 'disable-btn' : 'enable-btn'"
                @click.prevent="$emit('toggle', task.id, !task.enabled)"
              >
                {{ task.enabled ? '禁用' : '启用' }}
              </button>
              <button class="delete-btn" @click.prevent="$emit('delete', task.id)">删除</button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { formatTimeRange, getWeekdayText } from '../../utils/formatters'
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
.no-tasks { padding: 12px; background-color: var(--color-bg-light, #f8f9fa); border-radius: 4px; text-align: center; color: var(--color-text-muted, #6c757d); }
table { margin-top: 10px; width: 100%; border-collapse: collapse; }
th, td { border: 1px solid var(--color-border-table, #ccc); padding: 4px 8px; text-align: left; }
.send-times-list { display: flex; flex-direction: column; gap: 0.25rem; }
.send-time { font-size: 0.9em; color: var(--color-text-secondary, #666); }
.no-times { font-style: italic; color: #999; }
.promql-config-item { margin-bottom: 3px; }
.promql-name { font-weight: 500; }
.config-detail { font-size: 0.85em; margin-left: 4px; }
.promql-tag { font-size: 0.9em; color: var(--color-success, #28a745); background-color: #f0f9f0; padding: 2px 6px; border-radius: 3px; margin-right: 3px; display: inline-block; margin-bottom: 3px; }
.promql-names { display: flex; flex-wrap: wrap; max-width: 200px; gap: 3px; }
.no-promql-selected, .no-webhooks-bound { color: var(--color-text-muted, #6c757d); font-style: italic; }
.bound-webhooks { display: flex; flex-wrap: wrap; gap: 4px; }
.webhook-tag { font-size: 0.9em; color: var(--color-primary-hover, #0056b3); }
.task-actions { display: flex; gap: 4px; flex-wrap: wrap; }
.task-actions button { padding: 4px 8px; font-size: 0.9em; border-radius: 4px; border: none; cursor: pointer; color: white; }
.edit-btn { background-color: var(--color-info, #17a2b8); }
.copy-btn { background-color: var(--color-purple, #6610f2); }
.run-btn { background-color: var(--color-success, #28a745); }
.disable-btn { background-color: var(--color-warning, #ffc107); color: #212529; }
.enable-btn { background-color: var(--color-text-muted, #6c757d); }
.delete-btn { background-color: var(--color-danger, #dc3545); }
.task-actions button:hover { opacity: 0.9; }
tr.editing-task { background-color: #e2f2fd !important; }
</style>
