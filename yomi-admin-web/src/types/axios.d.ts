import 'axios'

declare module 'axios' {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- 必须与 axios 原声明保持相同的类型参数
  export interface AxiosRequestConfig<D = any> {
    /** 失败时不弹全局错误提示，由调用方自行处理 */
    silent?: boolean
  }
}
