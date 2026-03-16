import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getToken, setToken, getUser, setUser, clearAuth } from '../utils/storage'
import type { UserInfo } from '../types'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(getToken())
  const user = ref<UserInfo | null>(getUser())

  const isLoggedIn = computed(() => !!token.value)

  const userDisplayName = computed(() => {
    if (!user.value) return ''
    return user.value.displayName || user.value.username || ''
  })

  function login(newToken: string, newUser: UserInfo) {
    token.value = newToken
    user.value = newUser
    setToken(newToken)
    setUser(newUser)
  }

  function logout() {
    token.value = null
    user.value = null
    clearAuth()
  }

  function checkAuth() {
    token.value = getToken()
    user.value = getUser()
  }

  function updateUser(updatedUser: Partial<UserInfo>) {
    if (user.value) {
      user.value = { ...user.value, ...updatedUser }
      setUser(user.value)
    }
  }

  return {
    token,
    user,
    isLoggedIn,
    userDisplayName,
    login,
    logout,
    checkAuth,
    updateUser
  }
})
