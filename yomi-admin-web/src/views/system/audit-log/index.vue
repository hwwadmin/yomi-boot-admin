<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Refresh, Search } from '@element-plus/icons-vue'

import { getAuditLogPage } from '@/api/auditLog'
import type { AuditLogItem, AuditLogQuery } from '@/api/types'
import PageCard from '@/components/PageCard.vue'
import { formatDateTime } from '@/utils/datetime'

defineOptions({ name: 'SystemAuditLogPage' })

const { t, te } = useI18n()

// ———————— 列表查询 ————————
const loading = ref(false)
const list = ref<AuditLogItem[]>([])
const total = ref(0)

/** 默认时间范围是"今天"：与后端缺省口径保持一致，避免界面看着没过滤、实际只查了今天 */
function todayRange(): [Date, Date] {
  const start = new Date()
  start.setHours(0, 0, 0, 0)
  const end = new Date(start)
  end.setDate(end.getDate() + 1)
  end.setMilliseconds(end.getMilliseconds() - 1)
  return [start, end]
}

const timeRange = ref<[Date, Date]>(todayRange())
const query = reactive<AuditLogQuery>({
  page: 1,
  pageSize: 10,
  keyword: '',
  eventCode: '',
  phase: '',
  result: '',
  auditId: '',
})

/** 事件标签映射；新增事件时补充 i18n */
const EVENT_CODES = ['auth.login', 'auth.change_password', 'auth.reset_password'] as const

/** 事件编码含点号，需映射为 i18n 安全键 */
const EVENT_LABEL_KEYS: Record<string, string> = {
  'auth.login': 'login',
  'auth.change_password': 'changePassword',
  'auth.reset_password': 'resetPassword',
}

const eventOptions = computed(() =>
  EVENT_CODES.map((code) => ({ value: code, label: eventLabel(code) })),
)
const phaseOptions = computed(() => [
  { value: 'STARTED', label: t('system.auditLog.phases.started') },
  { value: 'FINISHED', label: t('system.auditLog.phases.finished') },
])
const resultOptions = computed(() => [
  { value: 'PENDING', label: t('system.auditLog.results.pending') },
  { value: 'SUCCESS', label: t('system.auditLog.results.success') },
  { value: 'FAILURE', label: t('system.auditLog.results.failure') },
])

/** 按事件 ID 查时忽略时间范围：那是精确定位单条事件，跨天也要能查到 */
const auditIdSearching = computed(() => (query.auditId ?? '').trim() !== '')

async function fetchList() {
  loading.value = true
  try {
    const result = await getAuditLogPage({
      page: query.page,
      pageSize: query.pageSize,
      keyword: query.keyword?.trim() || undefined,
      eventCode: query.eventCode || undefined,
      phase: query.phase || undefined,
      result: query.result || undefined,
      auditId: query.auditId?.trim() || undefined,
      startTime: auditIdSearching.value ? undefined : timeRange.value?.[0]?.getTime(),
      endTime: auditIdSearching.value ? undefined : timeRange.value?.[1]?.getTime(),
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
  query.eventCode = ''
  query.phase = ''
  query.result = ''
  query.auditId = ''
  timeRange.value = todayRange()
  handleSearch()
}

onMounted(fetchList)

// ———————— 展示辅助 ————————
function eventLabel(code: string): string {
  const labelKey = EVENT_LABEL_KEYS[code]
  if (!labelKey) return code
  const key = `system.auditLog.events.${labelKey}`
  return te(key) ? t(key) : code
}

function phaseLabel(phase: string): string {
  return phase === 'STARTED'
    ? t('system.auditLog.phases.started')
    : t('system.auditLog.phases.finished')
}

function phaseTagType(phase: string): 'info' | 'success' {
  return phase === 'STARTED' ? 'info' : 'success'
}

function resultLabel(result: string): string {
  switch (result) {
    case 'PENDING':
      return t('system.auditLog.results.pending')
    case 'SUCCESS':
      return t('system.auditLog.results.success')
    default:
      return t('system.auditLog.results.failure')
  }
}

function resultTagType(result: string): 'info' | 'success' | 'danger' {
  if (result === 'PENDING') return 'info'
  return result === 'SUCCESS' ? 'success' : 'danger'
}

/** 失败原因复用全局错误码翻译；未知码回退原文 */
function bizCodeText(code: string): string {
  if (!code || code === 'OK') return ''
  const key = `errors.codes.${code}`
  return te(key) ? t(key) : code
}

function operatorText(row: AuditLogItem): string {
  if (row.operatorName) return row.operatorName
  return row.operatorId !== '0' ? `#${row.operatorId}` : t('system.auditLog.unknownOperator')
}

function operatorHint(row: AuditLogItem): string {
  return row.operatorId !== '0' ? `#${row.operatorId}` : ''
}

function targetText(row: AuditLogItem): string {
  if (!row.targetType) return t('system.auditLog.noTarget')
  return row.targetName || (row.targetId !== '0' ? `#${row.targetId}` : t('system.auditLog.noTarget'))
}
</script>

<template>
  <div class="system-audit-log">
    <PageCard>
      <template #toolbar>
        <div class="page-toolbar">
          <div class="page-toolbar__filters">
            <el-input
              v-model="query.keyword"
              class="page-toolbar__input page-toolbar__input--compact"
              :placeholder="t('system.auditLog.searchPlaceholder')"
              :suffix-icon="Search"
              clearable
              @keyup.enter="handleSearch"
              @clear="handleSearch"
            />
            <el-input
              v-model="query.auditId"
              class="page-toolbar__input page-toolbar__input--compact"
              :placeholder="t('system.auditLog.auditIdPlaceholder')"
              clearable
              @keyup.enter="handleSearch"
              @clear="handleSearch"
            />
            <el-select
              v-model="query.eventCode"
              class="page-toolbar__select"
              :placeholder="t('system.auditLog.allEvents')"
              clearable
            >
              <el-option
                v-for="option in eventOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
            <el-select
              v-model="query.phase"
              class="page-toolbar__select"
              :placeholder="t('system.auditLog.allPhases')"
              clearable
            >
              <el-option
                v-for="option in phaseOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
            <el-select
              v-model="query.result"
              class="page-toolbar__select"
              :placeholder="t('system.auditLog.allResults')"
              clearable
            >
              <el-option
                v-for="option in resultOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
            <el-date-picker
              v-model="timeRange"
              class="page-toolbar__range"
              type="datetimerange"
              :disabled="auditIdSearching"
              :range-separator="t('common.to')"
              :start-placeholder="t('system.auditLog.startTime')"
              :end-placeholder="t('system.auditLog.endTime')"
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
          :label="t('system.auditLog.columns.createTime')"
          width="170"
          class-name="table-time"
        >
          <template #default="{ row }">
            {{ formatDateTime((row as AuditLogItem).createTime) }}
          </template>
        </el-table-column>
        <el-table-column
          :label="t('system.auditLog.columns.auditId')"
          width="180"
          class-name="table-code"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ (row as AuditLogItem).auditId }}
          </template>
        </el-table-column>
        <el-table-column :label="t('system.auditLog.columns.phase')" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="phaseTagType((row as AuditLogItem).phase)" size="small" effect="plain">
              {{ phaseLabel((row as AuditLogItem).phase) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('system.auditLog.columns.event')" min-width="150">
          <template #default="{ row }">
            <el-tag size="small" effect="plain">
              {{ eventLabel((row as AuditLogItem).eventCode) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('system.auditLog.columns.operator')" min-width="170">
          <template #default="{ row }">
            <div class="cell-primary">{{ operatorText(row as AuditLogItem) }}</div>
            <div v-if="operatorHint(row as AuditLogItem)" class="cell-secondary">
              {{ operatorHint(row as AuditLogItem) }}
            </div>
          </template>
        </el-table-column>
        <el-table-column :label="t('system.auditLog.columns.target')" min-width="160">
          <template #default="{ row }">
            {{ targetText(row as AuditLogItem) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('system.auditLog.columns.result')" min-width="200">
          <template #default="{ row }">
            <div class="result-cell">
              <el-tag :type="resultTagType((row as AuditLogItem).result)" size="small">
                {{ resultLabel((row as AuditLogItem).result) }}
              </el-tag>
              <span v-if="bizCodeText((row as AuditLogItem).bizCode)" class="result-reason">
                {{ bizCodeText((row as AuditLogItem).bizCode) }}
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column
          :label="t('system.auditLog.columns.clientIp')"
          width="150"
          prop="clientIp"
        />
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
  </div>
</template>

<style scoped>
.cell-primary {
  color: var(--el-text-color-primary);
}

.cell-secondary {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.result-cell {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.result-reason {
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.4;
}
</style>
