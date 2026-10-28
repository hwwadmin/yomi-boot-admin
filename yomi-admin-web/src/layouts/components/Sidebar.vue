<script setup lang="ts">
import type { Component } from 'vue'
import type { RouteRecordRaw } from 'vue-router'
import { Odometer, User } from '@element-plus/icons-vue'

import { useAppStore } from '@/stores/app'

// 菜单图标映射：路由 meta.icon 为 @element-plus/icons-vue 的组件名
const iconMap: Record<string, Component> = { Odometer, User }

function resolveIcon(name?: string): Component | undefined {
  return name ? iconMap[name] : undefined
}

const appStore = useAppStore()
const router = useRouter()
const route = useRoute()

// 取根路由 `/` 下需要展示在菜单里的子路由；接入动态路由后与 permission store 合并
const menus = computed<RouteRecordRaw[]>(() => {
  const root = router.options.routes.find((item) => item.path === '/')
  return (root?.children ?? []).filter((item) => item.meta?.title && !item.meta?.hidden)
})
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
      <template v-for="item in menus" :key="item.path">
        <!-- 预留二级菜单：子路由有标题时渲染为 el-sub-menu -->
        <el-sub-menu v-if="item.children?.some((child) => child.meta?.title)" :index="item.path">
          <template #title>
            <el-icon v-if="resolveIcon(item.meta?.icon)">
              <component :is="resolveIcon(item.meta?.icon)" />
            </el-icon>
            <span>{{ item.meta?.title }}</span>
          </template>
          <el-menu-item
            v-for="child in item.children"
            :key="child.path"
            :index="`${item.path}/${child.path}`"
          >
            {{ child.meta?.title }}
          </el-menu-item>
        </el-sub-menu>
        <el-menu-item v-else :index="item.path">
          <el-icon v-if="resolveIcon(item.meta?.icon)">
            <component :is="resolveIcon(item.meta?.icon)" />
          </el-icon>
          <template #title>{{ item.meta?.title }}</template>
        </el-menu-item>
      </template>
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
