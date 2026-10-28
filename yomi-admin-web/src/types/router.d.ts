import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    /** 页面标题，同时用于菜单与面包屑 */
    title?: string
    /** 菜单图标（@element-plus/icons-vue 组件名） */
    icon?: string
    /** 在菜单中隐藏 */
    hidden?: boolean
  }
}
