#!/usr/bin/env bash
# 整理各 module 的 go.mod / go.sum
# 用法: bash script/dev/tidy.sh

set -euo pipefail
root="$(cd "$(dirname "$0")/../.." && pwd)"
# shellcheck source=workspace-modules.sh
source "$(dirname "$0")/workspace-modules.sh"

for mod in "${WORKSPACE_MODULES[@]}"; do
  echo "[tidy] go mod tidy in ${mod}"
  (cd "${root}/${mod}" && go mod tidy)
done
