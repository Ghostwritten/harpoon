# Building Guide

This document describes how to build Harpoon (hpn) from source code on different platforms.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Build Scripts](#build-scripts)
- [Build Modes](#build-modes)
- [Version Information](#version-information)
- [Cross-Compilation](#cross-compilation)
- [Build Flags](#build-flags)
- [CI/CD Build Process](#cicd-build-process)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)

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

### Script Location

```
harpoon/
├── build.sh          # Main build script (in root directory)
├── Makefile          # Makefile wrapper for build.sh
├── dist/             # Build output directory (created automatically)
└── cmd/hpn/          # Source code
```

### Using Makefile (Recommended)

The `Makefile` provides a convenient interface to the build script:

```bash
make build        # Build for current platform
make build-all    # Build for all platforms
make clean        # Clean build artifacts
make test         # Run tests
make install      # Install to /usr/local/bin (requires sudo)
make help         # Show available targets
```

### Using build.sh Directly

You can also use the build script directly:

```bash
./build.sh [mode]
```

**Available modes:**

| Mode | Description |
|------|-------------|
| `current` (default) | Build for current platform only |
| `all` | Build for all supported platforms |
| `clean` | Remove all build artifacts |

**Note**: All binaries are output to the `dist/` directory.

## Build Modes

### Mode 1: Current Platform Build

Builds a binary optimized for your current development machine.

**Command:**
```bash
./build.sh current
```

**Process:**
1. Extracts version information from Git
2. Sets build flags with version metadata
3. Compiles for current OS/architecture
4. Outputs: `dist/hpn` binary

**Example Output:**

```
Building for current platform...
Version: v1.1, Commit: 6c1201a
✅ Built hpn
```

**Use Cases:**
- Local development and testing
- Quick iteration during development
- Testing new features

### Mode 2: Multi-Platform Build

Builds binaries for all supported platforms using cross-compilation.

**Command:**
```bash
./build.sh all
```

**Process:**
1. Builds Linux (amd64) - Static linking with `CGO_ENABLED=0`
2. Builds Linux (arm64) - Static linking with `CGO_ENABLED=0`
3. Builds macOS (amd64) - Intel Mac
4. Builds macOS (arm64) - Apple Silicon Mac
5. Builds Windows (amd64)

**Example Output:**
```
Building for all platforms...
Version: v1.1, Commit: 6c1201a
Building Linux (amd64) with CGO_ENABLED=0...
Building Linux (arm64) with CGO_ENABLED=0...
Building macOS (amd64)...
Building macOS (arm64)...
Building Windows (amd64)...
✅ Built all platforms
```

**Use Cases:**
- Preparing release binaries
- Testing cross-platform compatibility
- Distributing to multiple platforms

### Mode 3: Clean Build Artifacts

Removes all generated binary files.

**Command:**
```bash
./build.sh clean
```

**Output:**
```
✅ Cleaned
```

## Version Information

The build script automatically injects version information into the binary.

### Version Sources

1. **Git Tag**: If a tag exists (e.g., `v1.1`), it's used as the version
2. **Fallback**: If no tag exists, version is set to `dev`

### Injected Metadata

The following information is embedded in the binary:

| Field | Source | Example |
|-------|--------|---------|
| `Version` | Git tag or "dev" | `v1.1` or `dev` |
| `GitCommit` | Git commit short hash | `6c1201a` |
| `BuildDate` | UTC timestamp | `2024-12-19T10:30:00Z` |

### Viewing Version Information

```bash
# After building
./hpn --version

# Or use the version subcommand
./hpn version
```

## Cross-Compilation

Go natively supports cross-compilation, allowing you to build for different platforms from a single machine.

### Linux Builds (RHEL 8.x Compatible)

Linux binaries are built with special flags to ensure compatibility with older systems:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo,osusergo -ldflags "${LDFLAGS}" -o dist/hpn-linux-amd64 ./cmd/hpn
```

**Key Features:**
- `CGO_ENABLED=0`: Disables CGO for static linking
- `-tags netgo,osusergo`: Forces pure Go implementations
- **Result**: Statically linked binary compatible with RHEL 8.x+

### Platform-Specific Builds

#### macOS (Apple Silicon)
```bash
GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o dist/hpn-darwin-arm64 ./cmd/hpn
```

#### macOS (Intel)
```bash
GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o dist/hpn-darwin-amd64 ./cmd/hpn
```

#### Windows
```bash
GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o dist/hpn-windows-amd64.exe ./cmd/hpn
```

## Build Flags

### Linker Flags (LDFLAGS)

The build script uses the following linker flags:

```bash
LDFLAGS="-X github.com/harpoon/hpn/internal/version.Version=${VERSION}"
LDFLAGS="${LDFLAGS} -X github.com/harpoon/hpn/internal/version.GitCommit=${COMMIT}"
LDFLAGS="${LDFLAGS} -X github.com/harpoon/hpn/internal/version.BuildDate=${BUILD_DATE}"
```

**In Release Builds:**
```bash
LDFLAGS="-s -w"  # Strip symbols and debug info
```

### Build Tags

**For Linux builds:**
- `netgo`: Use pure Go DNS resolver (avoids CGO)
- `osusergo`: Use pure Go user/group lookup (avoids CGO)

**Usage:**
```bash
go build -tags netgo,osusergo ...
```

### Environment Variables

| Variable | Value | Purpose |
|----------|-------|---------|
| `CGO_ENABLED` | `0` | Disable CGO for static linking (Linux only) |
| `GOOS` | `linux`, `darwin`, `windows` | Target operating system |
| `GOARCH` | `amd64`, `arm64` | Target architecture |

## CI/CD Build Process

The project uses GitHub Actions for automated builds and releases.

### Release Workflow

**Trigger:** Push a tag matching `v*` pattern

**Location:** `.github/workflows/release.yml`

**Process:**
1. **Setup**: Checkout code and setup Go 1.21
2. **Test**: Run `go test -v ./...`
3. **Build**: Build binaries for all platforms
4. **Verify**: Check Linux binaries are statically linked
5. **Checksums**: Generate SHA256 checksums
6. **Release**: Create GitHub Release with binaries

### Build Verification

The CI process includes verification steps:

```bash
# Check if binary is statically linked
file dist/hpn-linux-amd64
# Expected: "statically linked"

# Verify no dynamic library dependencies
ldd dist/hpn-linux-amd64
# Expected: "not a dynamic executable"

# Test binary functionality
./dist/hpn-linux-amd64 --version
```

## Troubleshooting

### Common Issues

#### Issue: Go version mismatch

**Error:**
```
go: requires go >= 1.21
```

**Solution:**
```bash
# Update Go to version 1.21 or higher
brew upgrade go  # macOS with Homebrew
# Or download from https://golang.org/dl/
```

#### Issue: Permission denied

**Error:**
```
bash: ./build.sh: Permission denied
```

**Solution:**
```bash
chmod +x build.sh
```

#### Issue: Git not found

**Error:**
```
git: command not found
```

**Solution:**
- Install Git from https://git-scm.com/
- Or use Homebrew: `brew install git`

#### Issue: Build fails on Linux cross-compilation

**Error:**
```
#cgo CFLAGS: -g -O2
```

**Solution:**
This is normal. The build script sets `CGO_ENABLED=0` for Linux builds. If you see this error, ensure the build script is using the correct environment variables.

### Debugging Build Issues

**Enable verbose output:**
```bash
go build -v ./cmd/hpn
```

**Check build flags:**
```bash
go build -x ./cmd/hpn  # Shows all commands executed
```

**Verify dependencies:**
```bash
go mod verify
```

## Best Practices

### Development Workflow

1. **Local Development:**
   ```bash
   # Build for current platform
   ./build.sh current
   
   # Test the binary
   ./hpn --version
   ```

2. **Before Committing:**
   ```bash
   # Run tests
   go test ./...
   
   # Build to ensure it compiles
   ./build.sh current
   ```

3. **Before Release:**
   ```bash
   # Build all platforms
   ./build.sh all
   
   # Verify binaries
   ls -lh dist/hpn-*
   
   # Test on target platform if possible
   ```

### Build Optimization

**For Smaller Binaries:**
- Use `-ldflags "-s -w"` to strip symbols and debug info
- Already included in release builds

**For Faster Builds:**
- Use `-ldflags "-s"` (strip symbols only)
- Skip debug info stripping during development

### Version Management

**Tagging for Release:**
```bash
# Create a version tag
git tag v1.2.0

# Push the tag (triggers release workflow)
git push origin v1.2.0
```

**Semantic Versioning:**
- `v1.0.0` - Major version (breaking changes)
- `v1.1.0` - Minor version (new features)
- `v1.0.1` - Patch version (bug fixes)

### Platform-Specific Notes

#### macOS Development

**Apple Silicon (M1/M2/M3):**
- Native builds are ARM64
- Can cross-compile to Intel (amd64)
- Build time: ~5-10 seconds per platform

**Intel Mac:**
- Native builds are AMD64
- Can cross-compile to ARM64
- Build time: ~5-10 seconds per platform

#### Linux Development

- Can build for all platforms
- Linux builds use static linking by default
- Compatible with RHEL 8.x+ systems

#### Windows Development

- Requires Git Bash or WSL for build script
- Can cross-compile to all platforms
- Native builds produce `.exe` files

## Additional Resources

- [Go Build Documentation](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies)
- [Cross-Compilation Guide](https://golang.org/doc/install/source#environment)
- [Version Information](https://golang.org/cmd/link/)

---

**Last Updated:** 2024-12-19  
**Maintained by:** Harpoon Project
