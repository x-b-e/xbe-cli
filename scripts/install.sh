#!/usr/bin/env bash
set -euo pipefail

REPO="${REPO:-x-b-e/xbe-cli}"
OS=$(uname -s | tr "[:upper:]" "[:lower:]")
ARCH=$(uname -m)

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

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

curl -fsSL "https://github.com/${REPO}/releases/download/${TAG}/xbe_${VERSION}_${OS}_${ARCH}.tar.gz" \
  | tar -xz -C "$TMP_DIR"

sudo mv "$TMP_DIR/xbe" /usr/local/bin/xbe
xbe version
