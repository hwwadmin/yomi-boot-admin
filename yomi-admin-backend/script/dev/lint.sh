#!/usr/bin/env bash
# golangci-lint 各 module（读取仓库根 .golangci.yml）
# 用法: bash script/dev/lint.sh

set -euo pipefail
root="$(cd "$(dirname "$0")/../.." && pwd)"
golangci_version="${GOLANGCI_VERSION:-v2.12.2}"
# shellcheck source=workspace-modules.sh
source "$(dirname "$0")/workspace-modules.sh"

if [[ ! -f "${root}/.golangci.yml" ]]; then
  echo "[lint] missing .golangci.yml" >&2
  exit 1
fi

for mod in "${WORKSPACE_MODULES[@]}"; do
  echo "[lint] golangci-lint run in ${mod}"
  (
    cd "${root}/${mod}"
    if command -v golangci-lint >/dev/null 2>&1; then
      golangci-lint run --config "${root}/.golangci.yml" ./...
    else
      go run "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${golangci_version}" run --config "${root}/.golangci.yml" ./...
    fi
  )
done
