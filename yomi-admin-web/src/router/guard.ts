import type { Router } from 'vue-router'

import { usePermissionStore } from '@/stores/permission'
import { useUserStore } from '@/stores/user'

/** 无需登录即可访问的页面 */
const WHITE_LIST = new Set(['/login'])

/** 注册全局路由守卫；在 main.ts 中于 app.use(router) 之前调用 */
export function setupRouterGuards(router: Router) {
  router.beforeEach(async (to) => {
    const userStore = useUserStore()

    // 未登录：仅白名单页面放行，其余跳登录并记住来源
    if (!userStore.token) {
      if (WHITE_LIST.has(to.path)) return true
      return { path: '/login', query: to.path === '/' ? undefined : { redirect: to.fullPath } }
    }

    // 已登录访问登录页 → 回首页
    if (to.path === '/login') return { path: '/' }

    // 刷新页面后 token 还在但用户信息丢失，拉取一次恢复登录态
    if (!userStore.userInfo) {
      try {
        await userStore.fetchProfile()
      } catch {
        // Token 已失效时响应拦截器已清空登录态：不能返回 false（只会取消当前导航，
        // 依赖拦截器另起的 replace 容易被并发导航吞掉而停在原页面），
        // 改为守卫直接重定向，并带上原始目标地址。
        if (!userStore.token) {
          return { path: '/login', query: to.path === '/' ? undefined : { redirect: to.fullPath } }
        }
        // 其它失败（后端 5xx、网络抖动）保留 token 放行，避免有效登录态被误判登出
        return true
      }
    }

    // 后端菜单不覆盖当前路径时进入 403；/dashboard、/403 隐式放行
    const permissionStore = usePermissionStore()
    if (permissionStore.canAccess(to.path)) return true
    return { path: '/403' }
  })
}
