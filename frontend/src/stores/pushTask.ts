import { defineStore } from 'pinia'
import { ref } from 'vue'
import { get, post, put, del, fetchWithAuth } from '../utils/api'
import { useNotification } from '../composables/useNotification'
import type { PushTask, MetricsSource, FeishuWebhook, ChartTemplate, PromQL, PromQLConfig } from '../types'

export const usePushTaskStore = defineStore('pushTask', () => {
  const tasks = ref<PushTask[]>([])
  const sources = ref<MetricsSource[]>([])
  const webhooks = ref<FeishuWebhook[]>([])
  const chartTemplates = ref<ChartTemplate[]>([])
  const promqls = ref<PromQL[]>([])
  const loading = ref(false)

  const { showSuccess, showError } = useNotification()

  // 获取所有数据
  async function fetchAllData() {
    loading.value = true
    try {
      const [sourcesData, webhooksData, templatesData, promqlsData, tasksData] = await Promise.all([
        get<MetricsSource[]>('/api/metrics_source'),
        get<FeishuWebhook[]>('/api/feishu_webhook'),
        get<ChartTemplate[]>('/api/chart_template'),
        get<PromQL[]>('/api/promqls'),
        get<PushTask[]>('/api/push_task')
      ])

      sources.value = Array.isArray(sourcesData) ? sourcesData : []
      webhooks.value = Array.isArray(webhooksData) ? webhooksData : []
      chartTemplates.value = Array.isArray(templatesData) ? templatesData : []
      promqls.value = Array.isArray(promqlsData) ? promqlsData : []

      if (Array.isArray(tasksData)) {
        tasks.value = tasksData.map(task => processTask(task))
      } else {
        tasks.value = []
      }
    } catch (error) {
      const err = error instanceof Error ? error.message : '数据获取失败'
      console.error('数据获取失败:', error)
      showError(err)
      sources.value = []
      webhooks.value = []
      chartTemplates.value = []
      promqls.value = []
      tasks.value = []
    } finally {
      loading.value = false
    }
  }

  // 处理单个任务数据
  function processTask(task: PushTask): PushTask {
    const processed: PushTask = {
      ...task,
      id: task.id || 0,
      name: task.name || '未命名任务',
      source_id: task.source_id || 0,
      enabled: typeof task.enabled === 'boolean' ? task.enabled : task.enabled === 1,
      webhook_ids: Array.isArray(task.webhook_ids) ? task.webhook_ids : [],
      promql_ids: Array.isArray(task.promql_ids) ? task.promql_ids : [],
      send_times: Array.isArray(task.send_times) ? task.send_times : [],
      promql_configs: Array.isArray(task.promql_configs) ? task.promql_configs : [],
      chart_template_id: task.chart_template_id || null,
      bound_webhooks: []
    }

    // 处理 webhook 绑定
    if (Array.isArray(task.webhook_ids)) {
      processed.bound_webhooks = task.webhook_ids.map(id => {
        const webhook = webhooks.value.find(w => w.id === id)
        return webhook || { id, name: `WebHook #${id}`, url: '' }
      })
    }

    return processed
  }

  // 创建任务
  async function createTask(payload: Record<string, unknown>): Promise<boolean> {
    try {
      loading.value = true
      await post('/api/push_task', payload)
      await fetchAllData()
      showSuccess('任务创建成功')
      return true
    } catch (error) {
      const err = error instanceof Error ? error.message : '创建任务失败'
      showError(`创建任务失败: ${err}`)
      return false
    } finally {
      loading.value = false
    }
  }

  // 更新任务
  async function updateTask(taskId: number, payload: Record<string, unknown>): Promise<boolean> {
    try {
      loading.value = true
      await put(`/api/push_task/${taskId}`, payload)
      await fetchAllData()
      showSuccess('任务更新成功')
      return true
    } catch (error) {
      const err = error instanceof Error ? error.message : '更新任务失败'
      showError(`更新任务失败: ${err}`)
      return false
    } finally {
      loading.value = false
    }
  }

  // 删除任务
  async function deleteTask(taskId: number): Promise<boolean> {
    if (!confirm('确定要删除此任务吗？')) return false

    try {
      loading.value = true
      await del(`/api/push_task/${taskId}`)
      await fetchAllData()
      showSuccess('任务删除成功')
      return true
    } catch (error) {
      showError('删除任务失败')
      return false
    } finally {
      loading.value = false
    }
  }

  // 切换任务状态
  async function toggleTask(taskId: number, enabled: boolean): Promise<boolean> {
    const action = enabled ? '启用' : '禁用'
    try {
      await put(`/api/push_task/${taskId}/toggle`, { enabled: enabled ? 1 : 0 })
      await fetchAllData()
      showSuccess(`任务${action}成功`)
      return true
    } catch (error) {
      showError(`${action}任务失败`)
      return false
    }
  }

  // 立即运行任务
  async function runTask(taskId: number): Promise<boolean> {
    try {
      await post(`/api/push_task/${taskId}/run`, {})
      showSuccess('任务已触发运行')
      return true
    } catch (error) {
      showError('运行任务失败')
      return false
    }
  }

  // 复制任务
  async function copyTask(task: PushTask): Promise<boolean> {
    if (!task.chart_template_id) {
      showError('原任务没有设置图表模板，请先设置图表模板')
      return false
    }

    const templateExists = chartTemplates.value.some(t => t.id === task.chart_template_id)
    if (!templateExists) {
      showError('原任务的图表模板不存在，请先检查图表模板配置')
      return false
    }

    if (!task.promql_ids || task.promql_ids.length === 0) {
      showError('原任务没有关联的PromQL查询')
      return false
    }

    const queries = promqls.value
      .filter(p => task.promql_ids.includes(p.id))
      .map(p => p.query)

    let promqlConfigsList: Partial<PromQLConfig>[]
    if (task.promql_configs && task.promql_configs.length > 0) {
      promqlConfigsList = task.promql_configs.map(config => ({
        promql_id: config.promql_id,
        unit: config.unit || '',
        metric_label: config.metric_label || 'pod',
        custom_metric_label: config.custom_metric_label || '',
        initial_unit: config.initial_unit || '',
        display_order: config.display_order || 0,
        display_mode: config.display_mode || 'chart',
        chart_template_id: config.chart_template_id || task.chart_template_id
      }))
    } else {
      promqlConfigsList = task.promql_ids.map(promqlId => ({
        promql_id: promqlId,
        unit: task.unit || '',
        metric_label: task.metric_label || 'pod',
        custom_metric_label: task.custom_metric_label || '',
        initial_unit: '',
        display_order: 0,
        display_mode: 'chart' as const,
        chart_template_id: task.chart_template_id
      }))
    }

    const newTask = {
      name: `${task.name} (复制)`,
      source_id: task.source_id,
      promql_ids: task.promql_ids,
      promql_configs: promqlConfigsList,
      query: queries.join(', '),
      time_range: task.time_range,
      step: 0,
      chart_template_id: task.chart_template_id,
      webhook_ids: task.webhook_ids,
      card_title: task.card_title || '',
      card_template: task.card_template || 'blue',
      metric_label: task.metric_label || 'pod',
      custom_metric_label: task.custom_metric_label || '',
      unit: task.unit || '',
      button_text: task.button_text || '',
      button_url: task.button_url || '',
      show_data_label: task.show_data_label || false,
      enabled: 1,
      send_times: task.send_times.map(time => ({
        weekday: typeof time.weekday === 'string' ? parseInt(time.weekday) : time.weekday,
        send_time: time.send_time
      })),
      schedule_interval: task.schedule_interval || 3600,
      promql_chart_templates: task.promql_ids.map(promqlId => ({
        promql_id: promqlId,
        chart_template_id: task.chart_template_id
      }))
    }

    return createTask(newTask)
  }

  // 辅助函数
  function getSourceName(sourceId: number): string {
    const source = sources.value.find(s => s.id === sourceId)
    return source ? source.name : `数据源#${sourceId}`
  }

  function getPromqlName(promqlId: number): string {
    const promql = promqls.value.find(p => p.id === promqlId)
    return promql ? promql.name : `PromQL#${promqlId}`
  }

  return {
    tasks,
    sources,
    webhooks,
    chartTemplates,
    promqls,
    loading,
    fetchAllData,
    createTask,
    updateTask,
    deleteTask,
    toggleTask,
    runTask,
    copyTask,
    getSourceName,
    getPromqlName
  }
})
