# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.0.0] - 2025-01-23

### Added
- **Complete Go rewrite** of the original shell script with enhanced functionality
- **Multi-container runtime support**: Docker, Podman, and Nerdctl compatibility
- **Flexible operation modes**: Multiple modes for pull, save, load, and push operations
- **Configuration management**: YAML config files with environment variable overrides
- **Cross-platform support**: Native binaries for Linux (AMD64/ARM64), macOS (Intel/Apple Silicon), and Windows
- **Proxy support**: Built-in HTTP/HTTPS proxy configuration
- **Comprehensive logging**: Structured logging with multiple output formats
- **Batch operations**: Support for bulk image processing
- **Private registry support**: Complete private image registry push functionality

### Features
- **Pull operations**: Pull container images from external registries with proxy support
- **Save operations**: Save images to tar files with 3 different organization modes
- **Load operations**: Load images from tar files with flexible directory scanning
- **Push operations**: Push images to private registries with multiple tagging strategies
- **Configuration system**: Support for `~/.hpn/config.yaml` and environment variables
- **Error handling**: Comprehensive error messages with suggested solutions
- **Progress tracking**: Real-time progress reporting for all operations
- **Parallel processing**: Configurable concurrent operations for better performance

### Command Line Interface
- Maintains compatibility with original shell script interface
- Single command structure: `hpn -a <action> -f <file> [options]`
- Support for all original parameters and modes
- Enhanced help system with detailed usage examples

### Supported Platforms
- Linux AMD64 ✅
- Linux ARM64 ✅  
- macOS AMD64 (Intel) ✅
- macOS ARM64 (Apple Silicon) ✅
- Windows AMD64 ✅

### Documentation
- Complete English and Chinese documentation
- Comprehensive usage examples
- Configuration guide
- Troubleshooting section
- Migration guide from shell script

### Technical Details
- Built with Go 1.21+
- Uses Cobra for CLI framework
- Viper for configuration management
- Cross-platform binary distribution
- Optimized builds with version information

## [Unreleased]

### Planned Features
- Image signature verification
- OCI format support
- Progress bar display
- Incremental synchronization
- Performance optimizations