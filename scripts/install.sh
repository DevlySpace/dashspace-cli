#!/bin/bash

set -e

# Variables
REPO="dashspace/cli"
BINARY_NAME="dashspace"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Utility functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Detect OS and architecture
detect_os_arch() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case $OS in
        linux*)
            OS="linux"
            ;;
        darwin*)
            OS="darwin"
            ;;
        *)
            log_error "Unsupported OS: $OS"
            exit 1
            ;;
    esac

    case $ARCH in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            log_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    log_info "Detected OS: $OS"
    log_info "Detected architecture: $ARCH"
}

# Get latest version
get_latest_version() {
    log_info "Fetching latest version..."

    VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$VERSION" ]; then
        log_error "Failed to get latest version"
        exit 1
    fi

    log_info "Latest version: $VERSION"
}

# Download binary
download_binary() {
    BINARY_URL="https://github.com/$REPO/releases/download/$VERSION/${BINARY_NAME}-${OS}-${ARCH}"

    if [ "$OS" = "windows" ]; then
        BINARY_URL="${BINARY_URL}.exe"
        BINARY_NAME="${BINARY_NAME}.exe"
    fi

    log_info "Downloading from: $BINARY_URL"

    # Create temporary directory
    TMP_DIR=$(mktemp -d)
    TMP_FILE="$TMP_DIR/$BINARY_NAME"

    # Download with curl or wget
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$TMP_FILE" "$BINARY_URL"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$TMP_FILE" "$BINARY_URL"
    else
        log_error "curl or wget required for download"
        exit 1
    fi

    # Verify file was downloaded
    if [ ! -f "$TMP_FILE" ]; then
        log_error "Download failed"
        exit 1
    fi

    # Make binary executable
    chmod +x "$TMP_FILE"

    log_success "Binary downloaded: $TMP_FILE"
}

# Install binary
install_binary() {
    log_info "Installing to $INSTALL_DIR..."

    # Check permissions
    if [ ! -w "$INSTALL_DIR" ]; then
        log_warning "Admin permissions required for $INSTALL_DIR"
        sudo mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    else
        mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    fi

    # Cleanup
    rm -rf "$TMP_DIR"

    log_success "DashSpace CLI installed successfully!"
}

# Verify installation
verify_installation() {
    log_info "Verifying installation..."

    if command -v $BINARY_NAME >/dev/null 2>&1; then
        VERSION_OUTPUT=$($BINARY_NAME --version)
        log_success "Installation verified: $VERSION_OUTPUT"
    else
        log_error "Installation failed - binary not found in PATH"
        exit 1
    fi
}

# Show post-installation instructions
show_next_steps() {
    echo
    log_success "🎉 Installation complete!"
    echo
    echo "📚 Available commands:"
    echo "  dashspace login      # Login to DashSpace"
    echo "  dashspace create     # Create a new module"
    echo "  dashspace preview    # Preview in Buildy"
    echo "  dashspace publish    # Publish to store"
    echo
    echo "🔗 Documentation: https://docs.dashspace.dev/cli"
    echo "💬 Support: https://discord.gg/dashspace"
}

# Main script
main() {
    echo
    log_info "🚀 Installing DashSpace CLI"
    echo

    detect_os_arch
    get_latest_version
    download_binary
    install_binary
    verify_installation
    show_next_steps
}

# Run script
main "$@"