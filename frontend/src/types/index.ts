// 数据源
export interface MetricsSource {
  id: number
  name: string
  url: string
  type?: string
}

// 飞书 Webhook
export interface FeishuWebhook {
  id: number
  name: string
  url: string
}

// 图表模板
export interface ChartTemplate {
  id: number
  name: string
  chart_type: string
  template?: string
}

// PromQL 查询
export interface PromQL {
  id: number
  name: string
  description: string
  query: string
  category: string
  created_at: string
  updated_at: string
}

// PromQL 配置（每个 PromQL 查询的独立配置）
export interface PromQLConfig {
  promql_id: number
  promql_name?: string
  unit: string
  metric_label: string
  custom_metric_label: string
  initial_unit: string
  display_order: number
  display_mode: 'chart' | 'text' | 'both'
  chart_template_id?: number | null
}

// 发送时间
export interface SendTime {
  _id?: number
  weekday: number
  send_time: string
}

// 推送任务
export interface PushTask {
  id: number
  name: string
  source_id: number
  promql_ids: number[]
  promql_configs: PromQLConfig[]
  query: string
  time_range: string
  step: number
  chart_template_id: number | null
  webhook_ids: number[]
  bound_webhooks: FeishuWebhook[]
  card_title: string
  card_template: string
  metric_label: string
  custom_metric_label: string
  unit: string
  button_text: string
  button_url: string
  show_data_label: boolean
  enabled: boolean | number
  send_times: SendTime[]
  promql_names?: string[]
  schedule_interval?: number
}

// 推送任务表单数据
export interface PushTaskFormData {
  name: string
  source_id: string
  promql_ids: string[]
  promql_configs: Record<number, {
    unit: string
    metric_label: string
    custom_metric_label: string
    initial_unit: string
    display_order: number
    display_mode: string
  }>
  time_range: string
  chart_template_id: string
  card_title: string
  card_template: string
  metric_label: string
  custom_metric_label: string
  unit: string
  webhook_ids: number[]
  button_text: string
  button_url: string
  show_data_label: boolean
  send_times: SendTime[]
}

// 发送记录
export interface SendRecord {
  id: number
  timestamp: string
  status: string
  message: string
  webhook: string
  task_name: string
  query?: string
  time_range?: string
  button_text?: string
  button_url?: string
}

// 用户信息
export interface UserInfo {
  username: string
  displayName?: string
  display_name?: string
  email?: string
  role?: string
}

// API 响应
export interface ApiResponse<T> {
  code?: number
  data?: T
  message?: string
  error?: string
}

// 通知类型
export type NotificationType = 'success' | 'error' | 'warning' | 'info'

export interface Notification {
  id: number
  type: NotificationType
  message: string
  duration: number
}

// CRUD 实体基础接口
export interface CrudEntity {
  id: number
  [key: string]: unknown
}
