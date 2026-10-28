<script setup lang="ts">
import { Money, ShoppingCart, User, View } from '@element-plus/icons-vue'

import { getUserPage } from '@/api/user'
import { useUserStore } from '@/stores/user'

defineOptions({ name: 'DashboardPage' })

const userStore = useUserStore()

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 6) return '凌晨好'
  if (hour < 12) return '上午好'
  if (hour < 14) return '中午好'
  if (hour < 18) return '下午好'
  return '晚上好'
})

// 除用户总数来自真实接口外，其余为静态示例数据
const stats = ref([
  { label: '用户总数', value: 0, icon: User, color: '#409eff' },
  { label: '今日访问', value: 1286, icon: View, color: '#67c23a' },
  { label: '订单数', value: 328, icon: ShoppingCart, color: '#e6a23c' },
  { label: '销售额（元）', value: 92186, icon: Money, color: '#f56c6c' },
])

onMounted(async () => {
  try {
    const { total } = await getUserPage({ page: 1, pageSize: 1 })
    const first = stats.value[0]
    if (first) first.value = total
  } catch {
    // 错误提示已由 request 拦截器统一处理
  }
})
</script>

<template>
  <div class="dashboard">
    <el-card shadow="never" class="welcome-card">
      <div class="welcome-text">
        {{ greeting }}，{{ userStore.userInfo?.nickname || '管理员' }}！欢迎使用 Yomi Admin
        管理后台脚手架。
      </div>
      <div class="welcome-tags">
        <el-tag>vite</el-tag>
        <el-tag type="success">vue3 + ts</el-tag>
        <el-tag type="warning">element-plus</el-tag>
        <el-tag type="danger">vue-router(hash)</el-tag>
        <el-tag type="info">pinia</el-tag>
        <el-tag>axios</el-tag>
        <el-tag type="success">@vueuse/core</el-tag>
      </div>
    </el-card>

    <el-row :gutter="16" class="stat-row">
      <el-col v-for="item in stats" :key="item.label" :xs="12" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-body">
            <el-icon :size="36" :color="item.color">
              <component :is="item.icon" />
            </el-icon>
            <div class="stat-meta">
              <div class="stat-value">{{ item.value.toLocaleString() }}</div>
              <div class="stat-label">{{ item.label }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never">
      <template #header>下一步</template>
      <ol class="todo-list">
        <li>在 <code>src/api/</code> 下按模块新增接口，类型放 <code>src/api/types.ts</code></li>
        <li>
          在 <code>src/views/</code> 下新增页面，并在 <code>src/router/index.ts</code> 注册路由
        </li>
        <li>
          后端就绪后，把 <code>.env.development</code> 的 <code>VITE_USE_MOCK</code> 改为 false 联调
        </li>
        <li>需要按钮级权限时，使用 <code>v-permission</code> 指令</li>
      </ol>
    </el-card>
  </div>
</template>

<style scoped>
.welcome-card {
  margin-bottom: 16px;
}

.welcome-text {
  color: var(--el-text-color-primary);
  font-size: 16px;
  font-weight: 600;
}

.welcome-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

.stat-row {
  margin-bottom: 16px;
}

.stat-body {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-value {
  color: var(--el-text-color-primary);
  font-size: 24px;
  font-weight: 700;
}

.stat-label {
  margin-top: 4px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.todo-list {
  margin: 0;
  padding-left: 20px;
  color: var(--el-text-color-regular);
  font-size: 14px;
  line-height: 2;
}

.todo-list code {
  padding: 2px 6px;
  border-radius: 4px;
  background-color: var(--el-fill-color);
  font-size: 13px;
}
</style>
