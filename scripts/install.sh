#!/usr/bin/env bash
set -euo pipefail

REPO="${REPO:-x-b-e/xbe-cli}"
OS=$(uname -s | tr "[:upper:]" "[:lower:]")
ARCH=$(uname -m)

log() {
  printf '%s\n' "$*"
}

INSTALL_DIR="${INSTALL_DIR:-}"
USE_SUDO="${USE_SUDO:-0}"

case "$OS" in
  darwin|linux) ;;
  *) echo "unsupported OS: $OS" >&2; exit 1 ;;
esac

case "$ARCH" in
  x86_64|amd64) ARCH=amd64 ;;
  arm64|aarch64) ARCH=arm64 ;;
  *) echo "unsupported arch: $ARCH" >&2; exit 1 ;;
esac

TAG=${TAG:-}
if [ -z "$TAG" ]; then
  TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep -oE '"tag_name": "v[^"]+"' \
    | head -n1 \
    | cut -d'"' -f4)
fi

if [ -z "$TAG" ]; then
  echo "failed to resolve latest release tag" >&2
  exit 1
fi

VERSION=${TAG#v}

if [ -z "$INSTALL_DIR" ]; then
  if [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="${XDG_BIN_HOME:-$HOME/.local/bin}"
  fi
fi

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

log "Installing xbe ${TAG} for ${OS}/${ARCH}..."
log "Downloading release archive..."
curl -fsSL "https://github.com/${REPO}/releases/download/${TAG}/xbe_${VERSION}_${OS}_${ARCH}.tar.gz" \
  | tar -xz -C "$TMP_DIR"

if [ "$USE_SUDO" -eq 1 ]; then
  log "Installing to ${INSTALL_DIR} (sudo)..."
  sudo mv "$TMP_DIR/xbe" "${INSTALL_DIR}/xbe"
else
  log "Installing to ${INSTALL_DIR}..."
  mkdir -p "$INSTALL_DIR"
  mv "$TMP_DIR/xbe" "${INSTALL_DIR}/xbe"
fi

if ! command -v xbe >/dev/null 2>&1; then
  case ":$PATH:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
      log ""
      log "Note: ${INSTALL_DIR} is not on your PATH."
      log "Add this to your shell profile:"
      log "  export PATH=\"${INSTALL_DIR}:\$PATH\""
      ;;
  esac
fi

log ""
log "Done:"
"${INSTALL_DIR}/xbe" version
