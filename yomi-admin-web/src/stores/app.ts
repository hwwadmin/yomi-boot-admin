import { defineStore } from 'pinia'
import { useDark, useStorage, useToggle } from '@vueuse/core'

export const useAppStore = defineStore('app', () => {
  /** 侧边栏折叠（localStorage 持久化） */
  const sidebarCollapsed = useStorage('yomi:sidebar-collapsed', false)
  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  /** 暗黑模式：App.vue 中初始化（useDark 挂到 html.dark），Navbar 里切换 */
  const isDark = useDark()
  const toggleDark = useToggle(isDark)

  return { sidebarCollapsed, toggleSidebar, isDark, toggleDark }
})
