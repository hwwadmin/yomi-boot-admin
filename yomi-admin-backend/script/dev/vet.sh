#!/usr/bin/env bash
# go vet 各 module
# 用法: bash script/dev/vet.sh

set -euo pipefail
root="$(cd "$(dirname "$0")/../.." && pwd)"
# shellcheck source=workspace-modules.sh
source "$(dirname "$0")/workspace-modules.sh"

for mod in "${WORKSPACE_MODULES[@]}"; do
  echo "[vet] go vet ./... in ${mod}"
  (cd "${root}/${mod}" && go vet ./...)
done
