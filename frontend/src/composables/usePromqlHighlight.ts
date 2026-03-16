// PromQL 语法高亮
const KEYWORDS = [
  'sum', 'rate', 'irate', 'avg', 'max', 'min', 'count',
  'by', 'without', 'offset', 'bool', 'and', 'or', 'unless',
  'group', 'ignoring', 'on', 'topk', 'bottomk'
]

const HTML_ENTITIES: Record<string, string> = {
  '&': '&amp;',
  '<': '&lt;',
  '>': '&gt;'
}

export function usePromqlHighlight() {
  function highlightPromQL(query: string): string {
    if (!query) return ''

    // 转义 HTML 特殊字符
    let highlighted = query.replace(/[&<>]/g, char => HTML_ENTITIES[char] || char)

    // 高亮关键字
    KEYWORDS.forEach(keyword => {
      const regex = new RegExp(`\\b${keyword}\\b`, 'g')
      highlighted = highlighted.replace(regex, `<span class="keyword">${keyword}</span>`)
    })

    // 高亮标签和值
    highlighted = highlighted.replace(
      /(\{[^}]*\})/g,
      match => `<span class="label">${match}</span>`
    )

    // 高亮数字
    highlighted = highlighted.replace(
      /\b(\d+(\.\d+)?)\b/g,
      '<span class="number">$1</span>'
    )

    // 高亮时间单位
    highlighted = highlighted.replace(
      /\b(\d+)(s|m|h|d|w|y)\b/g,
      '<span class="number">$1</span><span class="unit">$2</span>'
    )

    return highlighted
  }

  return { highlightPromQL }
}
