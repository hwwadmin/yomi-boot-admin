import axios from 'axios'
import type { AxiosRequestConfig } from 'axios'

import router from '@/router'
import { useUserStore } from '@/stores/user'
import type { ApiResponse } from '@/api/types'

const SUCCESS_CODE = 200
const UNAUTHORIZED_CODE = 401

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
    if (body.code === UNAUTHORIZED_CODE) {
      useUserStore().reset()
      const current = router.currentRoute.value
      void router.replace({
        path: '/login',
        query: current.path === '/login' ? undefined : { redirect: current.fullPath },
      })
    }
    ElMessage.error(body.message || '请求失败')
    return Promise.reject(new Error(body.message || '请求失败'))
  },
  (error) => {
    ElMessage.error(error?.message || '网络异常，请稍后重试')
    return Promise.reject(error)
  },
)

/**
 * 业务请求：响应拦截器保证只有 code === 200 才会正常返回，这里直接解出 data。
 * 非 200（业务错误 / 401 / 网络错误）一律走 reject，调用方 catch 后无需重复提示。
 */
export async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const response = await service.request<ApiResponse<T>>(config)
  return response.data.data
}
