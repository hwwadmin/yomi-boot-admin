import { i18n } from '@/i18n'
import router from '@/router'
import { useUserStore } from '@/stores/user'

/** 会话失效跳转是否已在进行中，避免并发请求重复提示与重复导航 */
let redirecting = false

/**
 * 清理登录态并回到登录页。
 * 供响应拦截器（Token 失效）与路由守卫（拉取用户信息失败）共用：
 * 无论请求从哪个页面、哪条并发链路失败，最终状态都是“已登出 + 登录页”，不会停在半登录页面。
 */
export function expireSession() {
  const userStore = useUserStore()
  // 首条失败已清空 token，后续并发失败不再重复提示
  const hadToken = Boolean(userStore.token)
  userStore.reset()

  const current = router.currentRoute.value
  if (redirecting || current.path === '/login') return

  redirecting = true
  if (hadToken) ElMessage.error(i18n.global.t('errors.sessionExpired'))

  const done = () => {
    // 导航被取消等异常时释放标记，后续失败仍可再次触发
    redirecting = false
  }
  router
    .replace({
      path: '/login',
      query: current.path === '/' ? undefined : { redirect: current.fullPath },
    })
    .then(done, done)
}
