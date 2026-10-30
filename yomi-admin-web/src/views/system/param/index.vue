<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { Plus, Refresh, Search, Setting } from '@element-plus/icons-vue'

import {
  createSysParam,
  deleteSysParam,
  getSysParamPage,
  setSysParamStatus,
  updateSysParam,
} from '@/api/sysparam'
import type { SysParamQuery } from '@/api/sysparam'
import type { ParamValueType, Status, SysParamItem } from '@/api/types'
import NameCell from '@/components/NameCell.vue'
import PageCard from '@/components/PageCard.vue'
import { useParamValueTypeLabels } from '@/composables/useParamValueTypeLabels'
import { confirmAction } from '@/utils/confirm'
import { formatDateTime } from '@/utils/datetime'

defineOptions({ name: 'SystemParamPage' })

const { t } = useI18n()
const paramValueTypeLabels = useParamValueTypeLabels()

/** int64 取值范围，前端先挡一道，避免超范围值提交后才由后端报"参数不合法" */
const INT64_MAX = 2n ** 63n - 1n
const INT64_MIN = -(2n ** 63n)

// ———————— 列表查询 ————————
const loading = ref(false)
const list = ref<SysParamItem[]>([])
const total = ref(0)
const query = reactive<SysParamQuery>({
  page: 1,
  pageSize: 10,
  keyword: '',
  valueType: '',
  status: '',
})

async function fetchList() {
  loading.value = true
  try {
    const result = await getSysParamPage(query)
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
  query.valueType = ''
  query.status = ''
  handleSearch()
}

onMounted(fetchList)

// ———————— 新增 / 编辑 ————————
const dialogVisible = ref(false)
const submitting = ref(false)
const editingId = ref<string | null>(null)
const formRef = ref<FormInstance>()
const form = reactive({
  key: '',
  value: '',
  valueType: 'STRING' as ParamValueType,
  name: '',
  remark: '',
})

function validateValue(_rule: unknown, value: string, callback: (error?: Error) => void) {
  const raw = value ?? ''
  if (form.valueType === 'BOOL') {
    if (raw !== 'true' && raw !== 'false') {
      return callback(new Error(t('validation.param.bool')))
    }
    return callback()
  }
  if (form.valueType === 'INT') {
    if (!/^-?\d+$/.test(raw)) return callback(new Error(t('validation.param.int')))
    const parsed = BigInt(raw)
    if (parsed < INT64_MIN || parsed > INT64_MAX) {
      return callback(new Error(t('validation.param.intRange')))
    }
    return callback()
  }
  if (form.valueType === 'JSON') {
    try {
      JSON.parse(raw)
    } catch {
      return callback(new Error(t('validation.param.json')))
    }
    return callback()
  }
  if (raw.length > 2048) return callback(new Error(t('validation.param.stringMax')))
  callback()
}

const rules = computed<FormRules>(() => ({
  key: [
    { required: true, message: t('system.param.validation.keyRequired'), trigger: 'blur' },
    {
      pattern: /^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$/,
      message: t('validation.paramKeyPattern'),
      trigger: 'blur',
    },
  ],
  value: [{ validator: validateValue, trigger: 'blur' }],
  name: [{ required: true, message: t('system.param.validation.nameRequired'), trigger: 'blur' }],
}))

function openCreate() {
  editingId.value = null
  Object.assign(form, { key: '', value: '', valueType: 'STRING', name: '', remark: '' })
  dialogVisible.value = true
}

function openEdit(row: SysParamItem) {
  editingId.value = row.id
  Object.assign(form, {
    key: row.key,
    value: row.value,
    valueType: row.valueType,
    name: row.name,
    remark: row.remark,
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    if (editingId.value === null) {
      await createSysParam({
        key: form.key.trim(),
        value: form.value,
        valueType: form.valueType,
        name: form.name.trim(),
        remark: form.remark.trim(),
      })
      ElMessage.success(t('system.param.createSuccess'))
    } else {
      await updateSysParam(editingId.value, {
        value: form.value,
        name: form.name.trim(),
        remark: form.remark.trim(),
      })
      ElMessage.success(t('system.param.saveSuccess'))
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: SysParamItem) {
  const confirmed = await confirmAction({
    title: t('system.param.deleteConfirm.title'),
    message: t('system.param.deleteConfirm.message', { key: row.key }),
    confirmButtonText: t('system.param.deleteConfirm.confirm'),
  })
  if (!confirmed) return
  await deleteSysParam(row.id)
  ElMessage.success(t('system.param.deleteSuccess'))
  if (list.value.length === 1 && query.page > 1) {
    query.page -= 1
  }
  fetchList()
}

// ———————— 启停 ————————
function confirmStatusChange(key: string, next: Status): Promise<boolean> {
  const disabling = next === 'DISABLED'
  return confirmAction({
    title: t(
      disabling
        ? 'system.param.statusConfirm.disableTitle'
        : 'system.param.statusConfirm.enableTitle',
    ),
    message: t(
      disabling
        ? 'system.param.statusConfirm.disableMessage'
        : 'system.param.statusConfirm.enableMessage',
      { key },
    ),
  })
}

async function handleStatusChange(row: SysParamItem, status: Status) {
  try {
    await setSysParamStatus(row.id, status)
    row.status = status
    ElMessage.success(t('system.param.statusUpdated'))
  } catch {
    fetchList()
  }
}
</script>

<template>
  <div class="system-param">
    <PageCard>
      <template #toolbar>
        <div class="page-toolbar">
          <div class="page-toolbar__filters">
            <el-input
              v-model="query.keyword"
              class="page-toolbar__input"
              :placeholder="t('system.param.searchPlaceholder')"
              :suffix-icon="Search"
              clearable
              @keyup.enter="handleSearch"
              @clear="handleSearch"
            />
            <el-select
              v-model="query.valueType"
              class="page-toolbar__select"
              :placeholder="t('system.param.allTypes')"
              clearable
            >
              <el-option
                v-for="(label, type) in paramValueTypeLabels"
                :key="type"
                :label="label"
                :value="type"
              />
            </el-select>
            <el-select
              v-model="query.status"
              class="page-toolbar__select"
              :placeholder="t('system.param.allStatus')"
              clearable
            >
              <el-option :label="t('common.status.enabled')" value="ENABLED" />
              <el-option :label="t('common.status.disabled')" value="DISABLED" />
            </el-select>
            <el-button type="primary" :icon="Search" @click="handleSearch">
              {{ t('common.action.search') }}
            </el-button>
            <el-button :icon="Refresh" @click="handleReset">
              {{ t('common.action.reset') }}
            </el-button>
          </div>
          <div class="page-toolbar__actions">
            <el-button
              v-permission="'system:param:create'"
              type="primary"
              :icon="Plus"
              @click="openCreate"
            >
              {{ t('system.param.add') }}
            </el-button>
          </div>
        </div>
      </template>

      <el-table v-loading="loading" :data="list">
        <el-table-column
          :label="t('system.param.columns.name')"
          min-width="160"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <NameCell :label="(row as SysParamItem).name" :icon="Setting" tone="cyan" />
          </template>
        </el-table-column>
        <el-table-column
          prop="key"
          :label="t('system.param.columns.key')"
          min-width="180"
          class-name="table-code"
          show-overflow-tooltip
        />
        <el-table-column
          prop="value"
          :label="t('system.param.columns.value')"
          min-width="135"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.field.type')" width="90" align="center">
          <template #default="{ row }">
            <el-tag>{{ paramValueTypeLabels[(row as SysParamItem).valueType] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.field.status')" width="80" align="center">
          <template #default="{ row }">
            <el-switch
              v-permission="'system:param:status'"
              :model-value="(row as SysParamItem).status"
              active-value="ENABLED"
              inactive-value="DISABLED"
              :before-change="
                () =>
                  confirmStatusChange(
                    (row as SysParamItem).key,
                    (row as SysParamItem).status === 'ENABLED' ? 'DISABLED' : 'ENABLED',
                  )
              "
              @change="
                (value: string | number | boolean) =>
                  handleStatusChange(row as SysParamItem, value as Status)
              "
            />
          </template>
        </el-table-column>
        <el-table-column
          prop="remark"
          :label="t('common.field.remark')"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.field.updatedAt')" width="170" class-name="table-time">
          <template #default="{ row }">
            {{ formatDateTime((row as SysParamItem).updateTime) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.field.actions')" width="175" fixed="right">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button
                v-permission="'system:param:update'"
                link
                type="primary"
                @click="openEdit(row as SysParamItem)"
              >
                {{ t('common.action.edit') }}
              </el-button>
              <el-button
                v-permission="'system:param:delete'"
                link
                type="danger"
                @click="handleDelete(row as SysParamItem)"
              >
                {{ t('common.action.delete') }}
              </el-button>
            </div>
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

    <el-dialog
      v-model="dialogVisible"
      :title="
        editingId === null
          ? t('system.param.dialog.createTitle')
          : t('system.param.dialog.editTitle')
      "
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="130px"
        :validate-on-rule-change="false"
      >
        <el-form-item :label="t('system.param.dialog.name')" prop="name">
          <el-input v-model="form.name" maxlength="64" />
        </el-form-item>
        <el-form-item :label="t('system.param.dialog.key')" prop="key">
          <el-input
            v-model="form.key"
            :disabled="editingId !== null"
            :placeholder="t('system.param.dialog.keyPlaceholder')"
            maxlength="128"
          />
        </el-form-item>
        <el-form-item :label="t('system.param.dialog.valueType')" prop="valueType">
          <el-radio-group v-model="form.valueType" :disabled="editingId !== null">
            <el-radio v-for="(label, type) in paramValueTypeLabels" :key="type" :value="type">
              {{ label }}
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="t('system.param.dialog.value')" prop="value">
          <el-switch
            v-if="form.valueType === 'BOOL'"
            :model-value="form.value === 'true'"
            active-text="true"
            inactive-text="false"
            @change="(value: string | number | boolean) => (form.value = value ? 'true' : 'false')"
          />
          <el-input
            v-else-if="form.valueType === 'JSON'"
            v-model="form.value"
            type="textarea"
            :rows="4"
            :placeholder="t('system.param.dialog.jsonPlaceholder')"
          />
          <el-input
            v-else
            v-model="form.value"
            :maxlength="form.valueType === 'STRING' ? 2048 : undefined"
            :placeholder="
              form.valueType === 'INT'
                ? t('system.param.dialog.intPlaceholder')
                : t('system.param.dialog.valuePlaceholder')
            "
          />
        </el-form-item>
        <el-form-item :label="t('common.field.remark')">
          <el-input v-model="form.remark" type="textarea" maxlength="255" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ t('common.action.cancel') }}</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          {{ t('common.action.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>
