#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${DIST_DIR:-${ROOT_DIR}/dist}"

targets=(
  "linux amd64"
  "linux arm64"
  "darwin arm64"
)

for target in "${targets[@]}"; do
  read -r goos goarch <<<"${target}"
  GOOS="${goos}" GOARCH="${goarch}" DIST_DIR="${DIST_DIR}" "${ROOT_DIR}/scripts/package.sh"
done
