<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineOptions({ name: 'NotFoundPage' })

const router = useRouter()
const { t } = useI18n()
</script>

<template>
  <div class="not-found">
    <div class="code">404</div>
    <div class="text">{{ t('errorPage.notFound.message') }}</div>
    <el-button type="primary" @click="router.replace('/')">
      {{ t('errorPage.notFound.backHome') }}
    </el-button>
  </div>
</template>

<style scoped>
.not-found {
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
