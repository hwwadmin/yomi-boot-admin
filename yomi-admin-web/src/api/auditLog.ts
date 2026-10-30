import { requestPage } from '@/utils/request'
import type { AuditLogItem, AuditLogQuery, PageResult } from './types'

export type { AuditLogQuery }

/** 分页查询操作审计日志 */
export async function getAuditLogPage(params: AuditLogQuery): Promise<PageResult<AuditLogItem>> {
  const { list, total } = await requestPage<AuditLogItem>({
    url: '/system/audit-log',
    method: 'get',
    params,
  })
  return { list, total, page: params.page, pageSize: params.pageSize }
}
