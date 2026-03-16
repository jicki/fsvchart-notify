<template>
  <div>
    <div class="page-header">
      <div>
        <h3>数据源</h3>
        <p>管理 Prometheus 数据源连接</p>
      </div>
      <button class="btn btn-primary" @click="showAddForm = !showAddForm">
        <IconPlus :size="16" />
        添加数据源
      </button>
    </div>

    <div v-if="showAddForm" class="card" style="margin-bottom: var(--spacing-lg)">
      <h4 style="margin-top: 0; margin-bottom: var(--spacing-md)">添加数据源</h4>
      <div class="form-row">
        <div class="form-group" style="flex: 1">
          <label>名称</label>
          <input class="form-input" v-model="newName" placeholder="数据源名称" />
        </div>
        <div class="form-group" style="flex: 2">
          <label>URL</label>
          <input class="form-input" v-model="newURL" placeholder="Prometheus URL" />
        </div>
        <div class="form-actions-inline">
          <button class="btn btn-primary" @click="handleAdd">保存</button>
          <button class="btn btn-secondary" @click="showAddForm = false">取消</button>
        </div>
      </div>
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
            <td v-if="editingId === src.id">
              <input class="form-input" v-model="editName" />
            </td>
            <td v-else>{{ src.name }}</td>

            <td v-if="editingId === src.id">
              <input class="form-input" v-model="editURL" />
            </td>
            <td v-else>{{ src.url }}</td>

            <td>
              <div class="action-group" v-if="editingId === src.id">
                <button class="btn btn-primary btn-sm" @click="handleSave(src.id)">保存</button>
                <button class="btn btn-secondary btn-sm" @click="cancelEdit">取消</button>
              </div>
              <div class="action-group" v-else>
                <button class="btn-icon" @click="handleStartEdit(src)" title="编辑">
                  <IconEdit :size="16" />
                </button>
                <button class="btn-icon" @click="deleteItem(src.id)" title="删除" style="color: var(--color-danger)">
                  <IconTrash :size="16" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="items.length === 0" class="empty">暂无数据源</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useCrudList } from '../composables/useCrudList'
import { usePolling } from '../composables/usePolling'
import { IconPlus, IconEdit, IconTrash } from '../components/icons'
import type { MetricsSource } from '../types'

const { items, editingId, fetchList, addItem, updateItem, deleteItem, startEdit, cancelEdit, validateRequired } =
  useCrudList<MetricsSource>('/api/metrics_source', '数据源')

const showAddForm = ref(false)
const newName = ref('')
const newURL = ref('')
const editName = ref('')
const editURL = ref('')

async function handleAdd() {
  if (!validateRequired({ [newName.value]: '名称', [newURL.value]: 'URL' })) return
  const success = await addItem({ name: newName.value, url: newURL.value } as Partial<MetricsSource>)
  if (success) {
    newName.value = ''
    newURL.value = ''
    showAddForm.value = false
  }
}

function handleStartEdit(src: MetricsSource) {
  startEdit(src.id)
  editName.value = src.name
  editURL.value = src.url
}

async function handleSave(id: number) {
  if (!validateRequired({ [editName.value]: '名称', [editURL.value]: 'URL' })) return
  await updateItem(id, { name: editName.value, url: editURL.value } as Partial<MetricsSource>)
}

usePolling(fetchList, 30000)
</script>

<style scoped>
.form-row {
  display: flex;
  gap: var(--spacing-md);
  align-items: flex-end;
}

.form-input {
  width: 100%;
}

.form-actions-inline {
  display: flex;
  gap: 8px;
  padding-bottom: var(--spacing-md);
}

.action-group {
  display: flex;
  gap: 4px;
  align-items: center;
}
</style>
