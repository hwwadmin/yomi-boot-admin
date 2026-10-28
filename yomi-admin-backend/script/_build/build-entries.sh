#!/usr/bin/env bash
# 可编译入口列表，供 build.sh 引用
# 格式: name:pkg（pkg 相对仓库根目录）

BUILD_ENTRIES=(
  "yomi-admin-backend:./apps/server/cmd"
)
