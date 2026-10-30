<script setup lang="ts">
import { computed } from 'vue'

import { useAppStore } from '@/stores/app'
import Sidebar from './components/Sidebar.vue'
import Navbar from './components/Navbar.vue'

defineOptions({ name: 'AdminLayout' })

const appStore = useAppStore()

const sidebarWidth = computed(() => (appStore.sidebarCollapsed ? '64px' : '224px'))
</script>

<template>
  <el-container class="admin-layout">
    <el-aside class="layout-aside" :width="sidebarWidth">
      <div class="layout-logo" :class="{ 'is-collapsed': appStore.sidebarCollapsed }">
        <span class="logo-badge">Y</span>
        <span v-show="!appStore.sidebarCollapsed" class="logo-text">Yomi Admin</span>
      </div>
      <Sidebar />
    </el-aside>
    <el-container class="layout-body">
      <el-header class="layout-header" height="60px">
        <Navbar />
      </el-header>
      <el-main class="layout-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.admin-layout {
  height: 100vh;
}

.layout-aside {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: linear-gradient(180deg, var(--app-sider-bg) 0%, var(--app-sider-bg-end) 100%);
  box-shadow: inset -1px 0 0 rgb(255 255 255 / 5%);
  transition: width 0.2s ease;
}

.layout-logo {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  gap: 10px;
  height: var(--app-header-height);
  padding: 0 18px;
  overflow: hidden;
  color: #fff;
  white-space: nowrap;
}

.layout-logo.is-collapsed {
  justify-content: center;
  padding: 0;
}

.logo-badge {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 8px;
  background: linear-gradient(135deg, #4a9bff 0%, #2b6fe0 100%);
  box-shadow: 0 6px 14px -8px rgb(64 158 255 / 90%);
  font-size: 15px;
  font-weight: 600;
}

.logo-text {
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.3px;
}

.layout-header {
  padding: 0 var(--app-main-padding-x);
  background-color: var(--app-surface);
  border-bottom: 1px solid var(--app-border);
}

.layout-main {
  padding: var(--app-main-padding-y) var(--app-main-padding-x);
  overflow: auto;
  background-color: var(--app-bg);
}
</style>
