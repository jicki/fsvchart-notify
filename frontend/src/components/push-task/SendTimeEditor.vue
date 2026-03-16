<template>
  <div class="form-group">
    <label>发送时间设置</label>
    <div class="send-times">
      <div v-for="(time, index) in sendTimes" :key="time._id || index" class="send-time-item">
        <select class="form-input" v-model="time.weekday">
          <option value="1">周一</option>
          <option value="2">周二</option>
          <option value="3">周三</option>
          <option value="4">周四</option>
          <option value="5">周五</option>
          <option value="6">周六</option>
          <option value="7">周日</option>
        </select>
        <input class="form-input" type="time" v-model="time.send_time" required />
        <button type="button" class="btn-icon btn-icon-danger" @click="$emit('remove', index)" title="删除">
          <IconTrash :size="16" />
        </button>
      </div>
      <button type="button" class="btn btn-secondary btn-sm" @click="$emit('add')">
        <IconPlus :size="14" />
        添加发送时间
      </button>
    </div>
    <div class="form-hint">可以添加多个发送时间，每个时间点都会触发发送</div>
  </div>
</template>

<script setup lang="ts">
import { IconTrash, IconPlus } from '../icons'
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
.send-times {
  margin: var(--spacing-sm) 0;
}

.send-time-item {
  display: flex;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
  align-items: center;
}

.send-time-item .form-input {
  width: auto;
  min-width: 120px;
}
</style>
