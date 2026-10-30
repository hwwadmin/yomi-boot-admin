<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'

import {
  bindGroupPermissions,
  createPermissionGroup,
  deletePermissionGroup,
  getPermissionGroupDetail,
  getPermissionGroupPage,
  listAllPermissions,
  setPermissionGroupStatus,
  updatePermissionGroup,
} from '@/api/rbac'
import type { PermissionGroupQuery } from '@/api/rbac'
import type { PermissionGroupItem, PermissionItem, Status } from '@/api/types'
import { confirmAction } from '@/utils/confirm'
import { formatDateTime } from '@/utils/datetime'

defineOptions({ name: 'SystemPermissionGroupPage' })

const { t } = useI18n()

// ———————— 列表查询 ————————
const loading = ref(false)
const list = ref<PermissionGroupItem[]>([])
const total = ref(0)
const query = reactive<PermissionGroupQuery>({
  page: 1,
  pageSize: 10,
  keyword: '',
  status: '',
})

async function fetchList() {
  loading.value = true
  try {
    const result = await getPermissionGroupPage(query)
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
  remark: '',
})

const rules = computed<FormRules>(() => ({
  code: [
    {
      required: true,
      message: t('system.permissionGroup.validation.codeRequired'),
      trigger: 'blur',
    },
    {
      pattern: /^[a-z][a-z0-9_]{2,31}$/,
      message: t('validation.roleCodePattern'),
      trigger: 'blur',
    },
  ],
  name: [
    {
      required: true,
      message: t('system.permissionGroup.validation.nameRequired'),
      trigger: 'blur',
    },
  ],
}))

function openCreate() {
  editingId.value = null
  Object.assign(form, { code: '', name: '', remark: '' })
  dialogVisible.value = true
}

function openEdit(row: PermissionGroupItem) {
  editingId.value = row.id
  Object.assign(form, { code: row.code, name: row.name, remark: row.remark })
  dialogVisible.value = true
}

/** 启停是危险操作：确认后再提交 */
function confirmStatusChange(name: string, next: Status): Promise<boolean> {
  const disabling = next === 'DISABLED'
  return confirmAction({
    title: t(
      disabling
        ? 'system.permissionGroup.statusConfirm.disableTitle'
        : 'system.permissionGroup.statusConfirm.enableTitle',
    ),
    message: t(
      disabling
        ? 'system.permissionGroup.statusConfirm.disableMessage'
        : 'system.permissionGroup.statusConfirm.enableMessage',
      { name },
    ),
  })
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    if (editingId.value === null) {
      await createPermissionGroup({
        code: form.code.trim(),
        name: form.name.trim(),
        remark: form.remark.trim(),
      })
      ElMessage.success(t('system.permissionGroup.createSuccess'))
    } else {
      await updatePermissionGroup(editingId.value, {
        name: form.name.trim(),
        remark: form.remark.trim(),
      })
      ElMessage.success(t('system.permissionGroup.saveSuccess'))
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitting.value = false
  }
}

async function handleStatusChange(row: PermissionGroupItem, status: Status) {
  try {
    await setPermissionGroupStatus(row.id, status)
    row.status = status
    ElMessage.success(t('system.permissionGroup.statusUpdated'))
  } catch {
    fetchList()
  }
}

async function handleDelete(row: PermissionGroupItem) {
  const confirmed = await confirmAction({
    title: t('system.permissionGroup.deleteConfirm.title'),
    message: t('system.permissionGroup.deleteConfirm.message', { name: row.name }),
    confirmButtonText: t('system.permissionGroup.deleteConfirm.confirm'),
  })
  if (!confirmed) return
  await deletePermissionGroup(row.id)
  ElMessage.success(t('system.permissionGroup.deleteSuccess'))
  if (list.value.length === 1 && query.page > 1) {
    query.page -= 1
  }
  fetchList()
}

// ———————— 分配权限 ————————
const permissionDrawerVisible = ref(false)
const permissionLoading = ref(false)
const permissionSaving = ref(false)
const permissionTarget = ref<PermissionGroupItem | null>(null)
const permissions = ref<PermissionItem[]>([])
const selectedPermissionIds = ref<string[]>([])
const permissionTab = ref<'MENU' | 'API'>('MENU')
const permissionKeyword = ref('')

/** 按名称、编码、路径做前端模糊匹配 */
function matchPermission(item: PermissionItem, keyword: string): boolean {
  const normalized = keyword.trim().toLowerCase()
  if (!normalized) return true
  return [item.name, item.code, item.uiPath, item.httpPath].some((value) =>
    (value ?? '').toLowerCase().includes(normalized),
  )
}

const menuPermissions = computed(() =>
  permissions.value.filter(
    (item) => item.type === 'MENU' && matchPermission(item, permissionKeyword.value),
  ),
)
const apiPermissions = computed(() =>
  permissions.value.filter(
    (item) => item.type === 'API' && matchPermission(item, permissionKeyword.value),
  ),
)

async function openPermissionDrawer(row: PermissionGroupItem) {
  permissionTarget.value = row
  permissionDrawerVisible.value = true
  permissionLoading.value = true
  permissionTab.value = 'MENU'
  permissionKeyword.value = ''
  try {
    const [detail, allPermissions] = await Promise.all([
      getPermissionGroupDetail(row.id),
      listAllPermissions(),
    ])
    permissions.value = allPermissions
    selectedPermissionIds.value = detail.permissionIds
  } finally {
    permissionLoading.value = false
  }
}

async function savePermissions() {
  if (!permissionTarget.value) return
  permissionSaving.value = true
  try {
    await bindGroupPermissions(permissionTarget.value.id, selectedPermissionIds.value)
    ElMessage.success(t('system.permissionGroup.permissionDrawer.saved'))
    permissionDrawerVisible.value = false
  } finally {
    permissionSaving.value = false
  }
}
</script>

<template>
  <div class="system-permission-group">
    <el-card shadow="never">
      <div class="toolbar">
        <el-input
          v-model="query.keyword"
          class="toolbar-input"
          :placeholder="t('system.permissionGroup.searchPlaceholder')"
          clearable
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />
        <el-select
          v-model="query.status"
          class="toolbar-select"
          :placeholder="t('common.all')"
          clearable
        >
          <el-option :label="t('common.status.enabled')" value="ENABLED" />
          <el-option :label="t('common.status.disabled')" value="DISABLED" />
        </el-select>
        <el-button type="primary" @click="handleSearch">{{ t('common.action.search') }}</el-button>
        <el-button @click="handleReset">{{ t('common.action.reset') }}</el-button>
        <el-button
          v-permission="'system:permission-group:create'"
          type="primary"
          plain
          class="toolbar-add"
          @click="openCreate"
        >
          {{ t('system.permissionGroup.add') }}
        </el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column
          prop="name"
          :label="t('system.permissionGroup.columns.name')"
          min-width="140"
        />
        <el-table-column
          prop="code"
          :label="t('system.permissionGroup.columns.code')"
          min-width="160"
        />
        <el-table-column :label="t('common.field.status')" width="100" align="center">
          <template #default="{ row }">
            <el-switch
              v-permission="'system:permission-group:status'"
              :model-value="row.status"
              active-value="ENABLED"
              inactive-value="DISABLED"
              :before-change="
                () =>
                  confirmStatusChange(
                    (row as PermissionGroupItem).name,
                    (row as PermissionGroupItem).status === 'ENABLED' ? 'DISABLED' : 'ENABLED',
                  )
              "
              @change="
                (value: string | number | boolean) =>
                  handleStatusChange(row as PermissionGroupItem, value as Status)
              "
            />
          </template>
        </el-table-column>
        <el-table-column
          prop="remark"
          :label="t('common.field.remark')"
          min-width="200"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.field.createdAt')" width="180">
          <template #default="{ row }">
            {{ formatDateTime((row as PermissionGroupItem).createTime) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.field.actions')" width="270" fixed="right" align="center">
          <template #default="{ row }">
            <el-button
              v-permission="'system:permission-group:update'"
              link
              type="primary"
              @click="openEdit(row as PermissionGroupItem)"
            >
              {{ t('system.permissionGroup.actions.edit') }}
            </el-button>
            <el-button
              v-permission="'system:permission-group:delete'"
              link
              type="danger"
              @click="handleDelete(row as PermissionGroupItem)"
            >
              {{ t('system.permissionGroup.actions.delete') }}
            </el-button>
            <el-button
              v-permission="'system:permission-group:bind-permission'"
              link
              type="primary"
              @click="openPermissionDrawer(row as PermissionGroupItem)"
            >
              {{ t('system.permissionGroup.actions.assignPermissions') }}
            </el-button>
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

    <el-dialog
      v-model="dialogVisible"
      :title="
        editingId === null
          ? t('system.permissionGroup.dialog.createTitle')
          : t('system.permissionGroup.dialog.editTitle')
      "
      width="480px"
      :close-on-click-modal="false"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="130px">
        <el-form-item :label="t('system.permissionGroup.dialog.code')" prop="code">
          <el-input
            v-model="form.code"
            :disabled="editingId !== null"
            :placeholder="t('system.permissionGroup.dialog.codePlaceholder')"
            maxlength="32"
          />
        </el-form-item>
        <el-form-item :label="t('system.permissionGroup.dialog.name')" prop="name">
          <el-input v-model="form.name" maxlength="32" />
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

    <el-drawer
      v-model="permissionDrawerVisible"
      :title="t('system.permissionGroup.permissionDrawer.title')"
      size="560px"
    >
      <div v-loading="permissionLoading" class="permission-body">
        <div class="permission-meta">
          {{ t('system.permissionGroup.permissionDrawer.groupLabel') }}
          <strong>{{ permissionTarget?.name }}</strong>
          <span class="muted">{{ permissionTarget?.code }}</span>
          <span class="muted">
            {{
              t('system.permissionGroup.permissionDrawer.selected', {
                count: selectedPermissionIds.length,
              })
            }}
          </span>
        </div>
        <el-input
          v-model="permissionKeyword"
          class="permission-search"
          :placeholder="t('system.permissionGroup.permissionDrawer.searchPlaceholder')"
          clearable
        />
        <el-tabs v-model="permissionTab">
          <el-tab-pane
            :label="t('system.permissionGroup.permissionDrawer.menuTab')"
            name="MENU"
          >
            <el-checkbox-group v-model="selectedPermissionIds" class="permission-list">
              <el-checkbox
                v-for="item in menuPermissions"
                :key="item.id"
                :value="item.id"
                :disabled="item.status === 'DISABLED' && !selectedPermissionIds.includes(item.id)"
              >
                {{ item.name }}
                <span class="muted">{{ item.uiPath }}</span>
                <el-tag v-if="item.status === 'DISABLED'" type="info" size="small">
                  {{ t('system.permissionGroup.permissionDrawer.disabledTag') }}
                </el-tag>
              </el-checkbox>
            </el-checkbox-group>
            <el-empty
              v-if="!permissionLoading && menuPermissions.length === 0"
              :description="
                permissionKeyword
                  ? t('system.permissionGroup.permissionDrawer.menuEmptySearch')
                  : t('system.permissionGroup.permissionDrawer.menuEmpty')
              "
              :image-size="60"
            />
          </el-tab-pane>
          <el-tab-pane
            :label="t('system.permissionGroup.permissionDrawer.apiTab')"
            name="API"
          >
            <el-checkbox-group v-model="selectedPermissionIds" class="permission-list">
              <el-checkbox
                v-for="item in apiPermissions"
                :key="item.id"
                :value="item.id"
                :disabled="item.status === 'DISABLED' && !selectedPermissionIds.includes(item.id)"
              >
                {{ item.name }}
                <span class="muted">{{ item.httpMethod }} {{ item.httpPath }}</span>
                <el-tag v-if="item.status === 'DISABLED'" type="info" size="small">
                  {{ t('system.permissionGroup.permissionDrawer.disabledTag') }}
                </el-tag>
              </el-checkbox>
            </el-checkbox-group>
            <el-empty
              v-if="!permissionLoading && apiPermissions.length === 0"
              :description="
                permissionKeyword
                  ? t('system.permissionGroup.permissionDrawer.apiEmptySearch')
                  : t('system.permissionGroup.permissionDrawer.apiEmpty')
              "
              :image-size="60"
            />
          </el-tab-pane>
        </el-tabs>
        <span class="hint">{{ t('system.permissionGroup.permissionDrawer.hint') }}</span>
      </div>
      <template #footer>
        <el-button @click="permissionDrawerVisible = false">
          {{ t('common.action.cancel') }}
        </el-button>
        <el-button type="primary" :loading="permissionSaving" @click="savePermissions">
          {{ t('common.action.save') }}
        </el-button>
      </template>
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
  width: 220px;
}

.toolbar-select {
  width: 140px;
}

.toolbar-add {
  margin-left: auto;
}

.page-pagination {
  justify-content: flex-end;
  margin-top: 16px;
}

.permission-body {
  min-height: 320px;
}

.permission-meta {
  margin-bottom: 12px;
  color: var(--el-text-color-regular);
}

.permission-search {
  margin-bottom: 4px;
}

.permission-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.muted {
  margin-left: 6px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.hint {
  display: block;
  margin: 8px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
</style>
