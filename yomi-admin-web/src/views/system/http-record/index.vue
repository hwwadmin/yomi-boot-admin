<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Refresh, Search } from '@element-plus/icons-vue'

import { getHttpRecordDetail, getHttpRecordPage } from '@/api/httpRecord'
import type { HttpRecordDetail, HttpRecordItem, HttpRecordQuery } from '@/api/types'
import PageCard from '@/components/PageCard.vue'
import PageHeader from '@/components/PageHeader.vue'
import { formatDateTime } from '@/utils/datetime'

defineOptions({ name: 'SystemHttpRecordPage' })

const { t } = useI18n()

// ———————— 列表查询 ————————
const loading = ref(false)
const list = ref<HttpRecordItem[]>([])
const total = ref(0)

/** 默认时间范围是"今天"：后端要求列表查询必带时间范围，避免界面看着没过滤 */
function todayRange(): [Date, Date] {
  const start = new Date()
  start.setHours(0, 0, 0, 0)
  const end = new Date(start)
  end.setDate(end.getDate() + 1)
  end.setMilliseconds(end.getMilliseconds() - 1)
  return [start, end]
}

const timeRange = ref<[Date, Date]>(todayRange())
const query = reactive<HttpRecordQuery>({
  page: 1,
  pageSize: 10,
  keyword: '',
  method: '',
  statusClass: '',
})

const statusClassOptions = computed(() => [
  { value: '2xx', label: t('system.httpRecord.statusClass.success') },
  { value: '3xx', label: t('system.httpRecord.statusClass.redirect') },
  { value: '4xx', label: t('system.httpRecord.statusClass.clientError') },
  { value: '5xx', label: t('system.httpRecord.statusClass.serverError') },
])

async function fetchList() {
  loading.value = true
  try {
    // 选择器被清空时兜底回今天，保证请求始终带完整时间范围
    const range = timeRange.value ?? todayRange()
    const result = await getHttpRecordPage({
      page: query.page,
      pageSize: query.pageSize,
      keyword: query.keyword?.trim() || undefined,
      method: query.method || undefined,
      statusClass: query.statusClass || undefined,
      startTime: range[0].getTime(),
      endTime: range[1].getTime(),
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
  timeRange.value = todayRange()
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
    <PageHeader :title="t('menu.systemHttpRecord')" :subtitle="t('system.httpRecord.subtitle')" />

    <PageCard>
      <template #toolbar>
        <div class="page-toolbar">
          <div class="page-toolbar__filters">
            <el-input
              v-model="query.keyword"
              class="page-toolbar__input page-toolbar__input--compact"
              :placeholder="t('system.httpRecord.searchPlaceholder')"
              :suffix-icon="Search"
              clearable
              @keyup.enter="handleSearch"
              @clear="handleSearch"
            />
            <el-select
              v-model="query.method"
              class="page-toolbar__select"
              :placeholder="t('system.httpRecord.allMethods')"
              clearable
            >
              <el-option
                v-for="method in ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']"
                :key="method"
                :label="method"
                :value="method"
              />
            </el-select>
            <el-select
              v-model="query.statusClass"
              class="page-toolbar__select"
              :placeholder="t('system.httpRecord.allStatus')"
              clearable
            >
              <el-option
                v-for="option in statusClassOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
            <el-date-picker
              v-model="timeRange"
              class="page-toolbar__range"
              type="datetimerange"
              :clearable="false"
              :range-separator="t('common.to')"
              :start-placeholder="t('system.httpRecord.startTime')"
              :end-placeholder="t('system.httpRecord.endTime')"
            />
            <el-button type="primary" :icon="Search" @click="handleSearch">
              {{ t('common.action.search') }}
            </el-button>
            <el-button :icon="Refresh" @click="handleReset">
              {{ t('common.action.reset') }}
            </el-button>
          </div>
        </div>
      </template>

      <el-table v-loading="loading" :data="list">
        <el-table-column
          :label="t('system.httpRecord.columns.requestAt')"
          width="170"
          class-name="table-time"
        >
          <template #default="{ row }">
            {{ formatDateTime((row as HttpRecordItem).requestAt) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('system.httpRecord.columns.method')" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="methodTagType((row as HttpRecordItem).method)" size="small">
              {{ (row as HttpRecordItem).method }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          :label="t('system.httpRecord.columns.path')"
          min-width="240"
          class-name="table-code"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ (row as HttpRecordItem).path }}
          </template>
        </el-table-column>
        <el-table-column :label="t('system.httpRecord.columns.result')" width="210" align="center">
          <template #default="{ row }">
            <el-tag :type="resultTagType(row as HttpRecordItem)" size="small" class="result-tag">
              {{ (row as HttpRecordItem).bizCode || (row as HttpRecordItem).statusCode }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('system.httpRecord.columns.duration')" width="96" align="right">
          <template #default="{ row }">
            {{ (row as HttpRecordItem).durationMs }} {{ t('common.unit.ms') }}
          </template>
        </el-table-column>
        <el-table-column
          :label="t('system.httpRecord.columns.userId')"
          width="172"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{
              (row as HttpRecordItem).userId === '0'
                ? t('common.none')
                : (row as HttpRecordItem).userId
            }}
          </template>
        </el-table-column>
        <el-table-column :label="t('system.httpRecord.columns.actions')" width="90" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row as HttpRecordItem)">
              {{ t('system.httpRecord.detail') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <template #footer>
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
      </template>
    </PageCard>

    <el-drawer v-model="detailVisible" :title="t('system.httpRecord.drawer.title')" size="720px">
      <div v-loading="detailLoading" class="detail">
        <template v-if="detail">
          <el-descriptions :column="2" border>
            <el-descriptions-item :label="t('system.httpRecord.drawer.requestAt')">
              {{ formatDateTime(detail.requestAt) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.httpRecord.drawer.duration')">
              {{ detail.durationMs }} {{ t('common.unit.ms') }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.httpRecord.drawer.service')">
              {{ detail.serviceType }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.httpRecord.drawer.version')">
              {{ detail.serviceVersion || t('common.none') }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.httpRecord.drawer.method')">
              <el-tag :type="methodTagType(detail.method)" size="small">{{ detail.method }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.httpRecord.drawer.statusCode')">
              <el-tag :type="statusTagType(detail.statusCode)" size="small">
                {{ detail.statusCode }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.httpRecord.drawer.bizCode')">
              {{ detail.bizCode || t('common.none') }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.httpRecord.drawer.bizMsg')">
              {{ detail.bizMsg || t('common.none') }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.httpRecord.drawer.clientIp')">
              {{ detail.clientIp || t('common.none') }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.httpRecord.drawer.clientType')">
              {{ detail.clientType || t('common.none') }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.httpRecord.drawer.userId')">
              {{ detail.userId === '0' ? t('system.httpRecord.drawer.anonymous') : detail.userId }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.httpRecord.drawer.createdAt')">
              {{ formatDateTime(detail.createTime) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.httpRecord.drawer.traceId')" :span="2">
              <span class="detail-mono">{{ detail.traceId || t('common.none') }}</span>
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.httpRecord.drawer.path')" :span="2">
              <span class="detail-mono">{{ detail.path }}</span>
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.httpRecord.drawer.route')" :span="2">
              <span class="detail-mono">{{ detail.route || t('common.none') }}</span>
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.httpRecord.drawer.query')" :span="2">
              <span class="detail-mono">{{ detail.query || t('common.none') }}</span>
            </el-descriptions-item>
            <el-descriptions-item :label="t('system.httpRecord.drawer.userAgent')" :span="2">
              <span class="detail-mono">{{ detail.userAgent || t('common.none') }}</span>
            </el-descriptions-item>
          </el-descriptions>

          <section class="detail-section">
            <div class="detail-section__title">
              {{ t('system.httpRecord.drawer.requestHeaders') }}
            </div>
            <pre class="detail-pre">{{
              prettyJSON(detail.requestHeaders) || t('system.httpRecord.drawer.empty')
            }}</pre>
          </section>

          <section class="detail-section">
            <div class="detail-section__title">
              {{ t('system.httpRecord.drawer.requestBody') }}
              <el-tag v-if="detail.requestBodyTruncated" type="warning" size="small">
                {{ t('system.httpRecord.drawer.truncated') }}
              </el-tag>
            </div>
            <pre class="detail-pre">{{
              prettyJSON(detail.requestBody) || t('system.httpRecord.drawer.empty')
            }}</pre>
          </section>

          <section class="detail-section">
            <div class="detail-section__title">
              {{ t('system.httpRecord.drawer.responseBody') }}
              <el-tag v-if="detail.responseBodyTruncated" type="warning" size="small">
                {{ t('system.httpRecord.drawer.truncated') }}
              </el-tag>
            </div>
            <pre class="detail-pre">{{
              prettyJSON(detail.responseBody) || t('system.httpRecord.drawer.empty')
            }}</pre>
          </section>
        </template>
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
/* 业务码可能长达 40 字符，超出列宽时换行，避免省略号 */
.result-tag {
  max-width: 100%;
  height: auto;
  padding-top: 2px;
  padding-bottom: 2px;
  line-height: 18px;
  white-space: normal;
  word-break: break-all;
}

.detail {
  min-height: 200px;
}

.detail-mono {
  font-family: var(--app-font-mono);
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
  padding: 14px 16px;
  overflow: auto;
  border-radius: var(--app-radius-lg);
  background: var(--el-fill-color-light);
  font-family: var(--app-font-mono);
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
