<script setup lang="ts">
import { useAppStore } from '@/stores/app'
import { usePermissionStore } from '@/stores/permission'
import SidebarItem from './SidebarItem.vue'

const appStore = useAppStore()
const permissionStore = usePermissionStore()
const route = useRoute()
</script>

<template>
  <el-scrollbar class="sidebar-scrollbar">
    <el-menu
      class="sidebar-menu"
      router
      :default-active="route.path"
      :collapse="appStore.sidebarCollapsed"
      :collapse-transition="false"
    >
      <SidebarItem v-for="item in permissionStore.menus" :key="item.path" :item="item" />
    </el-menu>
  </el-scrollbar>
</template>

<style scoped>
.sidebar-scrollbar {
  flex: 1;
}

.sidebar-menu {
  border-right: none;
  background-color: #001529;
}

.sidebar-menu:not(.el-menu--collapse) {
  width: 100%;
}

/* 深色侧边栏（亮/暗主题下保持一致） */
.sidebar-menu :deep(.el-menu-item),
.sidebar-menu :deep(.el-sub-menu__title) {
  color: rgb(255 255 255 / 65%);
}

.sidebar-menu :deep(.el-menu-item:hover),
.sidebar-menu :deep(.el-sub-menu__title:hover) {
  color: #fff;
  background-color: rgb(255 255 255 / 8%);
}

.sidebar-menu :deep(.el-menu-item.is-active) {
  color: #fff;
  background-color: var(--el-color-primary);
}

.sidebar-menu :deep(.el-menu.el-menu--inline) {
  background-color: rgb(0 0 0 / 25%);
}
</style>
