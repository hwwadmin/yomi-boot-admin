import type { RouteRecordRaw } from 'vue-router'

import AdminLayout from '@/layouts/AdminLayout.vue'

/** 静态路由；后端菜单只决定可见性，不动态生成路由 */
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
      {
        path: 'system/role',
        name: 'system-role',
        component: () => import('@/views/system/role/index.vue'),
        meta: { title: '角色管理', icon: 'Avatar' },
      },
      {
        path: 'system/permission',
        name: 'system-permission',
        component: () => import('@/views/system/permission/index.vue'),
        meta: { title: '权限管理', icon: 'Key' },
      },
      {
        path: 'system/permission-group',
        name: 'system-permission-group',
        component: () => import('@/views/system/permission-group/index.vue'),
        meta: { title: '权限组', icon: 'Grid' },
      },
      {
        path: 'system/http-record',
        name: 'system-http-record',
        component: () => import('@/views/system/http-record/index.vue'),
        meta: { title: '访问记录', icon: 'Tickets' },
      },
    ],
  },
  {
    path: '/403',
    name: 'forbidden',
    component: () => import('@/views/error/Forbidden.vue'),
    meta: { title: '无权访问', hidden: true },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/views/error/NotFound.vue'),
    meta: { title: '404', hidden: true },
  },
]
