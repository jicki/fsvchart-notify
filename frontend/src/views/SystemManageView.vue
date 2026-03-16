<template>
  <div>
    <div class="page-header">
      <div>
        <h3>系统管理</h3>
        <p>管理数据源与飞书 WebHook 配置</p>
      </div>
    </div>

    <!-- Tab 切换 -->
    <div class="tab-bar">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        class="tab-btn"
        :class="{ active: activeTab === tab.key }"
        @click="activeTab = tab.key"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- 数据源面板 -->
    <div v-if="activeTab === 'datasource'" class="card">
      <div class="panel-header">
        <button class="btn btn-primary" @click="dsOpenAdd">
          <IconPlus :size="16" />
          添加数据源
        </button>
      </div>
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th><th>名称</th><th>URL</th><th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="src in dsItems" :key="src.id">
            <td>{{ src.id }}</td>
            <td>{{ src.name }}</td>
            <td>{{ src.url }}</td>
            <td>
              <div class="action-group">
                <button class="btn-icon" @click="dsOpenEdit(src)" title="编辑">
                  <IconEdit :size="16" />
                </button>
                <button class="btn-icon btn-icon-danger" @click="dsDelete(src.id)" title="删除">
                  <IconTrash :size="16" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="dsItems.length === 0" class="empty">暂无数据源</div>
    </div>

    <!-- WebHook 面板 -->
    <div v-if="activeTab === 'webhook'" class="card">
      <div class="panel-header">
        <button class="btn btn-primary" @click="whOpenAdd">
          <IconPlus :size="16" />
          添加 WebHook
        </button>
      </div>
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th><th>名称</th><th>URL</th><th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="wb in whItems" :key="wb.id">
            <td>{{ wb.id }}</td>
            <td>{{ wb.name }}</td>
            <td>{{ wb.url }}</td>
            <td>
              <div class="action-group">
                <button class="btn-icon" @click="whOpenEdit(wb)" title="编辑">
                  <IconEdit :size="16" />
                </button>
                <button class="btn-icon btn-icon-danger" @click="whDelete(wb.id)" title="删除">
                  <IconTrash :size="16" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="whItems.length === 0" class="empty">暂无 WebHook</div>
    </div>

    <!-- 数据源弹窗 -->
    <ModalDialog
      :visible="dsShowModal"
      :title="dsIsEditing ? '编辑数据源' : '添加数据源'"
      max-width="600px"
      @close="dsCloseModal"
    >
      <div class="form-group">
        <label>名称</label>
        <input class="form-input" v-model="dsFormName" placeholder="数据源名称" />
      </div>
      <div class="form-group">
        <label>URL</label>
        <input class="form-input" v-model="dsFormURL" placeholder="Prometheus URL" />
      </div>
      <div class="modal-actions">
        <button class="btn btn-primary" @click="dsHandleSave">保存</button>
        <button class="btn btn-secondary" @click="dsCloseModal">取消</button>
      </div>
    </ModalDialog>

    <!-- WebHook 弹窗 -->
    <ModalDialog
      :visible="whShowModal"
      :title="whIsEditing ? '编辑 WebHook' : '添加 WebHook'"
      max-width="600px"
      @close="whCloseModal"
    >
      <div class="form-group">
        <label>名称</label>
        <input class="form-input" v-model="whFormName" placeholder="WebHook 名称" />
      </div>
      <div class="form-group">
        <label>URL</label>
        <input class="form-input" v-model="whFormURL" placeholder="WebHook URL" />
      </div>
      <div class="modal-actions">
        <button class="btn btn-primary" @click="whHandleSave">保存</button>
        <button class="btn btn-secondary" @click="whCloseModal">取消</button>
      </div>
    </ModalDialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useCrudList } from '../composables/useCrudList'
import { usePolling } from '../composables/usePolling'
import ModalDialog from '../components/ModalDialog.vue'
import { IconPlus, IconEdit, IconTrash } from '../components/icons'
import type { MetricsSource, FeishuWebhook } from '../types'

interface Tab {
  key: string
  label: string
}

const tabs: Tab[] = [
  { key: 'datasource', label: '数据源' },
  { key: 'webhook', label: 'WebHook' },
]

const activeTab = ref('datasource')

// ====== 数据源 ======
const {
  items: dsItems,
  fetchList: dsFetch,
  addItem: dsAdd,
  updateItem: dsUpdate,
  deleteItem: dsDelete,
  validateRequired: dsValidate,
} = useCrudList<MetricsSource>('/api/metrics_source', '数据源')

const dsShowModal = ref(false)
const dsIsEditing = ref(false)
const dsEditingId = ref<number | null>(null)
const dsFormName = ref('')
const dsFormURL = ref('')

function dsOpenAdd() {
  dsFormName.value = ''
  dsFormURL.value = ''
  dsIsEditing.value = false
  dsEditingId.value = null
  dsShowModal.value = true
}

function dsOpenEdit(src: MetricsSource) {
  dsFormName.value = src.name
  dsFormURL.value = src.url
  dsIsEditing.value = true
  dsEditingId.value = src.id
  dsShowModal.value = true
}

function dsCloseModal() {
  dsShowModal.value = false
  dsIsEditing.value = false
  dsEditingId.value = null
}

async function dsHandleSave() {
  if (!dsValidate({ [dsFormName.value]: '名称', [dsFormURL.value]: 'URL' })) return
  if (dsIsEditing.value && dsEditingId.value !== null) {
    await dsUpdate(dsEditingId.value, { name: dsFormName.value, url: dsFormURL.value } as Partial<MetricsSource>)
  } else {
    await dsAdd({ name: dsFormName.value, url: dsFormURL.value } as Partial<MetricsSource>)
  }
  dsCloseModal()
}

// ====== WebHook ======
const {
  items: whItems,
  fetchList: whFetch,
  addItem: whAdd,
  updateItem: whUpdate,
  deleteItem: whDelete,
  validateRequired: whValidate,
} = useCrudList<FeishuWebhook>('/api/feishu_webhook', 'WebHook')

const whShowModal = ref(false)
const whIsEditing = ref(false)
const whEditingId = ref<number | null>(null)
const whFormName = ref('')
const whFormURL = ref('')

function whOpenAdd() {
  whFormName.value = ''
  whFormURL.value = ''
  whIsEditing.value = false
  whEditingId.value = null
  whShowModal.value = true
}

function whOpenEdit(wb: FeishuWebhook) {
  whFormName.value = wb.name
  whFormURL.value = wb.url
  whIsEditing.value = true
  whEditingId.value = wb.id
  whShowModal.value = true
}

function whCloseModal() {
  whShowModal.value = false
  whIsEditing.value = false
  whEditingId.value = null
}

async function whHandleSave() {
  if (!whValidate({ [whFormName.value]: '名称', [whFormURL.value]: 'URL' })) return
  if (whIsEditing.value && whEditingId.value !== null) {
    await whUpdate(whEditingId.value, { name: whFormName.value, url: whFormURL.value } as Partial<FeishuWebhook>)
  } else {
    await whAdd({ name: whFormName.value, url: whFormURL.value } as Partial<FeishuWebhook>)
  }
  whCloseModal()
}

// 轮询两个列表
usePolling(() => {
  dsFetch()
  whFetch()
}, 30000)
</script>

<style scoped>
.tab-bar {
  display: flex;
  gap: 0;
  margin-bottom: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border);
}

.tab-btn {
  padding: 10px 20px;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--color-text-secondary);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.tab-btn:hover {
  color: var(--color-text);
}

.tab-btn.active {
  color: var(--color-accent);
  border-bottom-color: var(--color-accent);
}

.panel-header {
  display: flex;
  justify-content: flex-end;
  margin-bottom: var(--spacing-md);
}
</style>
