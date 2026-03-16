<template>
  <div>
    <div class="page-header">
      <div>
        <h3>PromQL 查询管理</h3>
        <p>管理预定义的 PromQL 查询，可在创建任务时直接选择使用</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-secondary btn-sm" @click="expandable.toggleAll(promqls.map(p => p.id))">
          <IconChevronDown v-if="!expandable.isAllExpandedFor(promqls.length)" :size="16" />
          <IconChevronUp v-else :size="16" />
          {{ expandable.isAllExpandedFor(promqls.length) ? '收起所有' : '展开所有' }}
        </button>
        <button class="btn btn-primary" @click="openAddModal">
          <IconPlus :size="16" />
          添加查询
        </button>
      </div>
    </div>

    <!-- PromQL 列表 -->
    <div class="card">
      <table class="data-table">
        <thead>
          <tr>
            <th style="width: 5%">ID</th>
            <th style="width: 10%">名称</th>
            <th style="width: 10%">分类</th>
            <th style="width: 15%">描述</th>
            <th style="width: 30%">查询语句</th>
            <th style="width: 10%">创建时间</th>
            <th style="width: 10%">更新时间</th>
            <th style="width: 10%">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="promql in promqls" :key="promql.id">
            <td>{{ promql.id }}</td>
            <td>{{ promql.name }}</td>
            <td>{{ promql.category }}</td>
            <td class="text-ellipsis">{{ promql.description }}</td>
            <td>
              <div class="query-cell" :class="{ expanded: expandable.isExpanded(promql.id) }">
                <pre class="promql-code" v-html="highlightPromQL(promql.query)"></pre>
                <button class="btn-icon expand-toggle" @click="expandable.toggle(promql.id)">
                  <IconChevronDown v-if="!expandable.isExpanded(promql.id)" :size="14" />
                  <IconChevronUp v-else :size="14" />
                </button>
              </div>
            </td>
            <td>{{ formatDate(promql.created_at) }}</td>
            <td>{{ formatDate(promql.updated_at) }}</td>
            <td>
              <div class="action-group">
                <button class="btn-icon" @click="openEditModal(promql)" title="编辑">
                  <IconEdit :size="16" />
                </button>
                <button class="btn-icon" @click="copyPromQL(promql)" title="复制" style="color: var(--color-purple)">
                  <IconCopy :size="16" />
                </button>
                <button class="btn-icon" @click="deletePromQL(promql.id)" title="删除" style="color: var(--color-danger)">
                  <IconTrash :size="16" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="promqls.length === 0" class="empty">暂无 PromQL 查询</div>
    </div>

    <!-- 添加/编辑弹窗 -->
    <ModalDialog
      :visible="showModal"
      :title="isEditing ? '编辑 PromQL 查询' : '添加 PromQL 查询'"
      max-width="900px"
      @close="cancelEdit"
    >
      <div class="form-group">
        <label>名称</label>
        <input class="form-input" type="text" v-model="formData.name" placeholder="查询名称" />
      </div>
      <div class="form-group">
        <label>分类</label>
        <input class="form-input" type="text" v-model="formData.category" placeholder="查询分类" />
      </div>
      <div class="form-group">
        <label>描述</label>
        <textarea class="form-input" v-model="formData.description" placeholder="查询描述"></textarea>
      </div>
      <div class="form-group">
        <label>PromQL 查询</label>
        <textarea class="form-input promql-textarea" v-model="formData.query" placeholder="PromQL 查询语句" rows="5"></textarea>
        <div v-if="formData.query" class="promql-preview">
          <h5>语法高亮预览</h5>
          <pre class="promql-code" v-html="highlightPromQL(formData.query)"></pre>
        </div>
      </div>
      <div class="modal-actions">
        <button class="btn btn-primary" @click="savePromQL">保存</button>
        <button class="btn btn-secondary" @click="cancelEdit">取消</button>
      </div>
    </ModalDialog>
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
import ModalDialog from '../components/ModalDialog.vue'
import { IconPlus, IconEdit, IconTrash, IconCopy, IconChevronDown, IconChevronUp } from '../components/icons'
import type { PromQL } from '../types'

const { showSuccess, showError } = useNotification()
const expandable = useExpandable()
const { highlightPromQL } = usePromqlHighlight()

const promqls = ref<PromQL[]>([])
const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)

const formData = reactive({
  name: '',
  description: '',
  query: '',
  category: ''
})

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

function openAddModal() {
  formData.name = ''
  formData.description = ''
  formData.query = ''
  formData.category = ''
  isEditing.value = false
  editingId.value = null
  showModal.value = true
}

function openEditModal(promql: PromQL) {
  editingId.value = promql.id
  formData.name = promql.name
  formData.description = promql.description
  formData.query = promql.query
  formData.category = promql.category
  isEditing.value = true
  showModal.value = true
}

function cancelEdit() {
  editingId.value = null
  formData.name = ''
  formData.description = ''
  formData.query = ''
  formData.category = ''
  isEditing.value = false
  showModal.value = false
}

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
  } catch (err: unknown) {
    showError(err instanceof Error ? err.message : '保存PromQL查询失败')
  }
}

async function deletePromQL(id: number) {
  if (!confirm('确定要删除这个 PromQL 查询吗？如果有任务正在使用它，将无法删除。')) {
    return
  }

  try {
    await del(`/api/promql/${id}`)
    showSuccess('PromQL 删除成功')
    await fetchPromQLs()
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : '删除PromQL查询失败'
    showError(`删除失败: ${message}`)
  }
}

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
.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.promql-textarea {
  font-family: var(--font-mono);
  resize: vertical;
}

.promql-preview {
  margin-top: var(--spacing-sm);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  background-color: var(--color-bg-light);
}

.promql-preview h5 {
  margin: 0 0 8px;
  color: var(--color-text-secondary);
  font-size: 13px;
}

.query-cell {
  position: relative;
  max-width: 500px;
  background: var(--color-bg-light);
  border-radius: var(--radius-md);
  padding: var(--spacing-sm) var(--spacing-md);
  transition: all var(--transition-normal);
  margin: 4px 0;
  overflow: hidden;
  max-height: 80px;
}

.query-cell.expanded {
  max-width: none;
  max-height: none;
  min-width: 500px;
}

.expand-toggle {
  position: absolute;
  right: 4px;
  top: 4px;
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

.text-ellipsis {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.action-group {
  display: flex;
  gap: 2px;
  align-items: center;
}

.modal-actions {
  display: flex;
  gap: 8px;
  margin-top: var(--spacing-lg);
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--color-border);
}

/* PromQL 语法高亮 */
:deep(.keyword) {
  color: var(--color-primary);
  font-weight: 500;
}

:deep(.label) {
  color: #ec4899;
}

:deep(.number) {
  color: #059669;
}

:deep(.unit) {
  color: var(--color-primary);
  font-weight: 500;
}

@media (max-width: 1200px) {
  .query-cell { max-width: 300px; }
}

@media (max-width: 768px) {
  .query-cell { max-width: 200px; }
}
</style>
