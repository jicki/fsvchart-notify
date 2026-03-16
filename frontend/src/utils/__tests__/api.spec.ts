import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { fetchWithAuth, get, post, put, del } from '../api'
import * as storage from '../storage'

describe('api', () => {
  const mockResponse = (data: unknown, status = 200) => ({
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(data)
  })

  beforeEach(() => {
    vi.spyOn(storage, 'getToken').mockReturnValue('test-token')
    vi.spyOn(storage, 'clearAuth').mockImplementation(() => {})
    // 重置 window.location
    Object.defineProperty(window, 'location', {
      value: { href: '' },
      writable: true
    })
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('fetchWithAuth', () => {
    it('添加 Authorization 头', async () => {
      const mockFetch = vi.fn().mockResolvedValue(mockResponse({}))
      vi.stubGlobal('fetch', mockFetch)

      await fetchWithAuth('/api/test')

      expect(mockFetch).toHaveBeenCalledWith('/api/test', expect.objectContaining({
        headers: expect.objectContaining({
          'Authorization': 'Bearer test-token',
          'Content-Type': 'application/json'
        })
      }))
    })

    it('401 时清除认证并跳转登录', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue(mockResponse({}, 401)))

      await expect(fetchWithAuth('/api/test')).rejects.toThrow('认证已过期')
      expect(storage.clearAuth).toHaveBeenCalled()
    })

    it('无 token 时不添加 Authorization 头', async () => {
      vi.spyOn(storage, 'getToken').mockReturnValue(null)
      const mockFetch = vi.fn().mockResolvedValue(mockResponse({}))
      vi.stubGlobal('fetch', mockFetch)

      await fetchWithAuth('/api/test')

      const headers = mockFetch.mock.calls[0][1].headers
      expect(headers['Authorization']).toBeUndefined()
    })
  })

  describe('get', () => {
    it('返回解析后的 JSON', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue(mockResponse({ id: 1, name: 'test' })))

      const result = await get<{ id: number; name: string }>('/api/items')
      expect(result).toEqual({ id: 1, name: 'test' })
    })
  })

  describe('post', () => {
    it('发送 POST 请求并带上 body', async () => {
      const mockFetch = vi.fn().mockResolvedValue(mockResponse({ id: 1 }))
      vi.stubGlobal('fetch', mockFetch)

      await post('/api/items', { name: 'new' })

      expect(mockFetch).toHaveBeenCalledWith('/api/items', expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ name: 'new' })
      }))
    })
  })

  describe('put', () => {
    it('发送 PUT 请求', async () => {
      const mockFetch = vi.fn().mockResolvedValue(mockResponse({}))
      vi.stubGlobal('fetch', mockFetch)

      await put('/api/items/1', { name: 'updated' })

      expect(mockFetch).toHaveBeenCalledWith('/api/items/1', expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify({ name: 'updated' })
      }))
    })
  })

  describe('del', () => {
    it('发送 DELETE 请求', async () => {
      const mockFetch = vi.fn().mockResolvedValue(mockResponse({}))
      vi.stubGlobal('fetch', mockFetch)

      await del('/api/items/1')

      expect(mockFetch).toHaveBeenCalledWith('/api/items/1', expect.objectContaining({
        method: 'DELETE'
      }))
    })
  })
})
