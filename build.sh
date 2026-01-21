#!/bin/bash

# Automated build script for hpn
# Automatically detects and sets up Go environment, checks dependencies, and builds

set -e

BINARY_NAME="hpn"
REQUIRED_GO_VERSION="1.21"
DIST_DIR="dist"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Error handling function
error_exit() {
    echo -e "${RED}Error: $1${NC}" >&2
    echo ""
    echo "For help, see: docs/building.md"
    exit 1
}

# Setup Go environment automatically
setup_go_environment() {
    # Check if go command is available
    if command -v go >/dev/null 2>&1; then
        # Try to get GOROOT from go env
        GOROOT_VALUE=$(go env GOROOT 2>/dev/null || echo "")
        
        # If GOROOT is empty or invalid, try to find it
        if [ -z "$GOROOT_VALUE" ] || [ ! -d "$GOROOT_VALUE" ]; then
            # Try common Homebrew locations (macOS)
            if [ -d "/opt/homebrew/Cellar/go" ]; then
                GO_VERSION_DIR=$(ls -t /opt/homebrew/Cellar/go/ 2>/dev/null | head -1)
                if [ -n "$GO_VERSION_DIR" ] && [ -d "/opt/homebrew/Cellar/go/$GO_VERSION_DIR/libexec" ]; then
                    export GOROOT="/opt/homebrew/Cellar/go/$GO_VERSION_DIR/libexec"
                    export PATH="$GOROOT/bin:$PATH"
                fi
            # Try /usr/local for standard installations
            elif [ -d "/usr/local/go" ]; then
                export GOROOT="/usr/local/go"
                export PATH="$GOROOT/bin:$PATH"
            fi
        else
            export GOROOT="$GOROOT_VALUE"
        fi
    else
        # Go not in PATH, try to find it
        if [ -d "/opt/homebrew/Cellar/go" ]; then
            GO_VERSION_DIR=$(ls -t /opt/homebrew/Cellar/go/ 2>/dev/null | head -1)
            if [ -n "$GO_VERSION_DIR" ] && [ -d "/opt/homebrew/Cellar/go/$GO_VERSION_DIR/libexec" ]; then
                export GOROOT="/opt/homebrew/Cellar/go/$GO_VERSION_DIR/libexec"
                export PATH="$GOROOT/bin:$PATH"
            fi
        elif [ -d "/usr/local/go" ]; then
            export GOROOT="/usr/local/go"
            export PATH="$GOROOT/bin:$PATH"
        fi
    fi
    
    # Verify Go is now available
    if ! command -v go >/dev/null 2>&1; then
        error_exit "Go is not installed or not in PATH.

Please install Go ${REQUIRED_GO_VERSION} or higher:
  macOS (Homebrew): brew install go
  Linux: https://golang.org/dl/
  Or download from: https://golang.org/dl/"
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    GO_MAJOR=$(echo $GO_VERSION | cut -d. -f1)
    GO_MINOR=$(echo $GO_VERSION | cut -d. -f2)
    REQUIRED_MAJOR=$(echo $REQUIRED_GO_VERSION | cut -d. -f1)
    REQUIRED_MINOR=$(echo $REQUIRED_GO_VERSION | cut -d. -f2)
    
    if [ "$GO_MAJOR" -lt "$REQUIRED_MAJOR" ] || \
       ([ "$GO_MAJOR" -eq "$REQUIRED_MAJOR" ] && [ "$GO_MINOR" -lt "$REQUIRED_MINOR" ]); then
        error_exit "Go version ${GO_VERSION} is too old. Required: ${REQUIRED_GO_VERSION} or higher.

Current version: ${GO_VERSION}
Required version: ${REQUIRED_GO_VERSION}+

Please upgrade Go:
  macOS (Homebrew): brew upgrade go
  Or download from: https://golang.org/dl/"
    fi
    
    echo -e "${GREEN}✓${NC} Go environment detected: $(go version)"
}

# Check and download dependencies
check_dependencies() {
    if [ ! -f "go.mod" ]; then
        error_exit "go.mod not found. Are you in the project root directory?"
    fi
    
    echo "Checking dependencies..."
    if ! go mod download >/dev/null 2>&1; then
        echo -e "${YELLOW}Warning: Failed to download dependencies automatically.${NC}"
        echo "Trying to continue anyway..."
    else
        echo -e "${GREEN}✓${NC} Dependencies ready"
    fi
}

# Verify project structure
verify_project() {
    if [ ! -d "cmd/hpn" ]; then
        error_exit "cmd/hpn directory not found. Are you in the project root directory?"
    fi
    
    if [ ! -f "cmd/hpn/main.go" ]; then
        error_exit "cmd/hpn/main.go not found. Project structure may be corrupted."
    fi
}

# Get version information
get_version_info() {
    VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    # Build flags
    LDFLAGS="-X github.com/harpoon/hpn/internal/version.Version=${VERSION}"
    LDFLAGS="${LDFLAGS} -X github.com/harpoon/hpn/internal/version.GitCommit=${COMMIT}"
    LDFLAGS="${LDFLAGS} -X github.com/harpoon/hpn/internal/version.BuildDate=${BUILD_DATE}"
}

# Build for current platform
build_current() {
    echo ""
    echo "Building for current platform..."
    get_version_info
    echo "Version: ${VERSION}, Commit: ${COMMIT}"
    
    # Create dist directory if it doesn't exist
    mkdir -p "${DIST_DIR}"
    
    if go build -ldflags "${LDFLAGS}" -o ${DIST_DIR}/${BINARY_NAME} ./cmd/hpn; then
        echo -e "${GREEN}✅ Built ${BINARY_NAME}${NC}"
        echo "Binary location: $(pwd)/${DIST_DIR}/${BINARY_NAME}"
        echo "Binary size: $(ls -lh ${DIST_DIR}/${BINARY_NAME} | awk '{print $5}')"
    else
        error_exit "Build failed. Check the error messages above for details."
    fi
}

# Build for all platforms
build_all() {
    echo ""
    echo "Building for all platforms..."
    get_version_info
    echo "Version: ${VERSION}, Commit: ${COMMIT}"
    echo ""
    
    # Create dist directory if it doesn't exist
    mkdir -p "${DIST_DIR}"
    
    # Linux (static build with CGO disabled for RHEL 8.x compatibility)
    echo -n "Building Linux (amd64) with CGO_ENABLED=0... "
    if CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo,osusergo -ldflags "${LDFLAGS}" -o ${DIST_DIR}/${BINARY_NAME}-linux-amd64 ./cmd/hpn 2>/dev/null; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC}"
        error_exit "Failed to build Linux (amd64)"
    fi
    
    echo -n "Building Linux (arm64) with CGO_ENABLED=0... "
    if CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags netgo,osusergo -ldflags "${LDFLAGS}" -o ${DIST_DIR}/${BINARY_NAME}-linux-arm64 ./cmd/hpn 2>/dev/null; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC}"
        error_exit "Failed to build Linux (arm64)"
    fi
    
    # macOS
    echo -n "Building macOS (amd64)... "
    if GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${DIST_DIR}/${BINARY_NAME}-darwin-amd64 ./cmd/hpn 2>/dev/null; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC}"
        error_exit "Failed to build macOS (amd64)"
    fi
    
    echo -n "Building macOS (arm64)... "
    if GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o ${DIST_DIR}/${BINARY_NAME}-darwin-arm64 ./cmd/hpn 2>/dev/null; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC}"
        error_exit "Failed to build macOS (arm64)"
    fi
    
    # Windows
    echo -n "Building Windows (amd64)... "
    if GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${DIST_DIR}/${BINARY_NAME}-windows-amd64.exe ./cmd/hpn 2>/dev/null; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC}"
        error_exit "Failed to build Windows (amd64)"
    fi
    
    echo ""
    echo -e "${GREEN}✅ Built all platforms${NC}"
    echo ""
    echo "Generated binaries:"
    ls -lh ${DIST_DIR}/${BINARY_NAME}-* 2>/dev/null | awk '{print "  - " $9 ": " $5}'
}

# Clean build artifacts
clean_build() {
    echo "Cleaning build artifacts..."
    rm -rf ${DIST_DIR}
    rm -f ${BINARY_NAME} ${BINARY_NAME}-*
    echo -e "${GREEN}✅ Cleaned${NC}"
}

# Main function
main() {
    # Setup environment first
    setup_go_environment
    
    # Verify project structure
    verify_project
    
    # Check dependencies
    check_dependencies
    
    # Execute build command
    case "${1:-current}" in
        "current")
            build_current
            ;;
        "all")
            build_all
            ;;
        "clean")
            clean_build
            ;;
        *)
            echo "Usage: $0 [current|all|clean]"
            echo ""
            echo "Commands:"
            echo "  current  Build for current platform (default)"
            echo "  all      Build for all supported platforms"
            echo "  clean    Remove all build artifacts"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"