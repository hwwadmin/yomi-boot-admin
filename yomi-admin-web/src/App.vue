<script setup lang="ts">
import { useAppStore } from '@/stores/app'

// 初始化 app store：让暗黑模式（useDark）在包括登录页在内的所有页面生效
useAppStore()
</script>

<template>
  <router-view />
</template>
