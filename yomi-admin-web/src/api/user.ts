import { request, requestPage } from '@/utils/request'
import type { CreateUserPayload, PageQuery, PageResult, UserItem, UserPayload } from './types'

export interface UserQuery extends PageQuery {
  keyword?: string
}

/** 分页查询用户 */
export async function getUserPage(params: UserQuery): Promise<PageResult<UserItem>> {
  const { list, total } = await requestPage<UserItem>({
    url: '/system/users',
    method: 'get',
    params,
  })
  return { list, total, page: params.page, pageSize: params.pageSize }
}

/** 新增用户 */
export function createUser(data: CreateUserPayload) {
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
