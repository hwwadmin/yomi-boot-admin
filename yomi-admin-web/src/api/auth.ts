import { request } from '@/utils/request'
import type {
  CaptchaChallenge,
  ChangePasswordPayload,
  LoginConfig,
  LoginParams,
  LoginResult,
  UpdateProfilePayload,
  UserInfo,
} from './types'

/** 登录页公开配置（验证码开关等） */
export function getLoginConfig() {
  return request<LoginConfig>({ url: '/auth/config', method: 'get' })
}

/** 登录图形验证码；silent 用于登录失败后的静默换图 */
export function getCaptcha(options: { silent?: boolean } = {}) {
  return request<CaptchaChallenge>({ url: '/auth/captcha', method: 'get', ...options })
}

/** 登录 */
export function login(data: LoginParams) {
  return request<LoginResult>({ url: '/auth/login', method: 'post', data })
}

/** 当前登录用户信息 */
export function getProfile() {
  return request<UserInfo>({ url: '/auth/profile', method: 'get' })
}

/** 更新当前登录用户资料，返回更新后的身份信息 */
export function updateProfile(data: UpdateProfilePayload) {
  return request<UserInfo>({ url: '/auth/profile', method: 'post', data })
}

/** 退出登录 */
export function logout() {
  return request<null>({ url: '/auth/logout', method: 'post' })
}

/** 修改当前登录用户密码；成功后所有会话失效，需要重新登录 */
export function changePassword(data: ChangePasswordPayload) {
  return request<null>({ url: '/auth/password', method: 'post', data })
}
