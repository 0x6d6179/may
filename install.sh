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
  if [[ ":${PATH}:" == *":${dir}:"* ]]; then
    return
  fi

  export PATH="${dir}:${PATH}"

  local profile=""
  local shell_name
  shell_name=$(basename "${SHELL:-bash}")

  case "${shell_name}" in
    zsh)  profile="${HOME}/.zshrc" ;;
    bash)
      if [[ -f "${HOME}/.bashrc" ]]; then
        profile="${HOME}/.bashrc"
      elif [[ -f "${HOME}/.bash_profile" ]]; then
        profile="${HOME}/.bash_profile"
      else
        profile="${HOME}/.bashrc"
      fi
      ;;
    fish) profile="${HOME}/.config/fish/config.fish" ;;
  esac

  if [[ -n "${profile}" ]]; then
    local line
    case "${shell_name}" in
      fish) line="fish_add_path ${dir}" ;;
      *)    line="export PATH=\"${dir}:\$PATH\"" ;;
    esac

    if ! grep -qF "${dir}" "${profile}" 2>/dev/null; then
      printf '\n%s\n' "${line}" >> "${profile}"
      ok "added ${dir} to PATH in ${profile}"
    fi
  else
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

# ── detect existing install ────────────────────────────────────────────────────

check_existing() {
  local existing
  existing=$(command -v may 2>/dev/null) || return 0

  local source="unknown"
  case "${existing}" in
    */go/bin/*)              source="go install" ;;
    */homebrew/*|*/Cellar/*) source="homebrew" ;;
    */.local/bin/*)          source="install script" ;;
    /usr/local/bin/*)        source="install script" ;;
  esac

  if [[ -L "${existing}" ]]; then
    local target
    target=$(readlink -f "${existing}" 2>/dev/null || readlink "${existing}")
    if [[ -f "$(dirname "${target}")/go.mod" ]] || \
       [[ -f "$(dirname "${target}")/../go.mod" ]] || \
       [[ "${target}" == *"/Workspaces/"* ]]; then
      source="dev build"
    fi
  elif [[ -f "$(dirname "${existing}")/go.mod" ]] || \
       [[ -f "$(dirname "${existing}")/../go.mod" ]]; then
    source="dev build"
  fi

  info "existing installation detected: ${existing} (${source})"

  if [[ "${source}" == "dev build" ]]; then
    err "a dev build is installed at ${existing}.
       remove it or adjust your PATH before installing the release version."
  fi

  printf '  overwrite existing installation? [y/N] '
  read -r answer </dev/tty
  case "${answer}" in
    [yY]*) info "overwriting..." ;;
    *)     info "cancelled."; exit 0 ;;
  esac
}

# ── entry ─────────────────────────────────────────────────────────────────────

main() {
  printf '\n  \033[1mmay\033[0m — personal productivity toolkit\n\n'

  check_existing

  if command -v go >/dev/null 2>&1; then
    install_go
  elif install_binary 2>/dev/null; then
    :
  else
    err "go is not installed and no pre-built binary available.
       install go from https://go.dev/dl/ then re-run this script,
       or install via homebrew:  brew tap 0x6d6179/may && brew install may"
  fi

  printf '\n'
  ok "may installed"

  if [[ -t 0 ]] || [[ -e /dev/tty ]]; then
    info "running first-time setup..."
    may init </dev/tty
  else
    info "run 'may init' to complete setup"
  fi
}

main
