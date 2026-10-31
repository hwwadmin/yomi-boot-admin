<script setup lang="ts">
defineOptions({ name: 'PageHeader' })

defineProps<{
  /** 页面主标题，一般传菜单标题的翻译结果，与侧边栏菜单同名 */
  title: string
  /** 标题下的一行说明，可选 */
  subtitle?: string
}>()
</script>

<template>
  <div class="page-header">
    <div class="page-header__text">
      <h1 class="page-header__title">{{ title }}</h1>
      <p v-if="subtitle" class="page-header__subtitle">{{ subtitle }}</p>
    </div>
    <div v-if="$slots.actions" class="page-header__actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<style scoped>
.page-header {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 20px;
}

.page-header__text {
  flex: 1;
  min-width: 0;
}

.page-header__title {
  margin: 0;
  color: var(--app-text-1);
  font-size: 22px;
  font-weight: 700;
  line-height: 1.3;
  letter-spacing: 0.01em;
}

.page-header__subtitle {
  margin: 6px 0 0;
  color: var(--app-text-3);
  font-size: 13px;
  line-height: 1.4;
}

.page-header__actions {
  display: flex;
  flex-shrink: 0;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
</style>
