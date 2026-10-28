<script setup lang="ts">
const route = useRoute()

// 由 matched 路由生成面包屑，标题来自 meta.title
const items = computed(() => route.matched.filter((item) => item.meta?.title))
</script>

<template>
  <el-breadcrumb separator="/">
    <el-breadcrumb-item v-for="item in items" :key="item.path" :to="item.path">
      {{ item.meta?.title }}
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>
