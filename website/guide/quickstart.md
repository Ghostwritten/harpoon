# Quick Start Guide

Get up and running with Harpoon (hpn) in minutes.

## Installation

### Download Binary
```bash
# Download the latest release for your platform
curl -L https://github.com/Ghostwritten/harpoon/releases/latest/download/hpn-linux-amd64 -o hpn
chmod +x hpn
sudo mv hpn /usr/local/bin/
```

### Build from Source
```bash
git clone https://github.com/Ghostwritten/harpoon.git
cd harpoon
make build
sudo cp dist/hpn /usr/local/bin/
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
hpn pull -f images.txt
```

### 2. Save Images
Save images to tar files:
```bash
# Save to default directory (./images)
hpn save -f images.txt

# Save to custom directory
hpn save -f images.txt --path ./output/images
```

### 3. Load Images
Load images from tar files:
```bash
# Load from default directory (./images)
hpn load

# Load from custom directory
hpn load --path ./images

# Load recursively from subdirectories
hpn load --path ./images --recursive
```

### 4. List Images
Three modes:
```bash
# List images in local runtime (docker/podman/nerdctl)
hpn ls

# List saved tar files in path (IMAGE, SIZE, CHECKSUM)
hpn ls --path ./images

# Check images.txt against runtime or path
hpn ls -f images.txt
hpn ls -f images.txt --path ./images
```

### 5. Push Images
Push images to a registry:
```bash
# Preserve original paths
hpn push -f images.txt --registry harbor.company.com

# Push with unified project namespace
hpn push -f images.txt --registry harbor.company.com --project production
```

You can pass extra flags to the underlying runtime (docker/podman/nerdctl/skopeo) after `--`, e.g. `hpn pull -f images.txt -- --tls-verify=false`.

### 6. Login to Registry
Login to a container registry:
```bash
# Interactive login (most secure)
hpn login harbor.company.com

# Login with credentials
hpn login harbor.company.com -u admin -p password

# Login from stdin (for CI/CD)
echo "password" | hpn login harbor.company.com -u admin --password-stdin
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
paths:
  save_path: ./images
  load_path: ./images
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
hpn pull -f dev-images.txt

# Save for offline use
hpn save -f dev-images.txt --path ./images

# Load on another machine
hpn load --path ./images
```

### CI/CD Pipeline
```bash
# Build and push in CI
hpn --auto-fallback push -f production-images.txt --registry harbor.company.com --project production
```

### Image Migration
```bash
# Migrate from Docker Hub to private registry
hpn push -f dockerhub-images.txt --registry harbor.company.com --project production
```

## Runtime Selection

### Auto-detection (Default)
```bash
hpn pull -f images.txt
# Automatically detects and uses available runtime
```

### Specify Runtime
```bash
# Use Docker
hpn --runtime docker pull -f images.txt

# Use Podman
hpn --runtime podman pull -f images.txt

# Auto-fallback mode
hpn --auto-fallback pull -f images.txt
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
hpn login harbor.company.com
# Or use the native runtime command
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

- See [Examples](/guide/examples) for real-world scenarios
- Read [Installation](/guide/installation) for detailed setup
- Check [Building](/guide/building) for source builds
