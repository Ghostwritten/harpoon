# Harpoon (hpn) 🎯

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![Version](https://img.shields.io/badge/version-v1.1-green.svg)](https://github.com/your-org/harpoon/releases)
[![Build Status](https://github.com/your-org/harpoon/workflows/Enhanced%20Testing/badge.svg)](https://github.com/your-org/harpoon/actions)

**Harpoon** is a modern, efficient container image management CLI tool written in Go. It provides powerful operations for pulling, saving, loading, and pushing container images with support for multiple container runtimes and flexible operation modes.

## ✨ Features

- **Multi-Runtime Support**: Docker, Podman, Nerdctl with automatic detection
- **Smart Runtime Fallback**: Automatic fallback when preferred runtime unavailable
- **Flexible Operation Modes**: Multiple modes for different deployment scenarios
- **Cross-Platform**: Linux, macOS, Windows support (AMD64, ARM64)
- **Configuration Management**: YAML-based config with environment variables
- **Batch Operations**: Efficient bulk image processing
- **Enterprise Ready**: Proxy support, authentication, private registries

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

### Basic Usage

```bash
# Create image list
echo "nginx:latest" > images.txt
echo "alpine:3.18" >> images.txt

# Pull images
hpn -a pull -f images.txt

# Save to tar files
hpn -a save -f images.txt --save-mode 2

# Load from tar files
hpn -a load --load-mode 2

# Push to registry with smart project selection
hpn -a push -f images.txt -r harbor.company.com --push-mode 2
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
hpn -a pull -f images.txt

# Specify runtime explicitly
hpn --runtime podman -a pull -f images.txt

# Auto-fallback for CI/CD
hpn --auto-fallback -a pull -f images.txt
```

### Flexible Push Modes
```bash
# Simple push: registry/image:tag
hpn -a push -f images.txt -r harbor.com --push-mode 1

# Smart project selection: registry/project/image:tag
hpn -a push -f images.txt -r harbor.com -p production --push-mode 2
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
modes:
  push_mode: 2
EOF
```

## 🔨 Development

### Building from Source
```bash
# Clone repository
git clone https://github.com/your-org/harpoon.git
cd harpoon

# Build for current platform
make build

# Build for all platforms
make build-all

# Run development tests
make dev-test
```

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