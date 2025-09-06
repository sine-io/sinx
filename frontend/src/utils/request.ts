import axios from 'axios'
import type { AxiosError, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { getToken, clearToken } from './auth'

const baseURL = import.meta.env.VITE_API_BASE_URL || '/api'

const instance = axios.create({
  baseURL,
  timeout: 10000,
})

instance.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = getToken()
  if (token) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

instance.interceptors.response.use(
  (res: AxiosResponse<ApiResponse<any>>) => {
    const data = res.data
    // Unified response structure: { code, message, data }
    if (data && typeof data.code !== 'undefined' && data.code !== 0) {
      // 10003 unauthorized
      if (data.code === 10003) {
        clearToken()
        window.location.href = '/login'
      }
      return Promise.reject(new Error(data.message || 'Request Error'))
    }
  // Cast to any because we unwrap the unified envelope
  return data as any
  },
  (err: AxiosError) => {
    return Promise.reject(err)
  }
)

export default instance
