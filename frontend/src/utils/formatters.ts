// 格式化时间范围
export function formatTimeRange(timeRange: string): string {
  if (!timeRange) return '未设置'

  const match = timeRange.match(/^(\d+)([hHdDmM])$/)
  if (!match) return timeRange

  const [, value, unit] = match
  const unitMap: Record<string, string> = {
    h: '小时', H: '小时',
    d: '天', D: '天',
    m: '分钟', M: '月'
  }

  return `${value}${unitMap[unit] || unit}`
}

// 获取星期几的文字描述
export function getWeekdayText(weekday: number | string): string {
  const weekdays = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
  const num = typeof weekday === 'string' ? parseInt(weekday) : weekday
  return weekdays[num - 1] || '未知'
}

// 格式化日期
export function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleString()
}
