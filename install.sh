#!/bin/sh
# Install sensor-gen - IoT/OT sensor data generator
# Usage: curl -fsSL https://raw.githubusercontent.com/aronchick/sensor-gen/main/install.sh | sh
set -e

REPO="aronchick/sensor-gen"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY="sensor-gen"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  linux)  OS="linux" ;;
  darwin) OS="darwin" ;;
  *)      echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)  ARCH="amd64" ;;
  aarch64|arm64)  ARCH="arm64" ;;
  *)              echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

# Get latest release tag (or use specified version)
VERSION="${VERSION:-latest}"
if [ "$VERSION" = "latest" ]; then
  VERSION=$(curl -fsSL \
    "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | cut -d'"' -f4)
fi

if [ -z "$VERSION" ]; then
  echo "Error: could not determine latest version"
  exit 1
fi

ASSET="sensor-gen-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"

echo "Installing sensor-gen ${VERSION} (${OS}/${ARCH})..."
echo "  From: ${URL}"
echo "  To:   ${INSTALL_DIR}/${BINARY}"

# Download
TMP=$(mktemp)
if ! curl -fsSL "$URL" -o "$TMP"; then
  echo "Error: download failed. Check the version/platform."
  rm -f "$TMP"
  exit 1
fi

# Install
chmod +x "$TMP"
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP" "${INSTALL_DIR}/${BINARY}"
else
  echo "Need sudo to install to ${INSTALL_DIR}"
  sudo mv "$TMP" "${INSTALL_DIR}/${BINARY}"
fi

echo "Installed: $(${BINARY} -h 2>&1 | head -1 || echo "${BINARY} ${VERSION}")"
echo ""
echo "Quick start:"
echo "  sensor-gen -d 30s -v          # 30s of data, verbose"
echo "  sensor-gen -rate 50000 -d 10s # high throughput"
