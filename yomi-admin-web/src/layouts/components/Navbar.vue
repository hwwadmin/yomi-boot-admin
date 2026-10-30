<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useFullscreen } from '@vueuse/core'
import { Expand, Fold, FullScreen, Key, Moon, Sunny, SwitchButton } from '@element-plus/icons-vue'

import { changePassword as changePasswordApi, logout as logoutApi } from '@/api/auth'
import { useAppStore } from '@/stores/app'
import { useLocaleStore } from '@/stores/locale'
import { useUserStore } from '@/stores/user'
import { confirmAction } from '@/utils/confirm'
import Breadcrumb from './Breadcrumb.vue'

const appStore = useAppStore()
const localeStore = useLocaleStore()
const userStore = useUserStore()
const router = useRouter()
const { t } = useI18n()
const { isFullscreen, toggle: toggleFullscreen } = useFullscreen()

async function handleCommand(command: string) {
  if (command === 'change-password') {
    openChangePassword()
    return
  }
  if (command !== 'logout') return

  const confirmed = await confirmAction({
    title: t('layout.logout.confirmTitle'),
    message: t('layout.logout.confirmMessage'),
    confirmButtonText: t('layout.logout.confirmButton'),
  })
  if (!confirmed) return

  try {
    await logoutApi()
  } catch {
    // 退出接口失败不阻塞本地登出
  }
  userStore.reset()
  ElMessage.success(t('layout.logout.success'))
  router.replace('/login')
}

// ———————— 修改密码 ————————
const changePwdVisible = ref(false)
const changePwdSubmitting = ref(false)
const changePwdRef = ref<FormInstance>()
const changePwdForm = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })

const changePwdRules = computed<FormRules>(() => ({
  oldPassword: [
    { required: true, message: t('layout.changePassword.validation.oldRequired'), trigger: 'blur' },
  ],
  newPassword: [
    { required: true, message: t('layout.changePassword.validation.newRequired'), trigger: 'blur' },
    { min: 6, max: 72, message: t('layout.changePassword.validation.length'), trigger: 'blur' },
  ],
  confirmPassword: [
    {
      required: true,
      message: t('layout.changePassword.validation.confirmRequired'),
      trigger: 'blur',
    },
    {
      validator: (_rule, value: string, callback) => {
        if (value === changePwdForm.newPassword) callback()
        else callback(new Error(t('layout.changePassword.validation.mismatch')))
      },
      trigger: 'blur',
    },
  ],
}))

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
    ElMessage.success(t('layout.changePassword.success'))
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
      <el-tooltip :content="t('layout.language.switch')" placement="bottom">
        <span class="action-btn lang-btn" @click="localeStore.toggleLocale()">
          {{
            localeStore.locale === 'zh-CN'
              ? t('layout.language.shortEn')
              : t('layout.language.shortZh')
          }}
        </span>
      </el-tooltip>
      <el-tooltip
        :content="t(appStore.isDark ? 'layout.theme.toLight' : 'layout.theme.toDark')"
        placement="bottom"
      >
        <el-icon class="action-btn" :size="18" @click="appStore.toggleDark()">
          <Moon v-if="!appStore.isDark" />
          <Sunny v-else />
        </el-icon>
      </el-tooltip>
      <el-tooltip
        :content="t(isFullscreen ? 'layout.fullscreen.exit' : 'layout.fullscreen.enter')"
        placement="bottom"
      >
        <el-icon class="action-btn" :size="18" @click="toggleFullscreen()">
          <FullScreen />
        </el-icon>
      </el-tooltip>
      <el-dropdown trigger="click" @command="handleCommand">
        <span class="user-info">
          <el-avatar :size="28" class="user-avatar">
            {{ userStore.userInfo?.nickname?.charAt(0) || 'U' }}
          </el-avatar>
          <span class="user-name">
            {{ userStore.userInfo?.nickname || t('layout.user.fallbackName') }}
          </span>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="change-password" :icon="Key">
              {{ t('layout.menu.changePassword') }}
            </el-dropdown-item>
            <el-dropdown-item command="logout" :icon="SwitchButton" divided>
              {{ t('layout.menu.logout') }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <el-dialog
      v-model="changePwdVisible"
      :title="t('layout.changePassword.title')"
      width="460px"
      :close-on-click-modal="false"
      @closed="changePwdRef?.clearValidate()"
    >
      <el-form
        ref="changePwdRef"
        :model="changePwdForm"
        :rules="changePwdRules"
        label-width="110px"
      >
        <el-form-item :label="t('layout.changePassword.oldPassword')" prop="oldPassword">
          <el-input
            v-model="changePwdForm.oldPassword"
            type="password"
            show-password
            autocomplete="current-password"
            :placeholder="t('layout.changePassword.oldPlaceholder')"
            maxlength="72"
          />
        </el-form-item>
        <el-form-item :label="t('layout.changePassword.newPassword')" prop="newPassword">
          <el-input
            v-model="changePwdForm.newPassword"
            type="password"
            show-password
            autocomplete="new-password"
            :placeholder="t('layout.changePassword.newPlaceholder')"
            maxlength="72"
          />
        </el-form-item>
        <el-form-item :label="t('layout.changePassword.confirmPassword')" prop="confirmPassword">
          <el-input
            v-model="changePwdForm.confirmPassword"
            type="password"
            show-password
            autocomplete="new-password"
            :placeholder="t('layout.changePassword.confirmPlaceholder')"
            maxlength="72"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="changePwdVisible = false">{{ t('common.action.cancel') }}</el-button>
        <el-button type="primary" :loading="changePwdSubmitting" @click="handleChangePassword">
          {{ t('common.action.confirm') }}
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

.lang-btn {
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.5px;
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
