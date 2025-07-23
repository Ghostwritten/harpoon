#!/bin/bash

################################################################################
# Build Script for Harpoon (hpn)
# This script builds the hpn binary for different platforms
################################################################################

set -e

# Configuration
BINARY_NAME="hpn"
VERSION="v1.0"
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS="-ldflags=-X=main.version=${VERSION} -X=main.commit=${COMMIT} -X=main.date=${BUILD_DATE} -s -w"

echo "🎯 Building Harpoon (hpn) - Container Image Management Tool"
echo "Version: ${VERSION}"
echo "Commit: ${COMMIT}"
echo "Build Date: ${BUILD_DATE}"
echo ""

# Function to build for specific platform
build_platform() {
    local goos=$1
    local goarch=$2
    local output_name=$3
    
    echo "Building for ${goos}/${goarch}..."
    
    GOOS=${goos} GOARCH=${goarch} CGO_ENABLED=0 go build \
        -trimpath \
        -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE} -s -w" \
        -o ${output_name} \
        ./cmd/hpn
    
    if [ $? -eq 0 ]; then
        echo "✅ Successfully built ${output_name}"
        
        # Make executable on Unix-like systems
        if [[ "${goos}" != "windows" ]]; then
            chmod +x ${output_name}
        fi
        
        # Show file info
        ls -lh ${output_name}
    else
        echo "❌ Failed to build ${output_name}"
        exit 1
    fi
    echo ""
}

# Parse command line arguments
case "${1:-all}" in
    "linux")
        build_platform "linux" "amd64" "${BINARY_NAME}-linux-amd64"
        ;;
    "linux-arm64")
        build_platform "linux" "arm64" "${BINARY_NAME}-linux-arm64"
        ;;
    "darwin")
        build_platform "darwin" "amd64" "${BINARY_NAME}-darwin-amd64"
        ;;
    "darwin-arm64")
        build_platform "darwin" "arm64" "${BINARY_NAME}-darwin-arm64"
        ;;
    "windows")
        build_platform "windows" "amd64" "${BINARY_NAME}-windows-amd64.exe"
        ;;
    "current")
        echo "Building for current platform..."
        go build -trimpath -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE} -s -w" -o ${BINARY_NAME} ./cmd/hpn
        chmod +x ${BINARY_NAME}
        echo "✅ Successfully built ${BINARY_NAME}"
        ls -lh ${BINARY_NAME}
        ;;
    "all")
        echo "Building for all platforms..."
        build_platform "linux" "amd64" "${BINARY_NAME}-linux-amd64"
        build_platform "linux" "arm64" "${BINARY_NAME}-linux-arm64"
        build_platform "darwin" "amd64" "${BINARY_NAME}-darwin-amd64"
        build_platform "darwin" "arm64" "${BINARY_NAME}-darwin-arm64"
        build_platform "windows" "amd64" "${BINARY_NAME}-windows-amd64.exe"
        ;;
    "clean")
        echo "Cleaning build artifacts..."
        rm -f ${BINARY_NAME}*
        echo "✅ Cleaned"
        ;;
    "test")
        echo "Testing build..."
        go build -trimpath -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE} -s -w" -o ${BINARY_NAME} ./cmd/hpn
        chmod +x ${BINARY_NAME}
        echo ""
        echo "Testing help command:"
        ./${BINARY_NAME} --help
        echo ""
        echo "Testing version command:"
        ./${BINARY_NAME} --version
        ;;
    *)
        echo "Usage: $0 [linux|linux-arm64|darwin|darwin-arm64|windows|current|all|clean|test]"
        echo ""
        echo "Options:"
        echo "  linux        - Build for Linux AMD64"
        echo "  linux-arm64  - Build for Linux ARM64"
        echo "  darwin       - Build for macOS AMD64"
        echo "  darwin-arm64 - Build for macOS ARM64 (Apple Silicon)"
        echo "  windows      - Build for Windows AMD64"
        echo "  current      - Build for current platform"
        echo "  all          - Build for all platforms (default)"
        echo "  clean        - Clean build artifacts"
        echo "  test         - Build and test basic functionality"
        exit 1
        ;;
esac

echo "🎯 Build completed successfully!"