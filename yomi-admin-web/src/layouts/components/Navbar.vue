<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useFullscreen } from '@vueuse/core'
import {
  CaretBottom,
  Expand,
  Fold,
  FullScreen,
  Key,
  Moon,
  Sunny,
  SwitchButton,
  User,
} from '@element-plus/icons-vue'

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
  if (command === 'profile') {
    openProfile()
    return
  }
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

// ———————— 个人资料 ————————
const profileVisible = ref(false)
const profileSubmitting = ref(false)
const profileRef = ref<FormInstance>()
const profileForm = reactive({ nickname: '', email: '', phone: '' })

const profileRoles = computed(() => userStore.userInfo?.roles ?? [])

const profileRules = computed<FormRules>(() => ({
  nickname: [
    { required: true, message: t('layout.profile.validation.nicknameRequired'), trigger: 'blur' },
  ],
  email: [{ type: 'email', message: t('validation.email'), trigger: 'blur' }],
  phone: [{ pattern: /^1\d{10}$/, message: t('validation.phone'), trigger: 'blur' }],
}))

function openProfile() {
  Object.assign(profileForm, {
    nickname: userStore.userInfo?.nickname ?? '',
    email: userStore.userInfo?.email ?? '',
    phone: userStore.userInfo?.phone ?? '',
  })
  profileVisible.value = true
}

async function handleUpdateProfile() {
  const valid = await profileRef.value?.validate().catch(() => false)
  if (!valid) return
  profileSubmitting.value = true
  try {
    await userStore.updateProfile({
      nickname: profileForm.nickname.trim(),
      email: profileForm.email.trim(),
      phone: profileForm.phone.trim(),
    })
    ElMessage.success(t('layout.profile.success'))
    profileVisible.value = false
  } catch {
    // 失败提示由 request 拦截器统一处理；保存失败时保持弹窗与已填内容
  } finally {
    profileSubmitting.value = false
  }
}
</script>

<template>
  <div class="navbar">
    <div class="navbar-left">
      <button
        type="button"
        class="navbar-action"
        :aria-label="
          appStore.sidebarCollapsed ? t('layout.sidebar.expand') : t('layout.sidebar.collapse')
        "
        @click="appStore.toggleSidebar()"
      >
        <el-icon :size="18">
          <Fold v-if="!appStore.sidebarCollapsed" />
          <Expand v-else />
        </el-icon>
      </button>
      <Breadcrumb />
    </div>
    <div class="navbar-right">
      <el-tooltip :content="t('layout.language.switch')" placement="bottom">
        <button
          type="button"
          class="navbar-action navbar-action--text"
          :aria-label="t('layout.language.switch')"
          @click="localeStore.toggleLocale()"
        >
          {{
            localeStore.locale === 'zh-CN'
              ? t('layout.language.shortEn')
              : t('layout.language.shortZh')
          }}
        </button>
      </el-tooltip>
      <el-tooltip
        :content="t(appStore.isDark ? 'layout.theme.toLight' : 'layout.theme.toDark')"
        placement="bottom"
      >
        <button
          type="button"
          class="navbar-action"
          :aria-label="t(appStore.isDark ? 'layout.theme.toLight' : 'layout.theme.toDark')"
          @click="appStore.toggleDark()"
        >
          <el-icon :size="18">
            <Moon v-if="!appStore.isDark" />
            <Sunny v-else />
          </el-icon>
        </button>
      </el-tooltip>
      <el-tooltip
        :content="t(isFullscreen ? 'layout.fullscreen.exit' : 'layout.fullscreen.enter')"
        placement="bottom"
      >
        <button
          type="button"
          class="navbar-action"
          :aria-label="t(isFullscreen ? 'layout.fullscreen.exit' : 'layout.fullscreen.enter')"
          @click="toggleFullscreen()"
        >
          <el-icon :size="18"><FullScreen /></el-icon>
        </button>
      </el-tooltip>
      <span class="navbar-divider" />
      <el-dropdown trigger="click" @command="handleCommand">
        <span class="user-info">
          <el-avatar :size="30" class="user-avatar">
            {{ userStore.userInfo?.nickname?.charAt(0) || 'U' }}
          </el-avatar>
          <span class="user-name">
            {{ userStore.userInfo?.nickname || t('layout.user.fallbackName') }}
          </span>
          <el-icon class="user-caret" :size="12"><CaretBottom /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="profile" :icon="User">
              {{ t('layout.menu.profile') }}
            </el-dropdown-item>
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
      v-model="profileVisible"
      :title="t('layout.profile.title')"
      width="480px"
      :close-on-click-modal="false"
      @closed="profileRef?.clearValidate()"
    >
      <div class="profile-summary">
        <el-avatar :size="52" class="user-avatar">
          {{ userStore.userInfo?.nickname?.charAt(0) || 'U' }}
        </el-avatar>
        <div class="profile-summary-text">
          <span class="profile-summary-name">{{ userStore.userInfo?.nickname }}</span>
          <span class="profile-summary-account">@{{ userStore.userInfo?.username }}</span>
        </div>
      </div>
      <el-form
        ref="profileRef"
        :model="profileForm"
        :rules="profileRules"
        label-width="auto"
        :validate-on-rule-change="false"
      >
        <el-form-item :label="t('common.field.nickname')" prop="nickname">
          <el-input
            v-model="profileForm.nickname"
            maxlength="32"
            :placeholder="t('layout.profile.nicknamePlaceholder')"
          />
        </el-form-item>
        <el-form-item :label="t('common.field.email')" prop="email">
          <el-input
            v-model="profileForm.email"
            maxlength="64"
            :placeholder="t('layout.profile.emailPlaceholder')"
          />
        </el-form-item>
        <el-form-item :label="t('common.field.phone')" prop="phone">
          <el-input
            v-model="profileForm.phone"
            maxlength="11"
            :placeholder="t('layout.profile.phonePlaceholder')"
          />
        </el-form-item>
        <el-form-item :label="t('common.field.role')">
          <div class="profile-roles">
            <el-tag
              v-for="role in profileRoles"
              :key="role.code"
              :type="role.status === 'ENABLED' ? 'primary' : 'info'"
              effect="light"
              size="small"
            >
              {{ role.name }}
            </el-tag>
            <span v-if="profileRoles.length === 0" class="profile-empty">
              {{ t('common.none') }}
            </span>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="profileVisible = false">{{ t('common.action.cancel') }}</el-button>
        <el-button type="primary" :loading="profileSubmitting" @click="handleUpdateProfile">
          {{ t('common.action.save') }}
        </el-button>
      </template>
    </el-dialog>

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
        label-width="auto"
        :validate-on-rule-change="false"
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
          <span class="form-hint">{{ t('layout.changePassword.newHint') }}</span>
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
  gap: 16px;
  height: 100%;
}

.navbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.navbar-right {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  gap: 6px;
}

.navbar-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--app-text-2);
  cursor: pointer;
  transition:
    color 0.18s ease,
    background-color 0.18s ease;
}

.navbar-action:hover {
  color: var(--el-color-primary);
  background-color: var(--el-fill-color-light);
}

.navbar-action:focus-visible {
  outline: 2px solid var(--el-color-primary-light-5);
  outline-offset: 2px;
}

.navbar-action--text {
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.4px;
}

.navbar-divider {
  width: 1px;
  height: 20px;
  margin: 0 6px;
  background: var(--app-border);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px 4px 4px;
  border-radius: 8px;
  cursor: pointer;
  outline: none;
  transition: background-color 0.18s ease;
}

.user-info:hover {
  background-color: var(--el-fill-color-light);
}

.user-avatar {
  background: linear-gradient(135deg, #4a9bff 0%, #2b6fe0 100%);
  box-shadow: 0 4px 10px -6px rgb(43 111 224 / 80%);
  color: #fff;
  font-size: 13px;
  font-weight: 600;
}

.user-name {
  color: var(--app-text-1);
  font-size: 14px;
  font-weight: 500;
}

.user-caret {
  color: var(--app-text-3);
}

.profile-summary {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  margin-bottom: 18px;
  border: 1px solid var(--app-border);
  border-radius: 10px;
  background: var(--el-fill-color-lighter);
}

.profile-summary-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.profile-summary-name {
  overflow: hidden;
  color: var(--app-text-1);
  font-size: 15px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.profile-summary-account {
  color: var(--app-text-3);
  font-size: 12px;
}

.profile-roles {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
  min-height: 24px;
}

.profile-empty {
  color: var(--app-text-3);
  font-size: 13px;
}

.form-hint {
  display: block;
  margin-top: 4px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.4;
}

@media (max-width: 768px) {
  .user-name {
    display: none;
  }

  .navbar-divider {
    display: none;
  }
}
</style>
