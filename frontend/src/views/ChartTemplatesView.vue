<template>
  <div class="tab-content">
    <h3>图表模板</h3>
    <div>
      <label>模板名称: <input v-model="newName" /></label>
      <label>图表类型:
        <select v-model="newChartType">
          <option value="area">区域图</option>
          <option value="line">折线图</option>
          <option value="bar">柱状图</option>
        </select>
      </label>
      <button @click="handleAdd">添加模板</button>
    </div>

    <table>
      <thead>
        <tr>
          <th>ID</th><th>名称</th><th>图表类型</th><th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="tmpl in items" :key="tmpl.id">
          <td>{{ tmpl.id }}</td>
          <td v-if="editingId === tmpl.id">
            <input v-model="editName" />
          </td>
          <td v-else>{{ tmpl.name }}</td>

          <td v-if="editingId === tmpl.id">
            <select v-model="editChartType">
              <option value="area">区域图</option>
              <option value="line">折线图</option>
              <option value="bar">柱状图</option>
            </select>
          </td>
          <td v-else>{{ tmpl.chart_type }}</td>

          <td>
            <div v-if="editingId === tmpl.id">
              <button @click="handleSave(tmpl.id)">保存</button>
              <button @click="cancelEdit">取消</button>
            </div>
            <div v-else>
              <button @click="handleStartEdit(tmpl)">编辑</button>
              <button @click="deleteItem(tmpl.id)">删除</button>
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
import type { ChartTemplate } from '../types'

const { items, editingId, fetchList, addItem, updateItem, deleteItem, startEdit, cancelEdit, validateRequired } =
  useCrudList<ChartTemplate>('/api/chart_template', '模板')

const newName = ref('')
const newChartType = ref('area')
const editName = ref('')
const editChartType = ref('')

async function handleAdd() {
  if (!validateRequired({ [newName.value]: '模板名称' })) return
  const success = await addItem({ name: newName.value, chart_type: newChartType.value } as Partial<ChartTemplate>)
  if (success) {
    newName.value = ''
    newChartType.value = 'area'
  }
}

function handleStartEdit(tmpl: ChartTemplate) {
  startEdit(tmpl.id)
  editName.value = tmpl.name
  editChartType.value = tmpl.chart_type
}

async function handleSave(id: number) {
  if (!validateRequired({ [editName.value]: '模板名称' })) return
  await updateItem(id, { name: editName.value, chart_type: editChartType.value } as Partial<ChartTemplate>)
}

// 初始加载
import { onMounted } from 'vue'
onMounted(fetchList)
</script>

<style scoped>
.tab-content { padding: 20px; }
table { width: 100%; border-collapse: collapse; margin-top: 20px; }
th, td { padding: 8px; text-align: left; border-bottom: 1px solid var(--color-border, #ddd); }
th { background-color: var(--color-bg-light, #f5f5f5); }
button { margin: 0 5px; padding: 5px 10px; border: 1px solid var(--color-border, #ddd); border-radius: 4px; background-color: #fff; cursor: pointer; }
button:hover { background-color: var(--color-bg-light, #f5f5f5); }
input, select { padding: 5px; border: 1px solid var(--color-border, #ddd); border-radius: 4px; margin-right: 10px; }
label { margin-right: 15px; }
</style>
