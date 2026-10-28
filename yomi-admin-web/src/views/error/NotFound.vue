<script setup lang="ts">
defineOptions({ name: 'NotFoundPage' })

const router = useRouter()
</script>

<template>
  <div class="not-found">
    <div class="code">404</div>
    <div class="text">抱歉，您访问的页面不存在</div>
    <el-button type="primary" @click="router.replace('/')">返回首页</el-button>
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
