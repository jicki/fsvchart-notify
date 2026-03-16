// API 请求工具
import { getToken, clearAuth } from './storage'

// 基础请求函数，自动添加认证头
export async function fetchWithAuth(url: string, options: RequestInit = {}): Promise<Response> {
  const token = getToken()

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...options.headers as Record<string, string>
  }

  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const response = await fetch(url, {
    ...options,
    headers
  })

  // 处理 401 未授权错误
  if (response.status === 401) {
    clearAuth()
    window.location.href = '/login'
    throw new Error('认证已过期，请重新登录')
  }

  return response
}

// GET 请求
export async function get<T = unknown>(url: string): Promise<T> {
  const response = await fetchWithAuth(url)
  return response.json() as Promise<T>
}

// POST 请求
export async function post<T = unknown>(url: string, data: unknown): Promise<T> {
  const response = await fetchWithAuth(url, {
    method: 'POST',
    body: JSON.stringify(data)
  })
  return response.json() as Promise<T>
}

// PUT 请求
export async function put<T = unknown>(url: string, data: unknown): Promise<T> {
  const response = await fetchWithAuth(url, {
    method: 'PUT',
    body: JSON.stringify(data)
  })
  return response.json() as Promise<T>
}

// DELETE 请求
export async function del<T = unknown>(url: string): Promise<T> {
  const response = await fetchWithAuth(url, {
    method: 'DELETE'
  })
  return response.json() as Promise<T>
}
