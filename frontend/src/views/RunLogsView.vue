<template>
  <div>
    <div class="page-header">
      <div>
        <h3>发送记录</h3>
        <p>查看推送任务的执行历史记录</p>
      </div>
      <div class="header-actions">
        <label class="toggle">
          <input type="checkbox" v-model="autoRefresh">
          <span class="toggle-track"></span>
          <span class="toggle-label">自动刷新</span>
        </label>
        <button class="btn btn-secondary" @click="fetchRecords" :disabled="loading">
          <IconRefresh :size="16" />
          {{ loading ? '加载中...' : '刷新' }}
        </button>
      </div>
    </div>

    <div class="card">
      <div v-if="loading" class="loading">加载中...</div>
      <div v-else-if="error" class="error-state">{{ error }}</div>
      <div v-else-if="!records || records.length === 0" class="empty">暂无记录</div>
      <div v-else class="records">
        <div v-for="record in records"
             :key="record.id"
             class="record-card"
             :class="getStatusClass(record.status)">
          <div class="record-header">
            <span class="record-task-name">{{ record.task_name || '未命名任务' }}</span>
            <span :class="['badge', getBadgeClass(record.status)]">{{ record.status }}</span>
          </div>
          <div class="record-time">{{ formatDate(record.timestamp) }}</div>
          <div class="record-message">{{ record.message }}</div>
          <div class="record-detail">
            <div>Webhook: {{ record.webhook }}</div>
            <div v-if="record.query">查询: {{ record.query }}</div>
            <div v-if="record.time_range">时间范围: {{ record.time_range }}</div>
            <div v-if="record.button_text">按钮文本: {{ record.button_text }}</div>
            <div v-if="record.button_url">
              按钮链接: <a :href="record.button_url" target="_blank" class="link">{{ record.button_url }}</a>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { get } from '../utils/api'
import { formatDate } from '../utils/formatters'
import { IconRefresh } from '../components/icons'
import type { SendRecord } from '../types'

const records = ref<SendRecord[]>([])
const loading = ref(false)
const error = ref('')
const autoRefresh = ref(false)

let refreshInterval: number | null = null

async function fetchRecords() {
  loading.value = true
  error.value = ''

  try {
    records.value = await get<SendRecord[]>('/api/send_records')
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '获取记录失败'
  } finally {
    loading.value = false
  }
}

watch(autoRefresh, (newVal) => {
  if (newVal) {
    refreshInterval = window.setInterval(fetchRecords, 5000)
  } else if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
})

onMounted(fetchRecords)

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})

function getStatusClass(status: string): string {
  switch (status.toLowerCase()) {
    case 'success': return 'status-success'
    case 'failed': return 'status-error'
    case 'pending': return 'status-pending'
    default: return ''
  }
}

function getBadgeClass(status: string): string {
  switch (status.toLowerCase()) {
    case 'success': return 'badge-success'
    case 'failed': return 'badge-danger'
    case 'pending': return 'badge-warning'
    default: return 'badge-info'
  }
}
</script>

<style scoped>
.header-actions {
  display: flex;
  gap: var(--spacing-md);
  align-items: center;
}

.toggle-label {
  font-size: 14px;
  color: var(--color-text-secondary);
}

.records {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.record-card {
  padding: var(--spacing-md);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
  border-left: 4px solid var(--color-border);
  background: var(--color-bg-white);
  transition: border-color var(--transition-fast);
}

.record-card.status-success {
  border-left-color: var(--color-success);
}

.record-card.status-error {
  border-left-color: var(--color-danger);
}

.record-card.status-pending {
  border-left-color: var(--color-warning);
}

.record-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.record-task-name {
  font-weight: 600;
  font-size: 15px;
  color: var(--color-text);
}

.record-time {
  color: var(--color-text-muted);
  font-size: 13px;
  margin-bottom: 6px;
}

.record-message {
  margin-bottom: 8px;
  color: var(--color-text-secondary);
  font-size: 14px;
}

.record-detail {
  padding: var(--spacing-sm) var(--spacing-md);
  background: var(--color-bg-light);
  border-radius: var(--radius-md);
  font-size: 13px;
  color: var(--color-text-secondary);
  word-break: break-all;
}

.record-detail div {
  margin-bottom: 2px;
}

.link {
  color: var(--color-accent);
  text-decoration: underline;
  word-break: break-all;
}

.link:hover {
  text-decoration: none;
}

.error-state {
  padding: var(--spacing-lg);
  text-align: center;
  color: var(--color-danger);
}
</style>
