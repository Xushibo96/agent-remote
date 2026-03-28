#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${OUT_DIR:-${ROOT_DIR}/dist}"
GO_BIN="${GO_BIN:-go}"
GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"
GOARM="${GOARM:-}"
BIN_NAME="${BIN_NAME:-agent-remote}"

build_dir="${OUT_DIR}/${GOOS}_${GOARCH}"
bin_path="${build_dir}/${BIN_NAME}"

mkdir -p "${build_dir}"

(
  cd "${ROOT_DIR}"
  if [[ -n "${GOARM}" ]]; then
    GOOS="${GOOS}" GOARCH="${GOARCH}" GOARM="${GOARM}" "${GO_BIN}" build -trimpath -o "${bin_path}" ./cmd/agent-remote
  else
    GOOS="${GOOS}" GOARCH="${GOARCH}" "${GO_BIN}" build -trimpath -o "${bin_path}" ./cmd/agent-remote
  fi
)

printf '%s\n' "${bin_path}"
