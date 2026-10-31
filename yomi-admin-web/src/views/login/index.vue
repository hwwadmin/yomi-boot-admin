<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { ArrowRight, Key, Lock, User } from '@element-plus/icons-vue'

import { getCaptcha, getLoginConfig } from '@/api/auth'
import type { CaptchaChallenge } from '@/api/types'
import illustration from '@/assets/login-illustration.svg'
import logoUrl from '@/assets/logo.svg'
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
    <!-- 背景装饰层 -->
    <div class="bg-wash" aria-hidden="true"></div>
    <div class="bg-dots" aria-hidden="true"></div>
    <div class="bg-wave bg-wave--mid" aria-hidden="true"></div>
    <div class="bg-wave bg-wave--deep" aria-hidden="true"></div>

    <div class="login-body">
      <!-- 左侧品牌区 -->
      <section class="login-intro">
        <div class="intro-brand">
          <img class="intro-badge" :src="logoUrl" alt="" aria-hidden="true" />
          <h1 class="intro-title">Yomi Admin</h1>
        </div>
        <p class="intro-slogan">{{ t('auth.slogan') }}</p>
        <img class="intro-illustration" :src="illustration" :alt="t('auth.illustrationAlt')" />
      </section>

      <!-- 右侧登录卡片 -->
      <section class="login-panel">
        <div class="login-card">
          <img class="card-badge" :src="logoUrl" alt="" aria-hidden="true" />
          <div class="card-title">Yomi Admin</div>
          <div class="card-subtitle">{{ t('auth.subtitle') }}</div>

          <el-form
            ref="formRef"
            class="login-form"
            :model="form"
            :rules="rules"
            size="large"
            @keyup.enter="handleLogin"
          >
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
              <el-icon v-if="!loading" class="login-btn__arrow"><ArrowRight /></el-icon>
            </el-button>
          </el-form>

          <div v-if="showMockTip" class="mock-tip">{{ t('auth.mockTip') }}</div>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
/* ———————— 页面与背景装饰 ———————— */
.login-page {
  position: relative;
  display: flex;
  height: 100%;
  /* clip 禁止一切滚动（含 scrollIntoView），防止装饰层溢出把整页"平移"走；不支持 clip 的旧浏览器回退 hidden */
  overflow: hidden;
  overflow: clip;
  background: linear-gradient(158deg, #f3f7fd 0%, #e9f0fb 48%, #dfe9fa 100%);
}

/* 左上柔光 */
.bg-wash {
  position: absolute;
  top: -28%;
  left: -14%;
  width: 64%;
  height: 72%;
  background: radial-gradient(closest-side, rgb(255 255 255 / 80%), rgb(255 255 255 / 0%) 74%);
  pointer-events: none;
}

/* 右上点阵，向左下淡出 */
.bg-dots {
  position: absolute;
  top: 52px;
  right: 64px;
  width: 236px;
  height: 118px;
  background-image: radial-gradient(rgb(122 156 216 / 55%) 2.2px, transparent 2.6px);
  background-size: 28px 28px;
  -webkit-mask-image: linear-gradient(215deg, #000 30%, transparent 88%);
  mask-image: linear-gradient(215deg, #000 30%, transparent 88%);
  pointer-events: none;
}

/* 底部两层波浪 */
.bg-wave {
  position: absolute;
  border-radius: 50%;
  pointer-events: none;
}

.bg-wave--mid {
  bottom: -42%;
  left: -16%;
  width: 130%;
  height: 68%;
  background: rgb(205 222 248 / 78%);
  transform: rotate(-5deg);
}

.bg-wave--deep {
  bottom: -46%;
  right: -22%;
  width: 128%;
  height: 64%;
  background: rgb(184 208 244 / 62%);
  transform: rotate(4deg);
}

/* ———————— 主体两栏 ———————— */
.login-body {
  position: relative;
  z-index: 1;
  display: flex;
  flex: 1;
  align-items: center;
  justify-content: center;
  gap: clamp(48px, 7vw, 120px);
  padding: 48px clamp(24px, 6vw, 110px);
}

/* 左侧品牌区 */
.login-intro {
  display: flex;
  flex-direction: column;
  flex: 1 1 0;
  align-items: flex-start;
  max-width: 560px;
}

.intro-brand {
  display: flex;
  align-items: center;
  gap: 18px;
}

.intro-badge {
  flex-shrink: 0;
  width: 56px;
  height: 56px;
  filter: drop-shadow(0 6px 12px rgb(15 47 68 / 30%)) drop-shadow(0 20px 36px rgb(15 47 68 / 26%));
}

.intro-title {
  margin: 0;
  color: #17233d;
  font-size: 34px;
  font-weight: 800;
  letter-spacing: 0.5px;
}

.intro-slogan {
  margin: 14px 0 0 2px;
  color: #5f7396;
  font-size: 16px;
  letter-spacing: 0.04em;
}

.intro-illustration {
  width: min(440px, 92%);
  margin: 40px 0 0 6px;
  filter: drop-shadow(0 20px 32px rgb(43 111 224 / 10%));
}

/* ———————— 右侧登录卡片 ———————— */
.login-panel {
  display: flex;
  flex: 0 1 470px;
  justify-content: center;
  min-width: 0;
}

.login-card {
  --el-component-size-large: 44px;

  width: 100%;
  max-width: 470px;
  padding: 44px 44px 36px;
  background: #fff;
  border-radius: 16px;
  box-shadow:
    0 24px 60px -32px rgb(43 92 176 / 35%),
    0 2px 6px rgb(43 92 176 / 6%);
}

.card-badge {
  display: block;
  width: 46px;
  height: 46px;
  margin: 0 auto;
  filter: drop-shadow(0 5px 10px rgb(15 47 68 / 28%)) drop-shadow(0 16px 28px rgb(15 47 68 / 24%));
}

.card-title {
  margin-top: 14px;
  color: #1c2b4a;
  font-size: 22px;
  font-weight: 700;
  text-align: center;
  letter-spacing: 0.5px;
}

.card-subtitle {
  margin: 6px 0 26px;
  color: var(--app-text-3);
  font-size: 13px;
  text-align: center;
}

.login-form :deep(.el-form-item) {
  margin-bottom: 20px;
}

.login-form :deep(.el-input__wrapper) {
  border-radius: 8px;
}

/* ———————— 验证码 ———————— */
.captcha-row {
  display: flex;
  gap: 12px;
  width: 100%;
}

.captcha-image {
  position: relative;
  display: flex;
  flex: 0 0 122px;
  align-items: center;
  justify-content: center;
  height: 44px;
  padding: 0;
  overflow: hidden;
  color: var(--app-text-3);
  font-size: 12px;
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
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

/* ———————— 登录按钮 ———————— */
.login-btn {
  position: relative;
  width: 100%;
  margin-top: 6px;
  font-size: 16px;
  font-weight: 600;
  letter-spacing: 6px;
  text-indent: 6px;

  --el-button-bg-color: #2e7cf6;
  --el-button-border-color: #2e7cf6;
  --el-button-hover-bg-color: #4a90f8;
  --el-button-hover-border-color: #4a90f8;
  --el-button-active-bg-color: #2b6fe0;
  --el-button-active-border-color: #2b6fe0;
}

.login-btn__arrow {
  position: absolute;
  top: 50%;
  right: 16px;
  font-size: 17px;
  transform: translateY(-50%);
}

.mock-tip {
  margin: 14px 0 0;
  color: var(--app-text-3);
  font-size: 12px;
  text-align: center;
}

/* ———————— 暗色主题 ———————— */
html.dark .login-page {
  background: linear-gradient(158deg, #0b1220 0%, #0d1526 52%, #101c30 100%);
}

html.dark .bg-wash {
  background: radial-gradient(closest-side, rgb(64 158 255 / 10%), rgb(64 158 255 / 0%) 74%);
}

html.dark .bg-dots {
  background-image: radial-gradient(rgb(124 170 255 / 30%) 2.2px, transparent 2.6px);
}

html.dark .bg-wave--mid {
  background: rgb(64 158 255 / 7%);
}

html.dark .bg-wave--deep {
  background: rgb(64 158 255 / 5%);
}

html.dark .login-card {
  background: var(--app-surface);
  border: 1px solid var(--app-border);
  box-shadow: var(--app-shadow-pop);
}

html.dark .intro-title {
  color: #e6eaf2;
}

html.dark .card-title {
  color: #e6eaf2;
}

html.dark .intro-slogan {
  color: #8a97ad;
}

/* ———————— 响应式 ———————— */
@media (max-width: 1280px) {
  .intro-illustration {
    width: min(380px, 92%);
  }
}

@media (max-width: 992px) {
  .login-intro {
    display: none;
  }

  .bg-dots {
    right: 32px;
  }
}

@media (max-width: 480px) {
  .login-body {
    padding: 24px 16px;
  }

  .login-card {
    padding: 34px 24px 28px;
  }
}
</style>
