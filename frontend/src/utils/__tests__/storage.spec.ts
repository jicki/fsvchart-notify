import { describe, it, expect, beforeEach, vi } from 'vitest'
import { getToken, setToken, removeToken, getUser, setUser, removeUser, clearAuth } from '../storage'

// mock localStorage
const store: Record<string, string> = {}
const localStorageMock = {
  getItem: vi.fn((key: string) => store[key] ?? null),
  setItem: vi.fn((key: string, value: string) => { store[key] = value }),
  removeItem: vi.fn((key: string) => { delete store[key] }),
  clear: vi.fn(() => { Object.keys(store).forEach(k => delete store[k]) })
}

vi.stubGlobal('localStorage', localStorageMock)

describe('storage', () => {
  beforeEach(() => {
    Object.keys(store).forEach(k => delete store[k])
    vi.clearAllMocks()
  })

  describe('token 操作', () => {
    it('设置和获取 token', () => {
      expect(getToken()).toBeNull()
      setToken('test-token')
      expect(getToken()).toBe('test-token')
    })

    it('移除 token', () => {
      setToken('test-token')
      removeToken()
      expect(getToken()).toBeNull()
    })
  })

  describe('user 操作', () => {
    const testUser = { username: 'admin', displayName: '管理员' }

    it('设置和获取 user', () => {
      expect(getUser()).toBeNull()
      setUser(testUser)
      expect(getUser()).toEqual(testUser)
    })

    it('移除 user', () => {
      setUser(testUser)
      removeUser()
      expect(getUser()).toBeNull()
    })

    it('解析无效 JSON 返回 null', () => {
      store['user'] = 'invalid-json'
      expect(getUser()).toBeNull()
    })
  })

  describe('clearAuth', () => {
    it('清除 token 和 user', () => {
      setToken('test-token')
      setUser({ username: 'admin' })
      clearAuth()
      expect(getToken()).toBeNull()
      expect(getUser()).toBeNull()
    })
  })
})
