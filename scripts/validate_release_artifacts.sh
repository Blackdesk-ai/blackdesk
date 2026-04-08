#!/usr/bin/env bash

set -euo pipefail

VERSION="${1:?usage: validate_release_artifacts.sh <version>}"
DIST_DIR="${DIST_DIR:-dist}"
BINARY_NAME="${BINARY_NAME:-blackdesk}"

fail() {
  printf 'release validation: %s\n' "$*" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

need_cmd tar
need_cmd unzip
need_cmd dpkg-deb
need_cmd file

[ -d "${DIST_DIR}" ] || fail "dist directory not found: ${DIST_DIR}"

expect_file() {
  [ -f "$1" ] || fail "missing artifact: $1"
}

extract_dir="$(mktemp -d)"
trap 'rm -rf "${extract_dir}"' EXIT

mac_amd64="${DIST_DIR}/${BINARY_NAME}_${VERSION}_darwin_amd64.tar.gz"
mac_arm64="${DIST_DIR}/${BINARY_NAME}_${VERSION}_darwin_arm64.tar.gz"
linux_amd64="${DIST_DIR}/${BINARY_NAME}_${VERSION}_linux_amd64.tar.gz"
linux_arm64="${DIST_DIR}/${BINARY_NAME}_${VERSION}_linux_arm64.tar.gz"
windows_amd64="${DIST_DIR}/${BINARY_NAME}_${VERSION}_windows_amd64.zip"
deb_amd64="${DIST_DIR}/${BINARY_NAME}_${VERSION}_linux_amd64.deb"
deb_arm64="${DIST_DIR}/${BINARY_NAME}_${VERSION}_linux_arm64.deb"
rpm_amd64="${DIST_DIR}/${BINARY_NAME}_${VERSION}_linux_amd64.rpm"
rpm_arm64="${DIST_DIR}/${BINARY_NAME}_${VERSION}_linux_arm64.rpm"
checksums="${DIST_DIR}/${BINARY_NAME}_${VERSION}_SHA256SUMS.txt"
checksum_bundle="${checksums}.sigstore.json"

expect_file "${mac_amd64}"
expect_file "${mac_arm64}"
expect_file "${linux_amd64}"
expect_file "${linux_arm64}"
expect_file "${windows_amd64}"
expect_file "${deb_amd64}"
expect_file "${deb_arm64}"
expect_file "${rpm_amd64}"
expect_file "${rpm_arm64}"
expect_file "${checksums}"
expect_file "${checksum_bundle}"

tar -tzf "${mac_amd64}" | grep -qx "${BINARY_NAME}" || fail "macOS amd64 archive missing binary"
tar -tzf "${mac_arm64}" | grep -qx "${BINARY_NAME}" || fail "macOS arm64 archive missing binary"
tar -tzf "${linux_amd64}" | grep -qx "${BINARY_NAME}" || fail "linux amd64 archive missing binary"
tar -tzf "${linux_arm64}" | grep -qx "${BINARY_NAME}" || fail "linux arm64 archive missing binary"
unzip -Z1 "${windows_amd64}" | grep -qx "${BINARY_NAME}.exe" || fail "windows archive missing executable"

dpkg-deb --info "${deb_amd64}" >/dev/null
dpkg-deb --info "${deb_arm64}" >/dev/null
file "${rpm_amd64}" | grep -q "RPM" || fail "invalid rpm package: ${rpm_amd64}"
file "${rpm_arm64}" | grep -q "RPM" || fail "invalid rpm package: ${rpm_arm64}"

tar -xzf "${linux_amd64}" -C "${extract_dir}"
"${extract_dir}/${BINARY_NAME}" --version >/dev/null
