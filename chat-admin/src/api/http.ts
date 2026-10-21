import axios, {
  isAxiosError,
  type AxiosRequestConfig,
  type InternalAxiosRequestConfig,
} from 'axios'
import { buildClientContextHeaders } from '@/utils/clientContext'

let authLostHandler: (() => void) | null = null
let getAccessToken: () => string = () => ''
let tryRefreshToken: () => Promise<boolean> = async () => false

export function setAuthLostHandler(handler: () => void) {
  authLostHandler = handler
}

export function setAuthTokenProvider(getToken: () => string, refresh: () => Promise<boolean>) {
  getAccessToken = getToken
  tryRefreshToken = refresh
}

interface ApiResponse<T = unknown> {
  code: number
  data: T
  message: string
}

const api = axios.create({
  baseURL: '/',
  timeout: 10000,
})

api.interceptors.request.use((config) => {
  config.headers = config.headers || {}
  const meta = buildClientContextHeaders()
  for (const [k, v] of Object.entries(meta)) {
    config.headers[k] = v
  }
  const accessToken = getAccessToken()
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`
  }
  return config
})

function formatRequestError(error: unknown): Error {
  if (isAxiosError(error)) {
    if (!error.response) {
      if (error.code === 'ECONNABORTED') {
        return new Error('请求超时，请稍后重试')
      }
      return new Error('网络连接失败，请检查网络后重试')
    }

    const { status, data } = error.response
    if (typeof data === 'object' && data !== null && 'message' in data) {
      const message = (data as { message?: unknown }).message
      if (typeof message === 'string' && message.trim()) {
        return new Error(message)
      }
    }

    switch (status) {
      case 400:
        return new Error('请求参数错误')
      case 403:
        return new Error('没有权限执行此操作')
      case 404:
        return new Error('请求的资源不存在')
      case 500:
        return new Error('服务器内部错误，请稍后重试')
      case 502:
        return new Error('服务暂时不可用，请稍后重试')
      case 503:
        return new Error('服务维护中，请稍后重试')
      case 504:
        return new Error('网关超时，请稍后重试')
      default:
        return new Error(`请求失败 (${status})`)
    }
  }

  if (error instanceof Error) return error
  return new Error('请求失败')
}

api.interceptors.response.use(
  (response) => {
    const { data } = response
    if (!data || typeof data !== 'object' || !('code' in data)) {
      return Promise.reject(new Error('响应格式错误'))
    }
    if (data.code !== 0) {
      return Promise.reject(new Error(data.message || '请求失败'))
    }
    return data
  },
  async (error) => {
    const status = error.response?.status
    const config = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    if (status === 401 && config && !config._retry) {
      const url = config.url || ''
      if (!url.includes('/api/admin/auth/login') && !url.includes('/api/admin/auth/refresh')) {
        config._retry = true
        if (await tryRefreshToken()) {
          config.headers = config.headers || {}
          config.headers.Authorization = `Bearer ${getAccessToken()}`
          return api(config)
        }
      }
      authLostHandler?.()
      return Promise.reject(formatRequestError(error))
    }

    return Promise.reject(formatRequestError(error))
  },
)

export const request = {
  get: <T = unknown>(url: string, params?: object, config?: AxiosRequestConfig) =>
    api.get<unknown, ApiResponse<T>>(url, { params, ...config }),
  post: <T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
    api.post<unknown, ApiResponse<T>>(url, data, config),
}

export function uploadForm<T = unknown>(url: string, form: FormData, timeout = 60000) {
  return request.post<T>(url, form, { timeout })
}

export { formatRequestError }
