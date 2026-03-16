import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '../auth'

// mock localStorage
const store: Record<string, string> = {}
const localStorageMock = {
  getItem: vi.fn((key: string) => store[key] ?? null),
  setItem: vi.fn((key: string, value: string) => { store[key] = value }),
  removeItem: vi.fn((key: string) => { delete store[key] }),
  clear: vi.fn(() => { Object.keys(store).forEach(k => delete store[k]) })
}

vi.stubGlobal('localStorage', localStorageMock)

describe('auth store', () => {
  beforeEach(() => {
    Object.keys(store).forEach(k => delete store[k])
    vi.clearAllMocks()
    setActivePinia(createPinia())
  })

  it('初始状态未登录', () => {
    const authStore = useAuthStore()
    expect(authStore.isLoggedIn).toBe(false)
    expect(authStore.userDisplayName).toBe('')
  })

  it('login 设置 token 和用户信息', () => {
    const authStore = useAuthStore()
    authStore.login('test-token', { username: 'admin', displayName: '管理员' })

    expect(authStore.isLoggedIn).toBe(true)
    expect(authStore.userDisplayName).toBe('管理员')
    expect(authStore.token).toBe('test-token')
    expect(store['token']).toBe('test-token')
  })

  it('logout 清除状态', () => {
    const authStore = useAuthStore()
    authStore.login('test-token', { username: 'admin' })
    authStore.logout()

    expect(authStore.isLoggedIn).toBe(false)
    expect(authStore.token).toBeNull()
    expect(authStore.user).toBeNull()
    expect(store['token']).toBeUndefined()
  })

  it('checkAuth 从 localStorage 恢复状态', () => {
    store['token'] = 'saved-token'
    store['user'] = JSON.stringify({ username: 'admin', displayName: '管理员' })

    const authStore = useAuthStore()
    authStore.checkAuth()

    expect(authStore.isLoggedIn).toBe(true)
    expect(authStore.userDisplayName).toBe('管理员')
  })

  it('updateUser 更新部分用户信息', () => {
    const authStore = useAuthStore()
    authStore.login('token', { username: 'admin', displayName: '旧名称' })
    authStore.updateUser({ displayName: '新名称' })

    expect(authStore.userDisplayName).toBe('新名称')
    expect(authStore.user?.username).toBe('admin')
  })

  it('userDisplayName 回退到 username', () => {
    const authStore = useAuthStore()
    authStore.login('token', { username: 'admin' })

    expect(authStore.userDisplayName).toBe('admin')
  })
})
