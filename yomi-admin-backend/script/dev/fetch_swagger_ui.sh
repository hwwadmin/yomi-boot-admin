#!/usr/bin/env bash
# 重新下载内嵌的 Swagger UI 静态资源（swagger-ui-dist）。
# 用法: bash script/dev/fetch_swagger_ui.sh [版本号]   # 默认 5.34.0
set -euo pipefail

VERSION="${1:-5.34.0}"
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
TARGET="$ROOT/apps/server/internal/support/apidoc/swaggerui"

mkdir -p "$TARGET"
curl -fsSL --retry 3 -o "$TARGET/swagger-ui-bundle.js" \
  "https://unpkg.com/swagger-ui-dist@${VERSION}/swagger-ui-bundle.js"
curl -fsSL --retry 3 -o "$TARGET/swagger-ui.css" \
  "https://unpkg.com/swagger-ui-dist@${VERSION}/swagger-ui.css"

echo "swagger-ui-dist@${VERSION} 已下载到 ${TARGET}"
echo "如更换版本，请同步更新本脚本默认版本号与 api/openapi.yaml 无关，无需改代码。"
