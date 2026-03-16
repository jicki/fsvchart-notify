import { ref, computed } from 'vue'

export function useExpandable() {
  const expandedIds = ref<number[]>([])

  function toggle(id: number) {
    const index = expandedIds.value.indexOf(id)
    if (index === -1) {
      expandedIds.value.push(id)
    } else {
      expandedIds.value.splice(index, 1)
    }
  }

  function isExpanded(id: number): boolean {
    return expandedIds.value.includes(id)
  }

  function toggleAll(allIds: number[]) {
    if (expandedIds.value.length === allIds.length) {
      expandedIds.value = []
    } else {
      expandedIds.value = [...allIds]
    }
  }

  const isAllExpanded = computed(() => false) // 需要外部传入总数来判断

  function isAllExpandedFor(total: number): boolean {
    return expandedIds.value.length === total && total > 0
  }

  function expand(id: number) {
    if (!expandedIds.value.includes(id)) {
      expandedIds.value.push(id)
    }
  }

  function collapse(id: number) {
    const index = expandedIds.value.indexOf(id)
    if (index !== -1) {
      expandedIds.value.splice(index, 1)
    }
  }

  return {
    expandedIds,
    toggle,
    isExpanded,
    toggleAll,
    isAllExpanded,
    isAllExpandedFor,
    expand,
    collapse
  }
}
