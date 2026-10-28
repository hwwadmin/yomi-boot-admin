import { computed } from 'vue'
import { defineStore } from 'pinia'
import type { RouteRecordRaw } from 'vue-router'

import { staticRoutes } from '@/router/routes'
import { useUserStore } from '@/stores/user'

/** 不依赖菜单权限即可访问的路径 */
const implicitPaths = ['/dashboard', '/403']

/** 判断路径是否被菜单路径覆盖（含子路径） */
function covers(pathname: string, uiPath: string): boolean {
  if (!pathname || !uiPath) return false
  if (pathname === uiPath) return true
  const prefix = uiPath.endsWith('/') ? uiPath : `${uiPath}/`
  return pathname.startsWith(prefix)
}

/**
 * 菜单权限：后端下发可见菜单，前端过滤静态路由。
 * 页面是否存在由静态路由决定，后端菜单只控制可见性。
 */
export const usePermissionStore = defineStore('permission', () => {
  const userStore = useUserStore()

  const menuPaths = computed(() => (userStore.userInfo?.menus ?? []).map((menu) => menu.uiPath))

  function canAccess(pathname: string): boolean {
    if (implicitPaths.some((path) => covers(pathname, path))) {
      return true
    }
    return menuPaths.value.some((path) => covers(pathname, path))
  }

  /** 侧栏菜单：根路由下按菜单权限过滤后的子路由 */
  const menus = computed<RouteRecordRaw[]>(() => {
    const root = staticRoutes.find((route) => route.path === '/')
    return (root?.children ?? []).filter((route) => {
      if (!route.meta?.title || route.meta?.hidden) return false
      const fullPath = route.path.startsWith('/') ? route.path : `/${route.path}`
      return canAccess(fullPath)
    })
  })

  return { menuPaths, menus, canAccess }
})
