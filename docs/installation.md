# Installation Guide

This guide covers various methods to install Harpoon (hpn) on different platforms.

## System Requirements

- **Operating System**: Linux, macOS, or Windows
- **Architecture**: AMD64 or ARM64
- **Container Runtime**: Docker, Podman, or Nerdctl (at least one required)
- **Go Version**: 1.21+ (if building from source)

## Installation Methods

### 1. Download Pre-built Binary (Recommended)

#### Linux
```bash
# AMD64
curl -L https://github.com/your-org/harpoon/releases/latest/download/hpn-linux-amd64 -o hpn
chmod +x hpn
sudo mv hpn /usr/local/bin/

# ARM64
curl -L https://github.com/your-org/harpoon/releases/latest/download/hpn-linux-arm64 -o hpn
chmod +x hpn
sudo mv hpn /usr/local/bin/
```

#### macOS
```bash
# Intel Mac
curl -L https://github.com/your-org/harpoon/releases/latest/download/hpn-darwin-amd64 -o hpn
chmod +x hpn
sudo mv hpn /usr/local/bin/

# Apple Silicon Mac
curl -L https://github.com/your-org/harpoon/releases/latest/download/hpn-darwin-arm64 -o hpn
chmod +x hpn
sudo mv hpn /usr/local/bin/
```

#### Windows
```powershell
# Download from GitHub releases page
# https://github.com/your-org/harpoon/releases/latest
# Extract hpn.exe and add to PATH
```

### 2. Package Managers

#### Homebrew (macOS/Linux)
```bash
# Coming soon
brew install harpoon
```

#### Chocolatey (Windows)
```powershell
# Coming soon
choco install harpoon
```

#### APT (Ubuntu/Debian)
```bash
# Coming soon
sudo apt install harpoon
```

### 3. Build from Source

#### Prerequisites
- Go 1.21 or later
- Git

#### Build Steps
```bash
# Clone repository
git clone https://github.com/your-org/harpoon.git
cd harpoon

# Build for current platform
make build

# Install to system
sudo make install

# Or build for all platforms
make build-all
```

#### Custom Build Options
```bash
# Build with custom version
make build VERSION=v1.1.0

# Build with debug symbols
go build -o hpn ./cmd/hpn

# Cross-compile for specific platform
GOOS=linux GOARCH=amd64 make build
```

### 4. Container Image

#### Docker
```bash
# Pull image
docker pull ghcr.io/your-org/harpoon:latest

# Run as container
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd):/workspace \
  ghcr.io/your-org/harpoon:latest -a pull -f images.txt
```

#### Podman
```bash
# Pull image
podman pull ghcr.io/your-org/harpoon:latest

# Run as container
podman run --rm -v /run/podman/podman.sock:/var/run/docker.sock \
  -v $(pwd):/workspace \
  ghcr.io/your-org/harpoon:latest -a pull -f images.txt
```

## Container Runtime Setup

Harpoon requires at least one container runtime to be installed and configured.

### Docker

#### Linux
```bash
# Ubuntu/Debian
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# CentOS/RHEL/Fedora
sudo dnf install docker-ce docker-ce-cli containerd.io
sudo systemctl enable --now docker
sudo usermod -aG docker $USER
```

#### macOS
```bash
# Install Docker Desktop
brew install --cask docker
```

#### Windows
Download and install Docker Desktop from [docker.com](https://www.docker.com/products/docker-desktop)

### Podman

#### Linux
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install podman

# CentOS/RHEL/Fedora
sudo dnf install podman

# Arch Linux
sudo pacman -S podman
```

#### macOS
```bash
brew install podman
podman machine init
podman machine start
```

#### Windows
Download from [podman.io](https://podman.io/getting-started/installation#windows)

### Nerdctl

#### Linux
```bash
# Download latest release
curl -L https://github.com/containerd/nerdctl/releases/latest/download/nerdctl-full-linux-amd64.tar.gz -o nerdctl.tar.gz
sudo tar -xzf nerdctl.tar.gz -C /usr/local/
```

#### macOS
```bash
brew install nerdctl
```

## Verification

### Check Installation
```bash
# Verify hpn is installed
hpn --version

# Check available runtimes
hpn --help | grep runtime

# Test basic functionality
echo "hello-world:latest" > test-images.txt
hpn -a pull -f test-images.txt
```

### Runtime Detection
```bash
# Let hpn detect available runtimes
hpn -a pull -f test-images.txt

# Specify runtime explicitly
hpn --runtime docker -a pull -f test-images.txt
hpn --runtime podman -a pull -f test-images.txt
```

## Configuration

### Create Configuration Directory
```bash
mkdir -p ~/.hpn
```

### Basic Configuration
```bash
cat > ~/.hpn/config.yaml << EOF
registry: harbor.company.com
project: production
runtime:
  preferred: docker
  auto_fallback: true
modes:
  save_mode: 2
  load_mode: 2
  push_mode: 2
EOF
```

## Troubleshooting

### Common Issues

**Command not found:**
```bash
# Check if binary is in PATH
which hpn
echo $PATH

# Add to PATH if needed
export PATH=$PATH:/usr/local/bin
```

**Permission denied:**
```bash
# Make binary executable
chmod +x /usr/local/bin/hpn

# Or run with sudo
sudo hpn -a pull -f images.txt
```

**Runtime not found:**
```bash
# Check available runtimes
docker --version
podman --version
nerdctl --version

# Install at least one runtime (see above)
```

### Getting Help

- Check the [Troubleshooting Guide](troubleshooting.md)
- Review [Configuration Guide](configuration.md)
- Visit [GitHub Issues](https://github.com/your-org/harpoon/issues)

## Next Steps

- Read the [Quick Start Guide](quickstart.md)
- Explore [Configuration Options](configuration.md)
- Check out [Usage Examples](examples.md)