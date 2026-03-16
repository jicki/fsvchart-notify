import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { usePushTaskForm } from '../usePushTaskForm'
import type { PushTask } from '../../types'

// mock useNotification
vi.mock('../useNotification', () => ({
  useNotification: () => ({
    showWarning: vi.fn(),
    showSuccess: vi.fn(),
    showError: vi.fn(),
    showInfo: vi.fn(),
    notifications: { value: [] },
    removeNotification: vi.fn()
  })
}))

describe('usePushTaskForm', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('初始化默认值', () => {
    const form = usePushTaskForm()
    expect(form.name.value).toBe('')
    expect(form.sourceId.value).toBe('')
    expect(form.timeRange.value).toBe('7d')
    expect(form.sendTimes.value.length).toBe(1)
  })

  it('resetForm 恢复默认值', () => {
    const form = usePushTaskForm()
    form.name.value = '测试任务'
    form.sourceId.value = '1'
    form.selectedPromQLs.value = ['1', '2']

    form.resetForm()

    expect(form.name.value).toBe('')
    expect(form.sourceId.value).toBe('')
    expect(form.selectedPromQLs.value).toEqual([])
  })

  it('loadTask 加载任务数据', () => {
    const form = usePushTaskForm()
    const task: PushTask = {
      id: 1,
      name: '测试任务',
      source_id: 2,
      promql_ids: [10, 20],
      promql_configs: [
        { promql_id: 10, unit: 'MB', metric_label: 'pod', custom_metric_label: '', initial_unit: 'bytes', display_order: 1, display_mode: 'chart' }
      ],
      query: 'up',
      time_range: '24h',
      step: 0,
      chart_template_id: 5,
      webhook_ids: [1, 2],
      bound_webhooks: [],
      card_title: '标题',
      card_template: 'green',
      metric_label: 'pod',
      custom_metric_label: '',
      unit: 'MB',
      button_text: '查看',
      button_url: 'http://example.com',
      show_data_label: true,
      enabled: true,
      send_times: [
        { weekday: 1, send_time: '10:00' },
        { weekday: 5, send_time: '18:00' }
      ]
    }

    form.loadTask(task)

    expect(form.name.value).toBe('测试任务')
    expect(form.sourceId.value).toBe('2')
    expect(form.timeRange.value).toBe('24h')
    expect(form.chartTemplateId.value).toBe('5')
    expect(form.selectedPromQLs.value).toEqual(['10', '20'])
    expect(form.webhookIds.value).toEqual([1, 2])
    expect(form.sendTimes.value.length).toBe(2)
    expect(form.buttonText.value).toBe('查看')
    expect(form.showDataLabel.value).toBe(true)
    expect(form.promqlConfigs.value[10]?.unit).toBe('MB')
  })

  it('validate 缺少名称时失败', () => {
    const form = usePushTaskForm()
    form.sourceId.value = '1'
    form.selectedPromQLs.value = ['1']
    form.chartTemplateId.value = '1'

    expect(form.validate()).toBe(false)
  })

  it('validate 填写完整时通过', () => {
    const form = usePushTaskForm()
    form.name.value = '任务'
    form.sourceId.value = '1'
    form.selectedPromQLs.value = ['1']
    form.chartTemplateId.value = '1'

    expect(form.validate()).toBe(true)
  })

  it('buildPayload 构建正确的 payload', () => {
    const form = usePushTaskForm()
    form.name.value = '任务'
    form.sourceId.value = '2'
    form.selectedPromQLs.value = ['10']
    form.chartTemplateId.value = '5'
    form.timeRange.value = '7d'

    const payload = form.buildPayload([{ id: 10, query: 'up{job="test"}' }])

    expect(payload.name).toBe('任务')
    expect(payload.source_id).toBe(2)
    expect(payload.promql_ids).toEqual([10])
    expect(payload.time_range).toBe('7d')
    expect(payload.chart_template_id).toBe(5)
    expect(payload.query).toBe('up{job="test"}')
  })

  it('addSendTime 添加发送时间', () => {
    const form = usePushTaskForm()
    const initialCount = form.sendTimes.value.length
    form.addSendTime()
    expect(form.sendTimes.value.length).toBe(initialCount + 1)
  })
})
