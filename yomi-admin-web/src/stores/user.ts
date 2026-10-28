import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { useStorage } from '@vueuse/core'

import { getProfile, login as loginApi } from '@/api/auth'
import type { UserInfo } from '@/api/types'

const TOKEN_KEY = 'yomi:token'

export const useUserStore = defineStore('user', () => {
  const token = useStorage(TOKEN_KEY, '')
  const userInfo = ref<UserInfo | null>(null)

  const permissions = computed(() => userInfo.value?.permissions ?? [])

  /** 登录并保存 token 与用户信息（错误提示由 request 拦截器统一处理） */
  async function login(username: string, password: string) {
    const { token: newToken, user } = await loginApi({ username, password })
    token.value = newToken
    userInfo.value = user
  }

  /** 拉取当前登录用户信息（刷新页面后恢复登录态用） */
  async function fetchProfile() {
    userInfo.value = await getProfile()
  }

  /** 清空登录态；跳转由调用方或响应拦截器负责 */
  function reset() {
    token.value = ''
    userInfo.value = null
  }

  return { token, userInfo, permissions, login, fetchProfile, reset }
})
