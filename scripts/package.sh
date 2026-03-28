#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${DIST_DIR:-${ROOT_DIR}/dist}"
GO_BIN="${GO_BIN:-go}"
GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"
BIN_NAME="${BIN_NAME:-agent-remote}"
VERSION="${VERSION:-$(date +%Y%m%d%H%M%S)}"

build_path="${DIST_DIR}/${GOOS}_${GOARCH}/${BIN_NAME}"
archive_name="${BIN_NAME}_${VERSION}_${GOOS}_${GOARCH}.tar.gz"
archive_path="${DIST_DIR}/${archive_name}"

if [[ ! -x "${build_path}" ]]; then
  GOOS="${GOOS}" GOARCH="${GOARCH}" GO_BIN="${GO_BIN}" DIST_DIR="${DIST_DIR}" "${ROOT_DIR}/scripts/build.sh" >/dev/null
fi

mkdir -p "${DIST_DIR}"

tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT

install -m 0755 "${build_path}" "${tmp_dir}/${BIN_NAME}"

tar -C "${tmp_dir}" -czf "${archive_path}" "${BIN_NAME}"
(
  cd "${DIST_DIR}"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "${archive_name}" > "${archive_name}.sha256"
  else
    shasum -a 256 "${archive_name}" > "${archive_name}.sha256"
  fi
)

printf '%s\n' "${archive_path}"
