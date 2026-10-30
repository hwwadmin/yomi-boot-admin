import type { RouteRecordRaw } from 'vue-router'

import AdminLayout from '@/layouts/AdminLayout.vue'

/** 静态路由；后端菜单只决定可见性，不动态生成路由 */
export const staticRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/login/index.vue'),
    meta: { titleKey: 'menu.login' },
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
        meta: { titleKey: 'menu.dashboard', icon: 'Odometer' },
      },
      {
        // 纯分组路由：无组件，子路由使用绝对路径，子页面 URL 与后端菜单 uiPath 保持不变
        path: 'system',
        name: 'system',
        meta: { titleKey: 'menu.system', icon: 'Tools' },
        children: [
          {
            path: '/system/user',
            name: 'system-user',
            component: () => import('@/views/system/user/index.vue'),
            meta: { titleKey: 'menu.systemUser', icon: 'User' },
          },
          {
            path: '/system/rbac',
            name: 'system-rbac',
            meta: { titleKey: 'menu.systemRbac', icon: 'Lock' },
            children: [
              {
                path: '/system/role',
                name: 'system-role',
                component: () => import('@/views/system/role/index.vue'),
                meta: { titleKey: 'menu.systemRole', icon: 'Avatar' },
              },
              {
                path: '/system/permission-group',
                name: 'system-permission-group',
                component: () => import('@/views/system/permission-group/index.vue'),
                meta: { titleKey: 'menu.systemPermissionGroup', icon: 'Grid' },
              },
              {
                path: '/system/permission',
                name: 'system-permission',
                component: () => import('@/views/system/permission/index.vue'),
                meta: { titleKey: 'menu.systemPermission', icon: 'Key' },
              },
            ],
          },
          {
            path: '/system/http-record',
            name: 'system-http-record',
            component: () => import('@/views/system/http-record/index.vue'),
            meta: { titleKey: 'menu.systemHttpRecord', icon: 'Tickets' },
          },
          {
            path: '/system/param',
            name: 'system-param',
            component: () => import('@/views/system/param/index.vue'),
            meta: { titleKey: 'menu.systemParam', icon: 'Setting' },
          },
        ],
      },
    ],
  },
  {
    path: '/403',
    name: 'forbidden',
    component: () => import('@/views/error/Forbidden.vue'),
    meta: { titleKey: 'menu.forbidden', hidden: true },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/views/error/NotFound.vue'),
    meta: { titleKey: 'menu.notFound', hidden: true },
  },
]
