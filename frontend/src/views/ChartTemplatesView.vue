<template>
  <div class="tab-content">
    <h3>图表模板</h3>
    <div>
      <label>模板名称: <input v-model="newTemplateName" /></label>
      <label>图表类型:
        <select v-model="newChartType">
          <option value="area">区域图</option>
          <option value="line">折线图</option>
          <option value="bar">柱状图</option>
        </select>
      </label>
      <button @click="addChartTemplate">添加模板</button>
    </div>

    <table>
      <thead>
        <tr>
          <th>ID</th>
          <th>名称</th>
          <th>图表类型</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="tmpl in chartTemplates" :key="tmpl.id">
          <td>{{ tmpl.id }}</td>
          <td v-if="editTemplateId === tmpl.id">
            <input v-model="editTemplateName" />
          </td>
          <td v-else>{{ tmpl.name }}</td>
          
          <td v-if="editTemplateId === tmpl.id">
            <select v-model="editChartType">
              <option value="area">区域图</option>
              <option value="line">折线图</option>
              <option value="bar">柱状图</option>
            </select>
          </td>
          <td v-else>{{ tmpl.chart_type }}</td>
          
          <td>
            <div v-if="editTemplateId === tmpl.id">
              <button @click="saveEditTemplate(tmpl.id)">保存</button>
              <button @click="cancelEditTemplate">取消</button>
            </div>
            <div v-else>
              <button @click="startEditTemplate(tmpl)">编辑</button>
              <button @click="deleteTemplate(tmpl.id)">删除</button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { get, post, put, del } from '../utils/api'

const chartTemplates = ref<any[]>([])
const newTemplateName = ref('')
const newChartType = ref('area')

// 编辑状态
const editTemplateId = ref<number|null>(null)
const editTemplateName = ref('')
const editChartType = ref('')

async function fetchTemplates() {
  try {
    const data = await get('/api/chart_template')
    chartTemplates.value = data
  } catch (err) {
    console.error('fetchTemplates:', err)
  }
}

async function addChartTemplate() {
  if (!newTemplateName.value) {
    alert('模板名称不能为空')
    return
  }
  
  try {
    await post('/api/chart_template', {
      name: newTemplateName.value,
      chart_type: newChartType.value
    })
    
    newTemplateName.value = ''
    newChartType.value = 'area'
    fetchTemplates()
  } catch (err) {
    console.error('addChartTemplate:', err)
  }
}

function startEditTemplate(tmpl: any) {
  editTemplateId.value = tmpl.id
  editTemplateName.value = tmpl.name
  editChartType.value = tmpl.chart_type
}

function cancelEditTemplate() {
  editTemplateId.value = null
}

async function saveEditTemplate(id: number) {
  try {
    await put(`/api/chart_template/${id}`, {
      name: editTemplateName.value,
      chart_type: editChartType.value
    })
    
    editTemplateId.value = null
    fetchTemplates()
  } catch (err) {
    console.error('saveEditTemplate:', err)
  }
}

async function deleteTemplate(id: number) {
  if (!confirm(`确认删除模板ID=${id}?`)) return
  
  try {
    await del(`/api/chart_template/${id}`)
    fetchTemplates()
  } catch (err) {
    console.error('deleteTemplate:', err)
  }
}

onMounted(() => {
  fetchTemplates()
})
</script>

<style scoped>
</style>
