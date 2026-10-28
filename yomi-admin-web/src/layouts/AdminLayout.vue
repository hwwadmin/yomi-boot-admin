<script setup lang="ts">
import { useAppStore } from '@/stores/app'
import Sidebar from './components/Sidebar.vue'
import Navbar from './components/Navbar.vue'

const appStore = useAppStore()
</script>

<template>
  <el-container class="admin-layout">
    <el-aside class="layout-aside" :width="appStore.sidebarCollapsed ? '64px' : '210px'">
      <div class="layout-logo">
        <span class="logo-badge">Y</span>
        <span v-show="!appStore.sidebarCollapsed" class="logo-text">Yomi Admin</span>
      </div>
      <Sidebar />
    </el-aside>
    <el-container class="layout-body">
      <el-header class="layout-header" height="56px">
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
  background-color: #001529;
  transition: width 0.2s;
}

.layout-logo {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  gap: 10px;
  height: 56px;
  padding: 0 16px;
  overflow: hidden;
  color: #fff;
  font-size: 16px;
  font-weight: 600;
  white-space: nowrap;
}

.logo-badge {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  background-color: var(--el-color-primary);
  font-size: 15px;
}

.layout-header {
  padding: 0 16px;
  background-color: var(--el-bg-color);
  border-bottom: 1px solid var(--el-border-color-light);
}

.layout-main {
  padding: 16px;
  overflow: auto;
  background-color: var(--el-bg-color-page);
}
</style>
