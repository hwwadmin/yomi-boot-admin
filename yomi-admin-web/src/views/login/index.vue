<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import { Lock, User } from '@element-plus/icons-vue'

import { useUserStore } from '@/stores/user'

defineOptions({ name: 'LoginPage' })

const userStore = useUserStore()
const router = useRouter()
const route = useRoute()

const formRef = ref<FormInstance>()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

const showMockTip = import.meta.env.DEV && import.meta.env.VITE_USE_MOCK === 'true'

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    await userStore.login(form.username.trim(), form.password)
    ElMessage.success('登录成功')
    const redirect = (route.query.redirect as string) || '/'
    router.replace(redirect)
  } finally {
    loading.value = false
  }
}
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
          />
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
