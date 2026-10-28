#!/usr/bin/env bash
# yomi-admin-backend workspace 模块列表（相对仓库根目录）
# 用法: source script/dev/workspace-modules.sh

WORKSPACE_MODULES=(
  "apps/server"
  "domains/rbac"
  "domains/user"
)

# 集成测试 / 镜像构建针对编排层
SERVICE_MODULES=(
  "apps/server"
)
