<script setup lang="ts">
import { getHttpRecordDetail, getHttpRecordPage } from '@/api/httpRecord'
import type { HttpRecordDetail, HttpRecordItem, HttpRecordQuery } from '@/api/types'
import { formatDateTime } from '@/utils/datetime'

defineOptions({ name: 'SystemHttpRecordPage' })

// ———————— 列表查询 ————————
const loading = ref(false)
const list = ref<HttpRecordItem[]>([])
const total = ref(0)
const timeRange = ref<[Date, Date] | null>(null)
const query = reactive<HttpRecordQuery>({
  page: 1,
  pageSize: 10,
  keyword: '',
  method: '',
  statusClass: '',
})

async function fetchList() {
  loading.value = true
  try {
    const result = await getHttpRecordPage({
      page: query.page,
      pageSize: query.pageSize,
      keyword: query.keyword?.trim() || undefined,
      method: query.method || undefined,
      statusClass: query.statusClass || undefined,
      startTime: timeRange.value?.[0]?.getTime(),
      endTime: timeRange.value?.[1]?.getTime(),
    })
    list.value = result.list
    total.value = result.total
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  query.page = 1
  fetchList()
}

function handleReset() {
  query.keyword = ''
  query.method = ''
  query.statusClass = ''
  timeRange.value = null
  handleSearch()
}

onMounted(fetchList)

// ———————— 展示辅助 ————————
function statusTagType(statusCode: number): 'success' | 'info' | 'warning' | 'danger' {
  if (statusCode >= 500) return 'danger'
  if (statusCode >= 400) return 'warning'
  if (statusCode >= 300) return 'info'
  return 'success'
}

function methodTagType(method: string): 'success' | 'info' | 'warning' | 'danger' | 'primary' {
  switch (method) {
    case 'GET':
      return 'success'
    case 'POST':
      return 'primary'
    case 'PUT':
      return 'warning'
    case 'DELETE':
      return 'danger'
    default:
      return 'info'
  }
}

function resultTagType(row: HttpRecordItem): 'success' | 'info' | 'warning' | 'danger' {
  if (row.bizCode && row.bizCode !== 'OK') return 'danger'
  return statusTagType(row.statusCode)
}

function prettyJSON(value: string): string {
  if (!value) return ''
  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  } catch {
    return value
  }
}

// ———————— 详情 ————————
const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref<HttpRecordDetail | null>(null)

async function openDetail(row: HttpRecordItem) {
  detail.value = null
  detailVisible.value = true
  detailLoading.value = true
  try {
    detail.value = await getHttpRecordDetail(row.id)
  } finally {
    detailLoading.value = false
  }
}
</script>

<template>
  <div class="system-http-record">
    <el-card shadow="never">
      <div class="toolbar">
        <el-input
          v-model="query.keyword"
          class="toolbar-input"
          placeholder="路径 / TraceID / 客户端 IP"
          clearable
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />
        <el-select v-model="query.method" class="toolbar-select" placeholder="全部方法" clearable>
          <el-option
            v-for="method in ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']"
            :key="method"
            :label="method"
            :value="method"
          />
        </el-select>
        <el-select
          v-model="query.statusClass"
          class="toolbar-select"
          placeholder="全部状态"
          clearable
        >
          <el-option label="2xx 成功" value="2xx" />
          <el-option label="3xx 重定向" value="3xx" />
          <el-option label="4xx 客户端错误" value="4xx" />
          <el-option label="5xx 服务端错误" value="5xx" />
        </el-select>
        <el-date-picker
          v-model="timeRange"
          class="toolbar-range"
          type="datetimerange"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
        />
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column label="请求时间" width="170">
          <template #default="{ row }">
            {{ formatDateTime((row as HttpRecordItem).requestAt) }}
          </template>
        </el-table-column>
        <el-table-column label="方法" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="methodTagType((row as HttpRecordItem).method)" size="small">
              {{ (row as HttpRecordItem).method }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="路径" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="record-path">{{ (row as HttpRecordItem).path }}</span>
          </template>
        </el-table-column>
        <el-table-column label="结果" width="150" align="center">
          <template #default="{ row }">
            <el-tag :type="resultTagType(row as HttpRecordItem)" size="small">
              {{ (row as HttpRecordItem).bizCode || (row as HttpRecordItem).statusCode }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="耗时" width="90" align="right">
          <template #default="{ row }"> {{ (row as HttpRecordItem).durationMs }} ms </template>
        </el-table-column>
        <el-table-column label="用户 ID" width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ (row as HttpRecordItem).userId === '0' ? '-' : (row as HttpRecordItem).userId }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="70" fixed="right" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row as HttpRecordItem)"
              >详情</el-button
            >
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.pageSize"
        class="page-pagination"
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        :page-sizes="[10, 20, 50]"
        @current-change="fetchList"
        @size-change="handleSearch"
      />
    </el-card>

    <el-drawer v-model="detailVisible" title="访问记录详情" size="720px">
      <div v-loading="detailLoading" class="detail">
        <template v-if="detail">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="请求时间">
              {{ formatDateTime(detail.requestAt) }}
            </el-descriptions-item>
            <el-descriptions-item label="耗时">{{ detail.durationMs }} ms</el-descriptions-item>
            <el-descriptions-item label="服务">{{ detail.serviceType }}</el-descriptions-item>
            <el-descriptions-item label="版本">{{
              detail.serviceVersion || '-'
            }}</el-descriptions-item>
            <el-descriptions-item label="方法">
              <el-tag :type="methodTagType(detail.method)" size="small">{{ detail.method }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="状态码">
              <el-tag :type="statusTagType(detail.statusCode)" size="small">
                {{ detail.statusCode }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="业务码">{{ detail.bizCode || '-' }}</el-descriptions-item>
            <el-descriptions-item label="业务消息">{{ detail.bizMsg || '-' }}</el-descriptions-item>
            <el-descriptions-item label="客户端 IP">{{
              detail.clientIp || '-'
            }}</el-descriptions-item>
            <el-descriptions-item label="端类型">{{
              detail.clientType || '-'
            }}</el-descriptions-item>
            <el-descriptions-item label="用户 ID">
              {{ detail.userId === '0' ? '未登录' : detail.userId }}
            </el-descriptions-item>
            <el-descriptions-item label="落库时间">
              {{ formatDateTime(detail.createTime) }}
            </el-descriptions-item>
            <el-descriptions-item label="TraceID" :span="2">
              <span class="detail-mono">{{ detail.traceId || '-' }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="路径" :span="2">
              <span class="detail-mono">{{ detail.path }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="路由模板" :span="2">
              <span class="detail-mono">{{ detail.route || '-' }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="Query" :span="2">
              <span class="detail-mono">{{ detail.query || '-' }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="User-Agent" :span="2">
              <span class="detail-mono">{{ detail.userAgent || '-' }}</span>
            </el-descriptions-item>
          </el-descriptions>

          <section class="detail-section">
            <div class="detail-section__title">请求头</div>
            <pre class="detail-pre">{{ prettyJSON(detail.requestHeaders) || '（空）' }}</pre>
          </section>

          <section class="detail-section">
            <div class="detail-section__title">
              请求体
              <el-tag v-if="detail.requestBodyTruncated" type="warning" size="small">
                超出采集上限，正文未记录
              </el-tag>
            </div>
            <pre class="detail-pre">{{ prettyJSON(detail.requestBody) || '（空）' }}</pre>
          </section>

          <section class="detail-section">
            <div class="detail-section__title">
              响应体
              <el-tag v-if="detail.responseBodyTruncated" type="warning" size="small">
                超出采集上限，正文未记录
              </el-tag>
            </div>
            <pre class="detail-pre">{{ prettyJSON(detail.responseBody) || '（空）' }}</pre>
          </section>
        </template>
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
}

.toolbar-input {
  width: 240px;
}

.toolbar-select {
  width: 130px;
}

.toolbar-range {
  width: 360px;
}

.page-pagination {
  justify-content: flex-end;
  margin-top: 16px;
}

.record-path {
  font-family: var(--el-font-family-mono, monospace);
  font-size: 13px;
}

.detail {
  min-height: 200px;
}

.detail-mono {
  font-family: var(--el-font-family-mono, monospace);
  font-size: 12px;
  word-break: break-all;
}

.detail-section {
  margin-top: 20px;
}

.detail-section__title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-weight: 600;
}

.detail-pre {
  max-height: 280px;
  margin: 0;
  padding: 12px;
  overflow: auto;
  border-radius: 4px;
  background: var(--el-fill-color-light);
  font-family: var(--el-font-family-mono, monospace);
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
