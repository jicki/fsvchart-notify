<template>
  <div>
    <div class="page-header">
      <div>
        <h3>飞书 WebHook</h3>
        <p>管理飞书机器人 WebHook 地址</p>
      </div>
      <button v-if="isAdmin" class="btn btn-primary" @click="openAddModal">
        <IconPlus :size="16" />
        添加 WebHook
      </button>
    </div>

    <div class="card">
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th><th>名称</th><th>URL</th><th v-if="isAdmin">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="wb in items" :key="wb.id">
            <td>{{ wb.id }}</td>
            <td>{{ wb.name }}</td>
            <td>{{ wb.url }}</td>
            <td v-if="isAdmin">
              <div class="action-group">
                <button class="btn-icon" @click="openEditModal(wb)" title="编辑">
                  <IconEdit :size="16" />
                </button>
                <button class="btn-icon btn-icon-danger" @click="deleteItem(wb.id)" title="删除">
                  <IconTrash :size="16" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="items.length === 0" class="empty">暂无 WebHook</div>
    </div>

    <ModalDialog
      :visible="showModal"
      :title="isEditing ? '编辑 WebHook' : '添加 WebHook'"
      max-width="600px"
      @close="closeModal"
    >
      <div class="form-group">
        <label>名称</label>
        <input class="form-input" v-model="formName" placeholder="WebHook 名称" />
      </div>
      <div class="form-group">
        <label>URL</label>
        <input class="form-input" v-model="formURL" placeholder="WebHook URL" />
      </div>
      <div class="modal-actions">
        <button class="btn btn-primary" @click="handleSave">保存</button>
        <button class="btn btn-secondary" @click="closeModal">取消</button>
      </div>
    </ModalDialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useCrudList } from '../composables/useCrudList'
import { usePolling } from '../composables/usePolling'
import { useAuthStore } from '../stores/auth'
import ModalDialog from '../components/ModalDialog.vue'
import { IconPlus, IconEdit, IconTrash } from '../components/icons'
import type { FeishuWebhook } from '../types'

const { isAdmin } = useAuthStore()

const { items, fetchList, addItem, updateItem, deleteItem, validateRequired } =
  useCrudList<FeishuWebhook>('/api/feishu_webhook', 'WebHook')

const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)
const formName = ref('')
const formURL = ref('')

function openAddModal() {
  formName.value = ''
  formURL.value = ''
  isEditing.value = false
  editingId.value = null
  showModal.value = true
}

function openEditModal(wb: FeishuWebhook) {
  formName.value = wb.name
  formURL.value = wb.url
  isEditing.value = true
  editingId.value = wb.id
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  isEditing.value = false
  editingId.value = null
}

async function handleSave() {
  if (!validateRequired({ [formName.value]: '名称', [formURL.value]: 'URL' })) return
  if (isEditing.value && editingId.value !== null) {
    await updateItem(editingId.value, { name: formName.value, url: formURL.value } as Partial<FeishuWebhook>)
  } else {
    await addItem({ name: formName.value, url: formURL.value } as Partial<FeishuWebhook>)
  }
  closeModal()
}

usePolling(fetchList, 30000)
</script>
