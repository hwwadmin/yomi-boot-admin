<script setup lang="ts">
import type { Component } from 'vue'
import { Avatar, Grid, Key, Lock, Odometer, Setting, Tickets, User } from '@element-plus/icons-vue'

import type { MenuNode } from '@/stores/permission'

defineProps<{ item: MenuNode }>()

// 菜单图标映射：路由 meta.icon 为 @element-plus/icons-vue 的组件名
const iconMap: Record<string, Component> = { Odometer, User, Avatar, Key, Grid, Tickets, Lock, Setting }

function resolveIcon(name?: string): Component | undefined {
  return name ? iconMap[name] : undefined
}
</script>

<template>
  <el-sub-menu v-if="item.children?.length" :index="item.path">
    <template #title>
      <el-icon v-if="resolveIcon(item.icon)">
        <component :is="resolveIcon(item.icon)" />
      </el-icon>
      <span>{{ item.title }}</span>
    </template>
    <SidebarItem v-for="child in item.children" :key="child.path" :item="child" />
  </el-sub-menu>
  <el-menu-item v-else :index="item.path">
    <el-icon v-if="resolveIcon(item.icon)">
      <component :is="resolveIcon(item.icon)" />
    </el-icon>
    <template #title>{{ item.title }}</template>
  </el-menu-item>
</template>
