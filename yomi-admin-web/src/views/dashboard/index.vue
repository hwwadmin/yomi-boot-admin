<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import welcomeImage from '@/assets/welcome.svg'

import { useUserStore } from '@/stores/user'

defineOptions({ name: 'DashboardPage' })

const userStore = useUserStore()
const { t } = useI18n()

const greetingKey = computed(() => {
  const hour = new Date().getHours()
  if (hour < 6) return 'dashboard.greeting.dawn'
  if (hour < 12) return 'dashboard.greeting.morning'
  if (hour < 14) return 'dashboard.greeting.noon'
  if (hour < 18) return 'dashboard.greeting.afternoon'
  return 'dashboard.greeting.evening'
})

const greeting = computed(() => t(greetingKey.value))
const displayName = computed(() => userStore.userInfo?.nickname || t('dashboard.adminFallback'))
</script>

<template>
  <div class="dashboard">
    <section class="welcome-hero">
      <div class="welcome-copy">
        <span class="welcome-eyebrow" aria-hidden="true"></span>
        <h1 class="welcome-title">
          <span class="welcome-greeting">
            {{ t('dashboard.greetingLine', { greeting, name: displayName }) }}
          </span>
          <span class="welcome-subtitle">{{ t('dashboard.subtitle') }}</span>
        </h1>
      </div>
      <img class="welcome-image" :src="welcomeImage" :alt="t('dashboard.imageAlt')" />
    </section>
  </div>
</template>

<style scoped>
.welcome-hero {
  position: relative;
  display: flex;
  align-items: center;
  gap: 32px;
  overflow: hidden;
  padding: 34px 40px;
  border-radius: 12px;
  background:
    radial-gradient(80% 160% at 92% -30%, rgb(64 158 255 / 16%) 0%, rgb(64 158 255 / 0%) 60%),
    radial-gradient(70% 140% at 4% 130%, rgb(23 162 184 / 12%) 0%, rgb(23 162 184 / 0%) 55%),
    linear-gradient(118deg, #f7fbff 0%, #ffffff 46%, #f3faff 100%);
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 90%),
    inset 0 0 0 1px rgb(64 158 255 / 8%),
    var(--app-shadow-card);
}

/* 右上方点阵纹理，朝插画方向淡出 */
.welcome-hero::before {
  content: '';
  position: absolute;
  inset: 0;
  background-image: radial-gradient(rgba(64, 158, 255, 0.35) 1.2px, transparent 1.2px);
  background-size: 18px 18px;
  -webkit-mask-image: radial-gradient(420px 240px at 78% 18%, #000 0%, transparent 72%);
  mask-image: radial-gradient(420px 240px at 78% 18%, #000 0%, transparent 72%);
  pointer-events: none;
}

/* 左下装饰弧线，平衡右侧插画的视觉重量 */
.welcome-hero::after {
  content: '';
  position: absolute;
  left: -70px;
  bottom: -120px;
  width: 240px;
  height: 240px;
  border: 1px solid rgba(64, 158, 255, 0.14);
  border-radius: 50%;
  pointer-events: none;
}

.welcome-copy {
  position: relative;
  z-index: 1;
  flex: 1;
  min-width: 0;
}

.welcome-eyebrow {
  display: block;
  width: 40px;
  height: 4px;
  margin-bottom: 16px;
  border-radius: 999px;
  background: linear-gradient(90deg, #409eff, #17a2b8);
  box-shadow: 0 2px 8px rgb(64 158 255 / 35%);
}

.welcome-title {
  margin: 0;
  font-size: 30px;
  font-weight: 700;
  line-height: 1.3;
  letter-spacing: 0.01em;
}

.welcome-greeting {
  display: block;
  background: linear-gradient(96deg, #1554d1 0%, #409eff 48%, #17a2b8 100%);
  background-clip: text;
  -webkit-background-clip: text;
  color: transparent;
  -webkit-text-fill-color: transparent;
}

.welcome-subtitle {
  display: block;
  margin-top: 10px;
  color: var(--el-text-color-secondary);
  font-size: 15px;
  font-weight: 400;
  letter-spacing: 0.01em;
}

.welcome-image {
  position: relative;
  z-index: 1;
  flex-shrink: 0;
  width: 380px;
  height: auto;
  filter: drop-shadow(0 18px 28px rgba(64, 158, 255, 0.12));
  animation: welcome-float 7s ease-in-out infinite;
  will-change: transform;
}

@keyframes welcome-float {
  0%,
  100% {
    transform: translateY(0);
  }

  50% {
    transform: translateY(-8px);
  }
}

/* 暗色主题 */
html.dark .welcome-hero {
  background:
    radial-gradient(80% 160% at 92% -30%, rgb(64 158 255 / 24%) 0%, rgb(64 158 255 / 0%) 60%),
    radial-gradient(70% 140% at 4% 130%, rgb(23 162 184 / 14%) 0%, rgb(23 162 184 / 0%) 55%),
    linear-gradient(118deg, #101a2c 0%, #0d1524 46%, #0e1c2a 100%);
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 6%),
    inset 0 0 0 1px rgb(255 255 255 / 6%),
    var(--app-shadow-card);
}

html.dark .welcome-hero::before {
  background-image: radial-gradient(rgba(102, 177, 255, 0.3) 1.2px, transparent 1.2px);
}

html.dark .welcome-hero::after {
  border-color: rgba(124, 192, 255, 0.12);
}

html.dark .welcome-greeting {
  background: linear-gradient(96deg, #7cc0ff 0%, #b6e3ff 48%, #6fd6e6 100%);
  background-clip: text;
  -webkit-background-clip: text;
}

html.dark .welcome-image {
  filter: drop-shadow(0 18px 36px rgba(64, 158, 255, 0.25));
}

@media (max-width: 1280px) {
  .welcome-hero {
    gap: 24px;
    padding: 28px 32px;
  }

  .welcome-title {
    font-size: 26px;
  }

  .welcome-image {
    width: 320px;
  }
}

@media (max-width: 1024px) {
  .welcome-image {
    width: 260px;
  }
}

@media (max-width: 768px) {
  .welcome-hero {
    padding: 24px;
  }

  .welcome-title {
    font-size: 22px;
  }

  .welcome-image {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .welcome-image {
    animation: none;
  }
}
</style>
