#!/bin/bash

################################################################################
# Harpoon Installation Script
# This script downloads and installs the latest version of hpn
################################################################################

set -e

# Configuration
REPO="ghostwritten/harpoon"
BINARY_NAME="hpn"
INSTALL_DIR="/usr/local/bin"
VERSION="v1.0"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect platform
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case $os in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        *)
            log_error "Unsupported operating system: $os"
            exit 1
            ;;
    esac
    
    case $arch in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            log_error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
    
    PLATFORM="${OS}-${ARCH}"
    log_info "Detected platform: $PLATFORM"
}

# Download and install
install_hpn() {
    local download_url="https://github.com/${REPO}/releases/download/${VERSION}/hpn-${VERSION}-${PLATFORM}.tar.gz"
    local temp_dir=$(mktemp -d)
    local temp_file="${temp_dir}/hpn.tar.gz"
    
    log_info "Downloading hpn ${VERSION} for ${PLATFORM}..."
    log_info "URL: $download_url"
    
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$temp_file" "$download_url"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$temp_file" "$download_url"
    else
        log_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
    
    if [ ! -f "$temp_file" ]; then
        log_error "Failed to download hpn"
        exit 1
    fi
    
    log_info "Extracting archive..."
    tar -xzf "$temp_file" -C "$temp_dir"
    
    local binary_path="${temp_dir}/hpn-${PLATFORM}"
    if [ ! -f "$binary_path" ]; then
        log_error "Binary not found in archive"
        exit 1
    fi
    
    log_info "Installing to ${INSTALL_DIR}..."
    
    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        cp "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    else
        sudo cp "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    # Clean up
    rm -rf "$temp_dir"
    
    log_success "hpn ${VERSION} installed successfully!"
}

# Verify installation
verify_installation() {
    if command -v hpn >/dev/null 2>&1; then
        log_success "Installation verified!"
        log_info "Version: $(hpn --version)"
        echo ""
        echo "🎯 Harpoon is ready to use!"
        echo ""
        echo "Quick start:"
        echo "  hpn --help                    # Show help"
        echo "  hpn -a pull -f images.txt     # Pull images"
        echo "  hpn -a save -f images.txt     # Save images"
        echo ""
        echo "Documentation: https://github.com/${REPO}"
    else
        log_error "Installation verification failed. hpn command not found."
        log_warning "You may need to add ${INSTALL_DIR} to your PATH"
        exit 1
    fi
}

# Main installation process
main() {
    echo "🎯 Harpoon Installation Script"
    echo "=============================="
    echo ""
    
    # Check if already installed
    if command -v hpn >/dev/null 2>&1; then
        local current_version=$(hpn --version 2>/dev/null | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+' || echo "unknown")
        log_warning "hpn is already installed (version: $current_version)"
        echo -n "Do you want to reinstall? [y/N]: "
        read -r response
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            log_info "Installation cancelled"
            exit 0
        fi
    fi
    
    detect_platform
    install_hpn
    verify_installation
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Harpoon Installation Script"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --version, -v  Show version information"
        echo ""
        echo "Environment variables:"
        echo "  INSTALL_DIR    Installation directory (default: /usr/local/bin)"
        echo "  VERSION        Version to install (default: v1.0)"
        echo ""
        echo "Examples:"
        echo "  $0                           # Install latest version"
        echo "  VERSION=v1.0 $0              # Install specific version"
        echo "  INSTALL_DIR=~/bin $0         # Install to custom directory"
        exit 0
        ;;
    --version|-v)
        echo "Harpoon Installation Script"
        echo "Target version: ${VERSION}"
        exit 0
        ;;
    "")
        main
        ;;
    *)
        log_error "Unknown option: $1"
        echo "Use --help for usage information"
        exit 1
        ;;
esac