import { createRouter, createWebHashHistory } from 'vue-router'

import { staticRoutes } from './routes'

const router = createRouter({
  // hash 模式：管理后台无需 SEO，且部署时无需服务端 fallback 配置
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: staticRoutes,
})

export default router
