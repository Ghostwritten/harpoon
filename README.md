# Harpoon (hpn) 🎯

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org)
[![Version](https://img.shields.io/badge/version-v2.0.1-green.svg)](https://github.com/Ghostwritten/harpoon/releases)
[![Build Status](https://github.com/Ghostwritten/harpoon/workflows/Test/badge.svg)](https://github.com/Ghostwritten/harpoon/actions)

**Harpoon** is a modern, efficient container image management CLI tool written in Go. It provides powerful operations for pulling, saving, loading, and pushing container images with support for multiple container runtimes and flexible operation modes.

## ✨ Features

- **Multi-Runtime Support**: Docker, Podman, Nerdctl, Skopeo with automatic detection
- **Smart Runtime Fallback**: Automatic fallback when preferred runtime unavailable
- **Flexible Operation Modes**: Multiple modes for different deployment scenarios
- **Cross-Platform**: Linux, macOS, Windows support (AMD64, ARM64)
- **Wide Compatibility**: Statically linked Linux binaries compatible with RHEL 8.x, RHEL 9.x, Ubuntu, Debian, Alpine, and more
- **Configuration Management**: YAML-based config with environment variables
- **Batch Operations**: Efficient bulk image processing
- **Helm Chart Image List**: Extract image list from Helm charts (`list-images`) for offline or CI use
- **Enterprise Ready**: Proxy support, unified authentication, private registries
- **Secure Authentication**: Interactive password input, stdin support, environment variables

## 🚀 Quick Start

### Installation

```bash
# Download and install (Linux/macOS)
curl -L https://github.com/Ghostwritten/harpoon/releases/latest/download/hpn-linux-amd64 -o hpn
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

# Get image list from a Helm chart (requires Helm CLI), then pull
hpn list-images --chart bitnami/nginx --version 15.0.0 -o images.txt
hpn pull -f images.txt

# List images: runtime, path, or check file
hpn ls                              # List local runtime images
hpn ls --path ./images              # List saved tar files
hpn ls -f images.txt                # Check images.txt against runtime
```

## 📖 Documentation

### Essential Guides
- [📚 Quick Start Guide](docs/quickstart.md) - Get up and running in minutes
- [⚙️ Installation Guide](docs/installation.md) - Detailed installation instructions
- [💡 Examples](docs/examples.md) - Real-world usage examples

### Advanced Topics
- [🔨 Building Guide](docs/building.md) - Build from source and cross-compilation
- [🔄 Concurrency](docs/concurrency.md) - Workers and parallel operations
- [📦 Skopeo: Images](docs/skopeo-images.md) - Viewing and cleaning Skopeo images
- [📥 Skopeo: Pull vs Save](docs/skopeo-pull-vs-save.md) - Pull vs save behavior with Skopeo

### Release Information
- [📝 Changelog](docs/changelog.md) - Version history and changes

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

### Passthrough to Underlying Runtime
You can pass extra flags to the underlying runtime (docker/podman/nerdctl/skopeo) by putting them after `--`. Config file can also set `runtime.extra_args.pull`, `runtime.extra_args.save`, and `runtime.extra_args.push`.

```bash
# Example: skip TLS verify when pulling (podman/skopeo)
hpn pull -f images.txt -- --tls-verify=false

# Example: pass through to save/push
hpn save -f images.txt -- --tls-verify=false
hpn push -f images.txt --registry harbor.com -- --tls-verify=false
```

Interpretation of passthrough args is done by the underlying tool; see each runtime's documentation for valid options.

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
git clone https://github.com/Ghostwritten/harpoon.git
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
We welcome contributions! Please see the [Building Guide](docs/building.md) for building and testing, and the [Changelog](docs/changelog.md) for version history.

## 💼 Use Cases

- **Kubernetes Deployments**: Pre-pull and manage cluster images
- **Air-Gapped Environments**: Offline image distribution
- **Registry Migration**: Move images between registries
- **CI/CD Pipelines**: Automated image operations
- **Development Workflows**: Local image management

## 🆕 What's New in v2.0.1

- **list-images**: Extract image list from Helm charts (remote or local) for `hpn pull -f` / `hpn push -f`; requires [Helm](https://helm.sh) CLI
- **ls** (alias `list`): List runtime images, list tar files in `--path` with SIZE/CHECKSUM, or check `-f` list against runtime/path
- **rmi**: Remove images from a file in local runtime (Docker/Podman/Nerdctl); passthrough e.g. `-- -f` for force
- **Improved**: Chart image extraction filters out URLs, metric names, and socket paths to reduce false positives

See the [Changelog](docs/changelog.md) for complete details.

## 🤝 Community & Support

- **Documentation**: [docs/](docs/) - Guides and references
- **Issues**: [GitHub Issues](https://github.com/Ghostwritten/harpoon/issues) - Bug reports and feature requests
- **Contributing**: [Building Guide](docs/building.md) - Build and test; [Changelog](docs/changelog.md) - Version history

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Harpoon** - Modern container image management with precision and efficiency 🎯