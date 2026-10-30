<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import type { MenuNode } from '@/stores/permission'
import { resolveMenuIcon } from '@/utils/menuIcons'

defineOptions({ name: 'SidebarItem' })

defineProps<{ item: MenuNode }>()

const { t } = useI18n()
</script>

<template>
  <el-sub-menu v-if="item.children?.length" :index="item.path">
    <template #title>
      <el-icon v-if="resolveMenuIcon(item.icon)">
        <component :is="resolveMenuIcon(item.icon)" />
      </el-icon>
      <span>{{ t(item.titleKey) }}</span>
    </template>
    <SidebarItem v-for="child in item.children" :key="child.path" :item="child" />
  </el-sub-menu>
  <el-menu-item v-else :index="item.path">
    <el-icon v-if="resolveMenuIcon(item.icon)">
      <component :is="resolveMenuIcon(item.icon)" />
    </el-icon>
    <template #title>{{ t(item.titleKey) }}</template>
  </el-menu-item>
</template>
