<template>
  <div class="tab-content">
    <div class="header">
      <h3>发送记录</h3>
      <div class="controls">
        <button @click="fetchRecords" :disabled="loading">
          {{ loading ? '加载中...' : '刷新记录' }}
        </button>
        <label>
          <input type="checkbox" v-model="autoRefresh"> 自动刷新
        </label>
      </div>
    </div>
    
    <div class="records-container" ref="recordsContainer">
      <div v-if="loading" class="loading">加载中...</div>
      <div v-else-if="error" class="error">{{ error }}</div>
      <div v-else-if="!records || records.length === 0" class="empty">暂无记录</div>
      <div v-else class="records">
        <div v-for="record in records" 
             :key="record.id" 
             class="record-entry"
             :class="getStatusClass(record.status)">
          <div class="record-time">{{ formatTime(record.timestamp) }}</div>
          <div class="record-task">{{ record.task_name || '未命名任务' }}</div>
          <div class="record-status">{{ record.status }}</div>
          <div class="record-message">{{ record.message }}</div>
          <div class="record-detail">
            <div>Webhook: {{ record.webhook }}</div>
            <div v-if="record.query">查询: {{ record.query }}</div>
            <div v-if="record.time_range">时间范围: {{ record.time_range }}</div>
            <div v-if="record.button_text">按钮文本: {{ record.button_text }}</div>
            <div v-if="record.button_url">
              按钮链接: <a :href="record.button_url" target="_blank" class="button-link">{{ record.button_url }}</a>
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

interface SendRecord {
  id: number
  timestamp: string
  status: string
  message: string
  webhook: string
  task_name: string
  query?: string
  time_range?: string
  button_text?: string
  button_url?: string
}

const records = ref<SendRecord[]>([])
const loading = ref(false)
const error = ref('')
const autoRefresh = ref(false)
const recordsContainer = ref<HTMLElement | null>(null)

let refreshInterval: number | null = null

async function fetchRecords() {
  loading.value = true
  error.value = ''
  
  try {
    records.value = await get('/api/send_records')
    
    // 滚动到底部
    if (recordsContainer.value) {
      setTimeout(() => {
        if (recordsContainer.value) {
          recordsContainer.value.scrollTop = recordsContainer.value.scrollHeight
        }
      }, 100)
    }
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

watch(autoRefresh, (newVal) => {
  if (newVal) {
    refreshInterval = window.setInterval(fetchRecords, 5000)
  } else if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})

onMounted(() => {
  fetchRecords()
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})

function formatTime(timestamp: string): string {
  return new Date(timestamp).toLocaleString()
}

function getStatusClass(status: string): string {
  switch (status.toLowerCase()) {
    case 'success': return 'success'
    case 'failed': return 'error'
    case 'pending': return 'pending'
    default: return ''
  }
}
</script>

<style scoped>
.tab-content {
  padding: 20px;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.controls {
  display: flex;
  gap: 16px;
  align-items: center;
}

.records-container {
  max-height: 600px;
  overflow-y: auto;
  border: 1px solid #e9ecef;
  border-radius: 4px;
}

.record-entry {
  padding: 12px;
  border-bottom: 1px solid #e9ecef;
  border-left: 4px solid transparent;
}

.record-entry:last-child {
  border-bottom: none;
}

.record-time {
  color: #666;
  font-size: 0.9em;
}

.record-task {
  font-weight: bold;
  font-size: 1.1em;
  margin-bottom: 8px;
  color: #333;
}

.record-status {
  font-weight: bold;
  margin: 4px 0;
}

.record-message {
  margin: 4px 0;
}

.record-detail {
  margin-top: 8px;
  padding: 8px;
  background: #f8f9fa;
  border-radius: 4px;
  font-size: 0.9em;
  word-break: break-all;
}

.record-entry.success {
  border-left-color: #28a745;
}

.record-entry.error {
  border-left-color: #dc3545;
}

.record-entry.pending {
  border-left-color: #ffc107;
}

.loading, .error, .empty {
  padding: 20px;
  text-align: center;
  color: #666;
}

.error {
  color: #dc3545;
}

button {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  background: #007bff;
  color: white;
  cursor: pointer;
}

button:disabled {
  background: #ccc;
  cursor: not-allowed;
}

label {
  display: flex;
  align-items: center;
  gap: 4px;
}

.button-link {
  color: #007bff;
  text-decoration: underline;
  word-break: break-all;
}

.button-link:hover {
  text-decoration: none;
}
</style>
