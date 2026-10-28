<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'

import { bindUserRoles, getUserRoles } from '@/api/rbac'
import { createUser, deleteUser, getUserPage, updateUser } from '@/api/user'
import type { RoleItem, UserItem, UserStatus } from '@/api/types'
import { formatDateTime } from '@/utils/datetime'

defineOptions({ name: 'SystemUserPage' })

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
const editingId = ref<string | null>(null)
const formRef = ref<FormInstance>()
const form = reactive({
  username: '',
  password: '',
  nickname: '',
  email: '',
  phone: '',
  status: 'ENABLED' as UserStatus,
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 72, message: '密码长度 6-72 位', trigger: 'blur' },
  ],
  nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
  email: [{ type: 'email', message: '邮箱格式不正确', trigger: 'blur' }],
  phone: [{ pattern: /^1\d{10}$/, message: '手机号格式不正确', trigger: 'blur' }],
}

function openCreate() {
  editingId.value = null
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
      await createUser({ ...payload, password: form.password })
      ElMessage.success('新增成功')
    } else {
      await updateUser(editingId.value, payload)
      ElMessage.success('保存成功')
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
    ElMessage.success('角色已保存')
    roleDialogVisible.value = false
    fetchList()
  } finally {
    roleDialogSaving.value = false
  }
}

// ———————— 删除 ————————
async function handleDelete(row: UserItem) {
  try {
    await ElMessageBox.confirm(`确定删除用户「${row.nickname}」吗？`, '警告', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }
  await deleteUser(row.id)
  ElMessage.success('删除成功')
  // 当前页只剩一条时回退一页，避免删除后停留在空页
  if (list.value.length === 1 && query.page > 1) {
    query.page -= 1
  }
  fetchList()
}
</script>

<template>
  <div class="system-user">
    <el-card shadow="never">
      <div class="toolbar">
        <el-input
          v-model="query.keyword"
          class="toolbar-input"
          placeholder="用户名 / 昵称"
          clearable
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
        <el-button
          v-permission="'system:user:create'"
          type="primary"
          plain
          class="toolbar-add"
          @click="openCreate"
        >
          新增用户
        </el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="username" label="用户名" min-width="100" />
        <el-table-column prop="nickname" label="昵称" min-width="100" />
        <el-table-column label="角色" min-width="120">
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
        <el-table-column prop="email" label="邮箱" min-width="150" show-overflow-tooltip />
        <el-table-column prop="phone" label="手机号" min-width="110" />
        <el-table-column label="状态" width="72" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'ENABLED' ? 'success' : 'danger'">
              {{ row.status === 'ENABLED' ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">
            {{ formatDateTime((row as UserItem).createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="190" fixed="right" align="center">
          <template #default="{ row }">
            <el-button
              v-permission="'system:user:update'"
              link
              type="primary"
              @click="openEdit(row as UserItem)"
            >
              编辑
            </el-button>
            <el-button
              v-permission="'system:user:role:bind'"
              link
              type="primary"
              @click="openRoleDialog(row as UserItem)"
            >
              分配角色
            </el-button>
            <el-button
              v-permission="'system:user:delete'"
              link
              type="danger"
              @click="handleDelete(row as UserItem)"
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
      :title="editingId === null ? '新增用户' : '编辑用户'"
      width="480px"
      :close-on-click-modal="false"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="form.username"
            :disabled="editingId !== null"
            placeholder="登录账号，创建后不可修改"
            maxlength="32"
          />
        </el-form-item>
        <el-form-item v-if="editingId === null" label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            placeholder="6-72 位"
            maxlength="72"
          />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="form.nickname" maxlength="32" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" maxlength="64" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" maxlength="11" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio value="ENABLED">正常</el-radio>
            <el-radio value="DISABLED">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="roleDialogVisible" title="分配角色" width="460px">
      <div v-loading="roleDialogLoading">
        <div class="role-target">
          用户：<strong>{{ roleTarget?.nickname }}</strong>
          <span class="role-username">{{ roleTarget?.username }}</span>
        </div>
        <el-select
          v-model="selectedRoleIds"
          class="role-select"
          multiple
          filterable
          placeholder="请选择角色"
        >
          <el-option
            v-for="role in roleOptions"
            :key="role.id"
            :label="role.name"
            :value="role.id"
            :disabled="role.status === 'DISABLED'"
          >
            {{ role.name }} ({{ role.code }})
          </el-option>
        </el-select>
      </div>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="roleDialogSaving" @click="saveRoles">保存</el-button>
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
  width: 240px;
}

.toolbar-add {
  margin-left: auto;
}

.role-tag {
  margin: 2px 6px 2px 0;
}

.page-pagination {
  justify-content: flex-end;
  margin-top: 16px;
}

.role-target {
  margin-bottom: 12px;
  color: var(--el-text-color-regular);
}

.role-username {
  margin-left: 8px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.role-select {
  width: 100%;
}
</style>
