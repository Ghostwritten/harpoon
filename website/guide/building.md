# Building Guide

This document describes how to build Harpoon (hpn) from source code on different platforms.

## Prerequisites

### Required Tools

- **Go**: Version 1.21 or higher
- **Git**: For version information extraction
- **Bash**: For running build scripts (Unix-like systems)

### Verify Installation

```bash
# Check Go version
go version
# Output should show: go version go1.21.x or higher

# Check Git
git --version

# Check Bash (macOS/Linux)
bash --version
```

### Download Dependencies

Before building, ensure all dependencies are downloaded:

```bash
go mod download
```

## Quick Start

### Build for Current Platform

The simplest way to build hpn for your current platform:

```bash
./build.sh current
```

This will create a binary named `hpn` in the `dist/` directory.

### Build for All Platforms

To build binaries for all supported platforms:

```bash
./build.sh all
```

This will create multiple binaries in the `dist/` directory:
- `dist/hpn-linux-amd64` - Linux x86_64
- `dist/hpn-linux-arm64` - Linux ARM64
- `dist/hpn-darwin-amd64` - macOS Intel
- `dist/hpn-darwin-arm64` - macOS Apple Silicon
- `dist/hpn-windows-amd64.exe` - Windows x86_64

## Build Scripts

The project provides a `build.sh` script and a `Makefile` wrapper that automate the build process.

### Using Makefile (Recommended)

```bash
make build        # Build for current platform
make build-all    # Build for all platforms
make clean        # Clean build artifacts
make test         # Run tests
make install      # Install to /usr/local/bin (requires sudo)
make help         # Show available targets
```

### Using build.sh Directly

```bash
./build.sh [mode]
```

| Mode | Description |
|------|-------------|
| `current` (default) | Build for current platform only |
| `all` | Build for all supported platforms |
| `clean` | Remove all build artifacts |

## Version Information

The build script automatically injects version information into the binary.

### Version Sources

1. **Git Tag**: If a tag exists (e.g., `v1.1`), it's used as the version
2. **Fallback**: If no tag exists, version is set to `dev`

### Viewing Version Information

```bash
./hpn --version
./hpn version
```

## Cross-Compilation

Go natively supports cross-compilation. Linux binaries are built with static linking for RHEL 8.x+ compatibility:

- `CGO_ENABLED=0` for static linking
- `-tags netgo,osusergo` for pure Go implementations

## CI/CD Build Process

**Trigger:** Push a tag matching `v*` pattern

**Location:** `.github/workflows/release.yml`

The release workflow: tests, builds for all platforms, generates checksums, and creates GitHub Release.

## Troubleshooting

### Common Issues

**Go version mismatch:** Update Go to 1.21 or higher.

**Permission denied:** `chmod +x build.sh`

**Git not found:** Install Git from https://git-scm.com/

### Debugging

```bash
go build -v ./cmd/hpn
go build -x ./cmd/hpn  # Shows all commands
go mod verify
```

## See Also

- [Installation](/guide/installation) - Download binaries
- [Quick Start](/guide/quickstart) - Get started
- [GitHub Releases](https://github.com/Ghostwritten/harpoon/releases)
