import axios, { type AxiosRequestConfig, type InternalAxiosRequestConfig } from 'axios'
import {
  APP_VERSION,
  CLIENT_PLATFORM,
  ClientHeaders,
  getOrCreateDeviceId,
  newRequestId,
} from '@/constants/appVersion'

declare module 'axios' {
  export interface AxiosRequestConfig {
    silent?: boolean
  }
}

let getAccessToken: () => string = () => ''
let tryRefreshToken: () => Promise<boolean> = async () => false

export function setAuthTokenProvider(getToken: () => string, refresh: () => Promise<boolean>) {
  getAccessToken = getToken
  tryRefreshToken = refresh
}

const service = axios.create({
  baseURL: '/',
  timeout: 15000,
})

service.interceptors.request.use((config) => {
  config.headers = config.headers || {}
  config.headers[ClientHeaders.REQUEST_ID] = newRequestId()
  config.headers[ClientHeaders.CLIENT_PLATFORM] = CLIENT_PLATFORM
  config.headers[ClientHeaders.CLIENT_VERSION] = APP_VERSION
  config.headers[ClientHeaders.DEVICE_ID] = getOrCreateDeviceId()
  const token = getAccessToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

let refreshPromise: Promise<boolean> | null = null

function doRefresh(): Promise<boolean> {
  if (!refreshPromise) {
    refreshPromise = tryRefreshToken().finally(() => {
      refreshPromise = null
    })
  }
  return refreshPromise
}

type RequestConfig = InternalAxiosRequestConfig & {
  _retry?: boolean
  silent?: boolean
}

let errorNotifier: (message: string) => void = () => {}

export function setRequestErrorNotifier(notifier: (message: string) => void) {
  errorNotifier = notifier
}

function reportError(config: RequestConfig | undefined, message: string) {
  if (!config?.silent) errorNotifier(message)
}

let authLostHandler: (() => void) | null = null

export function setAuthLostHandler(handler: () => void) {
  authLostHandler = handler
}

export function notifyAuthLost() {
  authLostHandler?.()
}

service.interceptors.response.use(
  (resp) => {
    const data = resp.data
    const config = resp.config as RequestConfig
    if (data && typeof data === 'object' && 'code' in data) {
      if (data.code === 0) {
        return data
      }
      reportError(config, data.message || '操作失败')
      return Promise.reject(data)
    }
    return resp.data
  },
  async (err) => {
    const status = err?.response?.status
    const config = err?.config as RequestConfig

    if (status === 401 && config && !config._retry) {
      const url = config.url || ''
      if (!url.includes('/api/auth/login') && !url.includes('/api/auth/refresh')) {
        config._retry = true
        if (await doRefresh()) {
          config.headers = config.headers || {}
          config.headers.Authorization = `Bearer ${getAccessToken()}`
          return service(config)
        }
      }

      if (err?.response) {
        notifyAuthLost()
      }
      return Promise.reject(err)
    }

    if (status !== 401) {
      const msg = err?.response?.data?.message
      if (msg) reportError(config, msg)
      else if (!err?.response) reportError(config, '网络异常，请稍后重试')
      else reportError(config, '网络异常')
    }
    return Promise.reject(err)
  },
)

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

export function get<T>(url: string, params?: Record<string, unknown>, config?: AxiosRequestConfig) {
  return service.get<unknown, ApiResponse<T>>(url, { params, ...config })
}

export function post<T>(url: string, data?: unknown, config?: AxiosRequestConfig) {
  return service.post<unknown, ApiResponse<T>>(url, data, config)
}

export default service
