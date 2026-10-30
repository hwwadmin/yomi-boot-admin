<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { Key, Lock, User } from '@element-plus/icons-vue'

import { getCaptcha, getLoginConfig } from '@/api/auth'
import type { CaptchaChallenge } from '@/api/types'
import { useUserStore } from '@/stores/user'

defineOptions({ name: 'LoginPage' })

const userStore = useUserStore()
const router = useRouter()
const route = useRoute()
const { t } = useI18n()

const formRef = ref<FormInstance>()
const loading = ref(false)
const captcha = ref<CaptchaChallenge>()
const captchaLoading = ref(false)
const captchaStale = ref(false)
const form = reactive({ username: '', password: '', captchaCode: '' })

// 登录页配置未返回前按需要验证码处理，避免已启用部署出现表单闪变
const captchaRequired = ref(true)

const rules = computed<FormRules>(() => ({
  username: [{ required: true, message: t('auth.validation.usernameRequired'), trigger: 'blur' }],
  password: [{ required: true, message: t('auth.validation.passwordRequired'), trigger: 'blur' }],
  captchaCode: captchaRequired.value
    ? [{ required: true, message: t('auth.validation.captchaRequired'), trigger: 'blur' }]
    : [],
}))

const showMockTip = import.meta.env.DEV && import.meta.env.VITE_USE_MOCK === 'true'

/** 拉取登录页配置；失败按需要验证码处理，后续登录会暴露真实错误 */
async function loadLoginConfig() {
  const config = await getLoginConfig()
  captchaRequired.value = config.captchaRequired
}

/** 拉取新验证码；刷新失败（如触发限流）时保留当前图片 */
async function refreshCaptcha(silent = false) {
  if (captchaLoading.value) return
  captchaLoading.value = true
  try {
    captcha.value = await getCaptcha({ silent })
    captchaStale.value = false
  } catch {
    // 失败提示由 request 拦截器统一处理；静默换图时只保留失效蒙层
  } finally {
    captchaLoading.value = false
  }
}

async function handleLogin() {
  // 提交中直接忽略重复触发（回车键不经过按钮的 loading 禁用态）
  if (loading.value) return
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    await userStore.login(
      form.username.trim(),
      form.password,
      captchaRequired.value ? (captcha.value?.id ?? '') : '',
      captchaRequired.value ? form.captchaCode.trim() : '',
    )
    ElMessage.success(t('auth.success'))
    const redirect = (route.query.redirect as string) || '/'
    router.replace(redirect)
  } catch {
    // 登录失败时验证码已在服务端消费：清空输入并静默换一张，避免限流提示盖住真正的登录错误
    form.captchaCode = ''
    captchaStale.value = true
    if (captchaRequired.value) await refreshCaptcha(true)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try {
    await loadLoginConfig()
  } catch {
    // 配置拉取失败按需要验证码处理
  }
  if (captchaRequired.value) await refreshCaptcha()
})
</script>

<template>
  <div class="login-page">
    <el-card class="login-card" shadow="never">
      <div class="login-brand"><span class="login-brand__badge">Y</span></div>
      <div class="login-title">Yomi Admin</div>
      <div class="login-subtitle">{{ t('auth.subtitle') }}</div>
      <el-form ref="formRef" :model="form" :rules="rules" size="large" @keyup.enter="handleLogin">
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            :placeholder="t('auth.usernamePlaceholder')"
            :prefix-icon="User"
            autocomplete="username"
            maxlength="64"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            :placeholder="t('auth.passwordPlaceholder')"
            :prefix-icon="Lock"
            show-password
            autocomplete="current-password"
            maxlength="72"
          />
        </el-form-item>
        <el-form-item v-if="captchaRequired" prop="captchaCode">
          <div class="captcha-row">
            <el-input
              v-model="form.captchaCode"
              :placeholder="t('auth.captchaPlaceholder')"
              :prefix-icon="Key"
              maxlength="8"
              autocomplete="off"
            />
            <button
              type="button"
              class="captcha-image"
              :title="t('auth.captchaRefreshTitle')"
              :disabled="captchaLoading"
              @click="refreshCaptcha()"
            >
              <img v-if="captcha" :src="captcha.image" :alt="t('auth.captchaAlt')" />
              <span v-else>
                {{ captchaLoading ? t('auth.captchaLoading') : t('auth.captchaLoad') }}
              </span>
              <span v-if="captcha && captchaStale" class="captcha-stale-tip">
                {{ t('auth.captchaStale.line1') }}<br />{{ t('auth.captchaStale.line2') }}
              </span>
            </button>
          </div>
        </el-form-item>
        <el-button
          class="login-btn"
          type="primary"
          size="large"
          :loading="loading"
          @click="handleLogin"
        >
          {{ t('auth.submit') }}
        </el-button>
        <div v-if="showMockTip" class="mock-tip">{{ t('auth.mockTip') }}</div>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  background:
    radial-gradient(900px 480px at 10% -10%, rgb(64 158 255 / 16%), transparent 62%),
    radial-gradient(760px 420px at 100% 108%, rgb(34 184 207 / 12%), transparent 62%), var(--app-bg);
}

.login-card {
  width: 400px;
  border: 1px solid var(--app-border);
  border-radius: 12px;
  box-shadow: var(--app-shadow-pop);
}

.login-brand {
  display: flex;
  justify-content: center;
  margin-top: 4px;
}

.login-brand__badge {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: linear-gradient(135deg, #4a9bff 0%, #2b6fe0 100%);
  box-shadow: 0 10px 24px -12px rgb(43 111 224 / 85%);
  color: #fff;
  font-size: 20px;
  font-weight: 700;
}

.login-title {
  margin-top: 16px;
  color: var(--app-text-1);
  font-size: 22px;
  font-weight: 700;
  text-align: center;
  letter-spacing: 0.5px;
}

.login-subtitle {
  margin: 6px 0 26px;
  color: var(--app-text-3);
  font-size: 13px;
  text-align: center;
}

.captcha-row {
  display: flex;
  gap: 12px;
  width: 100%;
}

.captcha-image {
  position: relative;
  display: flex;
  flex: 0 0 120px;
  align-items: center;
  justify-content: center;
  height: 42px;
  padding: 0;
  overflow: hidden;
  color: var(--app-text-3);
  font-size: 12px;
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color);
  border-radius: var(--app-radius-base);
  cursor: pointer;
}

.captcha-image:disabled {
  cursor: wait;
}

.captcha-image img {
  width: 100%;
  height: 100%;
  object-fit: fill;
}

.captcha-stale-tip {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 12px;
  line-height: 1.3;
  text-align: center;
  background: rgb(0 0 0 / 45%);
}

.login-btn {
  width: 100%;
  margin-top: 4px;
  font-weight: 500;
  letter-spacing: 0.5px;
}

.mock-tip {
  margin: 14px 0 4px;
  color: var(--app-text-3);
  font-size: 12px;
  text-align: center;
}
</style>
