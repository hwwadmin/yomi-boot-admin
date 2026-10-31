<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'

import { Avatar, Plus, Refresh, Search } from '@element-plus/icons-vue'

import {
  bindRoleGroups,
  createRole,
  deleteRole,
  getRoleDetail,
  getRolePage,
  listAllPermissionGroups,
  listAllPermissions,
  refreshRbacCache,
  setRoleStatus,
  updateRole,
} from '@/api/rbac'
import type { RoleQuery } from '@/api/rbac'
import type { PermissionGroupItem, PermissionItem, RoleItem, Status } from '@/api/types'
import NameCell from '@/components/NameCell.vue'
import PageCard from '@/components/PageCard.vue'
import PageHeader from '@/components/PageHeader.vue'
import { confirmAction } from '@/utils/confirm'
import { formatDateTime } from '@/utils/datetime'
import { newIdempotencyKey } from '@/utils/idempotency'

defineOptions({ name: 'SystemRolePage' })

const { t } = useI18n()

// ———————— 列表查询 ————————
const loading = ref(false)
const list = ref<RoleItem[]>([])
const total = ref(0)
const query = reactive<RoleQuery>({ page: 1, pageSize: 10, keyword: '', status: '' })

async function fetchList() {
  loading.value = true
  try {
    const result = await getRolePage(query)
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

// ———————— 刷新鉴权缓存 ————————
// 角色、权限组与权限的变更不会自动生效，需从数据库重建一次 RBAC 投影
const refreshingCache = ref(false)

async function handleRefreshCache() {
  const confirmed = await confirmAction({
    title: t('system.role.cacheRefresh.title'),
    message: t('system.role.cacheRefresh.message'),
    confirmButtonText: t('system.role.cacheRefresh.confirm'),
  })
  if (!confirmed) return

  refreshingCache.value = true
  try {
    await refreshRbacCache()
    ElMessage.success(t('system.role.cacheRefresh.success'))
  } finally {
    refreshingCache.value = false
  }
}

// ———————— 新增 / 编辑 ————————
const dialogVisible = ref(false)
const submitting = ref(false)
/** 一次创建意图的幂等键：超时重试复用，成功后清空 */
const submitKey = ref('')
const editingId = ref<string | null>(null)
const formRef = ref<FormInstance>()
const form = reactive({
  code: '',
  name: '',
  remark: '',
})

const rules = computed<FormRules>(() => ({
  code: [
    { required: true, message: t('system.role.validation.codeRequired'), trigger: 'blur' },
    {
      pattern: /^[a-z][a-z0-9_]{2,31}$/,
      message: t('validation.roleCodePattern'),
      trigger: 'blur',
    },
  ],
  name: [{ required: true, message: t('system.role.validation.nameRequired'), trigger: 'blur' }],
}))

function openCreate() {
  editingId.value = null
  submitKey.value = ''
  Object.assign(form, { code: '', name: '', remark: '' })
  dialogVisible.value = true
}

function openEdit(row: RoleItem) {
  editingId.value = row.id
  submitKey.value = ''
  Object.assign(form, { code: row.code, name: row.name, remark: row.remark })
  dialogVisible.value = true
}

/** 启停是危险操作：确认后再提交 */
function confirmStatusChange(name: string, next: Status): Promise<boolean> {
  const disabling = next === 'DISABLED'
  return confirmAction({
    title: t(
      disabling
        ? 'system.role.statusConfirm.disableTitle'
        : 'system.role.statusConfirm.enableTitle',
    ),
    message: t(
      disabling
        ? 'system.role.statusConfirm.disableMessage'
        : 'system.role.statusConfirm.enableMessage',
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
      if (!submitKey.value) submitKey.value = newIdempotencyKey()
      await createRole(
        {
          code: form.code.trim(),
          name: form.name.trim(),
          remark: form.remark.trim(),
        },
        submitKey.value,
      )
      submitKey.value = ''
      ElMessage.success(t('system.role.createSuccess'))
    } else {
      await updateRole(editingId.value, { name: form.name.trim(), remark: form.remark.trim() })
      ElMessage.success(t('system.role.saveSuccess'))
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitting.value = false
  }
}

async function handleStatusChange(row: RoleItem, status: Status) {
  try {
    await setRoleStatus(row.id, status)
    row.status = status
    ElMessage.success(t('system.role.statusUpdated'))
  } catch {
    fetchList()
  }
}

async function handleDelete(row: RoleItem) {
  const confirmed = await confirmAction({
    title: t('system.role.deleteConfirm.title'),
    message: t('system.role.deleteConfirm.message', { name: row.name }),
    confirmButtonText: t('system.role.deleteConfirm.confirm'),
  })
  if (!confirmed) return
  await deleteRole(row.id)
  ElMessage.success(t('system.role.deleteSuccess'))
  if (list.value.length === 1 && query.page > 1) {
    query.page -= 1
  }
  fetchList()
}

// ———————— 分配权限组 ————————
const groupDrawerVisible = ref(false)
const groupLoading = ref(false)
const groupSaving = ref(false)
const groupTarget = ref<RoleItem | null>(null)
const groups = ref<PermissionGroupItem[]>([])
const permissions = ref<PermissionItem[]>([])
const selectedGroupIds = ref<string[]>([])
const aggregatedPermissionIds = ref<string[]>([])
const effectivePermissionIds = ref<string[]>([])

/** 角色权限来自权限组的去重聚合，这里只做只读展示 */
const aggregatedPermissions = computed(() =>
  permissions.value.filter((item) => aggregatedPermissionIds.value.includes(item.id)),
)

/** 判断权限是否真正生效：只有启用权限组下的启用权限才算 */
function isPermissionEffective(id: string): boolean {
  return effectivePermissionIds.value.includes(id)
}

/** 未生效的聚合权限数量，用于提示文案 */
const ineffectiveCount = computed(
  () => aggregatedPermissions.value.filter((item) => !isPermissionEffective(item.id)).length,
)

async function openGroupDrawer(row: RoleItem) {
  groupTarget.value = row
  groupDrawerVisible.value = true
  groupLoading.value = true
  try {
    const [detail, allGroups, allPermissions] = await Promise.all([
      getRoleDetail(row.id),
      listAllPermissionGroups(),
      listAllPermissions(),
    ])
    groups.value = allGroups
    permissions.value = allPermissions
    // 后端契约保证数组字段不为 null，这里兜底防御，避免异常响应冻结 loading 遮罩
    selectedGroupIds.value = detail.groupIds ?? []
    aggregatedPermissionIds.value = detail.permissionIds ?? []
    effectivePermissionIds.value = detail.effectivePermissionIds ?? []
  } finally {
    groupLoading.value = false
  }
}

async function saveGroups() {
  if (!groupTarget.value) return
  groupSaving.value = true
  try {
    await bindRoleGroups(groupTarget.value.id, selectedGroupIds.value)
    ElMessage.success(t('system.role.groupDrawer.saved'))
    groupDrawerVisible.value = false
  } finally {
    groupSaving.value = false
  }
}
</script>

<template>
  <div class="system-role">
    <PageHeader :title="t('menu.systemRole')" :subtitle="t('system.role.subtitle')">
      <template #actions>
        <el-button
          v-permission="'system:role:create'"
          type="primary"
          :icon="Plus"
          @click="openCreate"
        >
          {{ t('system.role.add') }}
        </el-button>
      </template>
    </PageHeader>

    <PageCard>
      <template #toolbar>
        <div class="page-toolbar">
          <div class="page-toolbar__filters">
            <el-input
              v-model="query.keyword"
              class="page-toolbar__input"
              :placeholder="t('system.role.searchPlaceholder')"
              :suffix-icon="Search"
              clearable
              @keyup.enter="handleSearch"
              @clear="handleSearch"
            />
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
              v-permission="'system:rbac_cache:refresh'"
              :icon="Refresh"
              :loading="refreshingCache"
              @click="handleRefreshCache"
            >
              {{ t('system.role.refreshCache') }}
            </el-button>
          </div>
        </div>
      </template>

      <el-table v-loading="loading" :data="list">
        <el-table-column :label="t('system.role.columns.name')" min-width="180">
          <template #default="{ row }">
            <NameCell :label="(row as RoleItem).name" :icon="Avatar" tone="violet" />
          </template>
        </el-table-column>
        <el-table-column
          prop="code"
          :label="t('system.role.columns.code')"
          min-width="170"
          class-name="table-code"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.field.builtin')" width="90" align="center">
          <template #default="{ row }">
            {{ row.builtin ? t('common.yes') : t('common.no') }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.field.status')" width="90" align="center">
          <template #default="{ row }">
            <el-switch
              v-permission="'system:role:status'"
              :model-value="row.status"
              active-value="ENABLED"
              inactive-value="DISABLED"
              :disabled="row.builtin"
              :before-change="
                () =>
                  confirmStatusChange(
                    (row as RoleItem).name,
                    (row as RoleItem).status === 'ENABLED' ? 'DISABLED' : 'ENABLED',
                  )
              "
              @change="
                (value: string | number | boolean) =>
                  handleStatusChange(row as RoleItem, value as Status)
              "
            />
          </template>
        </el-table-column>
        <el-table-column
          prop="remark"
          :label="t('common.field.remark')"
          min-width="190"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.field.createdAt')" width="170" class-name="table-time">
          <template #default="{ row }">
            {{ formatDateTime((row as RoleItem).createTime) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.field.actions')" width="220" fixed="right">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button
                v-permission="'system:role:update'"
                link
                type="primary"
                @click="openEdit(row as RoleItem)"
              >
                {{ t('system.role.actions.edit') }}
              </el-button>
              <el-button
                v-if="!row.builtin"
                v-permission="'system:role:delete'"
                link
                type="danger"
                @click="handleDelete(row as RoleItem)"
              >
                {{ t('system.role.actions.delete') }}
              </el-button>
              <el-button
                v-if="!row.builtin"
                v-permission="'system:role:bind-group'"
                link
                type="primary"
                @click="openGroupDrawer(row as RoleItem)"
              >
                {{ t('system.role.actions.assignGroups') }}
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
        editingId === null ? t('system.role.dialog.createTitle') : t('system.role.dialog.editTitle')
      "
      width="480px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
        :validate-on-rule-change="false"
      >
        <el-form-item :label="t('system.role.dialog.code')" prop="code">
          <el-input
            v-model="form.code"
            :disabled="editingId !== null"
            :placeholder="t('system.role.dialog.codePlaceholder')"
            maxlength="32"
          />
        </el-form-item>
        <el-form-item :label="t('system.role.dialog.name')" prop="name">
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
      v-model="groupDrawerVisible"
      :title="t('system.role.groupDrawer.title')"
      size="560px"
    >
      <div v-loading="groupLoading" class="group-body">
        <div class="group-meta">
          {{ t('system.role.groupDrawer.roleLabel') }}<strong>{{ groupTarget?.name }}</strong>
          <span class="muted">{{ groupTarget?.code }}</span>
        </div>
        <el-checkbox-group v-model="selectedGroupIds" class="group-list">
          <el-checkbox
            v-for="group in groups"
            :key="group.id"
            :value="group.id"
            :disabled="group.status === 'DISABLED' && !selectedGroupIds.includes(group.id)"
          >
            {{ group.name }}
            <span class="muted">{{ group.code }}</span>
            <el-tag v-if="group.status === 'DISABLED'" type="info" size="small">
              {{ t('system.role.groupDrawer.disabledTag') }}
            </el-tag>
          </el-checkbox>
        </el-checkbox-group>
        <span class="muted hint">{{ t('system.role.groupDrawer.hint') }}</span>
        <el-empty
          v-if="!groupLoading && groups.length === 0"
          :description="t('system.role.groupDrawer.empty')"
          :image-size="80"
        />
        <el-divider content-position="left">
          {{ t('system.role.groupDrawer.aggregated', { count: aggregatedPermissions.length }) }}
        </el-divider>
        <div class="permission-tags">
          <el-tag
            v-for="item in aggregatedPermissions"
            :key="item.id"
            size="small"
            :type="isPermissionEffective(item.id) ? 'info' : 'warning'"
          >
            {{ item.name }}
            <span v-if="!isPermissionEffective(item.id)">
              {{ t('system.role.groupDrawer.ineffectiveTag') }}
            </span>
          </el-tag>
          <span v-if="aggregatedPermissions.length === 0" class="muted">
            {{ t('system.role.groupDrawer.noGroups') }}
          </span>
        </div>
        <span v-if="ineffectiveCount > 0" class="muted hint">
          {{ t('system.role.groupDrawer.ineffective', { count: ineffectiveCount }) }}
        </span>
      </div>
      <template #footer>
        <el-button @click="groupDrawerVisible = false">{{ t('common.action.cancel') }}</el-button>
        <el-button type="primary" :loading="groupSaving" @click="saveGroups">
          {{ t('common.action.save') }}
        </el-button>
      </template>
    </el-drawer>
  </div>
</template>

<style scoped>
.group-body {
  min-height: 320px;
}

.group-meta {
  margin-bottom: 12px;
  color: var(--el-text-color-regular);
}

.group-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.permission-tags {
  display: flex;
  flex-wrap: wrap;
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
