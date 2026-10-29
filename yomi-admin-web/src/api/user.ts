import { request, requestPage } from '@/utils/request'
import type { CreateUserPayload, PageQuery, PageResult, UserItem, UserPayload } from './types'

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

/** 新增用户 */
export function createUser(data: CreateUserPayload) {
  return request<null>({ url: '/system/user/create', method: 'post', data })
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
  return request<null>({ url: '/system/user/reset-password', method: 'post', data: { id, newPassword } })
}
