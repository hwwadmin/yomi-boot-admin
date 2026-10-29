import { request } from '@/utils/request'
import type { CaptchaChallenge, ChangePasswordPayload, LoginParams, LoginResult, UserInfo } from './types'

/** 登录图像验证码 */
export function getCaptcha() {
  return request<CaptchaChallenge>({ url: '/auth/captcha', method: 'get' })
}

/** 登录 */
export function login(data: LoginParams) {
  return request<LoginResult>({ url: '/auth/login', method: 'post', data })
}

/** 当前登录用户信息 */
export function getProfile() {
  return request<UserInfo>({ url: '/auth/profile', method: 'get' })
}

/** 退出登录 */
export function logout() {
  return request<null>({ url: '/auth/logout', method: 'post' })
}

/** 修改当前登录用户密码；成功后所有会话失效，需要重新登录 */
export function changePassword(data: ChangePasswordPayload) {
  return request<null>({ url: '/auth/password', method: 'post', data })
}
