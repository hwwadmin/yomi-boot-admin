#!/usr/bin/env bash
# 编译 module-mono 编排层到 bin/
# 用法:
#   bash script/_build/build.sh
#   bash script/_build/build.sh linux_amd64
#   bash script/_build/build.sh windows_amd64 linux_amd64

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
BIN_DIR="$ROOT_DIR/bin"
# shellcheck source=build-entries.sh
source "$SCRIPT_DIR/build-entries.sh"

VALID_GOOS=("windows" "linux")
VALID_GOARCH=("amd64" "arm64")
TARGETS=()

if [ $# -eq 0 ]; then
  GOOS="${GOOS:-$(go env GOOS)}"
  GOARCH="${GOARCH:-$(go env GOARCH)}"
  TARGETS=("${GOOS}_${GOARCH}")
else
  for arg in "$@"; do
    IFS=',' read -r -a split_targets <<< "$arg"
    for split_target in "${split_targets[@]}"; do
      split_target="$(echo "$split_target" | tr '[:upper:]' '[:lower]' | xargs)"
      [ -n "$split_target" ] && TARGETS+=("$split_target")
    done
  done
fi

validate_target() {
  local t="$1"
  IFS='_' read -r goos goarch extra <<< "$t"
  if [ -n "${extra:-}" ] || [ -z "${goos:-}" ] || [ -z "${goarch:-}" ]; then
    echo "[ERROR] Invalid target format: '$t'. Use GOOS_GOARCH, e.g. windows_amd64" >&2
    exit 1
  fi
  local ok=false
  for v in "${VALID_GOOS[@]}"; do [ "$goos" = "$v" ] && ok=true && break; done
  [ "$ok" = true ] || { echo "[ERROR] Unsupported GOOS: $goos" >&2; exit 1; }
  ok=false
  for v in "${VALID_GOARCH[@]}"; do [ "$goarch" = "$v" ] && ok=true && break; done
  [ "$ok" = true ] || { echo "[ERROR] Unsupported GOARCH: $goarch" >&2; exit 1; }
}

for t in "${TARGETS[@]}"; do
  validate_target "$t"
done

mkdir -p "$BIN_DIR"
export CGO_ENABLED=0
cd "$ROOT_DIR"

for t in "${TARGETS[@]}"; do
  IFS='_' read -r goos goarch <<< "$t"
  export GOOS=$goos
  export GOARCH=$goarch

  for entry in "${BUILD_ENTRIES[@]}"; do
    name="${entry%%:*}"
    pkg="${entry##*:}"
    out_name="${name}-${goos}-${goarch}"
    [ "$goos" = "windows" ] && out_name="${out_name}.exe"

    echo "[build] ${name} [${goos}/${goarch}] -> ${out_name}"
    go build -trimpath -o "${BIN_DIR}/${out_name}" "$pkg"
  done
done

echo ""
echo "Build complete. Output directory: ${BIN_DIR}"
