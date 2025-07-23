#!/bin/bash

################################################################################
# Pre-release Check Script for Harpoon v1.0
# This script performs various checks before releasing
################################################################################

set -e

echo "🔍 Pre-release checks for Harpoon v1.0"
echo "======================================"

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Not in a git repository"
    exit 1
fi

# Check if working directory is clean
if ! git diff-index --quiet HEAD --; then
    echo "❌ Working directory is not clean. Please commit all changes."
    git status --porcelain
    exit 1
fi

echo "✅ Working directory is clean"

# Check if all required files exist
required_files=(
    "README.md"
    "LICENSE"
    "CHANGELOG.md"
    "config.yaml.example"
    "cmd/hpn/main.go"
    "cmd/hpn/root.go"
    "go.mod"
    "go.sum"
)

echo ""
echo "📁 Checking required files..."
for file in "${required_files[@]}"; do
    if [ -f "$file" ]; then
        echo "✅ $file"
    else
        echo "❌ Missing: $file"
        exit 1
    fi
done

# Check Go version
echo ""
echo "🔧 Checking Go version..."
go_version=$(go version | awk '{print $3}' | sed 's/go//')
echo "Go version: $go_version"

# Check if go mod tidy is needed
echo ""
echo "📦 Checking go.mod..."
go mod tidy
if ! git diff-index --quiet HEAD -- go.mod go.sum; then
    echo "❌ go.mod or go.sum needs to be updated. Run 'go mod tidy'"
    exit 1
fi
echo "✅ go.mod is up to date"

# Run tests
echo ""
echo "🧪 Running tests..."
if go test ./...; then
    echo "✅ All tests passed"
else
    echo "❌ Tests failed"
    exit 1
fi

# Check if version is consistent
echo ""
echo "🏷️  Checking version consistency..."
version_in_root=$(grep 'version = "v1.0"' cmd/hpn/root.go)
version_in_makefile=$(grep 'VERSION=v1.0' Makefile)
version_in_build_sh=$(grep 'VERSION="v1.0"' build.sh)

if [ -n "$version_in_root" ] && [ -n "$version_in_makefile" ] && [ -n "$version_in_build_sh" ]; then
    echo "✅ Version v1.0 is consistent across files"
else
    echo "❌ Version inconsistency detected"
    echo "root.go: $version_in_root"
    echo "Makefile: $version_in_makefile"
    echo "build.sh: $version_in_build_sh"
    exit 1
fi

# Test build
echo ""
echo "🔨 Testing build..."
if go build -o hpn-test ./cmd/hpn; then
    echo "✅ Build successful"
    rm -f hpn-test
else
    echo "❌ Build failed"
    exit 1
fi

# Test basic functionality
echo ""
echo "🧪 Testing basic functionality..."
go build -o hpn-test ./cmd/hpn
if ./hpn-test --help > /dev/null 2>&1; then
    echo "✅ Help command works"
else
    echo "❌ Help command failed"
    rm -f hpn-test
    exit 1
fi

if ./hpn-test --version > /dev/null 2>&1; then
    echo "✅ Version command works"
else
    echo "❌ Version command failed"
    rm -f hpn-test
    exit 1
fi

rm -f hpn-test

# Check documentation
echo ""
echo "📚 Checking documentation..."
if grep -q "v1.0" README.md; then
    echo "✅ README.md mentions v1.0"
else
    echo "❌ README.md doesn't mention v1.0"
    exit 1
fi

if grep -q "ghostwritten" README.md; then
    echo "✅ README.md has correct GitHub username"
else
    echo "❌ README.md doesn't have correct GitHub username"
    exit 1
fi

echo ""
echo "🎉 All pre-release checks passed!"
echo ""
echo "Ready to release v1.0! Next steps:"
echo "1. Run: ./release.sh"
echo "2. Test the built binaries"
echo "3. Create git tag: git tag v1.0"
echo "4. Push tag: git push origin v1.0"
echo "5. Create GitHub release with the packages"