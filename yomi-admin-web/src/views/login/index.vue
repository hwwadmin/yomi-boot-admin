<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import { Key, Lock, User } from '@element-plus/icons-vue'

import { getCaptcha, getLoginConfig } from '@/api/auth'
import type { CaptchaChallenge } from '@/api/types'
import { useUserStore } from '@/stores/user'

defineOptions({ name: 'LoginPage' })

const userStore = useUserStore()
const router = useRouter()
const route = useRoute()

const formRef = ref<FormInstance>()
const loading = ref(false)
const captcha = ref<CaptchaChallenge>()
const captchaLoading = ref(false)
const captchaStale = ref(false)
const form = reactive({ username: '', password: '', captchaCode: '' })

// 登录页配置未返回前按需要验证码处理，避免已启用部署出现表单闪变
const captchaRequired = ref(true)

const rules = computed<FormRules>(() => ({
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  captchaCode: captchaRequired.value
    ? [{ required: true, message: '请输入验证码', trigger: 'blur' }]
    : [],
}))

const showMockTip = import.meta.env.DEV && import.meta.env.VITE_USE_MOCK === 'true'

/** 拉取登录页配置；失败按需要验证码处理，后续登录会暴露真实错误 */
async function loadLoginConfig() {
  const config = await getLoginConfig()
  captchaRequired.value = config.captchaRequired
}

/** 拉取新验证码；刷新失败（如触发限流）时保留当前图片 */
async function refreshCaptcha() {
  if (captchaLoading.value) return
  captchaLoading.value = true
  try {
    captcha.value = await getCaptcha()
    captchaStale.value = false
  } catch {
    // 失败提示由 request 拦截器统一处理
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
    ElMessage.success('登录成功')
    const redirect = (route.query.redirect as string) || '/'
    router.replace(redirect)
  } catch {
    // 登录失败时验证码已在服务端消费：清空输入并尝试换一张
    form.captchaCode = ''
    captchaStale.value = true
    if (captchaRequired.value) await refreshCaptcha()
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
    <el-card class="login-card" shadow="always">
      <div class="login-title">Yomi Admin</div>
      <div class="login-subtitle">后台管理系统</div>
      <el-form ref="formRef" :model="form" :rules="rules" size="large" @keyup.enter="handleLogin">
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="用户名"
            :prefix-icon="User"
            autocomplete="username"
            maxlength="64"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
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
              placeholder="验证码"
              :prefix-icon="Key"
              maxlength="8"
              autocomplete="off"
            />
            <button
              type="button"
              class="captcha-image"
              title="点击刷新验证码"
              :disabled="captchaLoading"
              @click="refreshCaptcha"
            >
              <img v-if="captcha" :src="captcha.image" alt="验证码" />
              <span v-else>{{ captchaLoading ? '加载中' : '点击加载' }}</span>
              <span v-if="captcha && captchaStale" class="captcha-stale-tip">
                已失效<br />点击刷新
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
          登 录
        </el-button>
        <div v-if="showMockTip" class="mock-tip">mock 模式，演示账号：admin / admin123</div>
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
    radial-gradient(1200px 600px at 70% -10%, #36cfc933, transparent),
    radial-gradient(1000px 500px at 0% 100%, #409eff26, transparent), var(--el-bg-color-page);
}

.login-card {
  width: 380px;
  padding: 8px 8px 4px;
}

.login-title {
  margin-top: 8px;
  color: var(--el-text-color-primary);
  font-size: 24px;
  font-weight: 700;
  text-align: center;
  letter-spacing: 1px;
}

.login-subtitle {
  margin: 8px 0 28px;
  color: var(--el-text-color-secondary);
  font-size: 14px;
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
  height: 40px;
  padding: 0;
  overflow: hidden;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color);
  border-radius: var(--el-border-radius-base);
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
}

.mock-tip {
  margin: 14px 0 4px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  text-align: center;
}
</style>
