import { request, requestPage } from '@/utils/request'
import type {
  PageQuery,
  PageResult,
  PermissionGroupDetail,
  PermissionGroupItem,
  PermissionGroupPayload,
  PermissionItem,
  PermissionPayload,
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

export interface PermissionGroupQuery extends PageQuery {
  keyword?: string
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
  return request<RoleDetail>({ url: '/system/role/detail', method: 'get', params: { id } })
}

/** 创建角色 */
export function createRole(data: RolePayload) {
  return request<RoleItem>({ url: '/system/role/create', method: 'post', data })
}

/** 更新角色资料 */
export function updateRole(id: string, data: Omit<RolePayload, 'code'>) {
  return request<null>({ url: '/system/role/update', method: 'post', data: { id, ...data } })
}

/** 删除角色 */
export function deleteRole(id: string) {
  return request<null>({ url: '/system/role/delete', method: 'post', data: { id } })
}

/** 设置角色启停 */
export function setRoleStatus(id: string, status: Status) {
  return request<null>({ url: '/system/role/set-status', method: 'post', data: { id, status } })
}

/** 覆盖角色权限组 */
export function bindRoleGroups(id: string, groupIds: string[]) {
  return request<null>({
    url: '/system/role/bind-group',
    method: 'post',
    data: { roleId: id, groupIds },
  })
}

/** 分页查询权限组 */
export async function getPermissionGroupPage(
  params: PermissionGroupQuery,
): Promise<PageResult<PermissionGroupItem>> {
  const { list, total } = await requestPage<PermissionGroupItem>({
    url: '/system/permission-group',
    method: 'get',
    params,
  })
  return { list, total, page: params.page, pageSize: params.pageSize }
}

/** 查询权限组详情 */
export function getPermissionGroupDetail(id: string) {
  return request<PermissionGroupDetail>({
    url: '/system/permission-group/detail',
    method: 'get',
    params: { id },
  })
}

/** 创建权限组 */
export function createPermissionGroup(data: PermissionGroupPayload) {
  return request<PermissionGroupItem>({
    url: '/system/permission-group/create',
    method: 'post',
    data,
  })
}

/** 更新权限组资料 */
export function updatePermissionGroup(id: string, data: Omit<PermissionGroupPayload, 'code'>) {
  return request<null>({
    url: '/system/permission-group/update',
    method: 'post',
    data: { id, ...data },
  })
}

/** 删除权限组 */
export function deletePermissionGroup(id: string) {
  return request<null>({ url: '/system/permission-group/delete', method: 'post', data: { id } })
}

/** 设置权限组启停 */
export function setPermissionGroupStatus(id: string, status: Status) {
  return request<null>({
    url: '/system/permission-group/set-status',
    method: 'post',
    data: { id, status },
  })
}

/** 覆盖权限组权限 */
export function bindGroupPermissions(id: string, permissionIds: string[]) {
  return request<null>({
    url: '/system/permission-group/bind-permission',
    method: 'post',
    data: { groupId: id, permissionIds },
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

/** 创建权限 */
export function createPermission(data: PermissionPayload) {
  return request<PermissionItem>({ url: '/system/permission/create', method: 'post', data })
}

/** 更新权限资料；编码与类型不可变 */
export function updatePermission(id: string, data: Omit<PermissionPayload, 'code' | 'type'>) {
  return request<null>({ url: '/system/permission/update', method: 'post', data: { id, ...data } })
}

/** 删除权限 */
export function deletePermission(id: string) {
  return request<null>({ url: '/system/permission/delete', method: 'post', data: { id } })
}

/** 设置权限启停 */
export function setPermissionStatus(id: string, status: Status) {
  return request<null>({
    url: '/system/permission/set-status',
    method: 'post',
    data: { id, status },
  })
}

/** 查询用户角色绑定视图 */
export function getUserRoles(id: string) {
  return request<UserRoleBinding>({
    url: '/system/user/role',
    method: 'get',
    params: { userId: id },
  })
}

/** 覆盖用户角色 */
export function bindUserRoles(id: string, roleIds: string[]) {
  return request<null>({
    url: '/system/user/bind-role',
    method: 'post',
    data: { userId: id, roleIds },
  })
}
