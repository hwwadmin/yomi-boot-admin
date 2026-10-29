<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import { useFullscreen } from '@vueuse/core'
import { Expand, Fold, FullScreen, Key, Moon, Sunny, SwitchButton } from '@element-plus/icons-vue'

import { changePassword as changePasswordApi, logout as logoutApi } from '@/api/auth'
import { useAppStore } from '@/stores/app'
import { useUserStore } from '@/stores/user'
import Breadcrumb from './Breadcrumb.vue'

const appStore = useAppStore()
const userStore = useUserStore()
const router = useRouter()
const { isFullscreen, toggle: toggleFullscreen } = useFullscreen()

async function handleCommand(command: string) {
  if (command === 'change-password') {
    openChangePassword()
    return
  }
  if (command !== 'logout') return
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      type: 'warning',
      confirmButtonText: '退出',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }
  try {
    await logoutApi()
  } catch {
    // 退出接口失败不阻塞本地登出
  }
  userStore.reset()
  ElMessage.success('已退出登录')
  router.replace('/login')
}

// ———————— 修改密码 ————————
const changePwdVisible = ref(false)
const changePwdSubmitting = ref(false)
const changePwdRef = ref<FormInstance>()
const changePwdForm = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })

const changePwdRules: FormRules = {
  oldPassword: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 72, message: '密码长度 6-72 位', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (_rule, value: string, callback) => {
        if (value === changePwdForm.newPassword) callback()
        else callback(new Error('两次输入的新密码不一致'))
      },
      trigger: 'blur',
    },
  ],
}

function openChangePassword() {
  Object.assign(changePwdForm, { oldPassword: '', newPassword: '', confirmPassword: '' })
  changePwdVisible.value = true
}

async function handleChangePassword() {
  const valid = await changePwdRef.value?.validate().catch(() => false)
  if (!valid) return
  changePwdSubmitting.value = true
  try {
    // 后端会在改密成功后撤销该账号全部会话（含当前会话），必须重新登录
    await changePasswordApi({
      oldPassword: changePwdForm.oldPassword,
      newPassword: changePwdForm.newPassword,
    })
    ElMessage.success('密码已修改，请使用新密码重新登录')
    changePwdVisible.value = false
    userStore.reset()
    router.replace('/login')
  } catch {
    // 失败提示由 request 拦截器统一处理；改密失败保持当前登录态
  } finally {
    changePwdSubmitting.value = false
  }
}
</script>

<template>
  <div class="navbar">
    <div class="navbar-left">
      <el-icon class="action-btn" :size="18" @click="appStore.toggleSidebar()">
        <Fold v-if="!appStore.sidebarCollapsed" />
        <Expand v-else />
      </el-icon>
      <Breadcrumb />
    </div>
    <div class="navbar-right">
      <el-tooltip :content="appStore.isDark ? '切换为亮色' : '切换为暗色'" placement="bottom">
        <el-icon class="action-btn" :size="18" @click="appStore.toggleDark()">
          <Moon v-if="!appStore.isDark" />
          <Sunny v-else />
        </el-icon>
      </el-tooltip>
      <el-tooltip :content="isFullscreen ? '退出全屏' : '全屏'" placement="bottom">
        <el-icon class="action-btn" :size="18" @click="toggleFullscreen()">
          <FullScreen />
        </el-icon>
      </el-tooltip>
      <el-dropdown trigger="click" @command="handleCommand">
        <span class="user-info">
          <el-avatar :size="28" class="user-avatar">
            {{ userStore.userInfo?.nickname?.charAt(0) || 'U' }}
          </el-avatar>
          <span class="user-name">{{ userStore.userInfo?.nickname || '未登录' }}</span>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="change-password" :icon="Key">修改密码</el-dropdown-item>
            <el-dropdown-item command="logout" :icon="SwitchButton" divided>退出登录</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <el-dialog
      v-model="changePwdVisible"
      title="修改密码"
      width="420px"
      :close-on-click-modal="false"
      @closed="changePwdRef?.clearValidate()"
    >
      <el-form ref="changePwdRef" :model="changePwdForm" :rules="changePwdRules" label-width="80px">
        <el-form-item label="原密码" prop="oldPassword">
          <el-input
            v-model="changePwdForm.oldPassword"
            type="password"
            show-password
            autocomplete="current-password"
            maxlength="72"
          />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input
            v-model="changePwdForm.newPassword"
            type="password"
            show-password
            autocomplete="new-password"
            placeholder="6-72 位，修改后需重新登录"
            maxlength="72"
          />
        </el-form-item>
        <el-form-item label="确认新密码" prop="confirmPassword">
          <el-input
            v-model="changePwdForm.confirmPassword"
            type="password"
            show-password
            autocomplete="new-password"
            maxlength="72"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="changePwdVisible = false">取消</el-button>
        <el-button type="primary" :loading="changePwdSubmitting" @click="handleChangePassword">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.navbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
}

.navbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.navbar-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.action-btn {
  display: flex;
  align-items: center;
  color: var(--el-text-color-primary);
  cursor: pointer;
  transition: color 0.2s;
}

.action-btn:hover {
  color: var(--el-color-primary);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  outline: none;
}

.user-avatar {
  background-color: var(--el-color-primary);
  color: #fff;
  font-size: 14px;
}

.user-name {
  color: var(--el-text-color-primary);
  font-size: 14px;
}
</style>
