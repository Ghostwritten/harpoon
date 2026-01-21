.PHONY: build build-all build-current clean test help install

# Default target
help:
	@echo "Harpoon (hpn) Build System"
	@echo "=========================="
	@echo ""
	@echo "Available targets:"
	@echo "  make build        - Build for current platform"
	@echo "  make build-all    - Build for all platforms"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make test         - Run tests"
	@echo "  make install      - Install to /usr/local/bin (requires sudo)"
	@echo ""
	@echo "Note: All binaries are output to the dist/ directory"

build: build-current

build-current:
	@./build.sh current

build-all:
	@./build.sh all

clean:
	@./build.sh clean

test:
	@go test ./...

install: build-current
	@if [ -f "dist/hpn" ]; then \
		sudo cp dist/hpn /usr/local/bin/hpn && \
		sudo chmod +x /usr/local/bin/hpn && \
		echo "✅ Installed hpn to /usr/local/bin/hpn"; \
	else \
		echo "❌ Binary not found. Run 'make build' first."; \
		exit 1; \
	fi
