import { ref } from 'vue'
import { get, post, put, del } from '../utils/api'
import { useNotification } from './useNotification'
import type { CrudEntity } from '../types'

export function useCrudList<T extends CrudEntity>(endpoint: string, entityName: string) {
  const items = ref<T[]>([])
  const editingId = ref<number | null>(null)
  const loading = ref(false)
  const { showSuccess, showError, showWarning } = useNotification()

  async function fetchList() {
    try {
      const data = await get<T[]>(endpoint)
      if (Array.isArray(data)) {
        items.value = data as T[]
      } else {
        items.value = []
      }
    } catch (err) {
      console.error(`获取${entityName}失败:`, err)
      items.value = []
    }
  }

  async function addItem(body: Partial<T>): Promise<boolean> {
    try {
      loading.value = true
      await post(endpoint, body)
      await fetchList()
      showSuccess(`${entityName}添加成功`)
      return true
    } catch (err) {
      console.error(`添加${entityName}失败:`, err)
      showError(`添加${entityName}失败，请重试`)
      return false
    } finally {
      loading.value = false
    }
  }

  async function updateItem(id: number, body: Partial<T>): Promise<boolean> {
    try {
      loading.value = true
      await put(`${endpoint}/${id}`, body)
      editingId.value = null
      await fetchList()
      showSuccess(`${entityName}更新成功`)
      return true
    } catch (err) {
      console.error(`更新${entityName}失败:`, err)
      showError(`更新${entityName}失败，请重试`)
      return false
    } finally {
      loading.value = false
    }
  }

  async function deleteItem(id: number): Promise<boolean> {
    if (!confirm(`确认删除${entityName} ID=${id}？`)) return false

    try {
      loading.value = true
      await del(`${endpoint}/${id}`)
      await fetchList()
      showSuccess(`${entityName}删除成功`)
      return true
    } catch (err) {
      console.error(`删除${entityName}失败:`, err)
      showError(`删除${entityName}失败，请重试`)
      return false
    } finally {
      loading.value = false
    }
  }

  function startEdit(id: number) {
    editingId.value = id
  }

  function cancelEdit() {
    editingId.value = null
  }

  function validateRequired(fields: Record<string, string>): boolean {
    for (const [value, label] of Object.entries(fields)) {
      if (!value) {
        showWarning(`${label}不能为空`)
        return false
      }
    }
    return true
  }

  return {
    items,
    loading,
    editingId,
    fetchList,
    addItem,
    updateItem,
    deleteItem,
    startEdit,
    cancelEdit,
    validateRequired
  }
}
