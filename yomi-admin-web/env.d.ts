/// <reference types="vite/client" />

/** 由 vite define 注入：package.json 中的应用版本号 */
declare const __APP_VERSION__: string

interface ImportMetaEnv {
  /** 接口基础路径，如 /api */
  readonly VITE_API_BASE_URL: string
  /** 是否启用内置 mock：'true' | 'false' */
  readonly VITE_USE_MOCK: 'true' | 'false'
  /** 后端服务地址（仅 dev，vite 代理目标） */
  readonly VITE_PROXY_TARGET: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
