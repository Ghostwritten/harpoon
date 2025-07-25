# Changelog

All notable changes to Harpoon (hpn) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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