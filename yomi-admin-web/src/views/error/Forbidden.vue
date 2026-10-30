<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineOptions({ name: 'ForbiddenPage' })

const router = useRouter()
const { t } = useI18n()
</script>

<template>
  <div class="forbidden">
    <div class="code">403</div>
    <div class="text">{{ t('errorPage.forbidden.message') }}</div>
    <el-button type="primary" @click="router.replace('/dashboard')">
      {{ t('errorPage.forbidden.backHome') }}
    </el-button>
  </div>
</template>

<style scoped>
.forbidden {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  gap: 12px;
  background-color: var(--el-bg-color-page);
}

.code {
  color: var(--el-text-color-secondary);
  font-size: 96px;
  font-weight: 700;
  line-height: 1;
}

.text {
  margin-bottom: 8px;
  color: var(--el-text-color-regular);
  font-size: 15px;
}
</style>
