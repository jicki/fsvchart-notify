<template>
  <div class="tab-content">
    <h3>数据源</h3>
    <div>
      <label>名称: <input v-model="newSourceName"/></label>
      <label>URL: <input v-model="newSourceURL"/></label>
      <button @click="addMetricsSource">添加数据源</button>
    </div>

    <table>
      <thead>
        <tr>
          <th>ID</th><th>名称</th><th>URL</th><th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="src in metricsSources" :key="src.id">
          <td>{{ src.id }}</td>
          <td v-if="editSourceID===src.id">
            <input v-model="editSourceName"/>
          </td>
          <td v-else>
            {{ src.name }}
          </td>

          <td v-if="editSourceID===src.id">
            <input v-model="editSourceURL"/>
          </td>
          <td v-else>
            {{ src.url }}
          </td>

          <td>
            <div v-if="editSourceID===src.id">
              <button @click="saveEditSource(src.id)">保存</button>
              <button @click="cancelEditSource">取消</button>
            </div>
            <div v-else>
              <button @click="startEditSource(src)">编辑</button>
              <button @click="deleteMetricsSource(src.id)">删除</button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { get, post, put, del } from '../utils/api'

const emit = defineEmits(['source-updated'])

const metricsSources = ref<any[]>([])
const newSourceName = ref('')
const newSourceURL = ref('')
const editSourceID = ref<number|null>(null)
const editSourceName = ref('')
const editSourceURL = ref('')

async function fetchMetricsSources() {
  try {
    const data = await get('/api/metrics_source')
    if (Array.isArray(data)) {
      metricsSources.value = data
      console.log('成功获取数据源列表:', data.length)
    } else {
      console.error('获取数据源返回格式错误:', data)
    }
  } catch (err) {
    console.error('获取数据源失败:', err)
    metricsSources.value = [] // 确保失败时设置为空数组
  }
}

async function addMetricsSource() {
  if(!newSourceName.value || !newSourceURL.value){
    alert('名称或URL不能为空')
    return
  }
  
  try {
    const body = { name: newSourceName.value, url: newSourceURL.value }
    const result = await post('/api/metrics_source', body)
    console.log('添加数据源成功:', result)
    
    // 重置表单
    newSourceName.value = ''
    newSourceURL.value = ''
    
    // 刷新列表并通知更新
    await fetchMetricsSources()
    emit('source-updated')
  } catch (err) {
    console.error('添加数据源失败:', err)
    alert('添加数据源失败，请重试')
  }
}

function startEditSource(src:any){
  editSourceID.value = src.id
  editSourceName.value = src.name
  editSourceURL.value = src.url
}

function cancelEditSource(){
  editSourceID.value = null
  editSourceName.value = ''
  editSourceURL.value = ''
}

async function saveEditSource(id:number){
  if(!editSourceName.value || !editSourceURL.value){
    alert('名称或URL不能为空')
    return
  }

  try {
    const body = { name: editSourceName.value, url: editSourceURL.value }
    const result = await put(`/api/metrics_source/${id}`, body)
    console.log('更新数据源成功:', result)
    
    // 重置编辑状态
    editSourceID.value = null
    editSourceName.value = ''
    editSourceURL.value = ''
    
    // 刷新列表并通知更新
    await fetchMetricsSources()
    emit('source-updated')
  } catch (err) {
    console.error('更新数据源失败:', err)
    alert('更新数据源失败，请重试')
  }
}

async function deleteMetricsSource(id:number){
  if(!confirm(`确认删除数据源ID=${id}？`)) return
  
  try {
    await del(`/api/metrics_source/${id}`)
    console.log('删除数据源成功:', id)
    
    // 刷新列表并通知更新
    await fetchMetricsSources()
    emit('source-updated')
  } catch (err) {
    console.error('删除数据源失败:', err)
    alert('删除数据源失败，请重试')
  }
}

// 定期刷新数据
let refreshInterval: number

onMounted(() => {
  // 初始加载
  fetchMetricsSources()
  
  // 设置定期刷新
  refreshInterval = setInterval(() => {
    fetchMetricsSources()
  }, 30000) // 每30秒刷新一次
})

onUnmounted(() => {
  // 清理定时器
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})
</script>

<!-- 如果你需要子组件私有样式,可在此 <style scoped>, 不会影响父组件公共样式. -->
<style scoped>
</style>
