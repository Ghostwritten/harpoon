#!/bin/bash

# Simple build script for hpn
set -e

BINARY_NAME="hpn"
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS="-X github.com/harpoon/hpn/internal/version.Version=${VERSION}"
LDFLAGS="${LDFLAGS} -X github.com/harpoon/hpn/internal/version.GitCommit=${COMMIT}"
LDFLAGS="${LDFLAGS} -X github.com/harpoon/hpn/internal/version.BuildDate=${BUILD_DATE}"

case "${1:-current}" in
    "current")
        echo "Building for current platform..."
        echo "Version: ${VERSION}, Commit: ${COMMIT}"
        go build -ldflags "${LDFLAGS}" -o ${BINARY_NAME} ./cmd/hpn
        echo "✅ Built ${BINARY_NAME}"
        ;;
    "all")
        echo "Building for all platforms..."
        echo "Version: ${VERSION}, Commit: ${COMMIT}"
        
        GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BINARY_NAME}-linux-amd64 ./cmd/hpn
        GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BINARY_NAME}-darwin-amd64 ./cmd/hpn
        GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o ${BINARY_NAME}-darwin-arm64 ./cmd/hpn
        GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BINARY_NAME}-windows-amd64.exe ./cmd/hpn
        echo "✅ Built all platforms"
        ;;
    "clean")
        rm -f ${BINARY_NAME}*
        echo "✅ Cleaned"
        ;;
    *)
        echo "Usage: $0 [current|all|clean]"
        exit 1
        ;;
esac