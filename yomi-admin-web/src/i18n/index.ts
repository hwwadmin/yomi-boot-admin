import { createI18n } from 'vue-i18n'

import enUS from './locales/en-US'
import zhCN from './locales/zh-CN'
import { isLocaleCode, type LocaleCode } from './types'

const LOCALE_STORAGE_KEY = 'yomi:locale'

/** 有本地选择用本地选择，其次跟随浏览器，最终回退中文 */
function detectLocale(): LocaleCode {
  const stored = localStorage.getItem(LOCALE_STORAGE_KEY)
  if (isLocaleCode(stored)) return stored
  return navigator.language.toLowerCase().startsWith('en') ? 'en-US' : 'zh-CN'
}

export const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale: detectLocale(),
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS,
  },
})

export function setLocale(locale: LocaleCode) {
  i18n.global.locale.value = locale
  localStorage.setItem(LOCALE_STORAGE_KEY, locale)
}
