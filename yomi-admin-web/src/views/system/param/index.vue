<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'

import {
  createSysParam,
  deleteSysParam,
  getSysParamPage,
  paramValueTypeLabels,
  setSysParamStatus,
  updateSysParam,
} from '@/api/sysparam'
import type { SysParamQuery } from '@/api/sysparam'
import type { ParamValueType, Status, SysParamItem } from '@/api/types'
import { confirmAction } from '@/utils/confirm'
import { formatDateTime } from '@/utils/datetime'

defineOptions({ name: 'SystemParamPage' })

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
    if (raw !== 'true' && raw !== 'false') return callback(new Error('布尔取值仅支持 true / false'))
    return callback()
  }
  if (form.valueType === 'INT') {
    if (!/^-?\d+$/.test(raw)) return callback(new Error('请输入 64 位整数'))
    return callback()
  }
  if (form.valueType === 'JSON') {
    try {
      JSON.parse(raw)
    } catch {
      return callback(new Error('请输入合法的 JSON 文档'))
    }
    return callback()
  }
  if (raw.length > 2048) return callback(new Error('取值不超过 2048 字符'))
  callback()
}

const rules: FormRules = {
  key: [
    { required: true, message: '请输入参数键', trigger: 'blur' },
    {
      pattern: /^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$/,
      message: '小写点分命名，每段字母开头，可含数字与下划线',
      trigger: 'blur',
    },
  ],
  value: [{ validator: validateValue, trigger: 'blur' }],
  name: [{ required: true, message: '请输入参数名称', trigger: 'blur' }],
}

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
      ElMessage.success('新增成功，已对业务生效')
    } else {
      await updateSysParam(editingId.value, {
        value: form.value,
        name: form.name.trim(),
        remark: form.remark.trim(),
      })
      ElMessage.success('保存成功，已对业务生效')
    }
    dialogVisible.value = false
    fetchList()
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: SysParamItem) {
  const confirmed = await confirmAction({
    title: '删除参数',
    message: `确定删除参数「${row.key}」吗？删除后业务读取方会回退到代码内默认值。`,
    confirmButtonText: '删除',
  })
  if (!confirmed) return
  await deleteSysParam(row.id)
  ElMessage.success('删除成功，业务读取方将回退默认值')
  if (list.value.length === 1 && query.page > 1) {
    query.page -= 1
  }
  fetchList()
}

// ———————— 启停 ————————
function confirmStatusChange(key: string, next: Status): Promise<boolean> {
  const disabling = next === 'DISABLED'
  return confirmAction({
    title: disabling ? '停用参数' : '启用参数',
    message: disabling
      ? `确定停用参数「${key}」吗？停用后业务读取方回退到代码内默认值。`
      : `确定启用参数「${key}」吗？启用后对业务立即生效。`,
  })
}

async function handleStatusChange(row: SysParamItem, status: Status) {
  try {
    await setSysParamStatus(row.id, status)
    row.status = status
    ElMessage.success('状态已更新')
  } catch {
    fetchList()
  }
}
</script>

<template>
  <div class="system-param">
    <el-card shadow="never">
      <div class="toolbar">
        <el-input
          v-model="query.keyword"
          class="toolbar-input"
          placeholder="参数键 / 名称 / 备注"
          clearable
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />
        <el-select v-model="query.valueType" class="toolbar-select" placeholder="全部类型" clearable>
          <el-option v-for="(label, type) in paramValueTypeLabels" :key="type" :label="label" :value="type" />
        </el-select>
        <el-select v-model="query.status" class="toolbar-select" placeholder="全部状态" clearable>
          <el-option label="启用" value="ENABLED" />
          <el-option label="禁用" value="DISABLED" />
        </el-select>
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
        <el-button
          v-permission="'system:param:create'"
          type="primary"
          plain
          class="toolbar-add"
          @click="openCreate"
        >
          新增参数
        </el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="name" label="参数名称" min-width="130" show-overflow-tooltip />
        <el-table-column prop="key" label="参数键" min-width="220" show-overflow-tooltip />
        <el-table-column prop="value" label="参数值" min-width="160" show-overflow-tooltip />
        <el-table-column label="类型" width="90" align="center">
          <template #default="{ row }">
            <el-tag>{{ paramValueTypeLabels[(row as SysParamItem).valueType] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
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
        <el-table-column prop="remark" label="备注" min-width="140" show-overflow-tooltip />
        <el-table-column label="更新时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime((row as SysParamItem).updateTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="130" fixed="right" align="center">
          <template #default="{ row }">
            <el-button
              v-permission="'system:param:update'"
              link
              type="primary"
              @click="openEdit(row as SysParamItem)"
            >
              编辑
            </el-button>
            <el-button
              v-permission="'system:param:delete'"
              link
              type="danger"
              @click="handleDelete(row as SysParamItem)"
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
      :title="editingId === null ? '新增参数' : '编辑参数'"
      width="560px"
      :close-on-click-modal="false"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="参数名称" prop="name">
          <el-input v-model="form.name" maxlength="64" />
        </el-form-item>
        <el-form-item label="参数键" prop="key">
          <el-input
            v-model="form.key"
            :disabled="editingId !== null"
            placeholder="例如 order.auto_cancel_days，创建后不可修改"
            maxlength="128"
          />
        </el-form-item>
        <el-form-item label="值类型" prop="valueType">
          <el-radio-group v-model="form.valueType" :disabled="editingId !== null">
            <el-radio v-for="(label, type) in paramValueTypeLabels" :key="type" :value="type">
              {{ label }}
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="参数值" prop="value">
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
            placeholder='例如 {"max":10}'
          />
          <el-input
            v-else
            v-model="form.value"
            :maxlength="form.valueType === 'STRING' ? 2048 : undefined"
            :placeholder="form.valueType === 'INT' ? '64 位整数' : '参数取值'"
          />
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
</style>
