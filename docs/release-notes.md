# Release Notes

## Harpoon (hpn) v1.1 Release Notes

### Overview

Harpoon v1.1 is a significant feature enhancement release, primarily improving container runtime support, push mode logic, and user experience.

### New Features

#### 1. Enhanced Container Runtime Support
- **New `--runtime` parameter**: Manually specify container runtime (docker|podman|nerdctl)
- **Smart runtime detection**: Integrated complete runtime interfaces for more reliable detection
- **Auto-fallback mechanism**: Intelligently select available alternatives when configured runtime is unavailable
- **New `--auto-fallback` parameter**: Support automatic fallback in CI environments without user interaction

#### 2. Simplified Push Modes
- **Removed redundant Push Mode 3**: Simplified push modes to avoid feature duplication
- **Smart project name selection**: Push Mode 2 now supports intelligent project name selection
- **Priority mechanism**: Command line > config file > original image project name

#### 3. Improved User Experience
- **Concise error messages**: Removed verbose help information, showing only relevant errors
- **Strict parameter validation**: Prevents using unrelated mode parameters in wrong operations
- **Internationalization improvements**: Help text unified to English

### Feature Improvements

#### Push Mode Refactoring
```bash
# Mode 1: Simple push (no project name)
hpn -a push -f images.txt -r localhost:5001 --push-mode 1
# Result: localhost:5001/image:tag

# Mode 2: Smart project push
hpn -a push -f images.txt -r localhost:5001 --push-mode 2
# Result: localhost:5001/project/image:tag
```

#### Smart Project Name Selection Logic
1. **Command line specified project**: `-p myproject` → uses `myproject`
2. **Config file specified project**: `config.yaml: project: myproject` → uses `myproject`
3. **Original image project name**: `docker.io/ghostwritten/app:latest` → uses `ghostwritten`

#### Runtime Selection Examples
```bash
# Manually specify runtime
hpn --runtime podman -a pull -f images.txt

# Auto-fallback mode (suitable for CI environments)
hpn --auto-fallback -a pull -f images.txt

# Configuration file auto-fallback
# ~/.hpn/config.yaml
runtime:
  preferred: docker
  auto_fallback: true
```

### Technical Improvements

#### Architecture Optimization
- **Runtime interface integration**: Fully integrated `internal/runtime` package interfaces
- **Error handling refactoring**: Implemented better error classification and handling
- **Enhanced configuration management**: Support for more runtime configuration options

#### Code Quality
- **Type safety**: Used strongly-typed runtime interfaces instead of string operations
- **Error propagation**: Improved error information propagation and display
- **Parameter validation**: Added comprehensive command-line parameter validation

### Configuration Updates

#### New Configuration Options
```yaml
# ~/.hpn/config.yaml
runtime:
  preferred: docker          # Preferred runtime
  auto_fallback: false      # Whether to auto-fallback
  timeout: 5m               # Runtime operation timeout
  retry:
    max_attempts: 3
    delay: 1s
    max_delay: 30s

modes:
  push_mode: 1              # 1=simple push, 2=smart project push
  save_mode: 1
  load_mode: 1
```

### Breaking Changes

#### Push Mode Changes
- **Removed Push Mode 3**: Original `--push-mode 3` is no longer supported
- **Mode 2 behavior change**: Now uses smart project name selection instead of fixed specified project

#### Migration Guide
```bash
# v1.0 Push Mode 3
hpn -a push -f images.txt -r registry.com --push-mode 3

# v1.1 equivalent operation
hpn -a push -f images.txt -r registry.com --push-mode 2
# (Don't specify -p parameter, automatically uses original image project name)
```

### Bug Fixes

1. **Duplicate error messages**: Fixed issue where error messages displayed twice
2. **Parameter validation**: Fixed issue where cross-action mode parameter validation was not strict
3. **Runtime detection**: Improved reliability of runtime availability detection

### Usage Examples

#### Basic Operations
```bash
# Version information
hpn -v
hpn version

# Pull images
hpn -a pull -f images.txt

# Save images
hpn -a save -f images.txt --save-mode 2

# Load images
hpn -a load --load-mode 2

# Push images (smart project selection)
hpn -a push -f images.txt -r harbor.company.com --push-mode 2
```

#### Advanced Features
```bash
# Specify runtime
hpn --runtime podman -a pull -f images.txt

# Auto-fallback
hpn --auto-fallback -a pull -f images.txt

# Specify project push
hpn -a push -f images.txt -r registry.com -p production --push-mode 2
```

### Error Handling Examples

#### Parameter Validation
```bash
# Error: Using save-mode with push action
$ hpn -a push -f images.txt --save-mode 2
Error: --save-mode cannot be used with push action

# Error: Invalid push-mode value
$ hpn -a push -f images.txt --push-mode 3
Error: invalid push-mode '3'. Valid values: 1, 2
```

#### Runtime Handling
```bash
# Interactive prompt when configured runtime unavailable
Runtime 'docker' is not available
Found available runtime: podman
Use 'podman' instead of 'docker'? (y/N):
```

### Compatibility

#### Supported Platforms
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

#### Supported Container Runtimes
- Docker
- Podman
- Nerdctl

#### Go Version Requirements
- Go 1.21+

### Installation and Upgrade

#### Build from Source
```bash
git clone https://github.com/your-org/harpoon.git
cd harpoon
git checkout v1.1
go build -o hpn ./cmd/hpn
```

#### Configuration Migration
If you used Push Mode 3, update your scripts:
```bash
# Old version
hpn -a push -f images.txt --push-mode 3

# New version
hpn -a push -f images.txt --push-mode 2
```

### Acknowledgments

Thanks to all users for feedback and suggestions, especially regarding push mode simplification and runtime support improvements.

### Support

If you encounter issues or have suggestions:
1. Check documentation and examples
2. Submit GitHub Issue
3. Participate in community discussions

---

**Full Changelog**: [v1.0...v1.1](https://github.com/your-org/harpoon/compare/v1.0...v1.1)

## Previous Releases

### v1.0 - Initial Go Rewrite

#### Added
- Initial release of Harpoon Go rewrite
- Support for pull, save, load, push operations
- Multiple container runtime support (Docker, Podman, Nerdctl)
- Flexible operation modes for save, load, and push
- Configuration file support
- Cross-platform compatibility (Linux, macOS, Windows)
- Proxy support for corporate environments
- Parallel processing capabilities

#### Features
- **Pull**: Download images from registries
- **Save**: Export images to tar files with multiple modes
- **Load**: Import images from tar files with multiple modes  
- **Push**: Upload images to registries with multiple modes
- **Configuration**: YAML-based configuration with environment variable support
- **Logging**: Structured logging with multiple output formats

### See Also

- [Changelog](changelog.md) - Complete version history
- [Upgrade Guide](upgrade-guide.md) - Migration instructions
- [User Guide](user-guide.md) - Complete usage guide