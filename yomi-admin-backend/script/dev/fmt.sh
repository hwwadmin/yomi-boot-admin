#!/usr/bin/env bash
# 格式化各 module
# 用法: bash script/dev/fmt.sh

set -euo pipefail
root="$(cd "$(dirname "$0")/../.." && pwd)"
golangci_version="${GOLANGCI_VERSION:-v2.12.2}"
# shellcheck source=workspace-modules.sh
source "$(dirname "$0")/workspace-modules.sh"

for mod in "${WORKSPACE_MODULES[@]}"; do
  echo "[fmt] go fmt + golangci-lint fmt in ${mod}"
  (
    cd "${root}/${mod}"
    go fmt ./...
    if command -v golangci-lint >/dev/null 2>&1; then
      golangci-lint fmt --config "${root}/.golangci.yml" ./...
    else
      go run "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${golangci_version}" fmt --config "${root}/.golangci.yml" ./...
    fi
  )
done
