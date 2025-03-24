// API 请求工具

// 调试模式
const DEBUG = false;

// 基础请求函数，自动添加认证头
export async function fetchWithAuth(url: string, options: RequestInit = {}) {
  const token = localStorage.getItem('token')
  
  if (DEBUG) {
    console.log(`API Request: ${options.method || 'GET'} ${url}`, { 
      hasToken: !!token, 
      body: options.body 
    })
  }
  
  // 合并默认头和用户提供的头
  const headers = {
    'Authorization': token ? `Bearer ${token}` : '',
    'Content-Type': 'application/json',
    ...options.headers
  }
  
  try {
    const response = await fetch(url, {
      ...options,
      headers
    })
    
    if (DEBUG) {
      console.log(`API Response: ${response.status} ${response.statusText}`)
    }
    
    // 处理 401 未授权错误
    if (response.status === 401) {
      console.warn('Authentication failed, clearing credentials')
      // 清除本地存储的认证信息
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      
      // 重定向到登录页
      window.location.href = '/login'
      throw new Error('认证已过期，请重新登录')
    }
    
    return response
  } catch (error) {
    console.error('API 请求错误:', error)
    throw error
  }
}

// GET 请求
export async function get(url: string) {
  const response = await fetchWithAuth(url)
  const data = await response.json()
  if (DEBUG) console.log('GET response data:', data)
  return data
}

// POST 请求
export async function post(url: string, data: any) {
  const response = await fetchWithAuth(url, {
    method: 'POST',
    body: JSON.stringify(data)
  })
  const responseData = await response.json()
  if (DEBUG) console.log('POST response data:', responseData)
  return responseData
}

// PUT 请求
export async function put(url: string, data: any) {
  const response = await fetchWithAuth(url, {
    method: 'PUT',
    body: JSON.stringify(data)
  })
  const responseData = await response.json()
  if (DEBUG) console.log('PUT response data:', responseData)
  return responseData
}

// DELETE 请求
export async function del(url: string) {
  const response = await fetchWithAuth(url, {
    method: 'DELETE'
  })
  const data = await response.json()
  if (DEBUG) console.log('DELETE response data:', data)
  return data
} 