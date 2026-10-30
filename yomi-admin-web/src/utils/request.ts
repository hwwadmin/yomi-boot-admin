import axios from 'axios'
import type { AxiosRequestConfig, AxiosResponse } from 'axios'

import type { ApiResponse, PageData } from '@/api/types'
import { useUserStore } from '@/stores/user'
import {
  bizApiError,
  invalidResponseError,
  reportApiError,
  SESSION_EXPIRED_CODES,
  transportApiError,
} from '@/utils/apiError'
import { expireSession } from '@/utils/auth'

/** yomi-boot 默认协议的成功码 */
const SUCCESS_CODE = 'OK'

const service = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 15000,
})

/** POST 防重复提交窗口：窗口内相同「方法+地址+请求体」的请求共享同一次结果 */
const POST_DEDUP_WINDOW_MS = 500

interface PostRecord {
  promise: Promise<AxiosResponse<ApiResponse<unknown>>>
  /** 请求是否已落定；进行中的请求始终共享 */
  state: { settled: boolean; settledAt: number }
}

const postRecords = new Map<string, PostRecord>()

/** POST 去重键：方法 + 地址 + 请求体（同址同参才视为重复，不同实体的并发提交互不影响） */
function postKey(config: AxiosRequestConfig): string {
  const data = config.data
  const body = data == null ? '' : typeof data === 'string' ? data : JSON.stringify(data)
  return `POST ${config.url ?? ''} ${body}`
}

/**
 * 发送请求；POST 默认防重复提交：
 * 相同键的请求在进行中或落定后窗口内只发一次，重复调用共享同一次结果，
 * 双击 / 回车重复触发不再产生第二个请求；窗口过后的再次调用照常发起。
 *
 * 不用 VueUse 防抖的原因：trailing 防抖给每个请求加固定延迟、窗口外的重复点击仍会
 * 重发，且会把窗口内两次「不同参数」的 POST 折叠成最后一次（丢失首次意图）；
 * 这里的窗口跟随请求本身（进行中 + 落定后短窗复用），零额外延迟且不丢意图。
 * 输入触发的请求（如搜索）仍应在调用点用 useDebounceFn 防抖。
 */
function send<T>(config: AxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>> {
  if ((config.method ?? '').toLowerCase() !== 'post') {
    return service.request<ApiResponse<T>>(config)
  }

  const key = postKey(config)
  const now = Date.now()

  // 惰性清理窗口外的历史记录
  for (const [k, record] of postRecords) {
    if (record.state.settled && now - record.state.settledAt >= POST_DEDUP_WINDOW_MS) {
      postRecords.delete(k)
    }
  }

  const existing = postRecords.get(key)
  if (existing && (!existing.state.settled || now - existing.state.settledAt < POST_DEDUP_WINDOW_MS)) {
    return existing.promise as Promise<AxiosResponse<ApiResponse<T>>>
  }

  const state = { settled: false, settledAt: 0 }
  const promise = service.request<ApiResponse<unknown>>(config).finally(() => {
    state.settled = true
    state.settledAt = Date.now()
  })
  postRecords.set(key, { promise, state })
  return promise as Promise<AxiosResponse<ApiResponse<T>>>
}

function isBinaryResponse(config: AxiosRequestConfig): boolean {
  return config.responseType === 'blob' || config.responseType === 'arraybuffer'
}

function isApiResponse(value: unknown): value is ApiResponse {
  return (
    typeof value === 'object' &&
    value !== null &&
    typeof (value as { code?: unknown }).code === 'string'
  )
}

service.interceptors.request.use((config) => {
  const userStore = useUserStore()
  if (userStore.token) {
    config.headers.Authorization = `Bearer ${userStore.token}`
  }
  return config
})

service.interceptors.response.use(
  (response) => {
    // 文件流等非包络响应由调用方自行处理
    if (isBinaryResponse(response.config)) {
      return response
    }

    const body: unknown = response.data
    if (!isApiResponse(body)) {
      const error = invalidResponseError(response.config, response.status)
      reportApiError(error, { silent: response.config.silent })
      return Promise.reject(error)
    }

    if (body.code === SUCCESS_CODE) {
      return response
    }

    const error = bizApiError(body.code, body.msg ?? '', response.config, response.status)
    if (SESSION_EXPIRED_CODES.has(body.code)) {
      reportApiError(error, { silent: true })
      expireSession()
      return Promise.reject(error)
    }

    reportApiError(error, { silent: response.config.silent })
    return Promise.reject(error)
  },
  (reason: unknown) => {
    const error = transportApiError(reason)
    const silent = (reason as { config?: AxiosRequestConfig } | undefined)?.config?.silent

    if (error.httpStatus === 401) {
      reportApiError(error, { silent: true })
      expireSession()
      return Promise.reject(error)
    }

    reportApiError(error, { silent })
    return Promise.reject(error)
  },
)

/**
 * 业务请求：响应拦截器保证只有 code === 'OK' 才会正常返回，这里直接解出 data。
 * POST 默认防重复提交（见 send），业务错误 / Token 失效 / 网络错误一律走 reject，
 * 调用方 catch 后无需重复提示。
 */
export async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const response = await send<T>(config)
  return response.data.data
}

/**
 * 分页请求：yomi-boot 默认协议把分页载荷放在 data，形如 { items: [...], total }。
 * 统一映射成前端分页组件使用的 { list, total }。
 */
export async function requestPage<T>(config: AxiosRequestConfig): Promise<{ list: T[]; total: number }> {
  const response = await send<PageData<T>>(config)
  return {
    list: response.data.data?.items ?? [],
    total: response.data.data?.total ?? 0,
  }
}
