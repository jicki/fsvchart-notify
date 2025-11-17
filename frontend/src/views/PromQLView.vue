<template>
  <div>
    <h3>PromQL 查询管理</h3>
    <p>在这里管理预定义的 PromQL 查询，可以在创建任务时直接选择使用。</p>

    <!-- 添加全局展开/收起按钮 -->
    <div class="global-actions">
      <button class="action-btn" @click="toggleAllQueries">
        {{ isAllExpanded ? '收起所有查询' : '展开所有查询' }}
      </button>
    </div>

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
        <textarea v-model="newPromQL.query" placeholder="PromQL 查询语句" rows="5" @input="updateHighlightedPreview"></textarea>
        <!-- 添加语法高亮预览 -->
        <div v-if="newPromQL.query" class="promql-preview">
          <h5>语法高亮预览:</h5>
          <pre class="promql-code" v-html="highlightedPreview"></pre>
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
              <div class="query-cell" :class="{ 'expanded': expandedQueries.includes(promql.id) }">
                <pre class="promql-code" v-html="highlightPromQL(promql.query)"></pre>
                <button class="expand-btn" @click="toggleQueryExpand(promql.id)">
                  {{ expandedQueries.includes(promql.id) ? '收起' : '展开' }}
                </button>
              </div>
            </td>
            <td>{{ formatDate(promql.created_at) }}</td>
            <td>{{ formatDate(promql.updated_at) }}</td>
            <td class="action-column">
              <div class="action-buttons">
                <button class="action-btn edit" @click="editPromQL(promql)" title="编辑">
                  编辑
                </button>
                <button class="action-btn copy" @click="copyPromQL(promql)" title="复制">
                  复制
                </button>
                <button class="action-btn delete" @click="deletePromQL(promql.id)" title="删除">
                  删除
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
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

// 添加高亮预览状态
const highlightedPreview = ref('')

// 添加展开状态管理
const expandedQueries = ref<number[]>([])
const isAllExpanded = ref(false)

// PromQL 语法高亮函数
function highlightPromQL(query: string): string {
  if (!query) return ''
  
  // 基本的 PromQL 关键字和函数
  const keywords = [
    'sum', 'rate', 'irate', 'avg', 'max', 'min', 'count',
    'by', 'without', 'offset', 'bool', 'and', 'or', 'unless',
    'group', 'ignoring', 'on', 'topk', 'bottomk'
  ]
  
  // 转义 HTML 特殊字符
  let highlighted = query.replace(/[&<>]/g, char => {
    const entities: { [key: string]: string } = {
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;'
    }
    return entities[char] || char
  })
  
  // 高亮关键字
  keywords.forEach(keyword => {
    const regex = new RegExp(`\\b${keyword}\\b`, 'g')
    highlighted = highlighted.replace(regex, `<span class="keyword">${keyword}</span>`)
  })
  
  // 高亮标签和值
  highlighted = highlighted.replace(
    /(\{[^}]*\})/g,
    (match) => `<span class="label">${match}</span>`
  )
  
  // 高亮数字
  highlighted = highlighted.replace(
    /\b(\d+(\.\d+)?)\b/g,
    '<span class="number">$1</span>'
  )
  
  // 高亮时间单位
  highlighted = highlighted.replace(
    /\b(\d+)(s|m|h|d|w|y)\b/g,
    '<span class="number">$1</span><span class="unit">$2</span>'
  )
  
  return highlighted
}

// 更新高亮预览
const updateHighlightedPreview = () => {
  highlightedPreview.value = highlightPromQL(newPromQL.value.query)
}

// 展开/收起所有查询
function toggleAllQueries() {
  if (isAllExpanded.value) {
    expandedQueries.value = []
  } else {
    expandedQueries.value = promqls.value.map(p => p.id)
  }
  isAllExpanded.value = !isAllExpanded.value
}

// 展开/收起单个查询
function toggleQueryExpand(id: number) {
  const index = expandedQueries.value.indexOf(id)
  if (index === -1) {
    expandedQueries.value.push(id)
  } else {
    expandedQueries.value.splice(index, 1)
  }
  // 更新全局展开状态
  isAllExpanded.value = expandedQueries.value.length === promqls.value.length
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
  // 立即更新高亮预览
  updateHighlightedPreview()
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
    // 显示后端返回的具体错误信息
    const errorMessage = error.response?.data?.error || error.message || '删除PromQL查询失败'
    alert(`删除失败: ${errorMessage}\n\n如果此 PromQL 正在被任务使用，请先删除或修改相关任务。`)
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

.global-actions {
  margin: 20px 0;
  display: flex;
  justify-content: flex-end;
}

.action-column {
  width: 200px;
  white-space: nowrap;
  padding: 8px 4px;
}

.action-buttons {
  display: flex;
  gap: 8px;
  justify-content: flex-start;
  flex-wrap: nowrap;
}

.action-btn {
  padding: 4px 10px;
  border: 1px solid #ddd;
  border-radius: 3px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.3s ease;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 50px;
  background-color: #fff;
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.action-btn:hover {
  background-color: #f8f9fa;
  border-color: #adb5bd;
}

.action-btn.edit {
  color: #fff;
  background-color: #2196F3;
  border-color: #2196F3;
}

.action-btn.edit:hover {
  background-color: #1976D2;
  border-color: #1976D2;
}

.action-btn.copy {
  color: #fff;
  background-color: #4CAF50;
  border-color: #4CAF50;
}

.action-btn.copy:hover {
  background-color: #388E3C;
  border-color: #388E3C;
}

.action-btn.delete {
  color: #fff;
  background-color: #F44336;
  border-color: #F44336;
}

.action-btn.delete:hover {
  background-color: #D32F2F;
  border-color: #D32F2F;
}

.query-cell {
  position: relative;
  max-width: 500px;
  background: #f8f9fa;
  border-radius: 4px;
  padding: 12px;
  transition: all 0.3s ease;
  margin: 4px 0;
  overflow: hidden;
}

.query-cell.expanded {
  max-width: none;
  width: auto;
  min-width: 500px;
}

.promql-code {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.5;
  overflow-x: auto;
  padding: 4px;
  max-height: none;
}

.expand-btn {
  position: absolute;
  right: 8px;
  top: 8px;
  background: #e9ecef;
  border: 1px solid #ced4da;
  border-radius: 3px;
  padding: 4px 8px;
  font-size: 12px;
  cursor: pointer;
  opacity: 0.8;
  transition: all 0.2s ease;
  color: #495057;
  z-index: 1;
}

.expand-btn:hover {
  opacity: 1;
  background: #dee2e6;
  border-color: #adb5bd;
}

/* PromQL 语法高亮样式 */
:deep(.keyword) {
  color: #0066cc;
  font-weight: 500;
}

:deep(.label) {
  color: #e83e8c;
}

:deep(.number) {
  color: #2e7d32;
}

:deep(.unit) {
  color: #0066cc;
  font-weight: 500;
}

/* 添加 PromQL 预览区域样式 */
.promql-preview {
  margin-top: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 10px;
  background-color: #f8f9fa;
}

.promql-preview h5 {
  margin-top: 0;
  margin-bottom: 10px;
  color: #495057;
  font-size: 14px;
}

/* 确保表格单元格内容不会被截断 */
td {
  vertical-align: middle;
  padding: 8px 4px;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 优化表格布局 */
table {
  table-layout: fixed;
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  margin: 0;
  padding: 0;
}

thead th {
  padding: 8px 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 调整列宽度 */
th:nth-child(1) { width: 5%; }  /* ID */
th:nth-child(2) { width: 10%; } /* 名称 */
th:nth-child(3) { width: 10%; } /* 分类 */
th:nth-child(4) { width: 15%; } /* 描述 */
th:nth-child(5) { width: 30%; } /* 查询语句 */
th:nth-child(6) { width: 10%; } /* 创建时间 */
th:nth-child(7) { width: 10%; } /* 更新时间 */
th:nth-child(8) { width: 10%; }  /* 操作 */

thead th {
  background: #f8f9fa;
  position: sticky;
  top: 0;
  z-index: 1;
  padding: 12px 8px;
  border-bottom: 2px solid #dee2e6;
}

tbody tr:hover {
  background-color: #f8f9fa;
}

tbody td {
  border-bottom: 1px solid #dee2e6;
}

/* 添加响应式布局支持 */
@media (max-width: 1200px) {
  .query-cell {
    max-width: 300px;
  }
}

@media (max-width: 768px) {
  .query-cell {
    max-width: 200px;
  }
}
</style>