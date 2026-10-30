<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import { useAppStore } from '@/stores/app'
import { useLocaleStore } from '@/stores/locale'

// 初始化 app store：让暗黑模式（useDark）在包括登录页在内的所有页面生效
useAppStore()

const route = useRoute()
const { t, locale } = useI18n()
const localeStore = useLocaleStore()

// 标题与 html lang 跟随路由和当前语言变化
watchEffect(() => {
  const titleKey = route.meta.titleKey
  document.title = titleKey ? `${t(titleKey)} · Yomi Admin` : 'Yomi Admin'
  document.documentElement.lang = locale.value
})
</script>

<template>
  <el-config-provider :locale="localeStore.elementLocale">
    <router-view />
  </el-config-provider>
</template>
