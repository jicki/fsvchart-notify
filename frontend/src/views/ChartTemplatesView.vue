<template>
  <div>
    <div class="page-header">
      <div>
        <h3>图表模板</h3>
        <p>管理图表渲染模板</p>
      </div>
      <button class="btn btn-primary" @click="showAddForm = !showAddForm">
        <IconPlus :size="16" />
        添加模板
      </button>
    </div>

    <div v-if="showAddForm" class="card" style="margin-bottom: var(--spacing-lg)">
      <h4 style="margin-top: 0; margin-bottom: var(--spacing-md)">添加模板</h4>
      <div class="form-row">
        <div class="form-group" style="flex: 1">
          <label>模板名称</label>
          <input class="form-input" v-model="newName" placeholder="模板名称" />
        </div>
        <div class="form-group" style="flex: 1">
          <label>图表类型</label>
          <select class="form-input" v-model="newChartType">
            <option value="area">区域图</option>
            <option value="line">折线图</option>
            <option value="bar">柱状图</option>
          </select>
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
            <th>ID</th><th>名称</th><th>图表类型</th><th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="tmpl in items" :key="tmpl.id">
            <td>{{ tmpl.id }}</td>
            <td v-if="editingId === tmpl.id">
              <input class="form-input" v-model="editName" />
            </td>
            <td v-else>{{ tmpl.name }}</td>

            <td v-if="editingId === tmpl.id">
              <select class="form-input" v-model="editChartType">
                <option value="area">区域图</option>
                <option value="line">折线图</option>
                <option value="bar">柱状图</option>
              </select>
            </td>
            <td v-else>{{ tmpl.chart_type }}</td>

            <td>
              <div class="action-group" v-if="editingId === tmpl.id">
                <button class="btn btn-primary btn-sm" @click="handleSave(tmpl.id)">保存</button>
                <button class="btn btn-secondary btn-sm" @click="cancelEdit">取消</button>
              </div>
              <div class="action-group" v-else>
                <button class="btn-icon" @click="handleStartEdit(tmpl)" title="编辑">
                  <IconEdit :size="16" />
                </button>
                <button class="btn-icon" @click="deleteItem(tmpl.id)" title="删除" style="color: var(--color-danger)">
                  <IconTrash :size="16" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="items.length === 0" class="empty">暂无模板</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useCrudList } from '../composables/useCrudList'
import { IconPlus, IconEdit, IconTrash } from '../components/icons'
import type { ChartTemplate } from '../types'

const { items, editingId, fetchList, addItem, updateItem, deleteItem, startEdit, cancelEdit, validateRequired } =
  useCrudList<ChartTemplate>('/api/chart_template', '模板')

const showAddForm = ref(false)
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
    showAddForm.value = false
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

onMounted(fetchList)
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
