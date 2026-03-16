import { ref, watch } from 'vue'
import type { PushTask, SendTime } from '../types'
import { useNotification } from './useNotification'

interface PromQLConfigForm {
  unit: string
  metric_label: string
  custom_metric_label: string
  initial_unit: string
  display_order: number
  display_mode: string
}

export function usePushTaskForm() {
  const { showWarning } = useNotification()

  const name = ref('')
  const sourceId = ref('')
  const timeRange = ref('7d')
  const chartTemplateId = ref('')
  const cardTitle = ref('')
  const cardTemplate = ref('blue')
  const metricLabel = ref('pod')
  const customMetricLabel = ref('')
  const unit = ref('')
  const webhookIds = ref<number[]>([])
  const selectedPromQLs = ref<string[]>([])
  const buttonText = ref('')
  const buttonURL = ref('')
  const showDataLabel = ref(false)
  const sendTimes = ref<SendTime[]>([{ _id: Date.now(), weekday: 1, send_time: '09:00' }])
  const promqlConfigs = ref<Record<number, PromQLConfigForm>>({})

  // 监听 selectedPromQLs 变化，自动初始化配置
  watch(selectedPromQLs, (newVal, oldVal) => {
    newVal.forEach(id => {
      const numId = parseInt(id)
      if (!promqlConfigs.value[numId]) {
        promqlConfigs.value[numId] = {
          unit: unit.value || '',
          metric_label: metricLabel.value || 'pod',
          custom_metric_label: '',
          initial_unit: '',
          display_order: 0,
          display_mode: 'chart'
        }
      }
    })
  })

  function resetForm() {
    name.value = ''
    sourceId.value = ''
    timeRange.value = '7d'
    chartTemplateId.value = ''
    cardTitle.value = ''
    cardTemplate.value = 'blue'
    metricLabel.value = 'pod'
    customMetricLabel.value = ''
    unit.value = ''
    webhookIds.value = []
    selectedPromQLs.value = []
    buttonText.value = ''
    buttonURL.value = ''
    showDataLabel.value = false
    sendTimes.value = [{ _id: Date.now(), weekday: 1, send_time: '09:00' }]
    promqlConfigs.value = {}
  }

  function loadTask(task: PushTask) {
    name.value = task.name
    sourceId.value = task.source_id.toString()
    timeRange.value = task.time_range || '7d'
    chartTemplateId.value = task.chart_template_id?.toString() || ''
    cardTitle.value = task.card_title || ''
    cardTemplate.value = task.card_template || 'blue'
    metricLabel.value = task.metric_label || 'pod'
    customMetricLabel.value = task.custom_metric_label || ''
    unit.value = task.unit || ''
    webhookIds.value = task.webhook_ids || []
    selectedPromQLs.value = task.promql_ids?.map(id => id.toString()) || []
    buttonText.value = task.button_text || ''
    buttonURL.value = task.button_url || ''
    showDataLabel.value = task.show_data_label || false

    if (Array.isArray(task.send_times) && task.send_times.length > 0) {
      sendTimes.value = task.send_times.map(time => ({
        _id: Date.now() + Math.random(),
        weekday: time.weekday,
        send_time: time.send_time
      }))
    } else {
      sendTimes.value = [{ _id: Date.now(), weekday: 1, send_time: '09:00' }]
    }

    // 加载 PromQL 配置
    promqlConfigs.value = {}
    if (task.promql_configs && task.promql_configs.length > 0) {
      task.promql_configs.forEach(config => {
        promqlConfigs.value[config.promql_id] = {
          unit: config.unit || '',
          metric_label: config.metric_label || 'pod',
          custom_metric_label: config.custom_metric_label || '',
          initial_unit: config.initial_unit || '',
          display_order: config.display_order || 0,
          display_mode: config.display_mode || 'chart'
        }
      })
    } else if (task.promql_ids && task.promql_ids.length > 0) {
      task.promql_ids.forEach(promqlId => {
        promqlConfigs.value[promqlId] = {
          unit: task.unit || '',
          metric_label: task.metric_label || 'pod',
          custom_metric_label: task.custom_metric_label || '',
          initial_unit: '',
          display_order: 0,
          display_mode: 'chart'
        }
      })
    }
  }

  function validate(): boolean {
    if (!name.value) {
      showWarning('请输入任务名称')
      return false
    }
    if (!sourceId.value) {
      showWarning('请选择数据源')
      return false
    }
    if (selectedPromQLs.value.length === 0) {
      showWarning('请至少选择一个PromQL查询')
      return false
    }
    if (sendTimes.value.length === 0) {
      showWarning('请至少设置一个发送时间')
      return false
    }
    if (!chartTemplateId.value) {
      showWarning('请选择图表模板')
      return false
    }
    return true
  }

  function buildPayload(promqls: { id: number; query: string }[]) {
    const promqlConfigsList = selectedPromQLs.value.map(id => {
      const numId = parseInt(id)
      const config = promqlConfigs.value[numId] || {
        unit: '', metric_label: 'pod', custom_metric_label: '',
        initial_unit: '', display_order: 0, display_mode: 'chart'
      }
      return {
        promql_id: numId,
        unit: config.unit,
        metric_label: config.metric_label,
        custom_metric_label: config.custom_metric_label,
        initial_unit: config.initial_unit,
        display_order: config.display_order,
        display_mode: config.display_mode,
        chart_template_id: parseInt(chartTemplateId.value) || null
      }
    })

    return {
      name: name.value,
      source_id: parseInt(sourceId.value),
      promql_ids: selectedPromQLs.value.map(id => parseInt(id)),
      promql_configs: promqlConfigsList,
      query: selectedPromQLs.value.map(id => {
        const promql = promqls.find(p => p.id.toString() === id)
        return promql ? promql.query : ''
      }).filter(q => q).join(', '),
      time_range: timeRange.value,
      step: 0,
      chart_template_id: parseInt(chartTemplateId.value) || null,
      webhook_ids: webhookIds.value.map(id => typeof id === 'string' ? parseInt(id) : id),
      card_title: cardTitle.value,
      card_template: cardTemplate.value,
      metric_label: metricLabel.value,
      custom_metric_label: customMetricLabel.value,
      unit: unit.value,
      button_text: buttonText.value,
      button_url: buttonURL.value,
      show_data_label: showDataLabel.value,
      enabled: 1,
      send_times: sendTimes.value.map(time => ({
        weekday: typeof time.weekday === 'string' ? parseInt(time.weekday) : time.weekday,
        send_time: time.send_time
      })),
      promql_chart_templates: selectedPromQLs.value.map(promqlId => ({
        promql_id: parseInt(promqlId),
        chart_template_id: parseInt(chartTemplateId.value) || null
      }))
    }
  }

  function addSendTime() {
    sendTimes.value.push({
      _id: Date.now() + Math.random(),
      weekday: 1,
      send_time: '09:00'
    })
  }

  function removeSendTime(index: number) {
    if (sendTimes.value.length <= 1) {
      showWarning('至少需要保留一个发送时间')
      return
    }
    if (confirm(`确定要删除此发送时间吗？`)) {
      sendTimes.value.splice(index, 1)
    }
  }

  return {
    name, sourceId, timeRange, chartTemplateId, cardTitle, cardTemplate,
    metricLabel, customMetricLabel, unit, webhookIds, selectedPromQLs,
    buttonText, buttonURL, showDataLabel, sendTimes, promqlConfigs,
    resetForm, loadTask, validate, buildPayload, addSendTime, removeSendTime
  }
}
