<template>
  <div>
    <div class="page-header">
      <div>
        <h3>数据源</h3>
        <p>管理 Prometheus 数据源连接</p>
      </div>
      <button class="btn btn-primary" @click="openAddModal">
        <IconPlus :size="16" />
        添加数据源
      </button>
    </div>

    <div class="card">
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th><th>名称</th><th>URL</th><th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="src in items" :key="src.id">
            <td>{{ src.id }}</td>
            <td>{{ src.name }}</td>
            <td>{{ src.url }}</td>
            <td>
              <div class="action-group">
                <button class="btn-icon" @click="openEditModal(src)" title="编辑">
                  <IconEdit :size="16" />
                </button>
                <button class="btn-icon btn-icon-danger" @click="deleteItem(src.id)" title="删除">
                  <IconTrash :size="16" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="items.length === 0" class="empty">暂无数据源</div>
    </div>

    <ModalDialog
      :visible="showModal"
      :title="isEditing ? '编辑数据源' : '添加数据源'"
      max-width="600px"
      @close="closeModal"
    >
      <div class="form-group">
        <label>名称</label>
        <input class="form-input" v-model="formName" placeholder="数据源名称" />
      </div>
      <div class="form-group">
        <label>URL</label>
        <input class="form-input" v-model="formURL" placeholder="Prometheus URL" />
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
import ModalDialog from '../components/ModalDialog.vue'
import { IconPlus, IconEdit, IconTrash } from '../components/icons'
import type { MetricsSource } from '../types'

const { items, fetchList, addItem, updateItem, deleteItem, validateRequired } =
  useCrudList<MetricsSource>('/api/metrics_source', '数据源')

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

function openEditModal(src: MetricsSource) {
  formName.value = src.name
  formURL.value = src.url
  isEditing.value = true
  editingId.value = src.id
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
    await updateItem(editingId.value, { name: formName.value, url: formURL.value } as Partial<MetricsSource>)
  } else {
    await addItem({ name: formName.value, url: formURL.value } as Partial<MetricsSource>)
  }
  closeModal()
}

usePolling(fetchList, 30000)
</script>
