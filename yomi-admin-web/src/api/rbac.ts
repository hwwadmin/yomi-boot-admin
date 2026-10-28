import { request, requestPage } from '@/utils/request'
import type {
  PageQuery,
  PageResult,
  PermissionItem,
  PermissionType,
  RoleDetail,
  RoleItem,
  RolePayload,
  Status,
  UserRoleBinding,
} from './types'

export interface RoleQuery extends PageQuery {
  keyword?: string
  status?: Status | ''
}

export interface PermissionQuery extends PageQuery {
  keyword?: string
  type?: PermissionType | ''
  status?: Status | ''
}

/** 分页查询角色 */
export async function getRolePage(params: RoleQuery): Promise<PageResult<RoleItem>> {
  const { list, total } = await requestPage<RoleItem>({
    url: '/system/role',
    method: 'get',
    params,
  })
  return { list, total, page: params.page, pageSize: params.pageSize }
}

/** 查询角色详情 */
export function getRoleDetail(id: string) {
  return request<RoleDetail>({ url: `/system/role/${id}`, method: 'get' })
}

/** 创建角色 */
export function createRole(data: RolePayload) {
  return request<RoleItem>({ url: '/system/role', method: 'post', data })
}

/** 更新角色 */
export function updateRole(id: string, data: Omit<RolePayload, 'code'>) {
  return request<null>({ url: `/system/role/${id}`, method: 'put', data })
}

/** 覆盖角色权限 */
export function bindRolePermissions(id: string, permissionIds: string[]) {
  return request<null>({
    url: `/system/role/${id}/permission`,
    method: 'put',
    data: { permissionIds },
  })
}

/** 分页查询权限目录 */
export async function getPermissionPage(
  params: PermissionQuery,
): Promise<PageResult<PermissionItem>> {
  const { list, total } = await requestPage<PermissionItem>({
    url: '/system/permission',
    method: 'get',
    params,
  })
  return { list, total, page: params.page, pageSize: params.pageSize }
}

/** 设置权限启停 */
export function setPermissionStatus(id: string, status: Status) {
  return request<null>({
    url: `/system/permission/${id}/status`,
    method: 'put',
    data: { status },
  })
}

/** 查询用户角色绑定视图 */
export function getUserRoles(id: string) {
  return request<UserRoleBinding>({ url: `/system/user/${id}/role`, method: 'get' })
}

/** 覆盖用户角色 */
export function bindUserRoles(id: string, roleIds: string[]) {
  return request<null>({
    url: `/system/user/${id}/role`,
    method: 'put',
    data: { roleIds },
  })
}
