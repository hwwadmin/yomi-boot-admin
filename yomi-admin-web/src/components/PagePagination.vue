<script setup lang="ts">
defineOptions({ name: 'PagePagination' })

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

withDefaults(defineProps<{ total: number; pageSizes?: number[] }>(), {
  pageSizes: () => [10, 20, 50],
})

const emit = defineEmits<{ change: [] }>()

/** 切换每页条数时回到第一页，避免页码越界 */
function handleSizeChange() {
  page.value = 1
  emit('change')
}

function handleCurrentChange() {
  emit('change')
}
</script>

<template>
  <el-pagination
    v-model:current-page="page"
    v-model:page-size="pageSize"
    class="page-pagination"
    layout="total, sizes, prev, pager, next, jumper"
    :total="total"
    :page-sizes="pageSizes"
    @current-change="handleCurrentChange"
    @size-change="handleSizeChange"
  />
</template>

<style scoped>
.page-pagination {
  --el-pagination-font-size: 13px;
  --el-pagination-item-gap: 12px;
  --el-pagination-button-width: 30px;
  --el-pagination-button-height: 30px;

  width: 100%;
  color: var(--app-text-3);
}

.page-pagination :deep(.el-pagination__total) {
  margin-right: auto;
  color: var(--app-text-3);
  font-size: 13px;
}

/* 翻页按钮：幽灵态，悬浮才出现浅底 */
.page-pagination :deep(.btn-prev),
.page-pagination :deep(.btn-next) {
  min-width: 30px;
  height: 30px;
  color: var(--app-text-2);
  background: transparent;
  border-radius: var(--app-radius-lg);
  transition:
    color 0.15s ease,
    background-color 0.15s ease;
}

.page-pagination :deep(.btn-prev .el-icon),
.page-pagination :deep(.btn-next .el-icon) {
  font-size: 13px;
}

.page-pagination :deep(.btn-prev:hover),
.page-pagination :deep(.btn-next:hover) {
  color: var(--app-text-1);
  background: var(--app-surface-muted);
}

.page-pagination :deep(.btn-prev:disabled),
.page-pagination :deep(.btn-next:disabled) {
  color: var(--app-text-3);
  background: transparent;
  opacity: 0.4;
}

/* 页码：默认纯文字，悬浮浅底，当前页为蓝色胶囊 */
.page-pagination :deep(.el-pager li) {
  min-width: 30px;
  height: 30px;
  margin: 0 1px;
  color: var(--app-text-2);
  background: transparent;
  border-radius: var(--app-radius-lg);
  font-size: 13px;
  font-variant-numeric: tabular-nums;
  transition:
    color 0.15s ease,
    background-color 0.15s ease;
}

.page-pagination :deep(.el-pager li:hover) {
  color: var(--app-text-1);
  background: var(--app-surface-muted);
}

.page-pagination :deep(.el-pager li.is-active) {
  color: #fff;
  background: var(--app-primary);
  font-weight: 600;
}

.page-pagination :deep(.el-pager li.is-active:hover) {
  color: #fff;
  background: var(--app-primary);
}

.page-pagination :deep(.el-pager li.more) {
  color: var(--app-text-3);
  background: transparent;
  cursor: default;
}

.page-pagination :deep(.el-pager li.more:hover) {
  color: var(--app-text-3);
  background: transparent;
}

/* 每页条数：浅底无边框的轻量选择器 */
.page-pagination :deep(.el-select) {
  width: 98px;
}

.page-pagination :deep(.el-pagination__sizes .el-select__wrapper) {
  height: 30px;
  min-height: 30px;
  padding: 0 8px;
  background: var(--app-surface-muted);
  border-radius: var(--app-radius-lg);
  box-shadow: none;
  font-size: 13px;
  transition: background-color 0.15s ease;
}

.page-pagination :deep(.el-pagination__sizes .el-select__wrapper:hover),
.page-pagination :deep(.el-pagination__sizes .el-select__wrapper.is-focused) {
  background: var(--el-fill-color);
}

.page-pagination :deep(.el-pagination__sizes .el-select__selected-item) {
  color: var(--app-text-2);
  font-size: 13px;
}

/* 跳页：紧凑输入框，与整行同高 */
.page-pagination :deep(.el-pagination__jump) {
  color: var(--app-text-3);
  font-size: 13px;
}

.page-pagination :deep(.el-pagination__goto) {
  margin-right: 6px;
}

.page-pagination :deep(.el-pagination__classifier) {
  margin-left: 6px;
}

.page-pagination :deep(.el-pagination__editor.el-input) {
  width: 44px;
}

.page-pagination :deep(.el-pagination__editor .el-input__wrapper) {
  height: 30px;
  min-height: 30px;
  padding: 0 4px;
  background: var(--app-surface-muted);
  border-radius: var(--app-radius-lg);
  box-shadow: none;
  transition: background-color 0.15s ease;
}

.page-pagination :deep(.el-pagination__editor .el-input__wrapper:hover) {
  background: var(--el-fill-color);
}

.page-pagination :deep(.el-pagination__editor .el-input__wrapper.is-focus) {
  background: var(--app-surface);
  box-shadow: 0 0 0 1px var(--app-primary) inset;
}

.page-pagination :deep(.el-pagination__editor .el-input__inner) {
  color: var(--app-text-1);
  font-size: 13px;
}
</style>
