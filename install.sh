#!/bin/bash

# CuRe Code - Installer Script
# Usage: curl -fsSL https://raw.githubusercontent.com/broman0x/cure-code/main/install.sh | bash

set -e

REPO="broman0x/cure-code"
BINARY_NAME="curecode"

# [EN] Detect OS and Architecture
# [ID] Deteksi OS dan Arsitektur
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    linux) OS="linux" ;;
    darwin) OS="darwin" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# [EN] Get latest release version from GitHub
# [ID] Ambil versi rilis terbaru dari GitHub
echo "✦ Checking for latest version..."
LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_TAG" ]; then
    echo "Error: Could not find latest release."
    exit 1
fi

DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_TAG/curecode-$OS-$ARCH"
if [ "$OS" == "darwin" ] && [ "$ARCH" == "arm64" ]; then
    # Fallback if arm64 build doesn't exist yet
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_TAG/curecode-darwin-amd64"
fi

echo "✦ Downloading CuRe Code $LATEST_TAG for $OS-$ARCH..."
curl -L "$DOWNLOAD_URL" -o "$BINARY_NAME"
chmod +x "$BINARY_NAME"

# [EN] Install to /usr/local/bin
# [ID] Instal ke /usr/local/bin
INSTALL_DIR="/usr/local/bin"
echo "✦ Installing to $INSTALL_DIR (requires sudo)..."
sudo mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"

echo "✦ Installation complete!"
echo "✦ Running initial setup..."
"$INSTALL_DIR/$BINARY_NAME" --install < /dev/tty

echo "══════════════════════════════════════════"
echo "  CuRe Code installed successfully!"
echo "  Type 'curecode' to start."
echo "══════════════════════════════════════════"
