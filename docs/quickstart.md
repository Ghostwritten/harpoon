# Quick Start Guide

Get up and running with Harpoon (hpn) in minutes.

## Installation

### Download Binary
```bash
# Download the latest release for your platform
curl -L https://github.com/your-org/harpoon/releases/latest/download/hpn-linux-amd64 -o hpn
chmod +x hpn
sudo mv hpn /usr/local/bin/
```

### Build from Source
```bash
git clone https://github.com/your-org/harpoon.git
cd harpoon
make build
sudo cp hpn /usr/local/bin/
```

### Verify Installation
```bash
hpn --version
```

## Basic Usage

### 1. Pull Images
Create an image list file:
```bash
echo "nginx:latest" > images.txt
echo "alpine:3.18" >> images.txt
```

Pull images:
```bash
hpn -a pull -f images.txt
```

### 2. Save Images
Save images to tar files:
```bash
# Save to current directory
hpn -a save -f images.txt --save-mode 1

# Save to ./images/ directory
hpn -a save -f images.txt --save-mode 2
```

### 3. Load Images
Load images from tar files:
```bash
# Load from current directory
hpn -a load --load-mode 1

# Load from ./images/ directory
hpn -a load --load-mode 2
```

### 4. Push Images
Push images to a registry:
```bash
# Simple push
hpn -a push -f images.txt -r harbor.company.com --push-mode 1

# Push with project namespace
hpn -a push -f images.txt -r harbor.company.com -p production --push-mode 2
```

## Configuration

### Create Config File
```bash
mkdir -p ~/.hpn
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

### Environment Variables
```bash
export HPN_REGISTRY=harbor.company.com
export HPN_PROJECT=production
export HPN_RUNTIME_PREFERRED=docker
```

## Common Workflows

### Development Workflow
```bash
# Pull development images
hpn -a pull -f dev-images.txt

# Save for offline use
hpn -a save -f dev-images.txt --save-mode 2

# Load on another machine
hpn -a load --load-mode 2
```

### CI/CD Pipeline
```bash
# Build and push in CI
hpn --auto-fallback -a push -f production-images.txt -r harbor.company.com -p production --push-mode 2
```

### Image Migration
```bash
# Migrate from Docker Hub to private registry
hpn -a push -f dockerhub-images.txt -r harbor.company.com --push-mode 2
```

## Runtime Selection

### Auto-detection (Default)
```bash
hpn -a pull -f images.txt
# Automatically detects and uses available runtime
```

### Specify Runtime
```bash
# Use Docker
hpn --runtime docker -a pull -f images.txt

# Use Podman
hpn --runtime podman -a pull -f images.txt

# Auto-fallback mode
hpn --auto-fallback -a pull -f images.txt
```

## Troubleshooting

### Common Issues

**Runtime not found:**
```bash
Error: no container runtime found
```
Solution: Install Docker, Podman, or Nerdctl

**Permission denied:**
```bash
Error: permission denied
```
Solution: Add user to docker group or use sudo

**Registry authentication:**
```bash
Error: authentication failed
```
Solution: Login to registry first
```bash
docker login harbor.company.com
```

### Getting Help
```bash
# Show help
hpn --help

# Show version
hpn --version

# Show detailed version
hpn version
```

## Next Steps

- Read the [User Guide](user-guide.md) for detailed usage
- Check [Configuration Guide](configuration.md) for advanced settings
- See [Examples](examples.md) for real-world scenarios
- Review [Troubleshooting](troubleshooting.md) for common issues

## Support

- [GitHub Issues](https://github.com/your-org/harpoon/issues)
- [Documentation](https://github.com/your-org/harpoon/tree/main/docs)
- [Discussions](https://github.com/your-org/harpoon/discussions)