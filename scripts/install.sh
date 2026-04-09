#!/usr/bin/env bash

set -euo pipefail

APP="${BLACKDESK_BINARY_NAME:-blackdesk}"
REPO="${BLACKDESK_REPO:-Blackdesk-ai/blackdesk}"
INSTALL_DIR="${BLACKDESK_INSTALL_DIR:-$HOME/.local/bin}"
REQUESTED_VERSION="${BLACKDESK_VERSION:-${VERSION:-}}"
LOCAL_BINARY_PATH=""
NO_MODIFY_PATH=false

MUTED='\033[0;2m'
RED='\033[0;31m'
ORANGE='\033[38;5;214m'
NC='\033[0m'

usage() {
  cat <<EOF
Blackdesk Installer

Usage: install.sh [options]

Options:
  -h, --help                Display this help message
  -v, --version <version>   Install a specific version (e.g. 0.1.0)
  -b, --binary <path>       Install from a local binary instead of downloading
  -d, --install-dir <path>  Override install directory
      --no-modify-path      Don't modify shell config files

Environment:
  BLACKDESK_VERSION         Requested version
  VERSION                   Requested version alias
  BLACKDESK_INSTALL_DIR     Install directory
  BLACKDESK_REPO            GitHub repo in owner/name form
  BLACKDESK_BINARY_NAME     Binary name to install

Examples:
  curl -fsSL https://blackdesk.ai/install | bash
  curl -fsSL https://blackdesk.ai/install | bash -s -- --version 0.1.0
  curl -fsSL https://blackdesk.ai/install | bash -s -- --install-dir "\$HOME/bin"
  ./scripts/install.sh --binary ./bin/blackdesk --install-dir /tmp/blackdesk-bin

Installed binary commands:
  blackdesk --help
  blackdesk ?
  blackdesk --version
  blackdesk upgrade --check
EOF
}

print_message() {
  local level="$1"
  local message="$2"
  local color="${NC}"

  case "${level}" in
    warning)
      color="${ORANGE}"
      ;;
    error)
      color="${RED}"
      ;;
  esac

  printf '%b\n' "${color}${message}${NC}"
}

fail() {
  print_message error "blackdesk install: $*"
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    -v|--version)
      if [[ -n "${2:-}" ]]; then
        REQUESTED_VERSION="${2#v}"
        shift 2
      else
        fail "--version requires a version argument"
      fi
      ;;
    -b|--binary)
      if [[ -n "${2:-}" ]]; then
        LOCAL_BINARY_PATH="$2"
        shift 2
      else
        fail "--binary requires a path argument"
      fi
      ;;
    -d|--install-dir)
      if [[ -n "${2:-}" ]]; then
        INSTALL_DIR="$2"
        shift 2
      else
        fail "--install-dir requires a path argument"
      fi
      ;;
    --no-modify-path)
      NO_MODIFY_PATH=true
      shift
      ;;
    *)
      print_message warning "blackdesk install: ignoring unknown option '$1'"
      shift
      ;;
  esac
done

detect_os() {
  local raw_os
  raw_os="$(uname -s)"
  case "${raw_os}" in
    Darwin)
      printf 'darwin\n'
      ;;
    Linux)
      printf 'linux\n'
      ;;
    MINGW*|MSYS*|CYGWIN*)
      printf 'windows\n'
      ;;
    *)
      fail "unsupported operating system: ${raw_os}"
      ;;
  esac
}

detect_arch() {
  local raw_arch
  raw_arch="$(uname -m)"
  case "${raw_arch}" in
    x86_64|amd64)
      printf 'amd64\n'
      ;;
    arm64|aarch64)
      printf 'arm64\n'
      ;;
    *)
      fail "unsupported architecture: ${raw_arch}"
      ;;
  esac
}

adjust_arch_for_rosetta() {
  local os="$1"
  local arch="$2"

  if [[ "${os}" = "darwin" && "${arch}" = "amd64" ]]; then
    local rosetta_flag
    rosetta_flag="$(sysctl -n sysctl.proc_translated 2>/dev/null || echo 0)"
    if [[ "${rosetta_flag}" = "1" ]]; then
      printf 'arm64\n'
      return
    fi
  fi

  printf '%s\n' "${arch}"
}

latest_version() {
  local api_url
  api_url="https://api.github.com/repos/${REPO}/releases/latest"
  curl -fsSL "${api_url}" | awk -F'"' '/"tag_name"[[:space:]]*:/ { sub(/^v/, "", $4); print $4; exit }'
}

checksum_tool() {
  if command -v shasum >/dev/null 2>&1; then
    printf 'shasum -a 256'
    return
  fi
  if command -v sha256sum >/dev/null 2>&1; then
    printf 'sha256sum'
    return
  fi
  printf ''
}

verify_checksum() {
  local archive_path="$1"
  local sums_path="$2"
  local archive_name expected actual tool

  tool="$(checksum_tool)"
  [[ -n "${tool}" ]] || return 0

  archive_name="$(basename "${archive_path}")"
  expected="$(awk -v name="${archive_name}" '$2 == name { print $1 }' "${sums_path}")"
  [[ -n "${expected}" ]] || fail "checksum missing for ${archive_name}"

  actual="$(${tool} "${archive_path}" | awk '{print $1}')"
  [[ "${expected}" = "${actual}" ]] || fail "checksum mismatch for ${archive_name}"
}

installed_version() {
  local cmd="$1"
  "${cmd}" --version 2>/dev/null | awk 'NR == 1 { print $2 }'
}

check_existing_version() {
  local target_version="$1"

  if ! command -v "${APP}" >/dev/null 2>&1; then
    return 0
  fi

  local current_version
  current_version="$(installed_version "${APP}" || true)"
  [[ -n "${current_version}" ]] || return 0

  if [[ "${current_version}" = "${target_version}" ]]; then
    print_message info "${MUTED}${APP} ${target_version} is already installed${NC}"
    exit 0
  fi

  print_message info "${MUTED}Installed version:${NC} ${current_version}"
}

add_to_path() {
  local config_file="$1"
  local command="$2"

  if grep -Fxq "${command}" "${config_file}" 2>/dev/null; then
    print_message info "${MUTED}PATH entry already present in ${config_file}${NC}"
    return
  fi

  if [[ ! -e "${config_file}" ]]; then
    print_message warning "No shell config file found. Add this manually:"
    print_message info "  ${command}"
    return
  fi

  if [[ -w "${config_file}" ]]; then
    printf '\n# blackdesk\n%s\n' "${command}" >> "${config_file}"
    print_message info "${MUTED}Added ${INSTALL_DIR} to PATH in ${config_file}${NC}"
    return
  fi

  print_message warning "Can't write ${config_file}. Add this manually:"
  print_message info "  ${command}"
}

ensure_path() {
  local xdg_config_home current_shell config_file=""
  xdg_config_home="${XDG_CONFIG_HOME:-$HOME/.config}"
  current_shell="$(basename "${SHELL:-sh}")"

  case "${current_shell}" in
    fish)
      for file in "$HOME/.config/fish/config.fish"; do
        if [[ -f "${file}" ]]; then
          config_file="${file}"
          break
        fi
      done
      if [[ -n "${config_file}" ]]; then
        add_to_path "${config_file}" "fish_add_path ${INSTALL_DIR}"
      else
        print_message warning "No Fish config file found. Add this manually:"
        print_message info "  fish_add_path ${INSTALL_DIR}"
      fi
      ;;
    zsh)
      for file in "${ZDOTDIR:-$HOME}/.zshrc" "${ZDOTDIR:-$HOME}/.zshenv" "${xdg_config_home}/zsh/.zshrc" "${xdg_config_home}/zsh/.zshenv"; do
        if [[ -f "${file}" ]]; then
          config_file="${file}"
          break
        fi
      done
      config_file="${config_file:-${ZDOTDIR:-$HOME}/.zshrc}"
      add_to_path "${config_file}" "export PATH=${INSTALL_DIR}:\$PATH"
      ;;
    bash)
      for file in "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.profile" "${xdg_config_home}/bash/.bashrc" "${xdg_config_home}/bash/.bash_profile"; do
        if [[ -f "${file}" ]]; then
          config_file="${file}"
          break
        fi
      done
      config_file="${config_file:-$HOME/.bashrc}"
      add_to_path "${config_file}" "export PATH=${INSTALL_DIR}:\$PATH"
      ;;
    ash|sh)
      for file in "$HOME/.ashrc" "$HOME/.profile"; do
        if [[ -f "${file}" ]]; then
          config_file="${file}"
          break
        fi
      done
      config_file="${config_file:-$HOME/.profile}"
      add_to_path "${config_file}" "export PATH=${INSTALL_DIR}:\$PATH"
      ;;
    *)
      print_message warning "Unknown shell '${current_shell}'. Add this manually:"
      print_message info "  export PATH=${INSTALL_DIR}:\$PATH"
      ;;
  esac
}

download_release() {
  local version="$1"
  local os="$2"
  local arch="$3"
  local archive_ext archive_name base_url archive_url checksum_name checksum_url
  local tmpdir archive_path checksum_path

  if [[ "${os}" = "windows" ]]; then
    archive_ext="zip"
    need_cmd unzip
  else
    archive_ext="tar.gz"
    need_cmd tar
  fi

  archive_name="${APP}_${version}_${os}_${arch}.${archive_ext}"
  checksum_name="${APP}_${version}_SHA256SUMS.txt"
  base_url="https://github.com/${REPO}/releases/download/v${version}"
  archive_url="${base_url}/${archive_name}"
  checksum_url="${base_url}/${checksum_name}"

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "${tmpdir}"' RETURN

  archive_path="${tmpdir}/${archive_name}"
  checksum_path="${tmpdir}/${checksum_name}"

  print_message info "${MUTED}Installing ${APP} ${version} for ${os}/${arch}${NC}"
  curl -fL --progress-bar "${archive_url}" -o "${archive_path}" || fail "failed to download ${archive_url}"

  if curl -fsSL "${checksum_url}" -o "${checksum_path}"; then
    verify_checksum "${archive_path}" "${checksum_path}"
  fi

  mkdir -p "${tmpdir}/extract"
  if [[ "${archive_ext}" = "zip" ]]; then
    unzip -q "${archive_path}" -d "${tmpdir}/extract"
  else
    tar -xzf "${archive_path}" -C "${tmpdir}/extract"
  fi

  local extracted_binary install_path
  extracted_binary="${tmpdir}/extract/${APP}"
  install_path="${INSTALL_DIR}/${APP}"
  if [[ "${os}" = "windows" ]]; then
    extracted_binary="${tmpdir}/extract/${APP}.exe"
    install_path="${INSTALL_DIR}/${APP}.exe"
  fi

  [[ -f "${extracted_binary}" ]] || fail "archive did not contain $(basename "${extracted_binary}")"

  mkdir -p "${INSTALL_DIR}"
  cp "${extracted_binary}" "${install_path}"
  chmod 755 "${install_path}"

  print_message info "${MUTED}Installed to${NC} ${install_path}"
}

install_local_binary() {
  local os="$1"
  local install_path

  [[ -f "${LOCAL_BINARY_PATH}" ]] || fail "binary not found at ${LOCAL_BINARY_PATH}"

  install_path="${INSTALL_DIR}/${APP}"
  if [[ "${os}" = "windows" ]]; then
    install_path="${INSTALL_DIR}/${APP}.exe"
  fi

  mkdir -p "${INSTALL_DIR}"
  cp "${LOCAL_BINARY_PATH}" "${install_path}"
  chmod 755 "${install_path}"

  print_message info "${MUTED}Installed local binary to${NC} ${install_path}"
}

print_next_steps() {
  printf '\n'
  print_message info "${MUTED}Next:${NC}"
  print_message info "  ${APP} --help"
  print_message info "  ${APP} ?"
  print_message info "  ${APP} --version"
  print_message info "  ${APP} upgrade --check"
  print_message info "  ${APP}"
  printf '\n'
}

need_cmd curl
need_cmd mktemp
need_cmd chmod
need_cmd mkdir
need_cmd cp

OS="$(detect_os)"
ARCH="$(adjust_arch_for_rosetta "${OS}" "$(detect_arch)")"

if [[ -n "${LOCAL_BINARY_PATH}" ]]; then
  install_local_binary "${OS}"
else
  if [[ -z "${REQUESTED_VERSION}" || "${REQUESTED_VERSION}" = "latest" ]]; then
    SPECIFIC_VERSION="$(latest_version)"
    [[ -n "${SPECIFIC_VERSION}" ]] || fail "unable to determine latest release version"
  else
    SPECIFIC_VERSION="${REQUESTED_VERSION#v}"
  fi

  check_existing_version "${SPECIFIC_VERSION}"
  download_release "${SPECIFIC_VERSION}" "${OS}" "${ARCH}"
fi

if [[ "${NO_MODIFY_PATH}" != "true" ]]; then
  case ":${PATH}:" in
    *:"${INSTALL_DIR}":*)
      ;;
    *)
      ensure_path
      ;;
  esac
fi

if [[ -n "${GITHUB_ACTIONS:-}" && "${GITHUB_ACTIONS}" = "true" && -n "${GITHUB_PATH:-}" ]]; then
  printf '%s\n' "${INSTALL_DIR}" >> "${GITHUB_PATH}"
  print_message info "${MUTED}Added ${INSTALL_DIR} to GITHUB_PATH${NC}"
fi

print_next_steps
