<script setup lang="ts">
import { useAppStore } from '@/stores/app'
import { usePermissionStore } from '@/stores/permission'
import SidebarItem from './SidebarItem.vue'

defineOptions({ name: 'LayoutSidebar' })

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
  --el-menu-bg-color: transparent;
  --el-menu-text-color: var(--app-sider-text);
  --el-menu-active-color: var(--app-sider-text-active);
  --el-menu-hover-bg-color: var(--app-sider-hover);
  --el-menu-hover-text-color: var(--app-sider-text-active);
  --el-menu-item-height: 44px;
  --el-menu-sub-item-height: 44px;
  --el-menu-base-level-padding: 12px;
  --el-menu-level-padding: 14px;
  --el-menu-icon-width: 18px;
  padding: 10px 0;
  background: transparent;
  border-right: none;
}

.sidebar-menu:not(.el-menu--collapse) {
  width: 100%;
}

.sidebar-menu.el-menu--collapse {
  width: 100%;
}

.sidebar-menu :deep(.el-menu-item),
.sidebar-menu :deep(.el-sub-menu__title) {
  margin: 2px 10px;
  border-radius: 8px;
  transition:
    color 0.18s ease,
    background-color 0.18s ease;
}

.sidebar-menu :deep(.el-menu-item .el-icon),
.sidebar-menu :deep(.el-sub-menu__title .el-icon) {
  font-size: 17px;
}

.sidebar-menu :deep(.el-menu-item.is-active) {
  color: var(--app-sider-text-active);
  background: var(--app-sider-active-gradient);
  box-shadow: 0 6px 16px -6px rgb(64 158 255 / 55%);
}

.sidebar-menu :deep(.el-menu-item:hover),
.sidebar-menu :deep(.el-sub-menu__title:hover) {
  color: var(--app-sider-text-active);
}

/* 子菜单不加深底色，只靠缩进区分层级 */
.sidebar-menu :deep(.el-menu.el-menu--inline) {
  background: transparent;
}

.sidebar-menu.el-menu--collapse :deep(.el-menu-item),
.sidebar-menu.el-menu--collapse :deep(.el-sub-menu__title) {
  justify-content: center;
  margin: 2px 8px;
  padding: 0;
}

/* 收起时子项不可见，用父组的高亮提示当前页面所在分组 */
.sidebar-menu.el-menu--collapse :deep(.el-sub-menu.is-active > .el-sub-menu__title) {
  color: var(--app-sider-text-active);
  background: var(--app-sider-active-gradient);
  box-shadow: 0 6px 16px -6px rgb(64 158 255 / 55%);
}
</style>
