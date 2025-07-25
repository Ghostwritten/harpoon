# Harpoon Build Configuration
BINARY_NAME=hpn
VERSION=v1.1
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build flags
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE} -s -w"
BUILD_FLAGS=-trimpath $(LDFLAGS)

# Default target
.PHONY: all
all: build

# Build for current platform
.PHONY: build
build:
	@echo "Building ${BINARY_NAME} for current platform..."
	go build ${BUILD_FLAGS} -o ${BINARY_NAME} ./cmd/hpn

# Build for Linux AMD64 (most common server platform)
.PHONY: build-linux
build-linux:
	@echo "Building ${BINARY_NAME} for Linux AMD64..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ${BUILD_FLAGS} -o ${BINARY_NAME}-linux-amd64 ./cmd/hpn

# Build for Linux ARM64
.PHONY: build-linux-arm64
build-linux-arm64:
	@echo "Building ${BINARY_NAME} for Linux ARM64..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build ${BUILD_FLAGS} -o ${BINARY_NAME}-linux-arm64 ./cmd/hpn

# Build for macOS AMD64
.PHONY: build-darwin
build-darwin:
	@echo "Building ${BINARY_NAME} for macOS AMD64..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build ${BUILD_FLAGS} -o ${BINARY_NAME}-darwin-amd64 ./cmd/hpn

# Build for macOS ARM64 (Apple Silicon)
.PHONY: build-darwin-arm64
build-darwin-arm64:
	@echo "Building ${BINARY_NAME} for macOS ARM64..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build ${BUILD_FLAGS} -o ${BINARY_NAME}-darwin-arm64 ./cmd/hpn

# Build for Windows AMD64
.PHONY: build-windows
build-windows:
	@echo "Building ${BINARY_NAME} for Windows AMD64..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ${BUILD_FLAGS} -o ${BINARY_NAME}-windows-amd64.exe ./cmd/hpn

# Build for FreeBSD AMD64 (experimental)
.PHONY: build-freebsd
build-freebsd:
	@echo "Building ${BINARY_NAME} for FreeBSD AMD64..."
	GOOS=freebsd GOARCH=amd64 CGO_ENABLED=0 go build ${BUILD_FLAGS} -o ${BINARY_NAME}-freebsd-amd64 ./cmd/hpn

# Build all platforms
.PHONY: build-all
build-all: build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows build-freebsd
	@echo "All builds completed!"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f ${BINARY_NAME}*
	rm -f *.tar.gz

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	golangci-lint run

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Install binary to system
.PHONY: install
install: build
	@echo "Installing ${BINARY_NAME} to /usr/local/bin..."
	sudo cp ${BINARY_NAME} /usr/local/bin/
	sudo chmod +x /usr/local/bin/${BINARY_NAME}

# Uninstall binary from system
.PHONY: uninstall
uninstall:
	@echo "Uninstalling ${BINARY_NAME} from /usr/local/bin..."
	sudo rm -f /usr/local/bin/${BINARY_NAME}

# Create release packages
.PHONY: package
package: build-all
	@echo "Creating release packages..."
	mkdir -p dist
	tar -czf dist/${BINARY_NAME}-${VERSION}-linux-amd64.tar.gz ${BINARY_NAME}-linux-amd64 README.md LICENSE docs/
	tar -czf dist/${BINARY_NAME}-${VERSION}-linux-arm64.tar.gz ${BINARY_NAME}-linux-arm64 README.md LICENSE docs/
	tar -czf dist/${BINARY_NAME}-${VERSION}-darwin-amd64.tar.gz ${BINARY_NAME}-darwin-amd64 README.md LICENSE docs/
	tar -czf dist/${BINARY_NAME}-${VERSION}-darwin-arm64.tar.gz ${BINARY_NAME}-darwin-arm64 README.md LICENSE docs/
	zip -r dist/${BINARY_NAME}-${VERSION}-windows-amd64.zip ${BINARY_NAME}-windows-amd64.exe README.md LICENSE docs/

# Development setup
.PHONY: dev-setup
dev-setup: deps
	@echo "Setting up development environment..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Quick test build
.PHONY: quick-test
quick-test: build
	@echo "Testing basic functionality..."
	./${BINARY_NAME} --version
	./${BINARY_NAME} --help | head -15
	@echo "Quick test completed!"

# Development test (replaces old test scripts)
.PHONY: dev-test
dev-test: build test
	@echo "Running development tests..."
	./${BINARY_NAME} --version
	./${BINARY_NAME} -v
	./${BINARY_NAME} version
	@echo "Testing error handling..."
	./${BINARY_NAME} 2>&1 | head -3 || echo "Error handling works"
	@echo "Development test completed!"

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build for current platform"
	@echo "  build-linux    - Build for Linux AMD64"
	@echo "  build-all      - Build for all platforms"
	@echo "  test           - Run tests"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  clean          - Clean build artifacts"
	@echo "  install        - Install to system"
	@echo "  package        - Create release packages"
	@echo "  help           - Show this help"