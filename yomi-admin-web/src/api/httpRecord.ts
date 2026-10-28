import { request, requestPage } from '@/utils/request'
import type { HttpRecordDetail, HttpRecordItem, HttpRecordQuery, PageResult } from './types'

/** 分页查询访问记录 */
export async function getHttpRecordPage(
  params: HttpRecordQuery,
): Promise<PageResult<HttpRecordItem>> {
  const { list, total } = await requestPage<HttpRecordItem>({
    url: '/system/http-records',
    method: 'get',
    params,
  })
  return { list, total, page: params.page, pageSize: params.pageSize }
}

/** 查询访问记录详情 */
export function getHttpRecordDetail(id: string) {
  return request<HttpRecordDetail>({ url: `/system/http-records/${id}`, method: 'get' })
}
