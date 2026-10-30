<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const route = useRoute()
const { t } = useI18n()

// 由 matched 路由生成面包屑，标题来自 meta.titleKey；无组件的分组路由不可跳转
const items = computed(() =>
  route.matched
    .filter((item) => item.meta?.titleKey)
    .map((item) => ({
      path: item.path,
      titleKey: item.meta.titleKey as string,
      linkable: Boolean(item.components),
    })),
)
</script>

<template>
  <el-breadcrumb separator="/">
    <el-breadcrumb-item
      v-for="item in items"
      :key="item.path"
      :to="item.linkable ? item.path : undefined"
    >
      {{ t(item.titleKey) }}
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>
