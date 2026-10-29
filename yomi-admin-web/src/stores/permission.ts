import { computed } from 'vue'
import { defineStore } from 'pinia'
import type { RouteRecordRaw } from 'vue-router'

import { staticRoutes } from '@/router/routes'
import { useUserStore } from '@/stores/user'

/** 不依赖菜单权限即可访问的路径 */
const implicitPaths = ['/dashboard', '/403']

/** 侧栏菜单节点：分组节点只有 children，叶子节点 path 为可跳转的路由路径 */
export interface MenuNode {
  path: string
  title: string
  icon?: string
  children?: MenuNode[]
}

/** 判断路径是否被菜单路径覆盖（含子路径） */
function covers(pathname: string, uiPath: string): boolean {
  if (!pathname || !uiPath) return false
  if (pathname === uiPath) return true
  const prefix = uiPath.endsWith('/') ? uiPath : `${uiPath}/`
  return pathname.startsWith(prefix)
}

/** 拼接父路径与子路由路径，绝对路径原样返回 */
function resolvePath(parentPath: string, path: string): string {
  if (path.startsWith('/')) return path
  if (!parentPath) return `/${path}`
  return `${parentPath.replace(/\/$/, '')}/${path}`
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

  /** 递归过滤路由树：叶子按菜单权限裁剪，分组仅在还有可见子项时保留 */
  function buildMenu(routes: readonly RouteRecordRaw[], parentPath: string): MenuNode[] {
    const nodes: MenuNode[] = []
    for (const route of routes) {
      if (!route.meta?.title || route.meta?.hidden) continue
      const fullPath = resolvePath(parentPath, route.path)

      if (route.children?.length) {
        const children = buildMenu(route.children, fullPath)
        if (!children.length) continue
        nodes.push({ path: fullPath, title: route.meta.title, icon: route.meta.icon, children })
        continue
      }

      if (!canAccess(fullPath)) continue
      nodes.push({ path: fullPath, title: route.meta.title, icon: route.meta.icon })
    }
    return nodes
  }

  /** 侧栏菜单：根路由下按菜单权限过滤后的路由树 */
  const menus = computed<MenuNode[]>(() => {
    const root = staticRoutes.find((route) => route.path === '/')
    return buildMenu(root?.children ?? [], '')
  })

  return { menuPaths, menus, canAccess }
})
