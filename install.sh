#!/bin/sh
# Install sensor-gen + expanso-edge (debug) for IoT demo
# Usage: curl -fsSL https://raw.githubusercontent.com/aronchick/sensor-gen/main/install.sh | sh
set -e

SENSOR_REPO="aronchick/sensor-gen"
EDGE_REPO="aronchick/expanso-edge-debug-public"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

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

# Helper: get latest release tag from a repo
# Uses /releases (not /releases/latest) to include prereleases
get_latest_tag() {
  curl -fsSL \
    "https://api.github.com/repos/$1/releases?per_page=1" \
    | grep '"tag_name"' | head -1 | cut -d'"' -f4
}

# Helper: download and install a binary
install_binary() {
  _repo="$1"
  _asset="$2"
  _name="$3"
  _version="$4"

  _url="https://github.com/${_repo}/releases/download/${_version}/${_asset}"
  echo "  ${_name}: ${_url}"

  _tmp=$(mktemp)
  if ! curl -fsSL "$_url" -o "$_tmp"; then
    echo "Error: failed to download ${_name}"
    rm -f "$_tmp"
    return 1
  fi

  chmod +x "$_tmp"
  if [ -w "$INSTALL_DIR" ]; then
    mv "$_tmp" "${INSTALL_DIR}/${_name}"
  else
    sudo mv "$_tmp" "${INSTALL_DIR}/${_name}"
  fi
}

echo "=== IoT Sensor Demo Installer ==="
echo "Platform: ${OS}/${ARCH}"
echo "Install to: ${INSTALL_DIR}"
echo ""

# Check if sudo will be needed
if [ ! -w "$INSTALL_DIR" ]; then
  echo "Note: will need sudo for ${INSTALL_DIR}"
  echo ""
fi

# --- sensor-gen ---
SENSOR_VERSION="${SENSOR_VERSION:-latest}"
if [ "$SENSOR_VERSION" = "latest" ]; then
  SENSOR_VERSION=$(get_latest_tag "$SENSOR_REPO")
fi
if [ -z "$SENSOR_VERSION" ]; then
  echo "Error: could not determine sensor-gen version"
  exit 1
fi

echo "Installing sensor-gen ${SENSOR_VERSION}..."
install_binary "$SENSOR_REPO" \
  "sensor-gen-${OS}-${ARCH}" \
  "sensor-gen" \
  "$SENSOR_VERSION"

# --- expanso-edge (debug) ---
EDGE_VERSION="${EDGE_VERSION:-latest}"
if [ "$EDGE_VERSION" = "latest" ]; then
  EDGE_VERSION=$(get_latest_tag "$EDGE_REPO")
fi
if [ -z "$EDGE_VERSION" ]; then
  echo "Warning: could not determine expanso-edge version, skipping"
else
  echo "Installing expanso-edge ${EDGE_VERSION} (debug)..."
  install_binary "$EDGE_REPO" \
    "expanso-edge-${OS}-${ARCH}" \
    "expanso-edge" \
    "$EDGE_VERSION"
fi

echo ""
echo "Installed:"
echo "  sensor-gen   $(sensor-gen -h 2>&1 | head -1 || echo "$SENSOR_VERSION")"
echo "  expanso-edge $(expanso-edge version 2>&1 || echo "$EDGE_VERSION")"
echo ""
echo "Quick start:"
echo "  sensor-gen -d 30s -v -o output.jsonl"
echo "  INPUT_FILE=output.jsonl expanso-edge run --local pipeline.yaml"
