<template>
  <div class="home-container">
    <div class="dashboard-header">
      <h2>fsvchart-notify 仪表盘</h2>
    </div>

    <div class="tabs">
      <button :class="{active: currentTab==='datasource'}" @click="currentTab='datasource'">数据源</button>
      <button :class="{active: currentTab==='webhook'}" @click="currentTab='webhook'">飞书WebHooks</button>
      <button :class="{active: currentTab==='chartTemplates'}" @click="currentTab='chartTemplates'">图表模板</button>
      <button :class="{active: currentTab==='promql'}" @click="currentTab='promql'">PromQL 管理</button>
      <button :class="{active: currentTab==='pushTasks'}" @click="currentTab='pushTasks'">Push Tasks</button>
    </div>

    <div class="tab-panel">
      <component :is="tabs[currentTab]" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import DataSourceView from './DataSourceView.vue'
import WebhookView from './WebhookView.vue'
import ChartTemplatesView from './ChartTemplatesView.vue'
import PromQLView from './PromQLView.vue'
import PushTasksView from './PushTasksView.vue'

const currentTab = ref('datasource')

const tabs: Record<string, unknown> = {
  datasource: DataSourceView,
  webhook: WebhookView,
  chartTemplates: ChartTemplatesView,
  promql: PromQLView,
  pushTasks: PushTasksView
}
</script>

<style scoped>
.home-container {
  height: 100%;
}

.dashboard-header {
  margin-bottom: var(--spacing-lg);
}

.dashboard-header h2 {
  margin: 0;
  color: var(--color-text);
}

.tabs {
  margin-bottom: 12px;
  border-bottom: 1px solid var(--color-border-light);
  padding-bottom: 10px;
}

.tabs button {
  margin-right: 8px;
  padding: 6px 12px;
  cursor: pointer;
  background: var(--color-bg-white);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-md);
  transition: all 0.2s;
}

.tabs button.active {
  background: var(--color-primary);
  color: var(--color-text-white);
  border-color: var(--color-primary);
  font-weight: bold;
}

.tabs button:hover {
  background: var(--color-bg-hover);
}

.tabs button.active:hover {
  background: var(--color-primary-hover);
}

.tab-panel {
  padding: var(--spacing-lg);
  background: var(--color-bg-white);
}
</style>
