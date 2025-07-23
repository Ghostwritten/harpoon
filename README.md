# Harpoon 🎯

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.19+-blue.svg)](https://golang.org)
[![Version](https://img.shields.io/badge/version-v1.0-green.svg)](releases)

**Harpoon** is a powerful cloud-native container image management tool designed for Kubernetes environments and enterprise container workflows. It provides flexible image pulling, saving, loading, and pushing capabilities with multiple operation modes.

## 🌟 Features

- **Multi-Container Runtime Support**: Compatible with Docker, Podman, and Nerdctl
- **Flexible Operation Modes**: Each operation supports multiple modes for different scenarios
- **Cross-Platform**: Native Go binary for Linux, macOS, and Windows
- **Configuration Management**: YAML config files with environment variable overrides
- **Proxy Support**: Built-in HTTP/HTTPS proxy configuration
- **Batch Operations**: Support for bulk image processing
- **Private Registry Support**: Complete private image registry push functionality

## 🚀 Installation

### Quick Install (Recommended)

```bash
# Install latest version
curl -fsSL https://raw.githubusercontent.com/ghostwritten/harpoon/main/install.sh | bash

# Or with wget
wget -qO- https://raw.githubusercontent.com/ghostwritten/harpoon/main/install.sh | bash
```

### Download Binary

Choose your platform:

| Platform | Architecture | Download Command |
|----------|-------------|------------------|
| Linux | AMD64 | `wget https://github.com/ghostwritten/harpoon/releases/latest/download/hpn-linux-amd64` |
| Linux | ARM64 | `wget https://github.com/ghostwritten/harpoon/releases/latest/download/hpn-linux-arm64` |
| macOS | Intel | `wget https://github.com/ghostwritten/harpoon/releases/latest/download/hpn-darwin-amd64` |
| macOS | Apple Silicon | `wget https://github.com/ghostwritten/harpoon/releases/latest/download/hpn-darwin-arm64` |
| Windows | AMD64 | Download `hpn-windows-amd64.exe` from releases page |

After download:
```bash
chmod +x hpn-*
sudo mv hpn-* /usr/local/bin/hpn
```

### Build from Source

```bash
git clone https://github.com/ghostwritten/harpoon.git
cd harpoon
go build -o hpn ./cmd/hpn
```

## 🔧 Quick Start

### Basic Usage

```bash
# Create image list
echo "nginx:latest" > images.txt
echo "redis:alpine" >> images.txt

# Pull images
hpn -a pull -f images.txt

# Save to tar files
hpn -a save -f images.txt --save-mode 2

# Load from tar files
hpn -a load --load-mode 2

# Push to private registry
hpn -a push -f images.txt -r registry.example.com -p myproject --push-mode 2
```

## 📖 Command Reference

### Syntax
```bash
hpn -a <action> -f <image_list> [options]
```

### Actions
- `pull` - Pull images from external registry
- `save` - Save images into tar files
- `load` - Load images from tar files
- `push` - Push images to private registry

### Options
- `-a, --action` - Action (required)
- `-f, --file` - Image list file (required for pull/save/push)
- `-r, --registry` - Target registry (default from config)
- `-p, --project` - Target project namespace (default from config)
- `-c, --config` - Config file path

### Operation Modes

| Mode | Save | Load | Push |
|------|------|------|------|
| 1 | Current directory | Current directory | `registry/image:tag` |
| 2 | `./images/` | `./images/` | `registry/project/image:tag` |
| 3 | `./images/<project>/` | Recursive `./images/*/` | Preserve original path |

## ⚙️ Configuration

### Config File

Create `~/.hpn/config.yaml`:

```yaml
registry: registry.k8s.local
project: library
proxy:
  http: http://proxy.company.com:8080
  https: http://proxy.company.com:8080
  enabled: true
runtime:
  preferred: docker
  timeout: 5m
logging:
  level: info
  format: text
parallel:
  max_workers: 5
modes:
  save_mode: 2
  load_mode: 2
  push_mode: 2
```

### Environment Variables

```bash
export HPN_REGISTRY=registry.example.com
export HPN_PROJECT=myproject
export HPN_PROXY_HTTP=http://proxy.example.com:8080
```

## 🔨 Building

### Cross-Platform Build

```bash
# Current platform
go build -o hpn ./cmd/hpn

# Specific platforms
GOOS=linux GOARCH=amd64 go build -o hpn-linux-amd64 ./cmd/hpn
GOOS=linux GOARCH=arm64 go build -o hpn-linux-arm64 ./cmd/hpn
GOOS=darwin GOARCH=amd64 go build -o hpn-darwin-amd64 ./cmd/hpn
GOOS=darwin GOARCH=arm64 go build -o hpn-darwin-arm64 ./cmd/hpn
GOOS=windows GOARCH=amd64 go build -o hpn-windows-amd64.exe ./cmd/hpn

# Using Makefile
make build-all    # All platforms
make build-linux  # Linux AMD64

# Using build script
./build.sh all    # All platforms
./build.sh linux  # Linux AMD64
```

### Supported Platforms

| OS | Architecture | Status |
|---|---|---|
| Linux | AMD64 | ✅ Tested |
| Linux | ARM64 | ✅ Tested |
| macOS | AMD64 (Intel) | ✅ Tested |
| macOS | ARM64 (Apple Silicon) | ✅ Tested |
| Windows | AMD64 | ✅ Tested |

## 📋 Use Cases

### Kubernetes Cluster Setup
```bash
# Pull K8s system images
hpn -a pull -f k8s-system-images.txt
hpn -a save -f k8s-system-images.txt --save-mode 2
```

### Air-Gapped Deployment
```bash
# Prepare images for offline environment
hpn -a pull -f production-images.txt
hpn -a save -f production-images.txt --save-mode 2
tar -czf offline-images.tar.gz images/
```

### Private Registry Migration
```bash
# Migrate images to private registry
hpn -a pull -f public-images.txt
hpn -a push -f public-images.txt -r harbor.company.com -p production --push-mode 2
```

## 🐛 Troubleshooting

### Common Issues

1. **Binary won't execute**
   ```bash
   chmod +x hpn
   ```

2. **Container runtime not found**
   ```bash
   # Install Docker, Podman, or Nerdctl
   curl -fsSL https://get.docker.com | sh
   ```

3. **Network issues**
   ```bash
   # Configure proxy
   export HPN_PROXY_HTTP=http://proxy.example.com:8080
   ```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Harpoon** - Precision targeting for container image management challenges 🎯