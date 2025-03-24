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

    <!-- 不同 tab 显示不同子组件 -->
    <div class="tab-content">
      <component :is="tabs[currentTab]" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

// 导入子组件
import DataSourceView from './DataSourceView.vue'
import WebhookView from './WebhookView.vue'
import ChartTemplatesView from './ChartTemplatesView.vue'
import PromQLView from './PromQLView.vue'
import PushTasksView from './PushTasksView.vue'

const currentTab = ref('datasource')

// 定义组件选项
const tabs = {
  datasource: DataSourceView,
  webhook: WebhookView,
  chartTemplates: ChartTemplatesView,
  promql: PromQLView,
  pushTasks: PushTasksView
}
</script>

<!-- 注意：此处不使用 scoped，方便所有子组件继承这些公共样式 -->
<style>
.home-container {
  height: 100%;
}

.dashboard-header {
  margin-bottom: 20px;
}

.dashboard-header h2 {
  margin: 0;
  color: #333;
}

.tabs {
  margin-bottom:12px;
  border-bottom: 1px solid #e9ecef;
  padding-bottom: 10px;
}

button {
  margin-right:8px;
  padding:6px 12px;
  cursor:pointer;
  background: #fff;
  border: 1px solid #e9ecef;
  border-radius: 4px;
  transition: all 0.2s;
}

button.active {
  background: #007bff;
  color: #fff;
  border-color: #007bff;
  font-weight:bold;
}

button:hover {
  background: #f8f9fa;
}

button.active:hover {
  background: #0056b3;
}

/* 通用 .tab-content, table 样式 */
.tab-content {
  padding:16px;
  margin-bottom:24px;
  background: #fff;
}

table {
  margin-top:10px;
  width:100%;
  border-collapse:collapse;
}

th,td {
  border:1px solid #ccc;
  padding:4px 8px;
  text-align:left;
}
</style>
