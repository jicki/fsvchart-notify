<template>
  <div class="form-group">
    <label>选择预定义 PromQL 查询 (必选)</label>
    <div class="promql-selection">
      <div v-if="promqls.length > 0">
        <div class="promql-actions">
          <button class="btn btn-secondary btn-sm" type="button" @click="handleToggleAll">
            <IconChevronDown v-if="!isAllExpanded" :size="14" />
            <IconChevronUp v-else :size="14" />
            {{ isAllExpanded ? '收起所有' : '展开所有' }}
          </button>
        </div>
        <div v-for="promql in promqls" :key="promql.id" class="promql-item">
          <div class="promql-header">
            <input
              type="checkbox"
              :id="`${idPrefix}-promql-${promql.id}`"
              :value="promql.id.toString()"
              v-model="selectedIds"
            >
            <label :for="`${idPrefix}-promql-${promql.id}`" class="promql-label">
              {{ promql.name }}
              <span class="promql-category" v-if="promql.category">({{ promql.category }})</span>
            </label>
            <button class="btn-icon" type="button" @click="toggle(promql.id)">
              <IconChevronDown v-if="!isExpanded(promql.id)" :size="14" />
              <IconChevronUp v-else :size="14" />
            </button>
          </div>
          <div class="promql-details" :class="{ expanded: isExpanded(promql.id) }">
            <pre class="promql-query" v-html="highlightPromQL(promql.query)"></pre>
            <div class="promql-description" v-if="promql.description">
              {{ promql.description }}
            </div>

            <div v-if="selectedIds.includes(promql.id.toString())" class="promql-config card">
              <h5>PromQL 配置</h5>
              <div class="config-group">
                <label>显示顺序</label>
                <input
                  class="form-input"
                  type="number"
                  :value="getConfig(promql.id).display_order"
                  @input="updateConfig(promql.id, 'display_order', ($event.target as HTMLInputElement).valueAsNumber)"
                  placeholder="数字越小越靠前"
                  min="0" step="1"
                />
                <div class="form-hint">控制在卡片中的显示顺序，数字越小越靠前（默认为0）</div>
              </div>
              <div class="config-group">
                <label>展示模式</label>
                <select
                  class="form-input"
                  :value="getConfig(promql.id).display_mode"
                  @change="updateConfig(promql.id, 'display_mode', ($event.target as HTMLSelectElement).value)"
                >
                  <option value="chart">图表模式</option>
                  <option value="text">文本模式</option>
                  <option value="both">图表+文本</option>
                </select>
              </div>
              <div class="config-group">
                <label>初始单位 (可选)</label>
                <input
                  class="form-input"
                  type="text"
                  :value="getConfig(promql.id).initial_unit"
                  @input="updateConfig(promql.id, 'initial_unit', ($event.target as HTMLInputElement).value)"
                  placeholder="例如: bytes, ms, s (留空则不转换)"
                />
              </div>
              <div class="config-group">
                <label>单位</label>
                <input
                  class="form-input"
                  type="text"
                  :value="getConfig(promql.id).unit"
                  @input="updateConfig(promql.id, 'unit', ($event.target as HTMLInputElement).value)"
                  placeholder="例如: TiB, GiB, s, m"
                />
              </div>
              <div class="config-group">
                <label>指标标签</label>
                <select
                  class="form-input"
                  :value="getConfig(promql.id).metric_label"
                  @change="updateConfig(promql.id, 'metric_label', ($event.target as HTMLSelectElement).value)"
                >
                  <option value="pod">Pod 名称</option>
                  <option value="namespace">命名空间</option>
                  <option value="container">容器名称</option>
                  <option value="instance">实例</option>
                  <option value="job">任务名</option>
                  <option value="node">节点名称</option>
                  <option value="cluster">集群名称</option>
                </select>
              </div>
              <div class="config-group">
                <label>自定义指标标签</label>
                <input
                  class="form-input"
                  type="text"
                  :value="getConfig(promql.id).custom_metric_label"
                  @input="updateConfig(promql.id, 'custom_metric_label', ($event.target as HTMLInputElement).value)"
                  placeholder="留空则使用上方标准标签"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="empty">
        暂无预定义 PromQL 查询，请先在 PromQL 管理中创建
      </div>
    </div>
    <div class="form-hint">请在 PromQL 管理中创建并选择预定义的查询</div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { usePromqlHighlight } from '../../composables/usePromqlHighlight'
import { useExpandable } from '../../composables/useExpandable'
import { IconChevronDown, IconChevronUp } from '../icons'
import type { PromQL } from '../../types'

interface PromQLConfigForm {
  unit: string
  metric_label: string
  custom_metric_label: string
  initial_unit: string
  display_order: number
  display_mode: string
}

const props = defineProps<{
  promqls: PromQL[]
  idPrefix: string
}>()

const selectedIds = defineModel<string[]>('selectedIds', { default: () => [] })
const configs = defineModel<Record<number, PromQLConfigForm>>('configs', { default: () => ({}) })

const { highlightPromQL } = usePromqlHighlight()
const { toggle, isExpanded, toggleAll, isAllExpandedFor } = useExpandable()

const isAllExpanded = computed(() => isAllExpandedFor(props.promqls.length))

function handleToggleAll() {
  toggleAll(props.promqls.map(p => p.id))
}

function getConfig(promqlId: number): PromQLConfigForm {
  return configs.value[promqlId] || {
    unit: '', metric_label: 'pod', custom_metric_label: '',
    initial_unit: '', display_order: 0, display_mode: 'chart'
  }
}

function updateConfig(promqlId: number, field: string, value: string | number) {
  if (!configs.value[promqlId]) {
    configs.value[promqlId] = {
      unit: '', metric_label: 'pod', custom_metric_label: '',
      initial_unit: '', display_order: 0, display_mode: 'chart'
    }
  }
  ;(configs.value[promqlId] as Record<string, string | number>)[field] = value
}
</script>

<style scoped>
.promql-selection {
  max-height: 400px;
  overflow-y: auto;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  background: var(--color-bg-white);
}

.promql-actions {
  margin-bottom: var(--spacing-md);
  display: flex;
  justify-content: flex-end;
}

.promql-item {
  margin-bottom: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-light);
}

.promql-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.promql-label {
  flex: 1;
  font-weight: 500;
  cursor: pointer;
  font-size: 14px;
}

.promql-category {
  color: var(--color-text-secondary);
  font-size: 13px;
  font-weight: normal;
  margin-left: 4px;
}

.promql-details {
  display: none;
  margin-top: var(--spacing-sm);
  padding: var(--spacing-sm);
  background: var(--color-bg-white);
  border-radius: var(--radius-md);
}

.promql-details.expanded {
  display: block;
}

.promql-query {
  margin: 0;
  padding: var(--spacing-sm);
  background: var(--color-bg-light);
  border-radius: var(--radius-md);
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}

.promql-description {
  margin-top: var(--spacing-sm);
  color: var(--color-text-secondary);
  font-size: 13px;
}

.promql-config {
  margin-top: var(--spacing-md);
  padding: var(--spacing-md) !important;
}

.promql-config h5 {
  margin: 0 0 var(--spacing-sm);
  font-size: 14px;
}

.config-group {
  margin-bottom: var(--spacing-sm);
}

.config-group label {
  display: block;
  margin-bottom: 4px;
  font-weight: 500;
  font-size: 13px;
}

.config-group .form-input {
  width: 100%;
}

.form-hint {
  font-size: 13px;
  color: var(--color-text-muted);
  margin-top: 4px;
}

:deep(.keyword) { color: var(--color-primary); font-weight: 500; }
:deep(.label) { color: #ec4899; }
:deep(.number) { color: #059669; }
:deep(.unit) { color: var(--color-primary); font-weight: 500; }
</style>
