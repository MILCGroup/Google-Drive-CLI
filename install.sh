#!/bin/bash
set -e

# Company Google Drive CLI Installer
# Pre-bundled OAuth credentials - no configuration needed
# Usage: curl -fsSL https://raw.githubusercontent.com/MILCGroup/Google-Drive-CLI/master/install.sh | bash

VERSION="${GDRV_VERSION:-latest}"
INSTALL_DIR="${GDRV_INSTALL_DIR:-$HOME/.local/bin}"
REPO="MILCGroup/Google-Drive-CLI"

INSTALL_BINARY="gdrv"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case "$OS" in
    darwin)
        OS="darwin"
        ;;
    linux)
        OS="linux"
        ;;
    mingw*|msys*|cygwin*)
        OS="windows"
        ;;
    *)
        echo "Unsupported operating system: $OS"
        exit 1
        ;;
esac

ASSET_NAME="gdrv-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    ASSET_NAME="${ASSET_NAME}.exe"
    INSTALL_BINARY="gdrv.exe"
fi

echo "Company Google Drive CLI Installer"
echo "==================================="
echo ""
echo "Detected: ${OS}/${ARCH}"
echo "Install directory: ${INSTALL_DIR}"
echo ""

# Create install directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# Download pre-built binary from GitHub releases
echo "Downloading pre-built binary..."

if [ "$VERSION" = "latest" ]; then
    DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${ASSET_NAME}"
    CHECKSUMS_URL="https://github.com/${REPO}/releases/latest/download/checksums.txt"
else
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET_NAME}"
    CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"
fi

echo "Downloading from: $DOWNLOAD_URL"

# Helper: download a URL to a file
_download() {
    local url="$1" dest="$2"
    if command -v curl &> /dev/null; then
        curl -fsSL "$url" -o "$dest"
    elif command -v wget &> /dev/null; then
        wget -q "$url" -O "$dest"
    else
        echo "Error: curl or wget required"
        exit 1
    fi
}

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

_download "$DOWNLOAD_URL" "$TMPDIR/$INSTALL_BINARY" || {
    echo ""
    echo "Error: Failed to download binary."
    echo "Please check that the repository exists and has releases available."
    echo "Repository: https://github.com/${REPO}"
    exit 1
}

# Verify checksum if shasum/sha256sum is available
if command -v shasum &> /dev/null || command -v sha256sum &> /dev/null; then
    echo "Verifying checksum..."
    _download "$CHECKSUMS_URL" "$TMPDIR/checksums.txt" || {
        echo "Warning: Could not download checksums.txt — skipping verification."
    }

    if [ -f "$TMPDIR/checksums.txt" ]; then
        EXPECTED=$(grep -F "  ${ASSET_NAME}" "$TMPDIR/checksums.txt" | awk '{print $1}')
        if [ -z "$EXPECTED" ]; then
            echo "Warning: ${ASSET_NAME} not found in checksums.txt — skipping verification."
        else
            if command -v shasum &> /dev/null; then
                ACTUAL=$(shasum -a 256 "$TMPDIR/$INSTALL_BINARY" | awk '{print $1}')
            else
                ACTUAL=$(sha256sum "$TMPDIR/$INSTALL_BINARY" | awk '{print $1}')
            fi
            if [ "$ACTUAL" != "$EXPECTED" ]; then
                echo ""
                echo "Error: Checksum mismatch!"
                echo "  Expected: $EXPECTED"
                echo "  Actual:   $ACTUAL"
                echo "The downloaded binary may be corrupted or tampered with."
                exit 1
            fi
            echo "Checksum verified."
        fi
    fi
fi

# Move verified binary into place
cp "$TMPDIR/$INSTALL_BINARY" "$INSTALL_DIR/$INSTALL_BINARY"

# Make executable
chmod +x "$INSTALL_DIR/$INSTALL_BINARY"

# Verify installation
if [ -x "$INSTALL_DIR/$INSTALL_BINARY" ]; then
    echo ""
    echo "Installation successful!"
    echo ""
    "$INSTALL_DIR/$INSTALL_BINARY" version 2>/dev/null || echo "gdrv installed to $INSTALL_DIR/$INSTALL_BINARY"
    echo ""
    
    # Check if install dir is in PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo "Add the following to your shell profile (.bashrc, .zshrc, etc.):"
        echo ""
        echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
    fi
    
    echo "Quick start:"
    echo "  gdrv auth login --preset workspace-basic    # Authenticate with company Drive"
    echo "  gdrv files list --json                        # List your files"
    echo "  gdrv --help                                   # See all commands"
    echo ""
    echo "Note: This CLI has pre-bundled OAuth credentials."
    echo "      No configuration needed - just authenticate!"
else
    echo "Installation failed"
    exit 1
fi
