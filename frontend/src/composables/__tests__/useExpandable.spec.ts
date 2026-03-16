import { describe, it, expect } from 'vitest'
import { useExpandable } from '../useExpandable'

describe('useExpandable', () => {
  it('初始状态无展开项', () => {
    const { expandedIds } = useExpandable()
    expect(expandedIds.value).toEqual([])
  })

  it('toggle 切换展开状态', () => {
    const { toggle, isExpanded } = useExpandable()

    toggle(1)
    expect(isExpanded(1)).toBe(true)

    toggle(1)
    expect(isExpanded(1)).toBe(false)
  })

  it('toggleAll 全部展开/收起', () => {
    const { toggleAll, isExpanded, isAllExpandedFor } = useExpandable()

    toggleAll([1, 2, 3])
    expect(isExpanded(1)).toBe(true)
    expect(isExpanded(2)).toBe(true)
    expect(isExpanded(3)).toBe(true)
    expect(isAllExpandedFor(3)).toBe(true)

    toggleAll([1, 2, 3])
    expect(isExpanded(1)).toBe(false)
    expect(isAllExpandedFor(3)).toBe(false)
  })

  it('expand 和 collapse', () => {
    const { expand, collapse, isExpanded } = useExpandable()

    expand(1)
    expect(isExpanded(1)).toBe(true)

    expand(1) // 重复展开不会出错
    expect(isExpanded(1)).toBe(true)

    collapse(1)
    expect(isExpanded(1)).toBe(false)

    collapse(1) // 重复收起不会出错
    expect(isExpanded(1)).toBe(false)
  })

  it('isAllExpandedFor 正确判断', () => {
    const { toggle, isAllExpandedFor } = useExpandable()

    expect(isAllExpandedFor(0)).toBe(false)

    toggle(1)
    toggle(2)
    expect(isAllExpandedFor(2)).toBe(true)
    expect(isAllExpandedFor(3)).toBe(false)
  })
})
