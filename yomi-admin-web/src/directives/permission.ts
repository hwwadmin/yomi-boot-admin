import type { Directive } from 'vue'

import { useUserStore } from '@/stores/user'

/**
 * 按钮级权限指令（预留）：
 *   <el-button v-permission="'system:user:delete'">删除</el-button>
 *   <el-button v-permission="['system:user:edit', 'system:user:delete']">...</el-button>
 * 用户权限（userInfo.permissions）不含所需标识且无 '*' 时，直接移除元素。
 */
export const permission: Directive<HTMLElement, string | string[]> = {
  mounted(el, binding) {
    const { permissions } = useUserStore()
    const required = Array.isArray(binding.value) ? binding.value : [binding.value]
    const allowed = permissions.includes('*') || required.some((item) => permissions.includes(item))
    if (!allowed) {
      el.parentNode?.removeChild(el)
    }
  },
}
