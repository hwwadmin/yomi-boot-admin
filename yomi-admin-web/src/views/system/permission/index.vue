<script setup lang="ts">
import { getPermissionPage, setPermissionStatus } from '@/api/rbac'
import type { PermissionQuery } from '@/api/rbac'
import type { PermissionItem, Status } from '@/api/types'
import { confirmAction } from '@/utils/confirm'
import { formatDateTime } from '@/utils/datetime'

defineOptions({ name: 'SystemPermissionPage' })

// ———————— 列表查询 ————————
const loading = ref(false)
const list = ref<PermissionItem[]>([])
const total = ref(0)
const query = reactive<PermissionQuery>({ page: 1, pageSize: 10, keyword: '', type: '', status: '' })

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
              :before-change="() => confirmStatusChange((row as PermissionItem).name, (row as PermissionItem).status === 'ENABLED' ? 'DISABLED' : 'ENABLED')"
              @change="(value: string | number | boolean) => handleStatusChange(row as PermissionItem, value as Status)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="140" show-overflow-tooltip />
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime((row as PermissionItem).createTime) }}
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

.page-pagination {
  justify-content: flex-end;
  margin-top: 16px;
}

.method-tag {
  margin-right: 6px;
}
</style>
