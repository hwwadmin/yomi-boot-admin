import { createRouter, createWebHashHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

import AdminLayout from '@/layouts/AdminLayout.vue'

/** 静态路由；动态路由见 stores/permission.ts */
export const staticRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/login/index.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/',
    name: 'root',
    component: AdminLayout,
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'dashboard',
        component: () => import('@/views/dashboard/index.vue'),
        meta: { title: '首页', icon: 'Odometer' },
      },
      {
        path: 'system/user',
        name: 'system-user',
        component: () => import('@/views/system/user/index.vue'),
        meta: { title: '系统用户', icon: 'User' },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/views/error/NotFound.vue'),
    meta: { title: '404', hidden: true },
  },
]

const router = createRouter({
  // hash 模式：管理后台无需 SEO，且部署时无需服务端 fallback 配置
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: staticRoutes,
})

export default router
