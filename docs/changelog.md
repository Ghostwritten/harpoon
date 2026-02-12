# Changelog

All notable changes to Harpoon (hpn) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v2.0.1] - 2025-02-13

### Added

- **list-images**: New `hpn list-images` subcommand to extract container image lists from Helm charts (remote or local tgz/dir). Output is one image per line, compatible with `hpn pull -f` / `hpn push -f`. Requires [Helm](https://helm.sh) CLI; use `--chart`, `--version`, `-f/--values`, `-o/--output`, `--release-name`.
- **ls**: New `hpn ls` subcommand (alias `list`) with three modes: (1) `hpn ls` lists images in the local runtime (Docker/Podman/Nerdctl); (2) `hpn ls --path ./images` lists saved tar files in a directory with IMAGE, SIZE, CHECKSUM; (3) `hpn ls -f images.txt` checks whether images from a file exist in the runtime or in `--path`. Skopeo is not supported (no local image store).
- **rmi**: New `hpn rmi` subcommand to remove images from the local runtime using `-f <file>`. No interactive confirmation; supports passthrough args (e.g. `hpn rmi -f list.txt -- -f` for force). Supported runtimes: Docker, Podman, Nerdctl; Skopeo returns an error.
- **internal/chart**: New package for chart fetch, `helm template`, and YAML image extraction with filtering (embedded refs in args; excludes URLs, metric names, unix sockets).
- **Runtime interface**: New `RemoveImage(ctx, image, RmiOptions)` and `ListImages(ctx)` on `ContainerRuntime`; `ImageNameFromTarFilename` in `internal/runtime/checksum.go` for reversing tar filenames to image names.

### Improved

- **list-images filtering**: Image extraction now filters out non-image strings (URLs, Prometheus metric names, `unix://` / `tcp://` / `fd://` prefixes, port-only tags) to reduce false positives in chart image lists.

## [v2.0] - 2025-01-21

### 🎉 Major Release Highlights

This release introduces a major CLI architecture refactoring, adds unified registry authentication, integrates Skopeo support, and includes numerous improvements for better usability and maintainability.

### Added

#### Authentication & Security
- **Login Command**: New `hpn login` subcommand for unified registry authentication
  - Supports Docker, Podman, Skopeo, and Nerdctl runtimes
  - Interactive password input (most secure, hidden input)
  - Stdin password support for CI/CD pipelines (`--password-stdin`)
  - Environment variable support (`REGISTRY_USERNAME`, `REGISTRY_PASSWORD`)
  - Insecure registry support (`--insecure` flag for HTTP or self-signed certificates)
  - Automatic runtime detection or manual specification via `--runtime`
- **Authentication Utilities**: New `internal/runtime/auth.go` for managing authentication files
  - Support for Docker (`~/.docker/config.json`) and Podman/Skopeo (`~/.config/containers/auth.json`)
  - Automatic authentication file path detection based on runtime
  - Secure credential storage with proper file permissions (0600)

#### Runtime Support
- **Skopeo Integration**: Added Skopeo as a new container runtime option
  - Full support for pull, save, load, and push operations
  - Daemonless image operations (no Docker daemon required)
  - Multi-architecture image support
  - Automatic detection and fallback support

#### Developer Experience
- **Debug Mode**: Added `--debug` flag for detailed troubleshooting
  - Shows full stdout/stderr from underlying container commands
  - Enhanced error messages with detailed context
  - Useful for diagnosing runtime issues
- **Progress Bars**: Visual progress indicators for long-running operations
  - Real-time progress display for pull, save, load, and push operations
  - Multi-bar support for concurrent operations
  - Automatic terminal width detection and formatting

#### Build System
- **Makefile**: Added Makefile wrapper for build.sh
  - `make build` - Build for current platform
  - `make build-all` - Build for all platforms
  - `make clean` - Clean build artifacts
  - `make test` - Run tests
  - `make install` - Install to /usr/local/bin
- **Standardized Build Output**: All binaries now output to `dist/` directory
  - Follows industry-standard project structure
  - Cleaner project root directory
  - Better organization of build artifacts

### Changed

#### BREAKING: CLI Architecture Refactoring
- **Subcommand-based CLI**: Refactored from action-based to subcommand-based architecture
  - Removed `-a, --action` parameter
  - Commands now use subcommands: `hpn pull`, `hpn save`, `hpn load`, `hpn push`, `hpn login`
  - Follows industry standards (similar to docker, kubectl)
  - Better command discovery and help organization
- **Parameter Simplification**: Removed all `--*-mode` parameters
  - Removed `--save-mode`, `--load-mode`, `--push-mode`
  - Save/Load operations now use `--path` parameter (default: `./images`)
  - Save operation supports multi-level paths (e.g., `--path ./output/images/prod`)
  - Load operation supports `--recursive` flag for recursive loading
  - Push operation uses flexible naming with three intuitive modes:
    - Default: preserve original paths (`--registry new-registry.com`)
    - Append path: prepend path to original (`--registry new-registry.com/path/xx`)
    - Unified project: all images to single project (`--registry new-registry.com --project newproject`)
- **Parameter Scoping**: Command-specific parameters now scoped to their subcommands
  - Push-specific parameters (`--registry`, `--project`) only visible in `hpn push -h`
  - Cleaner root command help output
  - Better parameter organization

#### BREAKING: Configuration Structure
- **Configuration Schema Update**: Removed `modes` section, added `paths` section
  - Removed `modes.save_mode`, `modes.load_mode`, `modes.push_mode`
  - Added `paths.save_path` and `paths.load_path` fields
  - Default paths: `./images` for both save and load
  - More intuitive and flexible configuration

### Improved

- **CLI User Experience**: 
  - Clearer command structure with subcommands
  - More intuitive parameter organization
  - Enhanced help information with subcommand-specific parameters
  - Better error messages with actionable suggestions
- **Image Pull Strategy**:
  - Retry mechanism with exponential backoff
  - Configurable concurrency with worker pools
  - Platform parameter support (`--platform linux/amd64`, `linux/arm64`, `all`)
  - Timeout optimization for better reliability
- **File Operations**:
  - SHA256 checksum generation for saved tar files
  - Resume support: skip re-saving if valid tar and checksum files exist
  - Multi-level path support for better organization
- **Project Structure**:
  - Reorganized documentation (moved RELEASE.md, DOWNLOAD.md to docs/)
  - Removed unnecessary files (install.sh, demo/, benchmarks/)
  - Cleaner project root directory
  - Better separation of concerns

### Fixed

- Fixed empty `images` directory creation when custom `--path` is specified
- Fixed Skopeo save operation platform issues on macOS
- Improved configuration validation to prevent unnecessary directory creation
- Enhanced error handling for missing runtime dependencies

### Removed

- Removed `install.sh` script (users can download directly from GitHub Releases)
- Removed `demo/` directory (examples integrated into documentation)
- Removed `benchmarks/` directory (not needed for CLI tool)
- Removed `scripts/performance-analysis.sh` (depended on benchmarks)

### Migration Guide

#### Command Syntax Changes

**Pull Operations:**
```bash
# Old
hpn -a pull -f images.txt

# New
hpn pull -f images.txt
```

**Save Operations:**
```bash
# Old
hpn -a save -f images.txt --save-mode 2

# New
hpn save -f images.txt --path ./images
```

**Load Operations:**
```bash
# Old
hpn -a load --load-mode 2

# New
hpn load --path ./images
```

**Push Operations:**
```bash
# Old
hpn -a push -f images.txt -r registry.com --push-mode 2

# New
hpn push -f images.txt --registry registry.com --project project
```

#### Configuration File Changes

**Old config.yaml:**
```yaml
modes:
  save_mode: 2
  load_mode: 2
  push_mode: 2
```

**New config.yaml:**
```yaml
paths:
  save_path: ./images
  load_path: ./images
```

### Documentation

- Updated all documentation to reflect new CLI syntax
- Added comprehensive migration guide
- Updated examples and quick start guide
- Added login command documentation
- Updated building guide with Makefile instructions

## [v1.1] - 2024-12-19

### Added
- `--runtime` parameter to manually specify container runtime (docker|podman|nerdctl)
- `--auto-fallback` parameter for automatic runtime fallback in CI environments
- Smart runtime detection with fallback mechanism
- Interactive runtime selection when configured runtime is unavailable
- Strict parameter validation to prevent misuse of mode parameters

### Changed
- **BREAKING**: Removed Push Mode 3 (preserve original project path)
- **BREAKING**: Push Mode 2 now uses smart project name selection
- Project name selection priority: command line > config file > original image project
- Error messages are now concise without showing full help text
- Help text unified to English language

### Improved
- Enhanced container runtime support with full interface integration
- Better error handling and user experience
- More reliable runtime availability detection
- Cleaner command-line interface

### Fixed
- Duplicate error messages when validation fails
- Cross-action mode parameter validation
- Runtime detection reliability issues

### Technical
- Integrated `internal/runtime` package interfaces
- Replaced string-based runtime operations with type-safe interfaces
- Enhanced configuration management for runtime options
- Improved error propagation and display

## [v1.0] - 2024-12-01

### Added
- Initial release of Harpoon Go rewrite
- Support for pull, save, load, push operations
- Multiple container runtime support (Docker, Podman, Nerdctl)
- Flexible operation modes for save, load, and push
- Configuration file support
- Cross-platform compatibility (Linux, macOS, Windows)
- Proxy support for corporate environments
- Parallel processing capabilities

### Features
- **Pull**: Download images from registries
- **Save**: Export images to tar files with multiple modes
- **Load**: Import images from tar files with multiple modes  
- **Push**: Upload images to registries with multiple modes
- **Configuration**: YAML-based configuration with environment variable support
- **Logging**: Structured logging with multiple output formats