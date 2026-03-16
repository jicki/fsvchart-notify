<template>
  <div class="tab-content">
    <h3>飞书 WebHooks</h3>
    <div>
      <label>名称: <input v-model="newName" /></label>
      <label>URL: <input v-model="newURL" /></label>
      <button @click="handleAdd">添加 WebHook</button>
    </div>

    <table>
      <thead>
        <tr>
          <th>ID</th><th>名称</th><th>URL</th><th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="wb in items" :key="wb.id">
          <td>{{ wb.id }}</td>
          <td v-if="editingId === wb.id">
            <input v-model="editName" />
          </td>
          <td v-else>{{ wb.name }}</td>

          <td v-if="editingId === wb.id">
            <input v-model="editURL" />
          </td>
          <td v-else>{{ wb.url }}</td>

          <td>
            <div v-if="editingId === wb.id">
              <button @click="handleSave(wb.id)">保存</button>
              <button @click="cancelEdit">取消</button>
            </div>
            <div v-else>
              <button @click="handleStartEdit(wb)">编辑</button>
              <button @click="deleteItem(wb.id)">删除</button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useCrudList } from '../composables/useCrudList'
import { usePolling } from '../composables/usePolling'
import type { FeishuWebhook } from '../types'

const { items, editingId, fetchList, addItem, updateItem, deleteItem, startEdit, cancelEdit, validateRequired } =
  useCrudList<FeishuWebhook>('/api/feishu_webhook', 'WebHook')

const newName = ref('')
const newURL = ref('')
const editName = ref('')
const editURL = ref('')

async function handleAdd() {
  if (!validateRequired({ [newName.value]: '名称', [newURL.value]: 'URL' })) return
  const success = await addItem({ name: newName.value, url: newURL.value } as Partial<FeishuWebhook>)
  if (success) {
    newName.value = ''
    newURL.value = ''
  }
}

function handleStartEdit(wb: FeishuWebhook) {
  startEdit(wb.id)
  editName.value = wb.name
  editURL.value = wb.url
}

async function handleSave(id: number) {
  if (!validateRequired({ [editName.value]: '名称', [editURL.value]: 'URL' })) return
  await updateItem(id, { name: editName.value, url: editURL.value } as Partial<FeishuWebhook>)
}

usePolling(fetchList, 30000)
</script>

<style scoped>
.tab-content { padding: 20px; }
table { width: 100%; border-collapse: collapse; margin-top: 20px; }
th, td { padding: 8px; text-align: left; border-bottom: 1px solid var(--color-border, #ddd); }
th { background-color: var(--color-bg-light, #f5f5f5); }
button { margin: 0 5px; padding: 5px 10px; border: 1px solid var(--color-border, #ddd); border-radius: 4px; background-color: #fff; cursor: pointer; }
button:hover { background-color: var(--color-bg-light, #f5f5f5); }
input { padding: 5px; border: 1px solid var(--color-border, #ddd); border-radius: 4px; margin-right: 10px; }
label { margin-right: 15px; }
</style>
