<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'

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
import { confirmAction } from '@/utils/confirm'
import { formatDateTime } from '@/utils/datetime'

defineOptions({ name: 'SystemRolePage' })

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
    title: '刷新权限缓存',
    message: '将按数据库重建权限缓存，所有实例随后立即生效。确定刷新吗？',
    confirmButtonText: '刷新',
  })
  if (!confirmed) return

  refreshingCache.value = true
  try {
    await refreshRbacCache()
    ElMessage.success('刷新成功')
  } finally {
    refreshingCache.value = false
  }
}

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

const rules: FormRules = {
  code: [
    { required: true, message: '请输入角色编码', trigger: 'blur' },
    {
      pattern: /^[a-z][a-z0-9_]{2,31}$/,
      message: '小写字母开头，3-32 位小写字母/数字/下划线',
      trigger: 'blur',
    },
  ],
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
}

function openCreate() {
  editingId.value = null
  Object.assign(form, { code: '', name: '', remark: '' })
  dialogVisible.value = true
}

function openEdit(row: RoleItem) {
  editingId.value = row.id
  Object.assign(form, { code: row.code, name: row.name, remark: row.remark })
  dialogVisible.value = true
}

/** 启停是危险操作：确认后再提交 */
function confirmStatusChange(name: string, next: Status): Promise<boolean> {
  const disabling = next === 'DISABLED'
  return confirmAction({
    title: disabling ? '停用角色' : '启用角色',
    message: disabling
      ? `确定停用角色「${name}」吗？停用后该角色对应用户的权限会立即失效。`
      : `确定启用角色「${name}」吗？`,
  })
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    if (editingId.value === null) {
      await createRole({
        code: form.code.trim(),
        name: form.name.trim(),
        remark: form.remark.trim(),
      })
      ElMessage.success('新增成功')
    } else {
      await updateRole(editingId.value, { name: form.name.trim(), remark: form.remark.trim() })
      ElMessage.success('保存成功')
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
    ElMessage.success('状态已更新')
  } catch {
    fetchList()
  }
}

async function handleDelete(row: RoleItem) {
  const confirmed = await confirmAction({
    title: '删除角色',
    message: `确定删除角色「${row.name}」吗？已绑定用户的角色无法删除。`,
    confirmButtonText: '删除',
  })
  if (!confirmed) return
  await deleteRole(row.id)
  ElMessage.success('删除成功')
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

/** 角色权限来自权限组的去重聚合，这里只做只读展示 */
const aggregatedPermissions = computed(() =>
  permissions.value.filter((item) => aggregatedPermissionIds.value.includes(item.id)),
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
    selectedGroupIds.value = detail.groupIds
    aggregatedPermissionIds.value = detail.permissionIds
  } finally {
    groupLoading.value = false
  }
}

async function saveGroups() {
  if (!groupTarget.value) return
  groupSaving.value = true
  try {
    await bindRoleGroups(groupTarget.value.id, selectedGroupIds.value)
    ElMessage.success('权限组已保存')
    groupDrawerVisible.value = false
  } finally {
    groupSaving.value = false
  }
}
</script>

<template>
  <div class="system-role">
    <el-card shadow="never">
      <div class="toolbar">
        <el-input
          v-model="query.keyword"
          class="toolbar-input"
          placeholder="角色编码 / 名称"
          clearable
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />
        <el-select v-model="query.status" class="toolbar-select" placeholder="全部状态" clearable>
          <el-option label="启用" value="ENABLED" />
          <el-option label="禁用" value="DISABLED" />
        </el-select>
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
        <el-button
          v-permission="'system:role:create'"
          type="primary"
          plain
          class="toolbar-add"
          @click="openCreate"
        >
          新增角色
        </el-button>
        <el-button
          v-permission="'system:rbac_cache:refresh'"
          type="warning"
          :loading="refreshingCache"
          @click="handleRefreshCache"
        >
          刷新权限缓存
        </el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="name" label="角色名称" min-width="140" />
        <el-table-column prop="code" label="角色编码" min-width="140" />
        <el-table-column label="内置" width="80" align="center">
          <template #default="{ row }">{{ row.builtin ? '是' : '否' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
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
        <el-table-column prop="remark" label="备注" min-width="160" show-overflow-tooltip />
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime((row as RoleItem).createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right" align="center">
          <template #default="{ row }">
            <el-button
              v-permission="'system:role:update'"
              link
              type="primary"
              @click="openEdit(row as RoleItem)"
            >
              编辑
            </el-button>
            <el-button
              v-if="!row.builtin"
              v-permission="'system:role:delete'"
              link
              type="danger"
              @click="handleDelete(row as RoleItem)"
            >
              删除
            </el-button>
            <el-button
              v-if="!row.builtin"
              v-permission="'system:role:bind-group'"
              link
              type="primary"
              @click="openGroupDrawer(row as RoleItem)"
            >
              分配权限组
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
      :title="editingId === null ? '新增角色' : '编辑角色'"
      width="480px"
      :close-on-click-modal="false"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="角色编码" prop="code">
          <el-input
            v-model="form.code"
            :disabled="editingId !== null"
            placeholder="创建后不可修改"
            maxlength="32"
          />
        </el-form-item>
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="form.name" maxlength="32" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" maxlength="255" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="groupDrawerVisible" title="分配权限组" size="560px">
      <div v-loading="groupLoading" class="group-body">
        <div class="group-meta">
          角色：<strong>{{ groupTarget?.name }}</strong>
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
            <el-tag v-if="group.status === 'DISABLED'" type="info" size="small">已停用</el-tag>
          </el-checkbox>
        </el-checkbox-group>
        <el-empty
          v-if="!groupLoading && groups.length === 0"
          description="暂无权限组，请先在「权限组」中创建"
          :image-size="80"
        />
        <el-divider content-position="left">
          聚合权限（{{ aggregatedPermissions.length }} 项，自动去重）
        </el-divider>
        <div class="permission-tags">
          <el-tag v-for="item in aggregatedPermissions" :key="item.id" size="small" type="info">
            {{ item.name }}
          </el-tag>
          <span v-if="aggregatedPermissions.length === 0" class="muted">
            该角色尚未绑定任何权限组
          </span>
        </div>
      </div>
      <template #footer>
        <el-button @click="groupDrawerVisible = false">取消</el-button>
        <el-button type="primary" :loading="groupSaving" @click="saveGroups">保存</el-button>
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
</style>
