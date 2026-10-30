import type { ErrorCode } from '@/api/errorCodes'

import enUS from './locales/en-US/errors.json'
import zhCN from './locales/zh-CN/errors.json'

// 编译期保证两种语言的错误码字典完整：后端新增 code 后未补翻译会在这里报错。
const zhCoverage: Record<ErrorCode, string> = zhCN.codes
const enCoverage: Record<ErrorCode, string> = enUS.codes

export { enCoverage, zhCoverage }
