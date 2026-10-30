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
  <el-breadcrumb class="breadcrumb" separator="/">
    <el-breadcrumb-item
      v-for="item in items"
      :key="item.path"
      :to="item.linkable ? item.path : undefined"
    >
      {{ t(item.titleKey) }}
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>

<style scoped>
.breadcrumb {
  font-size: 13px;
}

.breadcrumb :deep(.el-breadcrumb__inner) {
  color: var(--app-text-3);
  font-weight: 400;
}

.breadcrumb :deep(.el-breadcrumb__inner.is-link:hover) {
  color: var(--el-color-primary);
}

.breadcrumb :deep(.el-breadcrumb__item:last-child .el-breadcrumb__inner) {
  color: var(--app-text-1);
  font-size: 14px;
  font-weight: 600;
}

.breadcrumb :deep(.el-breadcrumb__separator) {
  margin: 0 8px;
  color: var(--app-text-3);
  font-weight: 400;
}
</style>
