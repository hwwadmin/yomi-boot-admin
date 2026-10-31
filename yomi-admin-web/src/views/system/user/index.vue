<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { Plus, Refresh, Search, User } from '@element-plus/icons-vue'

import { bindUserRoles, getUserRoles } from '@/api/rbac'
import {
  createUser,
  deleteUser,
  getUserPage,
  getUserSessions,
  resetUserPassword,
  revokeUserSession,
  revokeUserSessions,
  updateUser,
} from '@/api/user'
import type { RoleItem, UserItem, UserSessionItem, UserStatus } from '@/api/types'
import NameCell from '@/components/NameCell.vue'
import PageCard from '@/components/PageCard.vue'
import PageHeader from '@/components/PageHeader.vue'
import { confirmAction } from '@/utils/confirm'
import { formatDateTime } from '@/utils/datetime'
import { newIdempotencyKey } from '@/utils/idempotency'

defineOptions({ name: 'SystemUserPage' })

const { t } = useI18n()

// ———————— 列表查询 ————————
const loading = ref(false)
const list = ref<UserItem[]>([])
const total = ref(0)
const query = reactive({ page: 1, pageSize: 10, keyword: '' })

async function fetchList() {
  loading.value = true
  try {
    const result = await getUserPage(query)
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
  handleSearch()
}

onMounted(fetchList)

// ———————— 新增 / 编辑 ————————
const dialogVisible = ref(false)
const submitting = ref(false)
/** 一次创建意图的幂等键：超时重试复用，成功后清空 */
const submitKey = ref('')
const editingId = ref<string | null>(null)
const editingOriginalStatus = ref<UserStatus | null>(null)
const formRef = ref<FormInstance>()
const form = reactive({
  username: '',
  password: '',
  nickname: '',
  email: '',
  phone: '',
  status: 'ENABLED' as UserStatus,
})

const rules = computed<FormRules>(() => ({
  username: [
    { required: true, message: t('system.user.validation.usernameRequired'), trigger: 'blur' },
  ],
  password: [
    { required: true, message: t('system.user.validation.passwordRequired'), trigger: 'blur' },
    { min: 6, max: 72, message: t('validation.passwordLength'), trigger: 'blur' },
  ],
  nickname: [
    { required: true, message: t('system.user.validation.nicknameRequired'), trigger: 'blur' },
  ],
  email: [{ type: 'email', message: t('validation.email'), trigger: 'blur' }],
  phone: [{ pattern: /^1\d{10}$/, message: t('validation.phone'), trigger: 'blur' }],
}))

function openCreate() {
  editingId.value = null
  editingOriginalStatus.value = null
  submitKey.value = ''
  Object.assign(form, {
    username: '',
    password: '',
    nickname: '',
    email: '',
    phone: '',
    status: 'ENABLED',
  })
  dialogVisible.value = true
}

function openEdit(row: UserItem) {
  editingId.value = row.id
  editingOriginalStatus.value = row.status
  submitKey.value = ''
  Object.assign(form, {
    username: row.username,
    password: '',
    nickname: row.nickname,
    email: row.email,
    phone: row.phone,
    status: row.status,
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (editingOriginalStatus.value && form.status !== editingOriginalStatus.value) {
    const disabling = form.status === 'DISABLED'
    const confirmed = await confirmAction({
      title: t(
        disabling
          ? 'system.user.statusConfirm.disableTitle'
          : 'system.user.statusConfirm.enableTitle',
      ),
      message: t(
        disabling
          ? 'system.user.statusConfirm.disableMessage'
          : 'system.user.statusConfirm.enableMessage',
        { name: form.nickname.trim() },
      ),
    })
    if (!confirmed) return
  }
  submitting.value = true
  try {
    const payload = {
      username: form.username.trim(),
      nickname: form.nickname.trim(),
      email: form.email.trim(),
      phone: form.phone.trim(),
      status: form.status,
    }
    if (editingId.value === null) {
      if (!submitKey.value) submitKey.value = newIdempotencyKey()
      await createUser({ ...payload, password: form.password }, submitKey.value)
      submitKey.value = ''
      ElMessage.success(t('system.user.createSuccess'))
    } else {
      await updateUser(editingId.value, payload)
      ElMessage.success(t('system.user.saveSuccess'))
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitting.value = false
  }
}

// ———————— 分配角色 ————————
const roleDialogVisible = ref(false)
const roleDialogLoading = ref(false)
const roleDialogSaving = ref(false)
const roleTarget = ref<UserItem | null>(null)
const roleOptions = ref<RoleItem[]>([])
const selectedRoleIds = ref<string[]>([])

async function openRoleDialog(row: UserItem) {
  roleTarget.value = row
  roleDialogVisible.value = true
  roleDialogLoading.value = true
  try {
    const binding = await getUserRoles(row.id)
    roleOptions.value = binding.roles
    selectedRoleIds.value = binding.roleIds
  } finally {
    roleDialogLoading.value = false
  }
}

async function saveRoles() {
  if (!roleTarget.value) return
  roleDialogSaving.value = true
  try {
    await bindUserRoles(roleTarget.value.id, selectedRoleIds.value)
    ElMessage.success(t('system.user.roleDialog.saved'))
    roleDialogVisible.value = false
    fetchList()
  } finally {
    roleDialogSaving.value = false
  }
}

// ———————— 删除 ————————
async function handleDelete(row: UserItem) {
  const confirmed = await confirmAction({
    title: t('system.user.deleteConfirm.title'),
    message: t('system.user.deleteConfirm.message', { name: row.nickname }),
    confirmButtonText: t('system.user.deleteConfirm.confirm'),
  })
  if (!confirmed) return
  await deleteUser(row.id)
  ElMessage.success(t('system.user.deleteSuccess'))
  // 当前页只剩一条时回退一页，避免删除后停留在空页
  if (list.value.length === 1 && query.page > 1) {
    query.page -= 1
  }
  fetchList()
}

// ———————— 重置密码（仅超级管理员） ————————
const resetPwdVisible = ref(false)
const resetPwdSubmitting = ref(false)
const resetPwdTarget = ref<UserItem | null>(null)
const resetPwdRef = ref<FormInstance>()
const resetPwdForm = reactive({ newPassword: '' })

const resetPwdRules = computed<FormRules>(() => ({
  newPassword: [
    {
      required: true,
      message: t('system.user.validation.newPasswordRequired'),
      trigger: 'blur',
    },
    { min: 6, max: 72, message: t('validation.passwordLength'), trigger: 'blur' },
  ],
}))

function openResetPassword(row: UserItem) {
  resetPwdTarget.value = row
  resetPwdForm.newPassword = ''
  resetPwdVisible.value = true
}

async function handleResetPassword() {
  const valid = await resetPwdRef.value?.validate().catch(() => false)
  if (!valid || !resetPwdTarget.value) return
  resetPwdSubmitting.value = true
  try {
    await resetUserPassword(resetPwdTarget.value.id, resetPwdForm.newPassword)
    ElMessage.success(t('system.user.resetPassword.success'))
    resetPwdVisible.value = false
  } finally {
    resetPwdSubmitting.value = false
  }
}

// ———————— 在线会话 ————————
const sessionDrawerVisible = ref(false)
const sessionLoading = ref(false)
const sessionKicking = ref(false)
const sessionTarget = ref<UserItem | null>(null)
const sessions = ref<UserSessionItem[]>([])
const hasCurrentSession = computed(() => sessions.value.some((session) => session.current))

function sessionName() {
  return sessionTarget.value?.nickname || sessionTarget.value?.username || ''
}

async function fetchSessions() {
  if (!sessionTarget.value) return
  sessionLoading.value = true
  try {
    sessions.value = await getUserSessions(sessionTarget.value.id)
  } finally {
    sessionLoading.value = false
  }
}

function openSessions(row: UserItem) {
  sessionTarget.value = row
  sessions.value = []
  sessionDrawerVisible.value = true
  fetchSessions()
}

async function handleRevokeSession(session: UserSessionItem) {
  if (!sessionTarget.value || session.current) return
  const confirmed = await confirmAction({
    title: t('system.user.session.revokeConfirm.title'),
    message: t('system.user.session.revokeConfirm.message', {
      name: sessionName(),
      ip: session.loginIp || '-',
    }),
    confirmButtonText: t('system.user.session.revokeConfirm.confirm'),
  })
  if (!confirmed) return
  sessionKicking.value = true
  try {
    await revokeUserSession(sessionTarget.value.id, session.sessionId)
    ElMessage.success(t('system.user.session.revoked'))
    await fetchSessions()
  } finally {
    sessionKicking.value = false
  }
}

async function handleRevokeAllSessions() {
  if (!sessionTarget.value || !sessions.value.length || hasCurrentSession.value) return
  const confirmed = await confirmAction({
    title: t('system.user.session.revokeAllConfirm.title'),
    message: t('system.user.session.revokeAllConfirm.message', {
      name: sessionName(),
      count: sessions.value.length,
    }),
    confirmButtonText: t('system.user.session.revokeAllConfirm.confirm'),
  })
  if (!confirmed) return
  sessionKicking.value = true
  try {
    await revokeUserSessions(sessionTarget.value.id)
    ElMessage.success(t('system.user.session.revokedAll'))
    await fetchSessions()
  } finally {
    sessionKicking.value = false
  }
}
</script>

<template>
  <div class="system-user">
    <PageHeader :title="t('menu.systemUser')" :subtitle="t('system.user.subtitle')">
      <template #actions>
        <el-button
          v-permission="'system:user:create'"
          type="primary"
          :icon="Plus"
          @click="openCreate"
        >
          {{ t('system.user.add') }}
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
              :placeholder="t('system.user.searchPlaceholder')"
              :prefix-icon="Search"
              clearable
              @keyup.enter="handleSearch"
              @clear="handleSearch"
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
        <el-table-column :label="t('common.field.username')" min-width="150">
          <template #default="{ row }">
            <NameCell :label="(row as UserItem).username" :icon="User" tone="blue" />
          </template>
        </el-table-column>
        <el-table-column prop="nickname" :label="t('common.field.nickname')" min-width="110" />
        <el-table-column :label="t('system.user.columns.roles')" min-width="120">
          <template #default="{ row }">
            <el-tag
              v-for="role in (row as UserItem).roles"
              :key="role.id"
              class="role-tag"
              :type="role.status === 'ENABLED' ? 'success' : 'info'"
            >
              {{ role.name }}
            </el-tag>
            <span v-if="!(row as UserItem).roles?.length">-</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="email"
          :label="t('common.field.email')"
          min-width="140"
          show-overflow-tooltip
        />
        <el-table-column prop="phone" :label="t('common.field.phone')" min-width="130" />
        <el-table-column :label="t('common.field.status')" width="80" align="center">
          <template #default="{ row }">
            <el-tag
              class="table-status-tag"
              :type="row.status === 'ENABLED' ? 'success' : 'danger'"
            >
              {{
                row.status === 'ENABLED'
                  ? t('system.user.status.normal')
                  : t('system.user.status.disabled')
              }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.field.createdAt')" width="160" class-name="table-time">
          <template #default="{ row }">
            {{ formatDateTime((row as UserItem).createdAt) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.field.actions')" width="360">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button
                v-permission="'system:user:update'"
                link
                type="primary"
                @click="openEdit(row as UserItem)"
              >
                {{ t('system.user.actions.edit') }}
              </el-button>
              <el-button
                v-permission="'system:user:role:bind'"
                link
                type="primary"
                @click="openRoleDialog(row as UserItem)"
              >
                {{ t('system.user.actions.assignRoles') }}
              </el-button>
              <el-button
                v-permission="'system:user:session:list'"
                link
                type="primary"
                @click="openSessions(row as UserItem)"
              >
                {{ t('system.user.actions.sessions') }}
              </el-button>
              <el-button
                v-permission="'system:user:reset-password'"
                link
                type="warning"
                @click="openResetPassword(row as UserItem)"
              >
                {{ t('system.user.actions.resetPassword') }}
              </el-button>
              <el-button
                v-permission="'system:user:delete'"
                link
                type="danger"
                @click="handleDelete(row as UserItem)"
              >
                {{ t('system.user.actions.delete') }}
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
        editingId === null ? t('system.user.dialog.createTitle') : t('system.user.dialog.editTitle')
      "
      width="480px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="110px"
        :validate-on-rule-change="false"
      >
        <el-form-item :label="t('common.field.username')" prop="username">
          <el-input
            v-model="form.username"
            :disabled="editingId !== null"
            :placeholder="t('system.user.dialog.usernamePlaceholder')"
            maxlength="32"
          />
        </el-form-item>
        <el-form-item v-if="editingId === null" :label="t('common.field.password')" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            :placeholder="t('system.user.dialog.passwordPlaceholder')"
            maxlength="72"
          />
        </el-form-item>
        <el-form-item :label="t('common.field.nickname')" prop="nickname">
          <el-input v-model="form.nickname" maxlength="32" />
        </el-form-item>
        <el-form-item :label="t('common.field.email')" prop="email">
          <el-input v-model="form.email" maxlength="64" />
        </el-form-item>
        <el-form-item :label="t('common.field.phone')" prop="phone">
          <el-input v-model="form.phone" maxlength="11" />
        </el-form-item>
        <el-form-item :label="t('common.field.status')" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio value="ENABLED">{{ t('system.user.dialog.enabled') }}</el-radio>
            <el-radio value="DISABLED">{{ t('system.user.dialog.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ t('common.action.cancel') }}</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          {{ t('common.action.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="roleDialogVisible" :title="t('system.user.roleDialog.title')" width="460px">
      <div v-loading="roleDialogLoading">
        <div class="role-target">
          {{ t('system.user.roleDialog.userLabel') }}<strong>{{ roleTarget?.nickname }}</strong>
          <span class="role-username">{{ roleTarget?.username }}</span>
        </div>
        <el-select
          v-model="selectedRoleIds"
          class="role-select"
          multiple
          filterable
          :placeholder="t('system.user.roleDialog.selectPlaceholder')"
        >
          <el-option
            v-for="role in roleOptions"
            :key="role.id"
            :label="role.name"
            :value="role.id"
            :disabled="role.status === 'DISABLED'"
          >
            {{ role.name }} ({{ role.code }})
            <span v-if="role.status === 'DISABLED'">
              {{ t('system.user.roleDialog.disabledSuffix') }}
            </span>
          </el-option>
        </el-select>
        <p class="role-hint">{{ t('system.user.roleDialog.hint') }}</p>
      </div>
      <template #footer>
        <el-button @click="roleDialogVisible = false">{{ t('common.action.cancel') }}</el-button>
        <el-button type="primary" :loading="roleDialogSaving" @click="saveRoles">
          {{ t('common.action.save') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="resetPwdVisible"
      :title="t('system.user.resetPassword.title')"
      width="460px"
      :close-on-click-modal="false"
      @closed="resetPwdRef?.clearValidate()"
    >
      <el-form
        ref="resetPwdRef"
        :model="resetPwdForm"
        :rules="resetPwdRules"
        label-width="auto"
        :validate-on-rule-change="false"
      >
        <el-form-item :label="t('system.user.resetPassword.userLabel')">
          <span class="reset-pwd-user">
            <strong>{{ resetPwdTarget?.nickname }}</strong>
            <span class="role-username">{{ resetPwdTarget?.username }}</span>
          </span>
        </el-form-item>
        <el-form-item :label="t('system.user.resetPassword.newPassword')" prop="newPassword">
          <el-input
            v-model="resetPwdForm.newPassword"
            type="password"
            show-password
            autocomplete="new-password"
            :placeholder="t('system.user.resetPassword.placeholder')"
            maxlength="72"
          />
          <span class="form-hint">{{ t('system.user.resetPassword.hint') }}</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetPwdVisible = false">{{ t('common.action.cancel') }}</el-button>
        <el-button type="primary" :loading="resetPwdSubmitting" @click="handleResetPassword">
          {{ t('common.action.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-drawer
      v-model="sessionDrawerVisible"
      :title="t('system.user.session.title', { name: sessionName() })"
      size="680px"
    >
      <div class="session-toolbar">
        <div class="session-toolbar__actions">
          <el-button :icon="Refresh" :loading="sessionLoading" @click="fetchSessions">
            {{ t('system.user.session.refresh') }}
          </el-button>
          <el-tooltip
            :disabled="!hasCurrentSession"
            :content="t('system.user.session.currentHint')"
            placement="top"
          >
            <span>
              <el-button
                v-permission="'system:user:session:revoke-all'"
                type="danger"
                plain
                :loading="sessionKicking"
                :disabled="!sessions.length || hasCurrentSession"
                @click="handleRevokeAllSessions"
              >
                {{ t('system.user.session.revokeAll') }}
              </el-button>
            </span>
          </el-tooltip>
        </div>
        <span class="session-count">
          {{ t('system.user.session.count', { count: sessions.length }) }}
        </span>
      </div>

      <el-table
        v-loading="sessionLoading"
        :data="sessions"
        :empty-text="t('system.user.session.empty')"
      >
        <el-table-column :label="t('system.user.session.columns.loginIp')" min-width="130">
          <template #default="{ row }">
            <span>{{ (row as UserSessionItem).loginIp || '-' }}</span>
            <el-tag
              v-if="(row as UserSessionItem).current"
              class="session-current"
              type="success"
              size="small"
            >
              {{ t('system.user.session.current') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="userAgent"
          :label="t('system.user.session.columns.client')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column
          :label="t('system.user.session.columns.loginAt')"
          width="170"
          class-name="table-time"
        >
          <template #default="{ row }">
            {{ formatDateTime((row as UserSessionItem).loginAt) }}
          </template>
        </el-table-column>
        <el-table-column
          :label="t('system.user.session.columns.expireAt')"
          width="170"
          class-name="table-time"
        >
          <template #default="{ row }">
            {{ formatDateTime((row as UserSessionItem).expireAt) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.field.actions')" width="90" fixed="right">
          <template #default="{ row }">
            <el-button
              v-permission="'system:user:session:revoke'"
              link
              type="danger"
              :disabled="(row as UserSessionItem).current || sessionKicking"
              @click="handleRevokeSession(row as UserSessionItem)"
            >
              {{ t('system.user.session.revoke') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-drawer>
  </div>
</template>

<style scoped>
.role-tag {
  margin: 2px 6px 2px 0;
}

.role-target {
  margin-bottom: 12px;
  color: var(--el-text-color-regular);
}

.reset-pwd-user {
  color: var(--el-text-color-regular);
}

.form-hint {
  display: block;
  margin-top: 4px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.4;
}

.role-username {
  margin-left: 8px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.role-select {
  width: 100%;
}

.role-hint {
  margin: 8px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.session-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.session-toolbar__actions {
  display: flex;
  gap: 8px;
}

.session-count {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.session-current {
  margin-left: 6px;
}
</style>
