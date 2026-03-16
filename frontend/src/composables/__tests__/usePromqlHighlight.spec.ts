import { describe, it, expect } from 'vitest'
import { usePromqlHighlight } from '../usePromqlHighlight'

describe('usePromqlHighlight', () => {
  const { highlightPromQL } = usePromqlHighlight()

  it('空字符串返回空', () => {
    expect(highlightPromQL('')).toBe('')
  })

  it('高亮关键字', () => {
    const result = highlightPromQL('sum by (pod)')
    expect(result).toContain('<span class="keyword">sum</span>')
    expect(result).toContain('<span class="keyword">by</span>')
  })

  it('高亮标签表达式', () => {
    const result = highlightPromQL('up{job="test"}')
    expect(result).toContain('<span class="label">')
  })

  it('高亮数字', () => {
    const result = highlightPromQL('rate(metric[5m])')
    expect(result).toContain('<span class="number">')
  })

  it('转义 HTML 特殊字符', () => {
    const result = highlightPromQL('a > b & c < d')
    expect(result).toContain('&gt;')
    expect(result).toContain('&amp;')
    expect(result).toContain('&lt;')
  })

  it('高亮时间单位', () => {
    const result = highlightPromQL('rate(metric[30s])')
    expect(result).toContain('<span class="unit">s</span>')
  })
})
