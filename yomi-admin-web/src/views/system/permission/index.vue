<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { Connection, Grid, Plus, Refresh, Search } from '@element-plus/icons-vue'

import {
  createPermission,
  deletePermission,
  getPermissionPage,
  setPermissionStatus,
  updatePermission,
} from '@/api/rbac'
import type { PermissionQuery } from '@/api/rbac'
import type { PermissionItem, PermissionType, Status } from '@/api/types'
import NameCell from '@/components/NameCell.vue'
import PageCard from '@/components/PageCard.vue'
import { confirmAction } from '@/utils/confirm'
import { formatDateTime } from '@/utils/datetime'

defineOptions({ name: 'SystemPermissionPage' })

const { t } = useI18n()

// ———————— 列表查询 ————————
const loading = ref(false)
const list = ref<PermissionItem[]>([])
const total = ref(0)
const query = reactive<PermissionQuery>({
  page: 1,
  pageSize: 10,
  keyword: '',
  type: '',
  status: '',
})

async function fetchList() {
  loading.value = true
  try {
    const result = await getPermissionPage(query)
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
  query.type = ''
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
  code: '',
  name: '',
  type: 'MENU' as PermissionType,
  uiPath: '',
  httpMethod: 'GET',
  httpPath: '',
  remark: '',
})

/** 只有菜单类型才校验菜单路径 */
function validateUIPath(_rule: unknown, value: string, callback: (error?: Error) => void) {
  if (form.type !== 'MENU') return callback()
  const uiPath = value?.trim() ?? ''
  if (!uiPath) return callback(new Error(t('system.permission.validation.uiPathRequired')))
  if (!uiPath.startsWith('/')) {
    return callback(new Error(t('system.permission.validation.uiPathPrefix')))
  }
  callback()
}

/** 只有接口类型才校验 HTTP 方法 */
function validateHTTPMethod(_rule: unknown, value: string, callback: (error?: Error) => void) {
  if (form.type !== 'API') return callback()
  if (!value) return callback(new Error(t('system.permission.validation.httpMethodRequired')))
  callback()
}

/** 只有接口类型才校验路由模板 */
function validateHTTPPath(_rule: unknown, value: string, callback: (error?: Error) => void) {
  if (form.type !== 'API') return callback()
  const httpPath = value?.trim() ?? ''
  if (!httpPath) return callback(new Error(t('system.permission.validation.httpPathRequired')))
  if (!httpPath.startsWith('/')) {
    return callback(new Error(t('system.permission.validation.httpPathPrefix')))
  }
  callback()
}

const rules = computed<FormRules>(() => ({
  code: [
    {
      required: true,
      message: t('system.permission.validation.codeRequired'),
      trigger: 'blur',
    },
    {
      pattern: /^[a-z][a-z0-9_]*(?::[a-z][a-z0-9_]*)*$/,
      message: t('validation.permissionCodePattern'),
      trigger: 'blur',
    },
  ],
  name: [
    { required: true, message: t('system.permission.validation.nameRequired'), trigger: 'blur' },
  ],
  uiPath: [{ validator: validateUIPath, trigger: 'blur' }],
  httpMethod: [{ validator: validateHTTPMethod, trigger: 'change' }],
  httpPath: [{ validator: validateHTTPPath, trigger: 'blur' }],
}))

function openCreate() {
  editingId.value = null
  Object.assign(form, {
    code: '',
    name: '',
    type: 'MENU',
    uiPath: '',
    httpMethod: 'GET',
    httpPath: '',
    remark: '',
  })
  dialogVisible.value = true
}

function openEdit(row: PermissionItem) {
  editingId.value = row.id
  Object.assign(form, {
    code: row.code,
    name: row.name,
    type: row.type,
    uiPath: row.uiPath ?? '',
    httpMethod: row.httpMethod || 'GET',
    httpPath: row.httpPath ?? '',
    remark: row.remark,
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    const resource =
      form.type === 'MENU'
        ? { uiPath: form.uiPath.trim(), httpMethod: '', httpPath: '' }
        : { uiPath: '', httpMethod: form.httpMethod, httpPath: form.httpPath.trim() }
    const payload = { name: form.name.trim(), ...resource, remark: form.remark.trim() }
    if (editingId.value === null) {
      await createPermission({ code: form.code.trim(), type: form.type, ...payload })
      ElMessage.success(t('system.permission.createSuccess'))
    } else {
      await updatePermission(editingId.value, payload)
      ElMessage.success(t('system.permission.saveSuccess'))
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: PermissionItem) {
  const confirmed = await confirmAction({
    title: t('system.permission.deleteConfirm.title'),
    message: t('system.permission.deleteConfirm.message', { name: row.name }),
    confirmButtonText: t('system.permission.deleteConfirm.confirm'),
  })
  if (!confirmed) return
  await deletePermission(row.id)
  ElMessage.success(t('system.permission.deleteSuccess'))
  if (list.value.length === 1 && query.page > 1) {
    query.page -= 1
  }
  fetchList()
}

// ———————— 启停 ————————
/** 启停是危险操作：确认后再提交 */
function confirmStatusChange(name: string, next: Status): Promise<boolean> {
  const disabling = next === 'DISABLED'
  return confirmAction({
    title: t(
      disabling
        ? 'system.permission.statusConfirm.disableTitle'
        : 'system.permission.statusConfirm.enableTitle',
    ),
    message: t(
      disabling
        ? 'system.permission.statusConfirm.disableMessage'
        : 'system.permission.statusConfirm.enableMessage',
      { name },
    ),
  })
}

async function handleStatusChange(row: PermissionItem, status: Status) {
  try {
    await setPermissionStatus(row.id, status)
    row.status = status
    ElMessage.success(t('system.permission.statusUpdated'))
  } catch {
    fetchList()
  }
}
</script>

<template>
  <div class="system-permission">
    <PageCard>
      <template #toolbar>
        <div class="page-toolbar">
          <div class="page-toolbar__filters">
            <el-input
              v-model="query.keyword"
              class="page-toolbar__input"
              :placeholder="t('system.permission.searchPlaceholder')"
              :suffix-icon="Search"
              clearable
              @keyup.enter="handleSearch"
              @clear="handleSearch"
            />
            <el-select
              v-model="query.type"
              class="page-toolbar__select"
              :placeholder="t('common.all')"
              clearable
            >
              <el-option :label="t('system.permission.type.menu')" value="MENU" />
              <el-option :label="t('system.permission.type.api')" value="API" />
            </el-select>
            <el-select
              v-model="query.status"
              class="page-toolbar__select"
              :placeholder="t('common.all')"
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
              v-permission="'system:permission:create'"
              type="primary"
              :icon="Plus"
              @click="openCreate"
            >
              {{ t('system.permission.add') }}
            </el-button>
          </div>
        </div>
      </template>

      <el-table v-loading="loading" :data="list">
        <el-table-column :label="t('system.permission.columns.name')" min-width="180">
          <template #default="{ row }">
            <NameCell
              :label="(row as PermissionItem).name"
              :icon="(row as PermissionItem).type === 'MENU' ? Grid : Connection"
              :tone="(row as PermissionItem).type === 'MENU' ? 'green' : 'orange'"
            />
          </template>
        </el-table-column>
        <el-table-column
          prop="code"
          :label="t('system.permission.columns.code')"
          min-width="200"
          class-name="table-code"
          show-overflow-tooltip
        />
        <el-table-column :label="t('system.permission.columns.type')" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.type === 'MENU' ? 'primary' : 'success'">
              {{
                row.type === 'MENU'
                  ? t('system.permission.type.menu')
                  : t('system.permission.type.api')
              }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          :label="t('system.permission.columns.resource')"
          min-width="260"
          class-name="table-code"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <template v-if="row.type === 'MENU'">{{ row.uiPath }}</template>
            <template v-else>
              <el-tag size="small" class="method-tag">{{ row.httpMethod }}</el-tag>
              {{ row.httpPath }}
            </template>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.field.status')" width="90" align="center">
          <template #default="{ row }">
            <el-switch
              v-permission="'system:permission:status'"
              :model-value="row.status"
              active-value="ENABLED"
              inactive-value="DISABLED"
              :before-change="
                () =>
                  confirmStatusChange(
                    (row as PermissionItem).name,
                    (row as PermissionItem).status === 'ENABLED' ? 'DISABLED' : 'ENABLED',
                  )
              "
              @change="
                (value: string | number | boolean) =>
                  handleStatusChange(row as PermissionItem, value as Status)
              "
            />
          </template>
        </el-table-column>
        <el-table-column
          prop="remark"
          :label="t('common.field.remark')"
          min-width="160"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.field.createdAt')" width="170" class-name="table-time">
          <template #default="{ row }">
            {{ formatDateTime((row as PermissionItem).createTime) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.field.actions')" width="175" fixed="right">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button
                v-permission="'system:permission:update'"
                link
                type="primary"
                @click="openEdit(row as PermissionItem)"
              >
                {{ t('system.permission.actions.edit') }}
              </el-button>
              <el-button
                v-permission="'system:permission:delete'"
                link
                type="danger"
                @click="handleDelete(row as PermissionItem)"
              >
                {{ t('system.permission.actions.delete') }}
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
          ? t('system.permission.dialog.createTitle')
          : t('system.permission.dialog.editTitle')
      "
      width="560px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="130px"
        :validate-on-rule-change="false"
      >
        <el-form-item :label="t('system.permission.dialog.type')" prop="type">
          <el-radio-group v-model="form.type" :disabled="editingId !== null">
            <el-radio value="MENU">{{ t('system.permission.type.menu') }}</el-radio>
            <el-radio value="API">{{ t('system.permission.type.api') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="t('system.permission.dialog.code')" prop="code">
          <el-input
            v-model="form.code"
            :disabled="editingId !== null"
            :placeholder="t('system.permission.dialog.codePlaceholder')"
            maxlength="96"
          />
        </el-form-item>
        <el-form-item :label="t('system.permission.dialog.name')" prop="name">
          <el-input v-model="form.name" maxlength="32" />
        </el-form-item>
        <el-form-item
          v-if="form.type === 'MENU'"
          :label="t('system.permission.dialog.uiPath')"
          prop="uiPath"
        >
          <el-input
            v-model="form.uiPath"
            :placeholder="t('system.permission.dialog.uiPathPlaceholder')"
            maxlength="191"
          />
        </el-form-item>
        <template v-else>
          <el-form-item :label="t('system.permission.dialog.httpMethod')" prop="httpMethod">
            <el-select v-model="form.httpMethod" class="method-select">
              <el-option
                v-for="method in ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']"
                :key="method"
                :label="method"
                :value="method"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('system.permission.dialog.httpPath')" prop="httpPath">
            <el-input
              v-model="form.httpPath"
              :placeholder="t('system.permission.dialog.httpPathPlaceholder')"
              maxlength="191"
            />
            <span class="form-hint">{{ t('system.permission.dialog.httpPathHint') }}</span>
          </el-form-item>
        </template>
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

<style scoped>
.method-tag {
  margin-right: 6px;
}

.method-select {
  width: 100%;
}

.form-hint {
  display: block;
  margin-top: 4px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.4;
}
</style>
