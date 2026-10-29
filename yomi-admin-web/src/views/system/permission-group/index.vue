<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'

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

const rules: FormRules = {
  code: [
    { required: true, message: '请输入权限组编码', trigger: 'blur' },
    {
      pattern: /^[a-z][a-z0-9_]{2,31}$/,
      message: '小写字母开头，3-32 位小写字母/数字/下划线',
      trigger: 'blur',
    },
  ],
  name: [{ required: true, message: '请输入权限组名称', trigger: 'blur' }],
}

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
    title: disabling ? '停用权限组' : '启用权限组',
    message: disabling
      ? `确定停用权限组「${name}」吗？使用该组的所有角色会立即失去组内权限。`
      : `确定启用权限组「${name}」吗？组内权限会立即对已绑角色生效。`,
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
      ElMessage.success('新增成功')
    } else {
      await updatePermissionGroup(editingId.value, {
        name: form.name.trim(),
        remark: form.remark.trim(),
      })
      ElMessage.success('保存成功')
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
    ElMessage.success('状态已更新')
  } catch {
    fetchList()
  }
}

async function handleDelete(row: PermissionGroupItem) {
  const confirmed = await confirmAction({
    title: '删除权限组',
    message: `确定删除权限组「${row.name}」吗？已绑定角色的权限组无法删除。`,
    confirmButtonText: '删除',
  })
  if (!confirmed) return
  await deletePermissionGroup(row.id)
  ElMessage.success('删除成功')
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
    ElMessage.success('权限已保存')
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
          placeholder="权限组编码 / 名称"
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
          v-permission="'system:permission-group:create'"
          type="primary"
          plain
          class="toolbar-add"
          @click="openCreate"
        >
          新增权限组
        </el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="name" label="权限组名称" min-width="140" />
        <el-table-column prop="code" label="权限组编码" min-width="160" />
        <el-table-column label="状态" width="100" align="center">
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
        <el-table-column prop="remark" label="备注" min-width="200" show-overflow-tooltip />
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime((row as PermissionGroupItem).createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right" align="center">
          <template #default="{ row }">
            <el-button
              v-permission="'system:permission-group:update'"
              link
              type="primary"
              @click="openEdit(row as PermissionGroupItem)"
            >
              编辑
            </el-button>
            <el-button
              v-permission="'system:permission-group:delete'"
              link
              type="danger"
              @click="handleDelete(row as PermissionGroupItem)"
            >
              删除
            </el-button>
            <el-button
              v-permission="'system:permission-group:bind-permission'"
              link
              type="primary"
              @click="openPermissionDrawer(row as PermissionGroupItem)"
            >
              分配权限
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
      :title="editingId === null ? '新增权限组' : '编辑权限组'"
      width="480px"
      :close-on-click-modal="false"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="权限组编码" prop="code">
          <el-input
            v-model="form.code"
            :disabled="editingId !== null"
            placeholder="创建后不可修改"
            maxlength="32"
          />
        </el-form-item>
        <el-form-item label="权限组名称" prop="name">
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

    <el-drawer v-model="permissionDrawerVisible" title="分配权限" size="560px">
      <div v-loading="permissionLoading" class="permission-body">
        <div class="permission-meta">
          权限组：<strong>{{ permissionTarget?.name }}</strong>
          <span class="muted">{{ permissionTarget?.code }}</span>
          <span class="muted">已选 {{ selectedPermissionIds.length }} 项</span>
        </div>
        <el-input
          v-model="permissionKeyword"
          class="permission-search"
          placeholder="按名称 / 编码 / 路径搜索"
          clearable
        />
        <el-tabs v-model="permissionTab">
          <el-tab-pane label="菜单权限" name="MENU">
            <el-checkbox-group v-model="selectedPermissionIds" class="permission-list">
              <el-checkbox
                v-for="item in menuPermissions"
                :key="item.id"
                :value="item.id"
                :disabled="item.status === 'DISABLED' && !selectedPermissionIds.includes(item.id)"
              >
                {{ item.name }}
                <span class="muted">{{ item.uiPath }}</span>
                <el-tag v-if="item.status === 'DISABLED'" type="info" size="small">已停用</el-tag>
              </el-checkbox>
            </el-checkbox-group>
            <el-empty
              v-if="!permissionLoading && menuPermissions.length === 0"
              :description="permissionKeyword ? '没有匹配的菜单权限' : '暂无菜单权限'"
              :image-size="60"
            />
          </el-tab-pane>
          <el-tab-pane label="接口权限" name="API">
            <el-checkbox-group v-model="selectedPermissionIds" class="permission-list">
              <el-checkbox
                v-for="item in apiPermissions"
                :key="item.id"
                :value="item.id"
                :disabled="item.status === 'DISABLED' && !selectedPermissionIds.includes(item.id)"
              >
                {{ item.name }}
                <span class="muted">{{ item.httpMethod }} {{ item.httpPath }}</span>
                <el-tag v-if="item.status === 'DISABLED'" type="info" size="small">已停用</el-tag>
              </el-checkbox>
            </el-checkbox-group>
            <el-empty
              v-if="!permissionLoading && apiPermissions.length === 0"
              :description="permissionKeyword ? '没有匹配的接口权限' : '暂无接口权限'"
              :image-size="60"
            />
          </el-tab-pane>
        </el-tabs>
        <span class="hint">停用的权限可以绑定，但不会授予访问能力。</span>
      </div>
      <template #footer>
        <el-button @click="permissionDrawerVisible = false">取消</el-button>
        <el-button type="primary" :loading="permissionSaving" @click="savePermissions"
          >保存</el-button
        >
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
