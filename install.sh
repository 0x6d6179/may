#!/usr/bin/env bash
set -euo pipefail

REPO="0x6d6179/may"
BINARY="may"
RELEASES="https://github.com/${REPO}/releases/latest/download"

# ── helpers ──────────────────────────────────────────────────────────────────

info()  { printf '\033[0;34m  %s\033[0m\n' "$*"; }
ok()    { printf '\033[0;32m  ✓ %s\033[0m\n' "$*"; }
err()   { printf '\033[0;31m  ✗ %s\033[0m\n' "$*" >&2; exit 1; }

need() {
  command -v "$1" >/dev/null 2>&1 || err "$1 is required but not installed"
}

install_dir() {
  if [[ -w /usr/local/bin ]]; then
    echo /usr/local/bin
  else
    echo "${HOME}/.local/bin"
  fi
}

add_to_path() {
  local dir="$1"
  if [[ ":${PATH}:" != *":${dir}:"* ]]; then
    info "${dir} is not in PATH — add it to your shell profile:"
    printf '\n  export PATH="%s:$PATH"\n\n' "${dir}"
  fi
}

# ── install via go install ────────────────────────────────────────────────────

install_go() {
  info "installing via go install..."
  local version
  version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    2>/dev/null | grep '"tag_name"' | sed 's/.*"v\([^"]*\)".*/\1/' || echo "latest")

  local ldflags="-X github.com/0x6d6179/may/internal/version.Version=${version}"
  GOFLAGS="" go install -ldflags "${ldflags}" "github.com/0x6d6179/may/cmd/may@latest"
  ok "installed may@${version} to $(go env GOPATH)/bin"
  add_to_path "$(go env GOPATH)/bin"
}

# ── install pre-built binary ──────────────────────────────────────────────────

install_binary() {
  local os arch asset dir
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  arch=$(uname -m)

  case "${arch}" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *) err "unsupported architecture: ${arch}" ;;
  esac

  case "${os}" in
    darwin|linux) ;;
    *) err "unsupported os: ${os}" ;;
  esac

  asset="${BINARY}_${os}_${arch}"
  dir=$(install_dir)

  info "downloading ${asset}..."
  need curl
  mkdir -p "${dir}"

  if ! curl -fsSL "${RELEASES}/${asset}" -o "${dir}/${BINARY}"; then
    return 1
  fi

  chmod +x "${dir}/${BINARY}"
  ok "installed to ${dir}/${BINARY}"
  add_to_path "${dir}"
}

# ── entry ─────────────────────────────────────────────────────────────────────

main() {
  printf '\n  \033[1mmay\033[0m — personal productivity toolkit\n\n'

  if install_binary 2>/dev/null; then
    :
  elif command -v go >/dev/null 2>&1; then
    install_go
  else
    err "no pre-built binary available and go is not installed.
       install go from https://go.dev/dl/ then re-run this script,
       or install via homebrew:  brew tap 0x6d6179/may && brew install may"
  fi

  printf '\n'
  ok "may installed"
  info "running first-time setup...\n"
  may init
}

main
