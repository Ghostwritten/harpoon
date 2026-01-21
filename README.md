# Harpoon (hpn) 🎯

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![Version](https://img.shields.io/badge/version-v1.1-green.svg)](https://github.com/your-org/harpoon/releases)
[![Build Status](https://github.com/your-org/harpoon/workflows/Enhanced%20Testing/badge.svg)](https://github.com/your-org/harpoon/actions)

**Harpoon** is a modern, efficient container image management CLI tool written in Go. It provides powerful operations for pulling, saving, loading, and pushing container images with support for multiple container runtimes and flexible operation modes.

## ✨ Features

- **Multi-Runtime Support**: Docker, Podman, Nerdctl, Skopeo with automatic detection
- **Smart Runtime Fallback**: Automatic fallback when preferred runtime unavailable
- **Flexible Operation Modes**: Multiple modes for different deployment scenarios
- **Cross-Platform**: Linux, macOS, Windows support (AMD64, ARM64)
- **Wide Compatibility**: Statically linked Linux binaries compatible with RHEL 8.x, RHEL 9.x, Ubuntu, Debian, Alpine, and more
- **Configuration Management**: YAML-based config with environment variables
- **Batch Operations**: Efficient bulk image processing
- **Enterprise Ready**: Proxy support, unified authentication, private registries
- **Secure Authentication**: Interactive password input, stdin support, environment variables

## 🚀 Quick Start

### Installation

```bash
# Download and install (Linux/macOS)
curl -L https://github.com/your-org/harpoon/releases/latest/download/hpn-linux-amd64 -o hpn
chmod +x hpn
sudo mv hpn /usr/local/bin/

# Verify installation
hpn --version
```

For detailed installation instructions, see the [Installation Guide](docs/installation.md).

### System Requirements

**Linux Binaries:**
- **Statically linked** - No glibc version dependencies
- Compatible with:
  - RHEL 8.x (glibc 2.28+)
  - RHEL 9.x (glibc 2.34+)
  - Ubuntu 18.04+ (glibc 2.27+)
  - Debian 10+ (glibc 2.28+)
  - Alpine Linux (musl)
  - CentOS 8/Stream
  - Other modern Linux distributions

**macOS:** macOS 10.14+ (Mojave or later)

**Windows:** Windows 10 or later

### Basic Usage

```bash
# Create image list
echo "nginx:latest" > images.txt
echo "alpine:3.18" >> images.txt

# Login to registry (interactive, most secure)
hpn login harbor.company.com

# Login with credentials (for CI/CD, use --password-stdin)
hpn login harbor.company.com -u admin -p password
echo "password" | hpn login harbor.company.com -u admin --password-stdin

# Login to insecure registry (HTTP or self-signed certificate)
hpn login http://registry.local -u admin -p password --insecure

# Pull images
hpn pull -f images.txt

# Save to tar files
hpn save -f images.txt --path ./images

# Load from tar files
hpn load --path ./images

# Push to registry (preserve original paths)
hpn push -f images.txt --registry harbor.company.com

# Push to registry with unified project
hpn push -f images.txt --registry harbor.company.com --project production
```

## 📖 Documentation

### Essential Guides
- [📚 Quick Start Guide](docs/quickstart.md) - Get up and running in minutes
- [⚙️ Installation Guide](docs/installation.md) - Detailed installation instructions
- [📖 User Guide](docs/user-guide.md) - Complete usage guide
- [🔧 Configuration Guide](docs/configuration.md) - Configuration options and examples

### Advanced Topics
- [🏗️ Architecture](docs/architecture.md) - System architecture and design
- [🐳 Runtime Support](docs/runtime-support.md) - Container runtime compatibility
- [🔒 Security Guide](docs/security.md) - Security best practices
- [🔨 Building Guide](docs/building.md) - Build from source and cross-compilation
- [🛠️ Development Guide](docs/development.md) - Contributing and development

### Reference & Support
- [📋 API Reference](docs/api-reference.md) - Command-line interface reference
- [💡 Examples](docs/examples.md) - Real-world usage examples
- [❓ FAQ](docs/faq.md) - Frequently asked questions
- [🔍 Troubleshooting](docs/troubleshooting.md) - Common issues and solutions

### Release Information
- [📝 Changelog](docs/changelog.md) - Version history and changes
- [🚀 Release Notes](docs/release-notes.md) - Latest release information
- [⬆️ Upgrade Guide](docs/upgrade-guide.md) - Version upgrade instructions

## 🎯 Key Features

### Smart Runtime Management
```bash
# Auto-detect available runtime
hpn pull -f images.txt

# Specify runtime explicitly
hpn --runtime podman pull -f images.txt

# Auto-fallback for CI/CD
hpn --auto-fallback pull -f images.txt
```

### Registry Authentication
```bash
# Interactive login (most secure, password hidden)
hpn login harbor.company.com

# Login with parameters
hpn login harbor.company.com -u admin -p password

# Login from stdin (for CI/CD scripts)
echo "password" | hpn login harbor.company.com -u admin --password-stdin

# Login using environment variables
export REGISTRY_USERNAME=admin
export REGISTRY_PASSWORD=password
hpn login harbor.company.com

# Login to insecure registry (HTTP or self-signed certificate)
hpn login http://registry.local -u admin -p password --insecure

# Specify runtime for login
hpn login harbor.company.com -u admin -p password --runtime podman
```

### Flexible Push Options
```bash
# Preserve original paths: registry/project/image:tag
hpn push -f images.txt --registry harbor.com

# Unified project: all images to registry/project/image:tag
hpn push -f images.txt --registry harbor.com --project production

# Append path: registry/path/xx/project/image:tag
hpn push -f images.txt --registry harbor.com/path/xx
```

### Configuration
```bash
# Create config directory
mkdir -p ~/.hpn

# Basic configuration
cat > ~/.hpn/config.yaml << EOF
registry: harbor.company.com
project: production
runtime:
  preferred: docker
  auto_fallback: true
paths:
  save_path: ./images
  load_path: ./images
EOF
```

## 🔨 Development

### Building from Source
```bash
# Clone repository
git clone https://github.com/your-org/harpoon.git
cd harpoon

# Build for current platform
./build.sh current

# Build for all platforms
./build.sh all

# Run tests
go test ./...
```

For detailed build instructions, see the [Building Guide](docs/building.md).

### Contributing
We welcome contributions! Please see our [Development Guide](docs/development.md) for:
- Setting up development environment
- Building and testing
- Submitting pull requests
- Code style guidelines

## 💼 Use Cases

- **Kubernetes Deployments**: Pre-pull and manage cluster images
- **Air-Gapped Environments**: Offline image distribution
- **Registry Migration**: Move images between registries
- **CI/CD Pipelines**: Automated image operations
- **Development Workflows**: Local image management

## 🆕 What's New in v1.1

- **Enhanced Runtime Support**: New `--runtime` parameter and smart detection
- **Simplified Push Modes**: Removed redundant mode, improved smart project selection
- **Better User Experience**: Concise error messages, parameter validation
- **Auto-fallback**: Automatic runtime fallback for CI environments
- **Improved Compatibility**: Statically linked Linux binaries for RHEL 8.x+ compatibility

See the [Changelog](docs/changelog.md) for complete details.

## 🤝 Community & Support

- **Documentation**: [docs/](docs/) - Comprehensive guides and references
- **Issues**: [GitHub Issues](https://github.com/your-org/harpoon/issues) - Bug reports and feature requests
- **Discussions**: [GitHub Discussions](https://github.com/your-org/harpoon/discussions) - Community discussions
- **Contributing**: [Development Guide](docs/development.md) - How to contribute

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Harpoon** - Modern container image management with precision and efficiency 🎯