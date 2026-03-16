<template>
  <div class="form-group">
    <label>选择预定义 PromQL 查询 (必选):</label>
    <div class="promql-selection">
      <div v-if="promqls.length > 0">
        <div class="promql-actions">
          <button class="action-btn" type="button" @click="handleToggleAll">
            {{ isAllExpanded ? '收起所有查询' : '展开所有查询' }}
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
            <button class="expand-btn" type="button" @click="toggle(promql.id)">
              {{ isExpanded(promql.id) ? '收起' : '展开' }}
            </button>
          </div>
          <div class="promql-details" :class="{ 'expanded': isExpanded(promql.id) }">
            <pre class="promql-query" v-html="highlightPromQL(promql.query)"></pre>
            <div class="promql-description" v-if="promql.description">
              {{ promql.description }}
            </div>

            <!-- 为选中的 PromQL 添加配置面板 -->
            <div v-if="selectedIds.includes(promql.id.toString())" class="promql-config">
              <h5>PromQL 配置</h5>
              <div class="config-group">
                <label>显示顺序:</label>
                <input
                  type="number"
                  :value="getConfig(promql.id).display_order"
                  @input="updateConfig(promql.id, 'display_order', ($event.target as HTMLInputElement).valueAsNumber)"
                  placeholder="数字越小越靠前"
                  min="0" step="1"
                />
                <small>控制在卡片中的显示顺序，数字越小越靠前（默认为0）</small>
              </div>
              <div class="config-group">
                <label>展示模式:</label>
                <select
                  :value="getConfig(promql.id).display_mode"
                  @change="updateConfig(promql.id, 'display_mode', ($event.target as HTMLSelectElement).value)"
                >
                  <option value="chart">图表模式</option>
                  <option value="text">文本模式</option>
                  <option value="both">图表+文本</option>
                </select>
              </div>
              <div class="config-group">
                <label>初始单位 (可选):</label>
                <input
                  type="text"
                  :value="getConfig(promql.id).initial_unit"
                  @input="updateConfig(promql.id, 'initial_unit', ($event.target as HTMLInputElement).value)"
                  placeholder="例如: bytes, ms, s (留空则不转换)"
                />
              </div>
              <div class="config-group">
                <label>单位:</label>
                <input
                  type="text"
                  :value="getConfig(promql.id).unit"
                  @input="updateConfig(promql.id, 'unit', ($event.target as HTMLInputElement).value)"
                  placeholder="例如: TiB, GiB, s, m"
                />
              </div>
              <div class="config-group">
                <label>指标标签:</label>
                <select
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
                <label>自定义指标标签:</label>
                <input
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
      <div v-else>
        <p>暂无预定义 PromQL 查询，请先在 PromQL 管理中创建</p>
      </div>
    </div>
    <div class="form-hint">请在 PromQL 管理中创建并选择预定义的查询</div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { usePromqlHighlight } from '../../composables/usePromqlHighlight'
import { useExpandable } from '../../composables/useExpandable'
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
  border: 1px solid var(--color-border, #ddd);
  border-radius: 4px;
  padding: 15px;
  background: #fff;
}
.promql-actions { margin-bottom: 15px; display: flex; justify-content: flex-end; }
.promql-item { margin-bottom: 12px; padding: 8px; border: 1px solid #eee; border-radius: 4px; background: var(--color-bg-light, #f8f9fa); }
.promql-header { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.promql-label { flex: 1; font-weight: 500; cursor: pointer; }
.promql-category { color: var(--color-text-secondary, #666); font-size: 0.9em; font-weight: normal; margin-left: 8px; }
.promql-details { display: none; margin-top: 8px; padding: 8px; background: #fff; border-radius: 4px; }
.promql-details.expanded { display: block; }
.promql-query { margin: 0; padding: 8px; background: var(--color-bg-light, #f8f9fa); border-radius: 4px; font-family: var(--font-mono); font-size: 13px; line-height: 1.5; white-space: pre-wrap; word-break: break-word; }
.promql-description { margin-top: 8px; color: var(--color-text-secondary, #666); font-size: 0.9em; }
.promql-config { margin-top: 12px; padding: 12px; border: 1px solid #e0e0e0; border-radius: 4px; background: #fafafa; }
.promql-config h5 { margin: 0 0 10px 0; }
.config-group { margin-bottom: 10px; }
.config-group label { display: block; margin-bottom: 4px; font-weight: 500; font-size: 13px; }
.config-group input, .config-group select { width: 100%; padding: 6px; border: 1px solid var(--color-border, #ddd); border-radius: 4px; }
.config-group small { color: var(--color-text-secondary, #666); font-size: 12px; display: block; margin-top: 4px; }
.expand-btn { padding: 4px 8px; background: #e9ecef; border: none; border-radius: 3px; font-size: 12px; cursor: pointer; }
.expand-btn:hover { background: #dee2e6; }
.action-btn { padding: 6px 12px; border: none; border-radius: 4px; cursor: pointer; font-size: 14px; background-color: #f0f0f0; }
.action-btn:hover { background-color: #e0e0e0; }
.form-hint { font-size: 0.85em; color: var(--color-text-secondary, #666); margin-top: 5px; }

:deep(.keyword) { color: #0066cc; font-weight: 500; }
:deep(.label) { color: #e83e8c; }
:deep(.number) { color: #2e7d32; }
:deep(.unit) { color: #0066cc; font-weight: 500; }
</style>
