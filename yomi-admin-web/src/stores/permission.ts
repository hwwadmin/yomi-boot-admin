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
  /** 标题语言键，渲染时翻译，切换语言后菜单同步更新 */
  titleKey: string
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
 * 目录分组路由的路径集合。
 * 目录权限（如 系统管理 /system）只作为可授权的分组节点下发，
 * 路由守卫仍按叶子页面路径匹配，避免授予目录后放行整个分支的未授权页面。
 * routes ↔ 布局 ↔ 本模块存在循环依赖，模块求值期 staticRoutes 可能尚未初始化，
 * 因此在 store 首次创建时再计算。
 */
function buildGroupRoutePaths(): Set<string> {
  const paths = new Set<string>()
  const walk = (routes: readonly RouteRecordRaw[], parentPath: string) => {
    for (const route of routes) {
      const fullPath = resolvePath(parentPath, route.path)
      if (route.children?.length) {
        paths.add(fullPath)
        walk(route.children, fullPath)
      }
    }
  }
  walk(staticRoutes, '')
  return paths
}

/**
 * 菜单权限：后端下发可见菜单，前端过滤静态路由。
 * 页面是否存在由静态路由决定，后端菜单只控制可见性。
 */
export const usePermissionStore = defineStore('permission', () => {
  const userStore = useUserStore()
  const groupRoutePaths = buildGroupRoutePaths()

  const menuPaths = computed(() =>
    (userStore.userInfo?.menus ?? [])
      .map((menu) => menu.uiPath)
      .filter((path) => !groupRoutePaths.has(path)),
  )

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
      if (!route.meta?.titleKey || route.meta?.hidden) continue
      const fullPath = resolvePath(parentPath, route.path)

      if (route.children?.length) {
        const children = buildMenu(route.children, fullPath)
        if (!children.length) continue
        nodes.push({
          path: fullPath,
          titleKey: route.meta.titleKey,
          icon: route.meta.icon,
          children,
        })
        continue
      }

      if (!canAccess(fullPath)) continue
      nodes.push({ path: fullPath, titleKey: route.meta.titleKey, icon: route.meta.icon })
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
