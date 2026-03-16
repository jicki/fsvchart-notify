<template>
  <div>
    <h3>PromQL 查询管理</h3>
    <p>在这里管理预定义的 PromQL 查询，可以在创建任务时直接选择使用。</p>

    <!-- 全局展开/收起按钮 -->
    <div class="global-actions">
      <button class="action-btn" @click="expandable.toggleAll(promqls.map(p => p.id))">
        {{ expandable.isAllExpandedFor(promqls.length) ? '收起所有查询' : '展开所有查询' }}
      </button>
    </div>

    <!-- 添加/编辑 PromQL 表单 -->
    <div class="form-container" v-if="showAddForm">
      <h4>{{ isEditing ? '编辑' : '添加' }} PromQL 查询</h4>
      <div class="form-group">
        <label>名称:</label>
        <input type="text" v-model="formData.name" placeholder="查询名称" />
      </div>
      <div class="form-group">
        <label>分类:</label>
        <input type="text" v-model="formData.category" placeholder="查询分类" />
      </div>
      <div class="form-group">
        <label>描述:</label>
        <textarea v-model="formData.description" placeholder="查询描述"></textarea>
      </div>
      <div class="form-group">
        <label>PromQL 查询:</label>
        <textarea v-model="formData.query" placeholder="PromQL 查询语句" rows="5"></textarea>
        <div v-if="formData.query" class="promql-preview">
          <h5>语法高亮预览:</h5>
          <pre class="promql-code" v-html="highlightPromQL(formData.query)"></pre>
        </div>
      </div>
      <div class="form-actions">
        <button @click="savePromQL">保存</button>
        <button @click="cancelEdit">取消</button>
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
              <div class="query-cell" :class="{ 'expanded': expandable.isExpanded(promql.id) }">
                <pre class="promql-code" v-html="highlightPromQL(promql.query)"></pre>
                <button class="expand-btn" @click="expandable.toggle(promql.id)">
                  {{ expandable.isExpanded(promql.id) ? '收起' : '展开' }}
                </button>
              </div>
            </td>
            <td>{{ formatDate(promql.created_at) }}</td>
            <td>{{ formatDate(promql.updated_at) }}</td>
            <td class="action-column">
              <div class="action-buttons">
                <button class="action-btn edit" @click="startEdit(promql)">编辑</button>
                <button class="action-btn copy" @click="copyPromQL(promql)">复制</button>
                <button class="action-btn delete" @click="deletePromQL(promql.id)">删除</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { get, post, put, del } from '../utils/api'
import { formatDate } from '../utils/formatters'
import { useNotification } from '../composables/useNotification'
import { usePolling } from '../composables/usePolling'
import { useExpandable } from '../composables/useExpandable'
import { usePromqlHighlight } from '../composables/usePromqlHighlight'
import type { PromQL } from '../types'

const emit = defineEmits<{
  'promql-updated': []
}>()

const { showSuccess, showError } = useNotification()
const expandable = useExpandable()
const { highlightPromQL } = usePromqlHighlight()

const promqls = ref<PromQL[]>([])
const showAddForm = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)

const formData = reactive({
  name: '',
  description: '',
  query: '',
  category: ''
})

// 获取 PromQL 列表
async function fetchPromQLs() {
  try {
    const data = await get<PromQL[]>('/api/promqls')
    if (Array.isArray(data)) {
      promqls.value = data
    }
  } catch (err: unknown) {
    showError(err instanceof Error ? err.message : '获取PromQL查询失败')
    promqls.value = []
  }
}

usePolling(fetchPromQLs, 30000)

// 保存 PromQL
async function savePromQL() {
  if (!formData.name || !formData.query) {
    showError('名称和查询语句不能为空')
    return
  }

  const payload = {
    name: formData.name,
    description: formData.description,
    query: formData.query,
    category: formData.category
  }

  try {
    if (isEditing.value && editingId.value !== null) {
      await put(`/api/promql/${editingId.value}`, payload)
      showSuccess('PromQL 更新成功')
    } else {
      await post('/api/promql', payload)
      showSuccess('PromQL 创建成功')
    }
    cancelEdit()
    await fetchPromQLs()
    emit('promql-updated')
  } catch (err: unknown) {
    showError(err instanceof Error ? err.message : '保存PromQL查询失败')
  }
}

// 开始编辑
function startEdit(promql: PromQL) {
  editingId.value = promql.id
  formData.name = promql.name
  formData.description = promql.description
  formData.query = promql.query
  formData.category = promql.category
  isEditing.value = true
  showAddForm.value = true
}

// 取消编辑
function cancelEdit() {
  editingId.value = null
  formData.name = ''
  formData.description = ''
  formData.query = ''
  formData.category = ''
  isEditing.value = false
  showAddForm.value = false
}

// 删除 PromQL
async function deletePromQL(id: number) {
  if (!confirm('确定要删除这个 PromQL 查询吗？如果有任务正在使用它，将无法删除。')) {
    return
  }

  try {
    await del(`/api/promql/${id}`)
    showSuccess('PromQL 删除成功')
    await fetchPromQLs()
    emit('promql-updated')
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : '删除PromQL查询失败'
    showError(`删除失败: ${message}`)
  }
}

// 复制 PromQL
async function copyPromQL(promql: PromQL) {
  try {
    await post('/api/promql', {
      name: `${promql.name} (复制)`,
      description: promql.description,
      query: promql.query,
      category: promql.category
    })
    showSuccess('PromQL 复制成功')
    await fetchPromQLs()
  } catch (err: unknown) {
    showError(err instanceof Error ? err.message : '复制PromQL查询失败')
  }
}
</script>

<style scoped>
.global-actions {
  margin: var(--spacing-lg) 0;
  display: flex;
  justify-content: flex-end;
}

.form-container {
  margin-bottom: var(--spacing-lg);
  padding: 15px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background-color: var(--color-bg-light);
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
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}

.form-actions {
  display: flex;
  gap: 10px;
}

.action-buttons {
  margin-bottom: var(--spacing-lg);
}

.promql-list {
  margin-top: var(--spacing-lg);
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
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
  transition: all 0.3s ease;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 50px;
  background-color: var(--color-bg-white);
  white-space: nowrap;
}

.action-btn:hover {
  background-color: var(--color-bg-hover);
  border-color: #adb5bd;
}

.action-btn.edit {
  color: var(--color-text-white);
  background-color: #2196F3;
  border-color: #2196F3;
}

.action-btn.edit:hover {
  background-color: #1976D2;
}

.action-btn.copy {
  color: var(--color-text-white);
  background-color: var(--color-success);
  border-color: var(--color-success);
}

.action-btn.copy:hover {
  background-color: var(--color-success-hover);
}

.action-btn.delete {
  color: var(--color-text-white);
  background-color: var(--color-danger);
  border-color: var(--color-danger);
}

.action-btn.delete:hover {
  background-color: var(--color-danger-hover);
}

.query-cell {
  position: relative;
  max-width: 500px;
  background: var(--color-bg-light);
  border-radius: var(--radius-md);
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
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 1.5;
  overflow-x: auto;
  padding: 4px;
}

.expand-btn {
  position: absolute;
  right: 8px;
  top: 8px;
  background: var(--color-border-light);
  border: 1px solid #ced4da;
  border-radius: var(--radius-sm);
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

.promql-preview {
  margin-top: 10px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: 10px;
  background-color: var(--color-bg-light);
}

.promql-preview h5 {
  margin-top: 0;
  margin-bottom: 10px;
  color: #495057;
  font-size: 14px;
}

td {
  vertical-align: middle;
  padding: 8px 4px;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

table {
  table-layout: fixed;
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
}

thead th {
  padding: 12px 8px;
  white-space: nowrap;
  background: var(--color-bg-light);
  position: sticky;
  top: 0;
  z-index: 1;
  border-bottom: 2px solid #dee2e6;
}

th:nth-child(1) { width: 5%; }
th:nth-child(2) { width: 10%; }
th:nth-child(3) { width: 10%; }
th:nth-child(4) { width: 15%; }
th:nth-child(5) { width: 30%; }
th:nth-child(6) { width: 10%; }
th:nth-child(7) { width: 10%; }
th:nth-child(8) { width: 10%; }

tbody tr:hover {
  background-color: var(--color-bg-hover);
}

tbody td {
  border-bottom: 1px solid #dee2e6;
}

@media (max-width: 1200px) {
  .query-cell { max-width: 300px; }
}

@media (max-width: 768px) {
  .query-cell { max-width: 200px; }
}
</style>
