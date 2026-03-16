<template>
  <div>
    <div class="page-header">
      <div>
        <h3>推送任务</h3>
        <p>管理定时图表推送任务</p>
      </div>
    </div>

    <div v-if="tasks.length === 0 && !store.loading" class="card empty-state">
      <p>暂无任务数据，请创建新任务</p>
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
import { ref, computed } from 'vue'
import { usePushTaskStore } from '../stores/pushTask'
import { usePushTaskForm } from '../composables/usePushTaskForm'
import { usePolling } from '../composables/usePolling'
import PushTaskForm from '../components/push-task/PushTaskForm.vue'
import PushTaskList from '../components/push-task/PushTaskList.vue'
import type { PushTask } from '../types'

const store = usePushTaskStore()
const tasks = computed(() => store.tasks)

const createForm = usePushTaskForm()
const editForm = usePushTaskForm()
const isEditing = ref(false)
const editingTaskId = ref<number | null>(null)

usePolling(() => store.fetchAllData(), 30000)

async function handleCreate(payload: Record<string, unknown>) {
  const success = await store.createTask(payload)
  if (success) {
    createForm.resetForm()
  }
}

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

function cancelEdit() {
  isEditing.value = false
  editingTaskId.value = null
  editForm.resetForm()
}

function handleCopy(task: PushTask) {
  store.copyTask(task)
}
</script>

<style scoped>
.empty-state {
  text-align: center;
  padding: var(--spacing-xl);
  color: var(--color-text-secondary);
  margin-bottom: var(--spacing-lg);
}
</style>
