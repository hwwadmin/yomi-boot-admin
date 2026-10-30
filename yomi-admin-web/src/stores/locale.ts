import { computed } from 'vue'
import { defineStore } from 'pinia'
import en from 'element-plus/es/locale/lang/en'
import zhCn from 'element-plus/es/locale/lang/zh-cn'

import { i18n, setLocale as applyLocale } from '@/i18n'
import type { LocaleCode } from '@/i18n/types'

export const useLocaleStore = defineStore('locale', () => {
  const locale = computed<LocaleCode>({
    get: () => i18n.global.locale.value as LocaleCode,
    set: (value) => applyLocale(value),
  })

  const elementLocale = computed(() => (locale.value === 'en-US' ? en : zhCn))

  function toggleLocale() {
    locale.value = locale.value === 'zh-CN' ? 'en-US' : 'zh-CN'
  }

  return { locale, elementLocale, toggleLocale }
})
