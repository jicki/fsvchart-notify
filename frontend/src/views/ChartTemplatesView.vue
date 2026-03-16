<template>
  <div>
    <div class="page-header">
      <div>
        <h3>图表模板</h3>
        <p>管理图表渲染模板</p>
      </div>
      <button v-if="isAdmin" class="btn btn-primary" @click="openAddModal">
        <IconPlus :size="16" />
        添加模板
      </button>
    </div>

    <div class="card">
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th><th>名称</th><th>图表类型</th><th v-if="isAdmin">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="tmpl in items" :key="tmpl.id">
            <td>{{ tmpl.id }}</td>
            <td>{{ tmpl.name }}</td>
            <td>{{ tmpl.chart_type }}</td>
            <td v-if="isAdmin">
              <div class="action-group">
                <button class="btn-icon" @click="openEditModal(tmpl)" title="编辑">
                  <IconEdit :size="16" />
                </button>
                <button class="btn-icon btn-icon-danger" @click="deleteItem(tmpl.id)" title="删除">
                  <IconTrash :size="16" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="items.length === 0" class="empty">暂无模板</div>
    </div>

    <ModalDialog
      :visible="showModal"
      :title="isEditing ? '编辑模板' : '添加模板'"
      max-width="600px"
      @close="closeModal"
    >
      <div class="form-group">
        <label>模板名称</label>
        <input class="form-input" v-model="formName" placeholder="模板名称" />
      </div>
      <div class="form-group">
        <label>图表类型</label>
        <select class="form-input" v-model="formChartType">
          <option value="area">区域图</option>
          <option value="line">折线图</option>
          <option value="bar">柱状图</option>
        </select>
      </div>
      <div class="modal-actions">
        <button class="btn btn-primary" @click="handleSave">保存</button>
        <button class="btn btn-secondary" @click="closeModal">取消</button>
      </div>
    </ModalDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useCrudList } from '../composables/useCrudList'
import { useAuthStore } from '../stores/auth'
import ModalDialog from '../components/ModalDialog.vue'
import { IconPlus, IconEdit, IconTrash } from '../components/icons'
import type { ChartTemplate } from '../types'

const { isAdmin } = useAuthStore()

const { items, fetchList, addItem, updateItem, deleteItem, validateRequired } =
  useCrudList<ChartTemplate>('/api/chart_template', '模板')

const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)
const formName = ref('')
const formChartType = ref('area')

function openAddModal() {
  formName.value = ''
  formChartType.value = 'area'
  isEditing.value = false
  editingId.value = null
  showModal.value = true
}

function openEditModal(tmpl: ChartTemplate) {
  formName.value = tmpl.name
  formChartType.value = tmpl.chart_type
  isEditing.value = true
  editingId.value = tmpl.id
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  isEditing.value = false
  editingId.value = null
}

async function handleSave() {
  if (!validateRequired({ [formName.value]: '模板名称' })) return
  if (isEditing.value && editingId.value !== null) {
    await updateItem(editingId.value, { name: formName.value, chart_type: formChartType.value } as Partial<ChartTemplate>)
  } else {
    await addItem({ name: formName.value, chart_type: formChartType.value } as Partial<ChartTemplate>)
  }
  closeModal()
}

onMounted(fetchList)
</script>
