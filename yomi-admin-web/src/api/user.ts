import { request } from '@/utils/request'
import type { PageQuery, PageResult, UserItem, UserPayload } from './types'

export interface UserQuery extends PageQuery {
  keyword?: string
}

/** 分页查询用户 */
export function getUserPage(params: UserQuery) {
  return request<PageResult<UserItem>>({ url: '/system/users', method: 'get', params })
}

/** 新增用户 */
export function createUser(data: UserPayload) {
  return request<null>({ url: '/system/users', method: 'post', data })
}

/** 编辑用户 */
export function updateUser(id: number, data: Partial<UserPayload>) {
  return request<null>({ url: `/system/users/${id}`, method: 'put', data })
}

/** 删除用户 */
export function deleteUser(id: number) {
  return request<null>({ url: `/system/users/${id}`, method: 'delete' })
}
