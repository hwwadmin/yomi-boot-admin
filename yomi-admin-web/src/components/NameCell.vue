<script setup lang="ts">
import type { Component } from 'vue'

defineOptions({ name: 'NameCell' })

withDefaults(
  defineProps<{
    label: string
    icon?: Component
    tone?: 'blue' | 'violet' | 'green' | 'orange' | 'cyan' | 'slate'
  }>(),
  { tone: 'blue' },
)
</script>

<template>
  <span class="name-cell">
    <span class="name-cell__badge" :class="`is-${tone}`">
      <el-icon v-if="icon" :size="16"><component :is="icon" /></el-icon>
      <span v-else class="name-cell__initial">{{ label.charAt(0) }}</span>
    </span>
    <span class="name-cell__label" :title="label">{{ label }}</span>
  </span>
</template>

<style scoped>
.name-cell {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.name-cell__badge {
  display: inline-flex;
  flex: none;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: rgb(64 158 255 / 12%);
  color: #3d8bfd;
  font-size: 14px;
  font-weight: 600;
}

.name-cell__badge.is-violet {
  background: rgb(124 111 240 / 12%);
  color: #7c6ff0;
}

.name-cell__badge.is-green {
  background: rgb(52 168 83 / 12%);
  color: #34a853;
}

.name-cell__badge.is-orange {
  background: rgb(245 158 11 / 14%);
  color: #e08c0a;
}

.name-cell__badge.is-cyan {
  background: rgb(23 162 184 / 12%);
  color: #17a2b8;
}

.name-cell__badge.is-slate {
  background: rgb(100 116 139 / 12%);
  color: #7b8794;
}

.name-cell__label {
  overflow: hidden;
  color: var(--app-text-1);
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
