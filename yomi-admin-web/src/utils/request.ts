import axios from 'axios'
import type { AxiosRequestConfig } from 'axios'

import { useUserStore } from '@/stores/user'
import { expireSession } from '@/utils/auth'
import type { ApiResponse } from '@/api/types'

/** yomi-boot 默认协议的成功码 */
const SUCCESS_CODE = 'OK'
/** Token 缺失 / 失效时后端返回的业务码 */
const UNAUTHORIZED_CODES = new Set(['TOKEN_REQUIRED', 'TOKEN_VALIDATE_FAILED', 'TOKEN_INVALID'])

const service = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 15000,
})

service.interceptors.request.use((config) => {
  const userStore = useUserStore()
  if (userStore.token) {
    config.headers.Authorization = `Bearer ${userStore.token}`
  }
  return config
})

service.interceptors.response.use(
  (response) => {
    const body = response.data as ApiResponse | undefined
    // 非统一响应体（如文件流、第三方接口）直接放行，由调用方自行处理
    if (typeof body !== 'object' || body === null || !('code' in body)) {
      return response
    }
    if (body.code === SUCCESS_CODE) {
      return response
    }
    if (UNAUTHORIZED_CODES.has(body.code)) {
      expireSession()
      return Promise.reject(new Error(body.code))
    }
    ElMessage.error(body.msg || '请求失败')
    return Promise.reject(new Error(body.msg || '请求失败'))
  },
  (error) => {
    // 后端默认用业务码表达鉴权失败，这里兼容通过 HTTP 401 表达的情况
    if (error?.response?.status === 401) {
      expireSession()
      return Promise.reject(error)
    }
    ElMessage.error(error?.message || '网络异常，请稍后重试')
    return Promise.reject(error)
  },
)

/**
 * 业务请求：响应拦截器保证只有 code === 'OK' 才会正常返回，这里直接解出 data。
 * 业务错误 / Token 失效 / 网络错误一律走 reject，调用方 catch 后无需重复提示。
 */
export async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const response = await service.request<ApiResponse<T>>(config)
  return response.data.data
}

/**
 * 分页请求：yomi-boot 默认协议把列表放在 data、总数放在顶层 total。
 * 统一映射成前端分页组件使用的 { list, total }。
 */
export async function requestPage<T>(config: AxiosRequestConfig): Promise<{ list: T[]; total: number }> {
  const response = await service.request<ApiResponse<T[]>>(config)
  return {
    list: response.data.data ?? [],
    total: response.data.total ?? 0,
  }
}
