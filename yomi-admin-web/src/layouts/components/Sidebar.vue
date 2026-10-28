<script setup lang="ts">
import type { Component } from 'vue'
import { Avatar, Key, Odometer, User } from '@element-plus/icons-vue'

import { useAppStore } from '@/stores/app'
import { usePermissionStore } from '@/stores/permission'

// 菜单图标映射：路由 meta.icon 为 @element-plus/icons-vue 的组件名
const iconMap: Record<string, Component> = { Odometer, User, Avatar, Key }

function resolveIcon(name?: string): Component | undefined {
  return name ? iconMap[name] : undefined
}

const appStore = useAppStore()
const permissionStore = usePermissionStore()
const route = useRoute()

const menus = computed(() =>
  permissionStore.menus.map((item) => ({
    path: item.path.startsWith('/') ? item.path : `/${item.path}`,
    title: item.meta?.title ?? '',
    icon: item.meta?.icon,
  })),
)
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
      <el-menu-item v-for="item in menus" :key="item.path" :index="item.path">
        <el-icon v-if="resolveIcon(item.icon)">
          <component :is="resolveIcon(item.icon)" />
        </el-icon>
        <template #title>{{ item.title }}</template>
      </el-menu-item>
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
