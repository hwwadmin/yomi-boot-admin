import type { Router } from 'vue-router'

import { usePermissionStore } from '@/stores/permission'
import { useUserStore } from '@/stores/user'

/** 无需登录即可访问的页面 */
const WHITE_LIST = new Set(['/login'])

/** 注册全局路由守卫；在 main.ts 中于 app.use(router) 之前调用 */
export function setupRouterGuards(router: Router) {
  router.beforeEach(async (to) => {
    const userStore = useUserStore()

    document.title = to.meta.title ? `${to.meta.title} · Yomi Admin` : 'Yomi Admin'

    // 未登录：仅白名单页面放行，其余跳登录并记住来源
    if (!userStore.token) {
      if (WHITE_LIST.has(to.path)) return true
      return { path: '/login', query: to.path === '/' ? undefined : { redirect: to.fullPath } }
    }

    // 已登录访问登录页 → 回首页
    if (to.path === '/login') return { path: '/' }

    // 刷新页面后 token 还在但用户信息丢失，拉取一次恢复登录态；
    // 401 时拦截器会清空 token 并跳转登录页
    if (!userStore.userInfo) {
      try {
        await userStore.fetchProfile()
      } catch {
        if (!useUserStore().token) return false
        return true
      }
    }

    // 预留：接入后端菜单后，在此调用 permissionStore.generateRoutes()，
    // 将返回的动态路由逐个 router.addRoute()，然后 return to.fullPath 重新进入本次导航
    void usePermissionStore()

    return true
  })
}
