import { describe, it, expect } from 'vitest'
import { formatTimeRange, getWeekdayText, formatDate } from '../formatters'

describe('formatTimeRange', () => {
  it('格式化小时', () => {
    expect(formatTimeRange('7h')).toBe('7小时')
    expect(formatTimeRange('24H')).toBe('24小时')
  })

  it('格式化天', () => {
    expect(formatTimeRange('7d')).toBe('7天')
    expect(formatTimeRange('30D')).toBe('30天')
  })

  it('格式化分钟', () => {
    expect(formatTimeRange('30m')).toBe('30分钟')
  })

  it('格式化月', () => {
    expect(formatTimeRange('3M')).toBe('3月')
  })

  it('空值返回未设置', () => {
    expect(formatTimeRange('')).toBe('未设置')
  })

  it('无法解析的值原样返回', () => {
    expect(formatTimeRange('invalid')).toBe('invalid')
  })
})

describe('getWeekdayText', () => {
  it('返回正确的星期几', () => {
    expect(getWeekdayText(1)).toBe('周一')
    expect(getWeekdayText(5)).toBe('周五')
    expect(getWeekdayText(7)).toBe('周日')
  })

  it('接受字符串参数', () => {
    expect(getWeekdayText('3')).toBe('周三')
  })

  it('超出范围返回未知', () => {
    expect(getWeekdayText(0)).toBe('未知')
    expect(getWeekdayText(8)).toBe('未知')
  })
})

describe('formatDate', () => {
  it('空字符串返回空', () => {
    expect(formatDate('')).toBe('')
  })

  it('格式化有效日期', () => {
    const result = formatDate('2024-01-15T10:30:00Z')
    expect(result).toBeTruthy()
    expect(typeof result).toBe('string')
  })
})
