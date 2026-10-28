import { ref } from 'vue'
import { defineStore } from 'pinia'
import type { RouteRecordRaw } from 'vue-router'

/**
 * 动态路由（后端下发菜单）预留。
 * 接入真实后端后：
 * 1. generateRoutes 内请求菜单/权限接口，把菜单树转换为 RouteRecordRaw[]（component 用 () => import('@/views/**') 路由表映射）；
 * 2. 路由守卫（router/guard.ts）中逐个 router.addRoute('root', route)；
 * 3. 404 通配路由要在动态路由全部 addRoute 之后注册（或注册后再补），避免刷新时先命中 404。
 */
export const usePermissionStore = defineStore('permission', () => {
  const dynamicRoutes = ref<RouteRecordRaw[]>([])

  async function generateRoutes(): Promise<RouteRecordRaw[]> {
    // TODO: 对接后端菜单/权限接口后在此生成动态路由
    return dynamicRoutes.value
  }

  return { dynamicRoutes, generateRoutes }
})
