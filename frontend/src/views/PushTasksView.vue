<template>
  <div>
    <div class="page-header">
      <div>
        <h3>推送任务</h3>
        <p>管理定时图表推送任务</p>
      </div>
      <button v-if="isAdmin" class="btn btn-primary" @click="openCreateModal">
        <IconPlus :size="16" />
        创建任务
      </button>
    </div>

    <div v-if="tasks.length === 0 && !store.loading" class="card empty-state">
      <p>暂无任务数据，请创建新任务</p>
    </div>

    <!-- 任务列表 -->
    <PushTaskList
      :tasks="tasks"
      :editing-task-id="editingTaskId"
      :is-admin="isAdmin"
      :get-source-name="store.getSourceName"
      :get-promql-name="store.getPromqlName"
      @edit="openEditModal"
      @copy="handleCopy"
      @run="store.runTask"
      @toggle="store.toggleTask"
      @delete="store.deleteTask"
    />

    <!-- 创建任务弹窗 -->
    <ModalDialog
      :visible="showCreateModal"
      title="创建新任务"
      max-width="900px"
      @close="showCreateModal = false"
    >
      <PushTaskForm
        :form="createForm"
        :sources="store.sources"
        :webhooks="store.webhooks"
        :chart-templates="store.chartTemplates"
        :promqls="store.promqls"
        :is-editing="false"
        :hide-actions="true"
      />
      <div class="modal-actions">
        <button class="btn btn-primary" @click="handleCreate">创建并发送</button>
        <button class="btn btn-secondary" @click="showCreateModal = false">取消</button>
      </div>
    </ModalDialog>

    <!-- 编辑任务弹窗 -->
    <ModalDialog
      :visible="showEditModal"
      title="编辑任务"
      max-width="900px"
      @close="cancelEdit"
    >
      <PushTaskForm
        :form="editForm"
        :sources="store.sources"
        :webhooks="store.webhooks"
        :chart-templates="store.chartTemplates"
        :promqls="store.promqls"
        :is-editing="true"
        :hide-actions="true"
      />
      <div class="modal-actions">
        <button class="btn btn-primary" @click="handleUpdate">保存修改</button>
        <button class="btn btn-secondary" @click="cancelEdit">取消</button>
      </div>
    </ModalDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { usePushTaskStore } from '../stores/pushTask'
import { usePushTaskForm } from '../composables/usePushTaskForm'
import { usePolling } from '../composables/usePolling'
import { useAuthStore } from '../stores/auth'
import ModalDialog from '../components/ModalDialog.vue'
import PushTaskForm from '../components/push-task/PushTaskForm.vue'
import PushTaskList from '../components/push-task/PushTaskList.vue'
import { IconPlus } from '../components/icons'
import type { PushTask } from '../types'

const { isAdmin } = useAuthStore()

const store = usePushTaskStore()
const tasks = computed(() => store.tasks)

const createForm = usePushTaskForm()
const editForm = usePushTaskForm()
const showCreateModal = ref(false)
const showEditModal = ref(false)
const editingTaskId = ref<number | null>(null)

usePolling(() => store.fetchAllData(), 30000)

function openCreateModal() {
  createForm.resetForm()
  showCreateModal.value = true
}

async function handleCreate() {
  if (!createForm.validate()) return
  const payload = createForm.buildPayload(store.promqls, store.chartTemplates)
  const success = await store.createTask(payload)
  if (success) {
    createForm.resetForm()
    showCreateModal.value = false
  }
}

function openEditModal(task: PushTask) {
  editingTaskId.value = task.id
  editForm.loadTask(task, store.chartTemplates)
  showEditModal.value = true
}

async function handleUpdate() {
  if (!editForm.validate()) return
  if (editingTaskId.value === null) return
  const payload = editForm.buildPayload(store.promqls, store.chartTemplates)
  const success = await store.updateTask(editingTaskId.value, {
    ...payload,
    id: editingTaskId.value
  })
  if (success) {
    cancelEdit()
  }
}

function cancelEdit() {
  showEditModal.value = false
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
