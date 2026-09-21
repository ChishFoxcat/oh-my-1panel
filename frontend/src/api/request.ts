import type { AxiosRequestConfig } from 'axios'
import type { Envelope } from './types'
import axios from 'axios'

/** 后端会话 Cookie 名与 CSRF 头，需与 backend/constant 保持一致。 */
const CSRF_COOKIE = 'ompcsrftoken'
const CSRF_HEADER = 'X-CSRF-Token'

/** ApiError 携带 HTTP 状态码与面板错误键，便于页面做分支处理。 */
export class ApiError extends Error {
  readonly status: number
  readonly detail: string

  constructor(message: string, status: number, detail = '') {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.detail = detail
  }
}

const http = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

// 写操作统一携带 CSRF 令牌（读 Cookie 后放进请求头）
http.interceptors.request.use((config) => {
  const method = (config.method ?? 'get').toUpperCase()
  if (!['GET', 'HEAD', 'OPTIONS'].includes(method)) {
    const token = readCookie(CSRF_COOKIE)
    if (token) {
      config.headers.set(CSRF_HEADER, token)
    }
  }
  return config
})

http.interceptors.response.use(
  response => response,
  (error) => {
    const envelope = error?.response?.data as Partial<Envelope<unknown>> | undefined
    const status = error?.response?.status ?? 0
    const message = envelope?.message || error?.message || '请求失败，请稍后重试！'
    return Promise.reject(new ApiError(message, status, envelope?.detail ?? ''))
  },
)

/** 发起请求并剥离信封，返回业务数据。 */
export async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const response = await http.request<Envelope<T>>(config)
  return response.data.data
}

function readCookie(name: string): string | undefined {
  const parts = `; ${document.cookie}`.split(`; ${name}=`)
  if (parts.length !== 2) {
    return undefined
  }
  return parts.pop()?.split(';').shift()
}
