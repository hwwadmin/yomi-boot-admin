<script setup lang="ts">
import { useFullscreen } from '@vueuse/core'
import { Expand, Fold, FullScreen, Moon, Sunny, SwitchButton } from '@element-plus/icons-vue'

import { logout as logoutApi } from '@/api/auth'
import { useAppStore } from '@/stores/app'
import { useUserStore } from '@/stores/user'
import Breadcrumb from './Breadcrumb.vue'

const appStore = useAppStore()
const userStore = useUserStore()
const router = useRouter()
const { isFullscreen, toggle: toggleFullscreen } = useFullscreen()

async function handleCommand(command: string) {
  if (command !== 'logout') return
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      type: 'warning',
      confirmButtonText: '退出',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }
  try {
    await logoutApi()
  } catch {
    // 退出接口失败不阻塞本地登出
  }
  userStore.reset()
  ElMessage.success('已退出登录')
  router.replace('/login')
}
</script>

<template>
  <div class="navbar">
    <div class="navbar-left">
      <el-icon class="action-btn" :size="18" @click="appStore.toggleSidebar()">
        <Fold v-if="!appStore.sidebarCollapsed" />
        <Expand v-else />
      </el-icon>
      <Breadcrumb />
    </div>
    <div class="navbar-right">
      <el-tooltip :content="appStore.isDark ? '切换为亮色' : '切换为暗色'" placement="bottom">
        <el-icon class="action-btn" :size="18" @click="appStore.toggleDark()">
          <Moon v-if="!appStore.isDark" />
          <Sunny v-else />
        </el-icon>
      </el-tooltip>
      <el-tooltip :content="isFullscreen ? '退出全屏' : '全屏'" placement="bottom">
        <el-icon class="action-btn" :size="18" @click="toggleFullscreen()">
          <FullScreen />
        </el-icon>
      </el-tooltip>
      <el-dropdown trigger="click" @command="handleCommand">
        <span class="user-info">
          <el-avatar :size="28" class="user-avatar">
            {{ userStore.userInfo?.nickname?.charAt(0) || 'U' }}
          </el-avatar>
          <span class="user-name">{{ userStore.userInfo?.nickname || '未登录' }}</span>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="logout" :icon="SwitchButton">退出登录</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<style scoped>
.navbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
}

.navbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.navbar-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.action-btn {
  display: flex;
  align-items: center;
  color: var(--el-text-color-primary);
  cursor: pointer;
  transition: color 0.2s;
}

.action-btn:hover {
  color: var(--el-color-primary);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  outline: none;
}

.user-avatar {
  background-color: var(--el-color-primary);
  color: #fff;
  font-size: 14px;
}

.user-name {
  color: var(--el-text-color-primary);
  font-size: 14px;
}
</style>
