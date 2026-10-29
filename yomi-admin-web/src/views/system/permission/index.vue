<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'

import {
  createPermission,
  deletePermission,
  getPermissionPage,
  setPermissionStatus,
  updatePermission,
} from '@/api/rbac'
import type { PermissionQuery } from '@/api/rbac'
import type { PermissionItem, PermissionType, Status } from '@/api/types'
import { confirmAction } from '@/utils/confirm'
import { formatDateTime } from '@/utils/datetime'

defineOptions({ name: 'SystemPermissionPage' })

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
  if (!uiPath) return callback(new Error('请输入菜单路径'))
  if (!uiPath.startsWith('/')) return callback(new Error('菜单路径需以 / 开头'))
  callback()
}

/** 只有接口类型才校验 HTTP 方法 */
function validateHTTPMethod(_rule: unknown, value: string, callback: (error?: Error) => void) {
  if (form.type !== 'API') return callback()
  if (!value) return callback(new Error('请选择 HTTP 方法'))
  callback()
}

/** 只有接口类型才校验路由模板 */
function validateHTTPPath(_rule: unknown, value: string, callback: (error?: Error) => void) {
  if (form.type !== 'API') return callback()
  const httpPath = value?.trim() ?? ''
  if (!httpPath) return callback(new Error('请输入接口路径'))
  if (!httpPath.startsWith('/api/')) return callback(new Error('接口路径需以 /api/ 开头'))
  callback()
}

const rules: FormRules = {
  code: [
    { required: true, message: '请输入权限编码', trigger: 'blur' },
    {
      pattern: /^[a-z][a-z0-9_]*(?::[a-z][a-z0-9_]*)*$/,
      message: '小写字母开头，用冒号分段，可含数字与下划线',
      trigger: 'blur',
    },
  ],
  name: [{ required: true, message: '请输入权限名称', trigger: 'blur' }],
  uiPath: [{ validator: validateUIPath, trigger: 'blur' }],
  httpMethod: [{ validator: validateHTTPMethod, trigger: 'change' }],
  httpPath: [{ validator: validateHTTPPath, trigger: 'blur' }],
}

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
      ElMessage.success('新增成功')
    } else {
      await updatePermission(editingId.value, payload)
      ElMessage.success('保存成功')
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: PermissionItem) {
  const confirmed = await confirmAction({
    title: '删除权限',
    message: `确定删除权限「${row.name}」吗？删除后会同步从所有权限组中移除。`,
    confirmButtonText: '删除',
  })
  if (!confirmed) return
  await deletePermission(row.id)
  ElMessage.success('删除成功')
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
    title: disabling ? '停用权限' : '启用权限',
    message: disabling
      ? `确定停用权限「${name}」吗？所有包含该权限的权限组会立即失去它。`
      : `确定启用权限「${name}」吗？包含该权限的权限组会立即恢复它。`,
  })
}

async function handleStatusChange(row: PermissionItem, status: Status) {
  try {
    await setPermissionStatus(row.id, status)
    row.status = status
    ElMessage.success('状态已更新')
  } catch {
    fetchList()
  }
}
</script>

<template>
  <div class="system-permission">
    <el-card shadow="never">
      <div class="toolbar">
        <el-input
          v-model="query.keyword"
          class="toolbar-input"
          placeholder="权限编码 / 名称"
          clearable
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />
        <el-select v-model="query.type" class="toolbar-select" placeholder="全部类型" clearable>
          <el-option label="菜单" value="MENU" />
          <el-option label="接口" value="API" />
        </el-select>
        <el-select v-model="query.status" class="toolbar-select" placeholder="全部状态" clearable>
          <el-option label="启用" value="ENABLED" />
          <el-option label="禁用" value="DISABLED" />
        </el-select>
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
        <el-button
          v-permission="'system:permission:create'"
          type="primary"
          plain
          class="toolbar-add"
          @click="openCreate"
        >
          新增权限
        </el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="name" label="权限名称" min-width="140" />
        <el-table-column prop="code" label="权限编码" min-width="200" />
        <el-table-column label="类型" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.type === 'MENU' ? 'primary' : 'success'">
              {{ row.type === 'MENU' ? '菜单' : '接口' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="资源" min-width="260" show-overflow-tooltip>
          <template #default="{ row }">
            <template v-if="row.type === 'MENU'">{{ row.uiPath }}</template>
            <template v-else>
              <el-tag size="small" class="method-tag">{{ row.httpMethod }}</el-tag>
              {{ row.httpPath }}
            </template>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
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
        <el-table-column prop="remark" label="备注" min-width="140" show-overflow-tooltip />
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime((row as PermissionItem).createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="130" fixed="right" align="center">
          <template #default="{ row }">
            <el-button
              v-permission="'system:permission:update'"
              link
              type="primary"
              @click="openEdit(row as PermissionItem)"
            >
              编辑
            </el-button>
            <el-button
              v-permission="'system:permission:delete'"
              link
              type="danger"
              @click="handleDelete(row as PermissionItem)"
            >
              删除
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
      :title="editingId === null ? '新增权限' : '编辑权限'"
      width="520px"
      :close-on-click-modal="false"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="权限类型" prop="type">
          <el-radio-group v-model="form.type" :disabled="editingId !== null">
            <el-radio value="MENU">菜单</el-radio>
            <el-radio value="API">接口</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="权限编码" prop="code">
          <el-input
            v-model="form.code"
            :disabled="editingId !== null"
            placeholder="创建后不可修改"
            maxlength="96"
          />
        </el-form-item>
        <el-form-item label="权限名称" prop="name">
          <el-input v-model="form.name" maxlength="32" />
        </el-form-item>
        <el-form-item v-if="form.type === 'MENU'" label="菜单路径" prop="uiPath">
          <el-input v-model="form.uiPath" placeholder="例如 /system/user" maxlength="191" />
        </el-form-item>
        <template v-else>
          <el-form-item label="HTTP 方法" prop="httpMethod">
            <el-select v-model="form.httpMethod" class="method-select">
              <el-option
                v-for="method in ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']"
                :key="method"
                :label="method"
                :value="method"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="接口路径" prop="httpPath">
            <el-input v-model="form.httpPath" placeholder="例如 /api/system/user" maxlength="191" />
          </el-form-item>
        </template>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" maxlength="255" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
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
  width: 130px;
}

.toolbar-add {
  margin-left: auto;
}

.page-pagination {
  justify-content: flex-end;
  margin-top: 16px;
}

.method-tag {
  margin-right: 6px;
}

.method-select {
  width: 100%;
}
</style>
