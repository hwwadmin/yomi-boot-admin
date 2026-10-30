import type { Component } from 'vue'
import {
  Avatar,
  Grid,
  Key,
  Lock,
  Notebook,
  Odometer,
  Setting,
  Tickets,
  Tools,
  User,
} from '@element-plus/icons-vue'

/** 路由 meta.icon 与权限组 code 共用的图标映射 */
export const menuIcons: Record<string, Component> = {
  Odometer,
  User,
  Avatar,
  Key,
  Grid,
  Tickets,
  Lock,
  Setting,
  Tools,
  Notebook,
}

export function resolveMenuIcon(name?: string): Component | undefined {
  return name ? menuIcons[name] : undefined
}
