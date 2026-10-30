import { request, requestPage } from '@/utils/request'
import type {
  PageResult,
  Status,
  SysParamCreatePayload,
  SysParamItem,
  SysParamQuery,
  SysParamUpdatePayload,
  ParamValueType,
} from './types'

export type { SysParamQuery }

/** 分页查询系统参数 */
export async function getSysParamPage(params: SysParamQuery): Promise<PageResult<SysParamItem>> {
  const { list, total } = await requestPage<SysParamItem>({
    url: '/system/param',
    method: 'get',
    params,
  })
  return { list, total, page: params.page, pageSize: params.pageSize }
}

/** 创建系统参数；键创建后不可变 */
export function createSysParam(data: SysParamCreatePayload) {
  return request<SysParamItem>({ url: '/system/param/create', method: 'post', data })
}

/** 更新系统参数；键与值类型不可变 */
export function updateSysParam(id: string, data: SysParamUpdatePayload) {
  return request<null>({ url: '/system/param/update', method: 'post', data: { id, ...data } })
}

/** 删除系统参数（物理删除） */
export function deleteSysParam(id: string) {
  return request<null>({ url: '/system/param/delete', method: 'post', data: { id } })
}

/** 设置参数启停；禁用后业务读取方回退默认值 */
export function setSysParamStatus(id: string, status: Status) {
  return request<null>({ url: '/system/param/set-status', method: 'post', data: { id, status } })
}

/** 参数值类型展示名；按展示顺序声明（对象键序遍历，字符串放第一位） */
export const paramValueTypeLabels: Record<ParamValueType, string> = {
  STRING: '字符串',
  BOOL: '布尔',
  INT: '整数',
  JSON: 'JSON',
}
