import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    /** 页面标题的语言键，同时用于菜单、面包屑与浏览器标题 */
    titleKey?: string
    /** 菜单图标（@element-plus/icons-vue 组件名） */
    icon?: string
    /** 在菜单中隐藏 */
    hidden?: boolean
  }
}
