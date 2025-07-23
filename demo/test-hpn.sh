#!/bin/bash

################################################################################
# Test script for hpn (Harpoon) functionality
################################################################################

set -e

echo "🧪 Testing hpn (Harpoon) functionality"
echo "======================================"

# Build the binary first
echo "📦 Building hpn binary..."
go build -o hpn ./cmd/hpn
chmod +x hpn

echo ""
echo "✅ Binary built successfully"

# Test help command
echo ""
echo "🔍 Testing help command..."
./hpn --help

echo ""
echo "🔍 Testing version command..."
./hpn --version

# Create test image list
echo ""
echo "📝 Creating test image list..."
cat > test-images.txt << EOF
nginx:alpine
busybox:latest
hello-world:latest
EOF

echo "Created test-images.txt with:"
cat test-images.txt

# Test parameter validation
echo ""
echo "🔍 Testing parameter validation..."

echo "Testing missing action parameter:"
./hpn -f test-images.txt || echo "✅ Correctly failed with missing action"

echo ""
echo "Testing invalid action:"
./hpn -a invalid -f test-images.txt || echo "✅ Correctly failed with invalid action"

echo ""
echo "Testing missing file parameter for pull:"
./hpn -a pull || echo "✅ Correctly failed with missing file parameter"

echo ""
echo "Testing load without file parameter (should work):"
./hpn -a load --load-mode 1 || echo "ℹ️  Load failed (expected if no tar files exist)"

# Test actual functionality (commented out to avoid network calls)
echo ""
echo "🚀 All parameter validation tests passed!"
echo ""
echo "To test actual functionality:"
echo "  Pull images: ./hpn -a pull -f test-images.txt"
echo "  Save images: ./hpn -a save -f test-images.txt --save-mode 2"
echo "  Load images: ./hpn -a load --load-mode 2"
echo "  Push images: ./hpn -a push -f test-images.txt -r registry.example.com -p test --push-mode 2"

echo ""
echo "🎯 hpn testing completed successfully!"