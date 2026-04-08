#!/usr/bin/env bash
# =============================================================================
# install.sh — portkill installer for macOS and Linux
# =============================================================================

set -euo pipefail

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

REPO="khanalsaroj/portkill"
BINARY_NAME="portkill"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# ---------------------------------------------------------------------------
# Colour helpers (degrade gracefully when not in a terminal)
# ---------------------------------------------------------------------------

if [ -t 1 ]; then
  RED="\033[0;31m"
  GREEN="\033[0;32m"
  YELLOW="\033[0;33m"
  CYAN="\033[0;36m"
  BOLD="\033[1m"
  RESET="\033[0m"
else
  RED="" GREEN="" YELLOW="" CYAN="" BOLD="" RESET=""
fi

info()    { printf "  ${CYAN}→${RESET}  %s\n" "$*"; }
success() { printf "  ${GREEN}✓${RESET}  %s\n" "$*"; }
warn()    { printf "  ${YELLOW}⚠${RESET}  %s\n" "$*" >&2; }
die()     { printf "  ${RED}✗${RESET}  %s\n" "$*" >&2; exit 1; }
header()  { printf "\n${BOLD}%s${RESET}\n" "$*"; }

# ---------------------------------------------------------------------------
# Preflight checks
# ---------------------------------------------------------------------------

header "portkill installer"

# Require bash 4+ (for associative arrays used below) or simply POSIX tools.
command -v curl  >/dev/null 2>&1 || command -v wget >/dev/null 2>&1 \
  || die "curl or wget is required but neither was found."

# ---------------------------------------------------------------------------
# Detect OS and architecture
# ---------------------------------------------------------------------------

detect_platform() {
  local os arch

  case "$(uname -s)" in
    Linux*)   os="linux"   ;;
    Darwin*)  os="darwin"  ;;
    FreeBSD*) os="freebsd" ;;
    *)        die "Unsupported OS: $(uname -s). Use install.ps1 on Windows." ;;
  esac

  case "$(uname -m)" in
    x86_64 | amd64)  arch="amd64" ;;
    arm64  | aarch64) arch="arm64" ;;
    armv7l)           arch="arm"   ;;
    *)                die "Unsupported architecture: $(uname -m)" ;;
  esac

  echo "${os}-${arch}"
}

PLATFORM="$(detect_platform)"
info "Detected platform: ${PLATFORM}"

# ---------------------------------------------------------------------------
# Resolve the version to install
# ---------------------------------------------------------------------------

resolve_version() {
  if [ -n "${VERSION:-}" ]; then
    echo "$VERSION"
    return
  fi

  info "Fetching latest release version..."

  local latest_url="https://api.github.com/repos/${REPO}/releases/latest"
  local version

  if command -v curl >/dev/null 2>&1; then
    version="$(curl -fsSL "$latest_url" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
  else
    version="$(wget -qO- "$latest_url" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
  fi

  [ -n "$version" ] || die "Could not determine latest release. Set VERSION=vX.Y.Z to install a specific version."
  echo "$version"
}

VERSION="$(resolve_version)"
info "Installing version: ${VERSION}"

# ---------------------------------------------------------------------------
# Build the download URL
#
# Expected GitHub release asset naming convention (set this in your
# release workflow):
#   portkill_linux_amd64
#   portkill_linux_arm64
#   portkill_darwin_amd64
#   portkill_darwin_arm64
# ---------------------------------------------------------------------------

ASSET_NAME="${BINARY_NAME}-${PLATFORM}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET_NAME}"

# ---------------------------------------------------------------------------
# Download
# ---------------------------------------------------------------------------

TMP_DIR="$(mktemp -d)"
TMP_BIN="${TMP_DIR}/${BINARY_NAME}"

# Ensure we clean up the temp directory on exit (success or failure).
trap 'rm -rf "$TMP_DIR"' EXIT

info "Downloading ${DOWNLOAD_URL} ..."

if command -v curl >/dev/null 2>&1; then
  curl -fSL --progress-bar "$DOWNLOAD_URL" -o "$TMP_BIN" \
    || die "Download failed. Check that ${VERSION} exists for platform ${PLATFORM}."
else
  wget -q --show-progress "$DOWNLOAD_URL" -O "$TMP_BIN" \
    || die "Download failed. Check that ${VERSION} exists for platform ${PLATFORM}."
fi

chmod +x "$TMP_BIN"

# ---------------------------------------------------------------------------
# Verify the binary runs (smoke test)
# ---------------------------------------------------------------------------

"$TMP_BIN" --help >/dev/null 2>&1 || "$TMP_BIN" help >/dev/null 2>&1 \
  || warn "Smoke test inconclusive — the binary may still work."

# ---------------------------------------------------------------------------
# Install
# ---------------------------------------------------------------------------

install_binary() {
  local dest="${INSTALL_DIR}/${BINARY_NAME}"

  # Create the install directory if it doesn't exist (e.g. ~/.local/bin).
  if [ ! -d "$INSTALL_DIR" ]; then
    info "Creating install directory: ${INSTALL_DIR}"
    mkdir -p "$INSTALL_DIR" 2>/dev/null \
      || sudo mkdir -p "$INSTALL_DIR" \
      || die "Could not create ${INSTALL_DIR}"
  fi

  # Try a direct move first; fall back to sudo if permission is denied.
  if mv "$TMP_BIN" "$dest" 2>/dev/null; then
    : # success
  elif command -v sudo >/dev/null 2>&1; then
    info "Elevated permissions required — running sudo mv ..."
    sudo mv "$TMP_BIN" "$dest" \
      || die "Installation failed even with sudo. Try: INSTALL_DIR=~/.local/bin bash install.sh"
  else
    die "Permission denied writing to ${INSTALL_DIR} and sudo is not available.
Try: INSTALL_DIR=~/.local/bin bash install.sh"
  fi

  echo "$dest"
}

INSTALLED_PATH="$(install_binary)"
success "Installed to ${INSTALLED_PATH}"

# ---------------------------------------------------------------------------
# PATH check — warn if the install dir isn't on PATH
# ---------------------------------------------------------------------------

if ! echo ":${PATH}:" | grep -q ":${INSTALL_DIR}:"; then
  warn "${INSTALL_DIR} is not in your PATH."
  printf "\n  Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):\n"
  printf "\n    ${BOLD}export PATH=\"\$PATH:${INSTALL_DIR}\"${RESET}\n\n"
  printf "  Then reload your shell:\n"
  printf "\n    ${BOLD}source ~/.bashrc${RESET}  (or open a new terminal)\n\n"
fi

# ---------------------------------------------------------------------------
# Done
# ---------------------------------------------------------------------------

printf "\n${BOLD}All done!${RESET} Run it:\n\n"
printf "  ${BOLD}portkill kill 8080${RESET}\n\n"