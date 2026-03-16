<template>
  <div class="tab-content">
    <h3>Push Tasks</h3>

    <div v-if="tasks.length === 0 && !store.loading" class="no-data-message">
      <p>暂无任务数据</p>
    </div>

    <!-- 创建任务表单 -->
    <PushTaskForm
      :form="createForm"
      :sources="store.sources"
      :webhooks="store.webhooks"
      :chart-templates="store.chartTemplates"
      :promqls="store.promqls"
      :is-editing="false"
      @submit="handleCreate"
    />

    <hr />

    <!-- 编辑任务表单 -->
    <PushTaskForm
      v-if="isEditing"
      id="edit-task-form"
      :form="editForm"
      :sources="store.sources"
      :webhooks="store.webhooks"
      :chart-templates="store.chartTemplates"
      :promqls="store.promqls"
      :is-editing="true"
      @submit="handleUpdate"
      @cancel="cancelEdit"
    />

    <!-- 任务列表 -->
    <PushTaskList
      :tasks="tasks"
      :editing-task-id="editingTaskId"
      :get-source-name="store.getSourceName"
      :get-promql-name="store.getPromqlName"
      @edit="startEdit"
      @copy="handleCopy"
      @run="store.runTask"
      @toggle="store.toggleTask"
      @delete="store.deleteTask"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'
import { usePushTaskStore } from '../stores/pushTask'
import { usePushTaskForm } from '../composables/usePushTaskForm'
import { usePolling } from '../composables/usePolling'
import PushTaskForm from '../components/push-task/PushTaskForm.vue'
import PushTaskList from '../components/push-task/PushTaskList.vue'
import type { PushTask } from '../types'

const store = usePushTaskStore()
const tasks = computed(() => store.tasks)

// 创建表单
const createForm = usePushTaskForm()

// 编辑表单
const editForm = usePushTaskForm()
const isEditing = ref(false)
const editingTaskId = ref<number | null>(null)

// 定期刷新
usePolling(() => store.fetchAllData(), 30000)

// 创建任务
async function handleCreate(payload: Record<string, unknown>) {
  const success = await store.createTask(payload)
  if (success) {
    createForm.resetForm()
  }
}

// 开始编辑
function startEdit(task: PushTask) {
  isEditing.value = true
  editingTaskId.value = task.id
  editForm.loadTask(task)

  setTimeout(() => {
    const editFormEl = document.getElementById('edit-task-form')
    if (editFormEl) {
      editFormEl.scrollIntoView({ behavior: 'smooth' })
    }
  }, 100)
}

// 更新任务
async function handleUpdate(payload: Record<string, unknown>) {
  if (editingTaskId.value === null) return
  const success = await store.updateTask(editingTaskId.value, {
    ...payload,
    id: editingTaskId.value
  })
  if (success) {
    cancelEdit()
  }
}

// 取消编辑
function cancelEdit() {
  isEditing.value = false
  editingTaskId.value = null
  editForm.resetForm()
}

// 复制任务
function handleCopy(task: PushTask) {
  store.copyTask(task)
}
</script>

<style scoped>
.tab-content { border: 1px solid var(--color-border-table, #ccc); padding: 16px; margin-bottom: 24px; }
.no-data-message { text-align: center; padding: 20px; background-color: var(--color-bg-page, #f5f5f5); border-radius: 4px; margin: 20px 0; }
</style>
