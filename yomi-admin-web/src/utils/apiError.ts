import type { AxiosError, AxiosRequestConfig } from 'axios'

import { i18n } from '@/i18n'

export type ApiErrorKind = 'biz' | 'http' | 'network' | 'timeout' | 'invalid-response'

interface ApiErrorOptions {
  code: string
  msg: string
  kind: ApiErrorKind
  httpStatus?: number
  method?: string
  url?: string
}

/**
 * 接口失败统一对象。
 * code 是前端唯一允许用于判断和本地化的字段；msg 是后端调试信息，只进 console。
 */
export class ApiError extends Error {
  readonly code: string
  readonly msg: string
  readonly kind: ApiErrorKind
  readonly httpStatus?: number
  readonly method?: string
  readonly url?: string

  constructor(options: ApiErrorOptions) {
    super(`[${options.code}] ${options.msg}`)
    this.name = 'ApiError'
    this.code = options.code
    this.msg = options.msg
    this.kind = options.kind
    this.httpStatus = options.httpStatus
    this.method = options.method
    this.url = options.url
  }
}

/** 登录态失效：由拦截器统一清会话并跳登录，不再弹通用错误 */
export const SESSION_EXPIRED_CODES = new Set([
  'TOKEN_REQUIRED',
  'TOKEN_VALIDATE_FAILED',
  'TOKEN_INVALID',
])

const TOAST_DEDUPE_WINDOW_MS = 1500
let lastToast = { code: '', at: 0 }

function requestMeta(config?: AxiosRequestConfig) {
  return {
    method: config?.method?.toUpperCase(),
    url: config?.url,
  }
}

export function bizApiError(
  code: string,
  msg: string,
  config?: AxiosRequestConfig,
  httpStatus?: number,
): ApiError {
  return new ApiError({ code, msg, kind: 'biz', httpStatus, ...requestMeta(config) })
}

export function invalidResponseError(
  config?: AxiosRequestConfig,
  httpStatus?: number,
): ApiError {
  return new ApiError({
    code: 'INVALID_RESPONSE',
    msg: 'unexpected response body',
    kind: 'invalid-response',
    httpStatus,
    ...requestMeta(config),
  })
}

export function transportApiError(error: unknown): ApiError {
  if (error instanceof ApiError) return error

  const axiosError = error as AxiosError | undefined
  const config = axiosError?.config
  const status = axiosError?.response?.status
  const msg = axiosError?.message ?? ''
  const meta = requestMeta(config)

  if (status) {
    return new ApiError({ code: `HTTP_${status}`, msg, kind: 'http', httpStatus: status, ...meta })
  }
  if (axiosError?.code === 'ECONNABORTED' || axiosError?.code === 'ETIMEDOUT') {
    return new ApiError({ code: 'TIMEOUT', msg, kind: 'timeout', ...meta })
  }
  return new ApiError({ code: 'NETWORK_ERROR', msg, kind: 'network', ...meta })
}

/** 按 code 取当前语言文案；未知 code 只给通用提示 */
export function resolveApiErrorMessage(error: ApiError): string {
  const { t, te } = i18n.global

  if (error.kind === 'biz') {
    const key = `errors.codes.${error.code}`
    if (te(key)) return t(key)
    return t('errors.unknown')
  }

  if (error.kind === 'http' && error.httpStatus) {
    const statusKey = `errors.http.${error.httpStatus}`
    if (te(statusKey)) return t(statusKey)
    if (error.httpStatus >= 500) return t('errors.http.server')
    if (error.httpStatus >= 400) return t('errors.http.client')
    return t('errors.unknown')
  }

  if (error.kind === 'timeout') return t('errors.network.timeout')
  if (error.kind === 'network') return t('errors.network.offline')
  if (error.kind === 'invalid-response') return t('errors.invalidResponse')
  return t('errors.unknown')
}

/**
 * 失败统一出口：msg 永远只写 console；界面文案只由 code 决定。
 * 同一 code 在短窗口内只弹一次，避免并发请求失败刷屏。
 */
export function reportApiError(error: ApiError, options: { silent?: boolean } = {}) {
  console.warn('[api] request failed', {
    code: error.code,
    msg: error.msg,
    method: error.method,
    url: error.url,
    status: error.httpStatus,
  })

  if (options.silent) return

  const now = Date.now()
  if (error.code === lastToast.code && now - lastToast.at < TOAST_DEDUPE_WINDOW_MS) return

  lastToast = { code: error.code, at: now }
  ElMessage.error(resolveApiErrorMessage(error))
}
