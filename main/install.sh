#!/usr/bin/env bash
set -euo pipefail


# ---------- Core Identity ----------
APP_NAME="portkill"
CTL_NAME="portkill"
GITHUB_ORG="khanalsaroj"
CTL_REPO="portkill"

# ---------- Paths (IMPORTANT SEPARATION) ----------
BIN_DIR="/usr/local/bin"
APP_HOME="/opt/${APP_NAME}"



# -------- ASCII Art & Branding --------
show_banner() {
  cat <<"CONFIGEOF"

    ░█▀█░█▀█░█▀▄░▀█▀░█░█░▀█▀░█░░░█░░
    ░█▀▀░█░█░█▀▄░░█░░█▀▄░░█░░█░░░█░░
    ░▀░░░▀▀▀░▀░▀░░▀░░▀░▀░▀▀▀░▀▀▀░▀▀▀

     🌟 Installation System | v1.0.0 🌟

CONFIGEOF
}

# ---------- Versioning ----------
DEFAULT_VERSION="latest"
MIN_BASH_VERSION=4
SUPPORTED_OS=("linux" "darwin")

# ---------- Logging ----------
info()    { printf "%s\n" "$*"; }
success() { printf "[OK]   %s\n" "$*"; }
error()   { printf "[ERR]  %s\n" "$*" >&2; exit 1; }

# ---------- Preconditions ----------
require_cmd() {
  command -v "$1" >/dev/null 2>&1 || error "Missing required command: $1"
}

check_bash() {
  (( BASH_VERSINFO[0] >= MIN_BASH_VERSION )) || \
    error "Bash ${MIN_BASH_VERSION}+ required"
}

require_root() {
  [[ "$(id -u)" -eq 0 ]] || error "Run as root (use sudo)"
}

require_docker() {
  command -v docker >/dev/null 2>&1 || error "Docker is not installed. Please install Docker."

  if ! docker info >/dev/null 2>&1; then
    error "Docker daemon is not running. Start Docker before proceeding."
  fi
}


# ---------- Spinner ----------
spinner() {
  local pid=$1
  local frames='|/-\'
  local i=0

  while kill -0 "$pid" 2>/dev/null; do
    printf "\r[%c] Working..." "${frames:i++%4:1}"
    sleep 0.1
  done

  printf "\r[✓] Done            \n"
}

run_with_spinner() {
  ("$@" >/dev/null 2>&1) &
  spinner $!
}

# ---------- System Detection ----------
detect_system() {
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"

  case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) error "Unsupported architecture: $ARCH" ;;
  esac

  [[ " ${SUPPORTED_OS[*]} " =~ " ${OS} " ]] || \
    error "Unsupported OS: $OS"
}

# ---------- Version Resolution ----------
resolve_version() {
  if [[ "$DEFAULT_VERSION" != "latest" ]]; then
    echo "$DEFAULT_VERSION"
    return
  fi

  curl -sfL \
    "https://api.github.com/repos/${GITHUB_ORG}/${CTL_REPO}/releases/latest" |
    grep -o '"tag_name": *"[^"]*"' |
    cut -d'"' -f4 |
    sed 's/^v//'
}



find_binary() {
  local root="$1"
  local bin=""

  # Strategy 1: exact match
  bin="$(find "$root" -type f -name "$CTL_NAME" -print -quit)"

  # Strategy 2: OS/ARCH suffixed
  if [[ -z "$bin" ]]; then
    bin="$(find "$root" -type f -name "${CTL_NAME}-${OS}-${ARCH}" -print -quit)"
  fi

  # Strategy 3: any executable named like the binary
  if [[ -z "$bin" ]]; then
    bin="$(find "$root" -type f -executable -name "*${CTL_NAME}*" -print -quit)"
  fi

  # Strategy 4: bin/ directory
  if [[ -z "$bin" && -d "$root/bin" ]]; then
    bin="$(find "$root/bin" -type f -print -quit)"
  fi

  # Strategy 5: last resort — first regular file
  if [[ -z "$bin" ]]; then
    bin="$(find "$root" -type f -print -quit)"
  fi

  [[ -n "$bin" ]] || return 1
  echo "$bin"
}



# ---------- Main ----------
main() {
  show_banner
  check_bash
  require_root
  require_docker

  require_cmd curl
  require_cmd tar
  require_cmd install

  detect_system

  info "Resolving version"
  VERSION="$(resolve_version)"
  [[ -n "$VERSION" ]] || error "Failed to resolve version"

  info "Installing ${CTL_NAME} v${VERSION}"

  URL="https://github.com/${GITHUB_ORG}/${CTL_REPO}/releases/download/v${VERSION}/${CTL_NAME}-${OS}-${ARCH}.tar.gz"

  TMP_DIR="$(mktemp -d)"
  trap 'rm -rf "$TMP_DIR"' EXIT

  info "Downloading binary"
  run_with_spinner curl -fL "$URL" -o "$TMP_DIR/pkg.tar.gz"

  info "Extracting archive"
  run_with_spinner tar -xzf "$TMP_DIR/pkg.tar.gz" -C "$TMP_DIR"

  info "Locating binary in archive"

  BIN="$(find_binary "$TMP_DIR")" || {
    error "Failed to locate binary in archive. Contents were: $(find "$TMP_DIR" -type f)"
  }

  chmod +x "$BIN"

  info "Installing binary to $BIN_DIR"
  install -m 755 "$BIN" "$BIN_DIR/$CTL_NAME"

  success "Binary: $BIN_DIR/$CTL_NAME"
  success "App home: $APP_HOME"

  if command -v "$CTL_NAME" >/dev/null 2>&1; then
      echo ""
      success "Checking installed version..."
      $CTL_NAME --version
  else
      echo ""
      error "Installation failed or $CTL_NAME is not in your PATH"
  fi

  echo "✨ ProtKill Installation Complete! ✨"
}

main "$@"