<template>
  <div class="form-group">
    <label>发送时间设置:</label>
    <div class="send-times">
      <div v-for="(time, index) in sendTimes" :key="time._id || index" class="send-time-item">
        <select v-model="time.weekday" class="form-control">
          <option value="1">周一</option>
          <option value="2">周二</option>
          <option value="3">周三</option>
          <option value="4">周四</option>
          <option value="5">周五</option>
          <option value="6">周六</option>
          <option value="7">周日</option>
        </select>
        <input type="time" v-model="time.send_time" class="form-control" required />
        <button type="button" @click="$emit('remove', index)" class="remove-time-btn">删除</button>
      </div>
      <button type="button" @click="$emit('add')" class="add-time-btn">添加发送时间</button>
    </div>
    <div class="form-hint">可以添加多个发送时间，每个时间点都会触发发送</div>
  </div>
</template>

<script setup lang="ts">
import type { SendTime } from '../../types'

defineProps<{
  sendTimes: SendTime[]
}>()

defineEmits<{
  add: []
  remove: [index: number]
}>()
</script>

<style scoped>
.send-times { margin: 1rem 0; }
.send-time-item { display: flex; gap: 1rem; margin-bottom: 0.5rem; align-items: center; }
.send-time-item select, .send-time-item input { padding: 0.5rem; border: 1px solid var(--color-border, #ddd); border-radius: 4px; min-width: 120px; }
.remove-time-btn { padding: 0.5rem 1rem; background-color: var(--color-danger, #dc3545); color: white; border: none; border-radius: 4px; cursor: pointer; }
.remove-time-btn:hover { background-color: var(--color-danger-hover, #c82333); }
.add-time-btn { margin-top: 0.5rem; padding: 0.5rem 1rem; background-color: var(--color-success, #28a745); color: white; border: none; border-radius: 4px; cursor: pointer; }
.add-time-btn:hover { background-color: var(--color-success-hover, #218838); }
.form-hint { font-size: 0.85em; color: var(--color-text-secondary, #666); margin-top: 5px; }
</style>
