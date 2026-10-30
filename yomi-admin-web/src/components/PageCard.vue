<script setup lang="ts">
defineOptions({ name: 'PageCard' })
</script>

<template>
  <el-card class="page-card" shadow="never">
    <div v-if="$slots.toolbar" class="page-card__toolbar">
      <slot name="toolbar" />
    </div>
    <slot />
    <div v-if="$slots.footer" class="page-card__footer">
      <slot name="footer" />
    </div>
  </el-card>
</template>

<style scoped>
.page-card {
  border-color: var(--app-border);
  border-radius: var(--app-radius-card);
  background-color: var(--app-surface);
  box-shadow: var(--app-shadow-card);
}

.page-card :deep(.el-card__body) {
  padding: 24px;
}

.page-card__toolbar {
  margin-bottom: 20px;
}

.page-card__footer {
  display: flex;
  align-items: center;
  margin-top: 16px;
}
</style>
