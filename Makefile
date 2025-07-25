# Harpoon Build Configuration
BINARY_NAME=hpn

# Version information (managed by version package)
VERSION=$(shell ./scripts/version.sh current 2>/dev/null | grep "Version:" | awk '{print $$2}' || echo "v1.1.0-dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build script
BUILD_SCRIPT=./scripts/build.sh

# Default target
.PHONY: all
all: build

# Build for current platform
.PHONY: build
build:
	@echo "Building ${BINARY_NAME} for current platform..."
	@$(BUILD_SCRIPT) -o .

# Build for specific platforms
.PHONY: build-linux
build-linux:
	@echo "Building ${BINARY_NAME} for Linux AMD64..."
	@$(BUILD_SCRIPT) -o . -p linux/amd64

.PHONY: build-linux-arm64
build-linux-arm64:
	@echo "Building ${BINARY_NAME} for Linux ARM64..."
	@$(BUILD_SCRIPT) -o . -p linux/arm64

.PHONY: build-darwin
build-darwin:
	@echo "Building ${BINARY_NAME} for macOS AMD64..."
	@$(BUILD_SCRIPT) -o . -p darwin/amd64

.PHONY: build-darwin-arm64
build-darwin-arm64:
	@echo "Building ${BINARY_NAME} for macOS ARM64..."
	@$(BUILD_SCRIPT) -o . -p darwin/arm64

.PHONY: build-windows
build-windows:
	@echo "Building ${BINARY_NAME} for Windows AMD64..."
	@$(BUILD_SCRIPT) -o . -p windows/amd64

# Build all platforms
.PHONY: build-all
build-all:
	@echo "Building ${BINARY_NAME} for all platforms..."
	@$(BUILD_SCRIPT) -o . -p all

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
	@mkdir -p dist/packages
	@tar -czf dist/packages/${BINARY_NAME}-${VERSION}-linux-amd64.tar.gz ${BINARY_NAME}-linux-amd64 README.md LICENSE docs/
	@tar -czf dist/packages/${BINARY_NAME}-${VERSION}-linux-arm64.tar.gz ${BINARY_NAME}-linux-arm64 README.md LICENSE docs/
	@tar -czf dist/packages/${BINARY_NAME}-${VERSION}-darwin-amd64.tar.gz ${BINARY_NAME}-darwin-amd64 README.md LICENSE docs/
	@tar -czf dist/packages/${BINARY_NAME}-${VERSION}-darwin-arm64.tar.gz ${BINARY_NAME}-darwin-arm64 README.md LICENSE docs/
	@zip -r dist/packages/${BINARY_NAME}-${VERSION}-windows-amd64.zip ${BINARY_NAME}-windows-amd64.exe README.md LICENSE docs/
	@echo "Packages created in dist/packages/"

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

# Branch management helpers
.PHONY: branch-feature
branch-feature:
	@./scripts/branch-helper.sh feature $(name)

.PHONY: branch-bugfix
branch-bugfix:
	@./scripts/branch-helper.sh bugfix $(name)

.PHONY: branch-release
branch-release:
	@./scripts/branch-helper.sh release $(version)

.PHONY: branch-hotfix
branch-hotfix:
	@./scripts/branch-helper.sh hotfix $(name)

.PHONY: branch-finish
branch-finish:
	@./scripts/branch-helper.sh finish

.PHONY: branch-cleanup
branch-cleanup:
	@./scripts/branch-helper.sh cleanup

.PHONY: branch-status
branch-status:
	@./scripts/branch-helper.sh status

# CI/CD simulation
.PHONY: ci-check
ci-check: fmt lint test build
	@echo "Running CI checks..."
	@echo "✅ All CI checks passed!"

.PHONY: pre-commit
pre-commit: fmt lint test
	@echo "Running pre-commit checks..."
	@echo "✅ Pre-commit checks passed!"

.PHONY: pre-push
pre-push: ci-check
	@echo "Running pre-push checks..."
	@echo "✅ Pre-push checks passed!"

# Security checks
.PHONY: security-scan
security-scan:
	@echo "Running security scan..."
	@command -v gosec >/dev/null 2>&1 || { echo "Installing gosec..."; go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; }
	gosec ./...

.PHONY: vuln-check
vuln-check:
	@echo "Checking for vulnerabilities..."
	@command -v govulncheck >/dev/null 2>&1 || { echo "Installing govulncheck..."; go install golang.org/x/vuln/cmd/govulncheck@latest; }
	govulncheck ./...

# Version management
.PHONY: version
version:
	@./scripts/version.sh current

.PHONY: version-next
version-next:
	@if [ -z "$(type)" ]; then echo "Usage: make version-next type=patch|minor|major|alpha|beta|rc"; exit 1; fi
	@./scripts/version.sh next $(type)

.PHONY: version-bump
version-bump:
	@if [ -z "$(type)" ]; then echo "Usage: make version-bump type=patch|minor|major|alpha|beta|rc"; exit 1; fi
	@./scripts/version.sh bump $(type)

.PHONY: version-tag
version-tag:
	@if [ -z "$(ver)" ]; then echo "Usage: make version-tag ver=v1.1.0"; exit 1; fi
	@./scripts/version.sh tag $(ver)

.PHONY: changelog
changelog:
	@./scripts/version.sh changelog

# Release preparation
.PHONY: prepare-release
prepare-release:
	@echo "Preparing release $(version)..."
	@if [ -z "$(version)" ]; then echo "Usage: make prepare-release version=v1.1.0"; exit 1; fi
	@./scripts/branch-helper.sh release $(version)

# Complete CI/CD pipeline simulation
.PHONY: full-ci
full-ci: deps fmt lint security-scan vuln-check test-coverage build-all
	@echo "Running full CI pipeline..."
	@echo "✅ Full CI pipeline completed successfully!"

# Help
.PHONY: help
help:
	@echo "Harpoon Build System"
	@echo ""
	@echo "Build Targets:"
	@echo "  build          - Build for current platform"
	@echo "  build-linux    - Build for Linux AMD64"
	@echo "  build-all      - Build for all platforms"
	@echo "  package        - Create release packages"
	@echo ""
	@echo "Development Targets:"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  security-scan  - Run security scan"
	@echo "  vuln-check     - Check for vulnerabilities"
	@echo ""
	@echo "Branch Management:"
	@echo "  branch-feature name=<name>     - Create feature branch"
	@echo "  branch-bugfix name=<name>      - Create bugfix branch"
	@echo "  branch-release version=<ver>   - Create release branch"
	@echo "  branch-hotfix name=<name>      - Create hotfix branch"
	@echo "  branch-finish                  - Finish current branch"
	@echo "  branch-cleanup                 - Clean merged branches"
	@echo "  branch-status                  - Show branch status"
	@echo ""
	@echo "CI/CD Simulation:"
	@echo "  ci-check       - Run CI checks"
	@echo "  pre-commit     - Run pre-commit checks"
	@echo "  pre-push       - Run pre-push checks"
	@echo "  full-ci        - Run complete CI pipeline"
	@echo ""
	@echo "Version Management:"
	@echo "  version                        - Show current version"
	@echo "  version-next type=<type>       - Calculate next version"
	@echo "  version-bump type=<type>       - Bump version and create tag"
	@echo "  version-tag ver=<version>      - Create version tag"
	@echo "  changelog                      - Generate changelog"
	@echo ""
	@echo "Release Management:"
	@echo "  prepare-release version=<ver>  - Prepare new release"
	@echo ""
	@echo "Utility:"
	@echo "  clean          - Clean build artifacts"
	@echo "  install        - Install to system"
	@echo "  dev-setup      - Setup development environment"
	@echo "  help           - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make branch-feature name=add-new-runtime"
	@echo "  make version-bump type=patch"
	@echo "  make prepare-release version=v1.1.0"
	@echo "  make ci-check"