#!/bin/bash

################################################################################
# Release Script for Harpoon v1.0
# This script builds binaries for all platforms and creates release packages
################################################################################

set -e

VERSION="v1.0"
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "🎯 Building Harpoon ${VERSION} Release"
echo "Commit: ${COMMIT}"
echo "Build Date: ${BUILD_DATE}"
echo ""

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf dist/
rm -f hpn-*
mkdir -p dist

# Build for all platforms
echo "🔨 Building binaries for all platforms..."

platforms=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    
    if [ "$GOOS" = "windows" ]; then
        output_name="hpn-${GOOS}-${GOARCH}.exe"
    else
        output_name="hpn-${GOOS}-${GOARCH}"
    fi
    
    echo "Building ${output_name}..."
    
    GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build \
        -trimpath \
        -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE} -s -w" \
        -o ${output_name} \
        ./cmd/hpn
    
    if [ $? -eq 0 ]; then
        echo "✅ Successfully built ${output_name}"
        
        # Make executable on Unix-like systems
        if [[ "${GOOS}" != "windows" ]]; then
            chmod +x ${output_name}
        fi
        
        # Show file info
        ls -lh ${output_name}
    else
        echo "❌ Failed to build ${output_name}"
        exit 1
    fi
done

echo ""
echo "📦 Creating release packages..."

# Create packages
mkdir -p dist/packages

# Linux AMD64
echo "Creating Linux AMD64 package..."
tar -czf dist/packages/hpn-${VERSION}-linux-amd64.tar.gz \
    hpn-linux-amd64 README.md LICENSE config.yaml.example docs/

# Linux ARM64
echo "Creating Linux ARM64 package..."
tar -czf dist/packages/hpn-${VERSION}-linux-arm64.tar.gz \
    hpn-linux-arm64 README.md LICENSE config.yaml.example docs/

# macOS AMD64
echo "Creating macOS AMD64 package..."
tar -czf dist/packages/hpn-${VERSION}-darwin-amd64.tar.gz \
    hpn-darwin-amd64 README.md LICENSE config.yaml.example docs/

# macOS ARM64
echo "Creating macOS ARM64 package..."
tar -czf dist/packages/hpn-${VERSION}-darwin-arm64.tar.gz \
    hpn-darwin-arm64 README.md LICENSE config.yaml.example docs/

# Windows AMD64
echo "Creating Windows AMD64 package..."
zip -r dist/packages/hpn-${VERSION}-windows-amd64.zip \
    hpn-windows-amd64.exe README.md LICENSE config.yaml.example docs/

# Move binaries to dist
mv hpn-* dist/

echo ""
echo "📋 Release Summary:"
echo "=================="
ls -lh dist/
echo ""
ls -lh dist/packages/

echo ""
echo "🎉 Release ${VERSION} build completed successfully!"
echo ""
echo "Next steps:"
echo "1. Test the binaries"
echo "2. Create git tag: git tag ${VERSION}"
echo "3. Push tag: git push origin ${VERSION}"
echo "4. Create GitHub release with the packages in dist/packages/"