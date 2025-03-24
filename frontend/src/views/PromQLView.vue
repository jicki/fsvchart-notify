<template>
  <div>
    <h3>PromQL 查询管理</h3>
    <p>在这里管理预定义的 PromQL 查询，可以在创建任务时直接选择使用。</p>

    <!-- 添加 PromQL 表单 -->
    <div class="form-container" v-if="showAddForm">
      <h4>{{ isEditing ? '编辑' : '添加' }} PromQL 查询</h4>
      <div class="form-group">
        <label>名称:</label>
        <input type="text" v-model="newPromQL.name" placeholder="查询名称" />
      </div>
      <div class="form-group">
        <label>分类:</label>
        <input type="text" v-model="newPromQL.category" placeholder="查询分类" />
      </div>
      <div class="form-group">
        <label>描述:</label>
        <textarea v-model="newPromQL.description" placeholder="查询描述"></textarea>
      </div>
      <div class="form-group">
        <label>PromQL 查询:</label>
        <div class="query-input-container">
          <textarea 
            v-model="newPromQL.query" 
            placeholder="PromQL 查询语句" 
            class="query-textarea"
            @input="autoAdjustHeight"
            ref="queryTextarea"
          ></textarea>
        </div>
      </div>
      <div class="form-actions">
        <button @click="savePromQL">保存</button>
        <button @click="cancelAdd">取消</button>
      </div>
    </div>

    <!-- 添加按钮 -->
    <div class="action-buttons" v-if="!showAddForm">
      <button @click="showAddForm = true">添加 PromQL 查询</button>
    </div>

    <!-- PromQL 列表 -->
    <div class="promql-list">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>名称</th>
            <th>分类</th>
            <th>描述</th>
            <th>查询语句</th>
            <th>创建时间</th>
            <th>更新时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="promql in promqls" :key="promql.id">
            <td>{{ promql.id }}</td>
            <td>{{ promql.name }}</td>
            <td>{{ promql.category }}</td>
            <td>{{ promql.description }}</td>
            <td>
              <div class="query-cell">
                <pre class="query-content">{{ promql.query }}</pre>
              </div>
            </td>
            <td>{{ formatDate(promql.created_at) }}</td>
            <td>{{ formatDate(promql.updated_at) }}</td>
            <td>
              <button @click="editPromQL(promql)">编辑</button>
              <button @click="copyPromQL(promql)">复制</button>
              <button @click="deletePromQL(promql.id)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { get, post, put, del } from '../utils/api'

const emit = defineEmits(['promql-updated'])

// 定义 PromQL 类型
interface PromQL {
  id: number
  name: string
  description: string
  query: string
  category: string
  created_at: string
  updated_at: string
}

// 状态变量
const promqls = ref<PromQL[]>([])
const showAddForm = ref(false)
const isEditing = ref(false)
const newPromQL = ref({
  id: 0,
  name: '',
  description: '',
  query: '',
  category: ''
})

// 添加自动调整高度的方法
const queryTextarea = ref<HTMLTextAreaElement | null>(null)

const autoAdjustHeight = () => {
  if (queryTextarea.value) {
    queryTextarea.value.style.height = 'auto'
    queryTextarea.value.style.height = queryTextarea.value.scrollHeight + 'px'
  }
}

// 获取所有 PromQL 查询
const fetchPromQLs = async () => {
  try {
    const data = await get('/api/promqls')
    if (Array.isArray(data)) {
      promqls.value = data
      console.log('成功获取PromQL列表:', data.length)
    } else {
      console.error('获取PromQL返回格式错误:', data)
    }
  } catch (error) {
    console.error('获取PromQL查询失败:', error)
    promqls.value = [] // 确保失败时设置为空数组
    alert('获取PromQL查询失败，请刷新页面重试')
  }
}

// 保存 PromQL 查询
const savePromQL = async () => {
  try {
    if (!newPromQL.value.name || !newPromQL.value.query) {
      alert('名称和查询语句不能为空')
      return
    }

    const payload = {
      name: newPromQL.value.name,
      description: newPromQL.value.description,
      query: newPromQL.value.query,
      category: newPromQL.value.category
    }

    if (isEditing.value) {
      // 更新现有 PromQL
      const result = await put(`/api/promql/${newPromQL.value.id}`, payload)
      console.log('更新PromQL成功:', result)
    } else {
      // 创建新 PromQL
      const result = await post('/api/promql', payload)
      console.log('创建PromQL成功:', result)
    }

    // 重置表单并刷新列表
    resetForm()
    await fetchPromQLs()
    emit('promql-updated')
  } catch (error) {
    console.error('保存PromQL查询失败:', error)
    alert('保存PromQL查询失败，请重试')
  }
}

// 编辑 PromQL 查询
const editPromQL = (promql: PromQL) => {
  newPromQL.value = {
    id: promql.id,
    name: promql.name,
    description: promql.description,
    query: promql.query,
    category: promql.category
  }
  isEditing.value = true
  showAddForm.value = true
  
  // 等待 DOM 更新后调整高度
  nextTick(() => {
    autoAdjustHeight()
  })
}

// 删除 PromQL 查询
const deletePromQL = async (id: number) => {
  if (!confirm('确定要删除这个 PromQL 查询吗？如果有任务正在使用它，将无法删除。')) {
    return
  }

  try {
    await del(`/api/promql/${id}`)
    console.log('删除PromQL成功:', id)
    await fetchPromQLs()
    emit('promql-updated')
  } catch (error: any) {
    console.error('删除PromQL查询失败:', error)
    alert('删除PromQL查询失败，可能正在被任务使用')
  }
}

// 复制 PromQL 查询
const copyPromQL = async (promql) => {
  try {
    const newPromQL = {
      name: `${promql.name} (复制)`,
      description: promql.description,
      query: promql.query,
      category: promql.category
    }

    if (!newPromQL.name || !newPromQL.query) {
      alert('名称和查询语句不能为空')
      return
    }

    const result = await post('/api/promql', newPromQL)
    console.log('复制PromQL成功:', result)
    await fetchPromQLs()
    alert('PromQL复制成功')
  } catch (error) {
    console.error('复制PromQL查询失败:', error)
    alert('复制PromQL查询失败，请重试')
  }
}

// 取消添加/编辑
const cancelAdd = () => {
  resetForm()
}

// 重置表单
const resetForm = () => {
  newPromQL.value = {
    id: 0,
    name: '',
    description: '',
    query: '',
    category: ''
  }
  isEditing.value = false
  showAddForm.value = false
}

// 格式化日期
const formatDate = (dateStr: string) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleString()
}

// 定期刷新数据
let refreshInterval: number

// 组件挂载时获取数据
onMounted(() => {
  // 初始加载
  fetchPromQLs()
  
  // 设置定期刷新
  refreshInterval = setInterval(() => {
    fetchPromQLs()
  }, 30000) // 每30秒刷新一次
})

onUnmounted(() => {
  // 清理定时器
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})
</script>

<style scoped>
.form-container {
  margin-bottom: 20px;
  padding: 15px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background-color: #f9f9f9;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
}

.form-group input,
.form-group textarea,
.form-group select {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.form-actions {
  display: flex;
  gap: 10px;
}

.action-buttons {
  margin-bottom: 20px;
}

.promql-list {
  margin-top: 20px;
}

.query-input-container {
  position: relative;
  width: 100%;
}

.query-textarea {
  width: 100%;
  min-height: 80px;
  padding: 8px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 14px;
  line-height: 1.5;
  border: 1px solid #ddd;
  border-radius: 4px;
  resize: none;
  overflow-y: hidden;
  background-color: #f8f9fa;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.query-textarea:focus {
  border-color: #007bff;
  box-shadow: 0 0 0 2px rgba(0,123,255,0.25);
  outline: none;
}

.query-cell {
  position: relative;
  max-width: none;
  overflow: visible;
}

.query-content {
  margin: 0;
  padding: 8px;
  background-color: #f8f9fa;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.4;
  white-space: pre-wrap;
  word-break: break-all;
  overflow-x: auto;
}

/* 添加悬停效果 */
.query-cell:hover .query-content {
  background-color: #e9ecef;
}

/* 适配移动设备 */
@media (max-width: 768px) {
  .query-textarea {
    font-size: 13px;
  }
  
  .query-content {
    font-size: 12px;
    padding: 6px;
  }
}
</style>