#!/usr/bin/env bash
# 单元测试：各 module；排除 apps/server/test 集成目录
# 用法: bash script/dev/test.sh
# 集成测试请先启动服务后执行: go test ./apps/server/test -v（在 module-mono 根或 apps/server 下）

set -euo pipefail
root="$(cd "$(dirname "$0")/../.." && pwd)"
# shellcheck source=workspace-modules.sh
source "$(dirname "$0")/workspace-modules.sh"

for mod in "${WORKSPACE_MODULES[@]}"; do
  mapfile -t packages < <(
    cd "${root}/${mod}" && go list -f '{{if or (len .TestGoFiles) (len .XTestGoFiles)}}{{.ImportPath}}{{end}}' ./... \
      | grep -Ev '/test($|/)' || true
  )
  if ((${#packages[@]} == 0)); then
    echo "[test] skip ${mod}: no unit tests"
    continue
  fi
  echo "[test] go test in ${mod} (${#packages[@]} packages)"
  (cd "${root}/${mod}" && go test "${packages[@]}")
done
