#!/usr/bin/env bash
# portkill installer for Linux, macOS and FreeBSD.
#
#   curl -fsSL https://raw.githubusercontent.com/khanalsaroj/portkill/main/main/install.sh | bash
#
# Environment overrides:
#   PORTKILL_VERSION       install a specific version (e.g. v1.2.3), default: latest
#   PORTKILL_INSTALL_DIR   install location, default: /usr/local/bin (falls back to ~/.local/bin)
set -euo pipefail

REPO="khanalsaroj/portkill"
BIN_NAME="portkill"
VERSION="${PORTKILL_VERSION:-latest}"
INSTALL_DIR="${PORTKILL_INSTALL_DIR:-}"

# ---------- Output helpers ----------
info() { printf '  %s\n' "$*"; }
ok()   { printf '  \033[32m✓\033[0m %s\n' "$*"; }
warn() { printf '  \033[33m!\033[0m %s\n' "$*" >&2; }
die()  { printf '  \033[31m✗\033[0m %s\n' "$*" >&2; exit 1; }

need() { command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"; }

banner() {
  printf '\n'
  printf '    ░█▀█░█▀█░█▀▄░▀█▀░█░█░▀█▀░█░░░█░░\n'
  printf '    ░█▀▀░█░█░█▀▄░░█░░█▀▄░░█░░█░░░█░░\n'
  printf '    ░▀░░░▀▀▀░▀░▀░░▀░░▀░▀░▀▀▀░▀▀▀░▀▀▀\n'
  printf '\n    kill the process holding a port — instantly\n\n'
}

# ---------- Platform detection ----------
detect_platform() {
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"

  case "$OS" in
    linux | darwin | freebsd) ;;
    *) die "unsupported operating system: $OS" ;;
  esac

  case "$ARCH" in
    x86_64 | amd64) ARCH="amd64" ;;
    arm64 | aarch64) ARCH="arm64" ;;
    i386 | i686) ARCH="386" ;;
    *) die "unsupported architecture: $ARCH" ;;
  esac
}

# ---------- Version resolution ----------
resolve_version() {
  if [ "$VERSION" != "latest" ]; then
    printf '%s' "${VERSION#v}"
    return
  fi
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" |
    grep -o '"tag_name"[ ]*:[ ]*"[^"]*"' | head -1 | cut -d'"' -f4 | sed 's/^v//'
}

# ---------- Checksum verification (best effort) ----------
verify_checksum() {
  local dir="$1" asset="$2" version="$3" tool expected actual
  if command -v sha256sum >/dev/null 2>&1; then
    tool="sha256sum"
  elif command -v shasum >/dev/null 2>&1; then
    tool="shasum -a 256"
  else
    warn "no sha256 tool found — skipping checksum verification"
    return
  fi

  if ! curl -fsSL "https://github.com/${REPO}/releases/download/v${version}/checksums.txt" -o "$dir/checksums.txt"; then
    warn "checksums.txt unavailable — skipping checksum verification"
    return
  fi

  expected="$(awk -v f="$asset" '$2 == f {print $1}' "$dir/checksums.txt" | head -1)"
  if [ -z "$expected" ]; then
    warn "no checksum entry for ${asset} — skipping verification"
    return
  fi
  actual="$($tool "$dir/$asset" | awk '{print $1}')"
  [ "$expected" = "$actual" ] || die "checksum mismatch for ${asset} (expected ${expected}, got ${actual})"
  ok "checksum verified"
}

# ---------- Install location ----------
choose_install_dir() {
  if [ -n "$INSTALL_DIR" ]; then
    printf '%s' "$INSTALL_DIR"
    return
  fi
  for d in /usr/local/bin /opt/homebrew/bin; do
    if [ -d "$d" ]; then
      printf '%s' "$d"
      return
    fi
  done
  printf '%s' "$HOME/.local/bin"
}

install_binary() {
  local src="$1" dir="$2"
  mkdir -p "$dir" 2>/dev/null || true
  if [ -w "$dir" ]; then
    install -m 0755 "$src" "$dir/$BIN_NAME"
  elif command -v sudo >/dev/null 2>&1; then
    warn "elevated permissions required to write to ${dir}"
    sudo install -m 0755 "$src" "$dir/$BIN_NAME"
  else
    die "cannot write to ${dir} and sudo is unavailable — set PORTKILL_INSTALL_DIR to a writable directory"
  fi
}

# ---------- Main ----------
main() {
  banner
  need curl
  need tar
  need uname

  detect_platform

  local version
  version="$(resolve_version)"
  [ -n "$version" ] || die "could not resolve the latest version"
  info "Installing ${BIN_NAME} v${version} for ${OS}/${ARCH}"

  local tmp asset url
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT
  asset="${BIN_NAME}-${OS}-${ARCH}.tar.gz"
  url="https://github.com/${REPO}/releases/download/v${version}/${asset}"

  info "Downloading ${url}"
  curl -fSL "$url" -o "$tmp/$asset" || die "download failed — does a release exist for ${OS}/${ARCH}?"

  verify_checksum "$tmp" "$asset" "$version"

  tar -xzf "$tmp/$asset" -C "$tmp" || die "failed to extract archive"

  local bin="$tmp/$BIN_NAME"
  if [ ! -f "$bin" ]; then
    bin="$(find "$tmp" -type f -name "$BIN_NAME" 2>/dev/null | head -1)"
  fi
  [ -n "$bin" ] && [ -f "$bin" ] || die "could not find the ${BIN_NAME} binary inside the archive"
  chmod +x "$bin"

  local dir
  dir="$(choose_install_dir)"
  install_binary "$bin" "$dir"

  hash -r 2>/dev/null || true
  if command -v "$BIN_NAME" >/dev/null 2>&1; then
    ok "installed: $(command -v "$BIN_NAME")"
    "$BIN_NAME" version || true
  else
    ok "installed to ${dir}/${BIN_NAME}"
    warn "${dir} is not on your PATH yet. Add it with:"
    warn "    export PATH=\"${dir}:\$PATH\""
  fi
  printf '\n'
  ok "Done! Try:  portkill kill 8080"
}

main "$@"
