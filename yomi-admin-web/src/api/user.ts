import { request, requestPage } from '@/utils/request'
import { idempotencyHeaders } from '@/utils/idempotency'
import type {
  CreateUserPayload,
  PageQuery,
  PageResult,
  UserItem,
  UserPayload,
  UserSessionItem,
} from './types'

export interface UserQuery extends PageQuery {
  keyword?: string
}

/** 分页查询用户 */
export async function getUserPage(params: UserQuery): Promise<PageResult<UserItem>> {
  const { list, total } = await requestPage<UserItem>({
    url: '/system/user',
    method: 'get',
    params,
  })
  return { list, total, page: params.page, pageSize: params.pageSize }
}

/** 新增用户；idempotencyKey 由调用方在一次提交意图内保持稳定 */
export function createUser(data: CreateUserPayload, idempotencyKey: string) {
  return request<null>({
    url: '/system/user/create',
    method: 'post',
    data,
    headers: idempotencyHeaders(idempotencyKey),
  })
}

/** 编辑用户 */
export function updateUser(id: string, data: Partial<UserPayload>) {
  return request<null>({ url: '/system/user/update', method: 'post', data: { id, ...data } })
}

/** 删除用户 */
export function deleteUser(id: string) {
  return request<null>({ url: '/system/user/delete', method: 'post', data: { id } })
}

/** 重置用户密码（仅超级管理员）；成功后该用户全部会话失效 */
export function resetUserPassword(id: string, newPassword: string) {
  return request<null>({
    url: '/system/user/reset-password',
    method: 'post',
    data: { id, newPassword },
  })
}

/** 查询用户在线会话；sessionId 是令牌摘要，不是明文令牌 */
export function getUserSessions(userId: string) {
  return request<UserSessionItem[]>({
    url: '/system/user/session',
    method: 'get',
    params: { userId },
  })
}

/** 强制下线用户的单个会话 */
export function revokeUserSession(userId: string, sessionId: string) {
  return request<null>({
    url: '/system/user/session/revoke',
    method: 'post',
    data: { userId, sessionId },
  })
}

/** 强制下线用户的全部会话 */
export function revokeUserSessions(userId: string) {
  return request<null>({
    url: '/system/user/session/revoke-all',
    method: 'post',
    data: { userId },
  })
}
