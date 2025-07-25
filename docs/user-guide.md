# User Guide

Complete guide for using Harpoon (hpn) container image management tool.

## Table of Contents

- [Installation](#installation)
- [Basic Operations](#basic-operations)
- [Operation Modes](#operation-modes)
- [Runtime Selection](#runtime-selection)
- [Configuration](#configuration)
- [Advanced Usage](#advanced-usage)
- [Best Practices](#best-practices)

## Installation

### System Requirements
- Go 1.21+ (for building from source)
- One of: Docker, Podman, or Nerdctl
- Linux, macOS, or Windows

### Install from Binary
```bash
# Linux/macOS
curl -L https://github.com/your-org/harpoon/releases/latest/download/hpn-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) -o hpn
chmod +x hpn
sudo mv hpn /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/your-org/harpoon/releases/latest/download/hpn-windows-amd64.exe" -OutFile "hpn.exe"
```

### Build from Source
```bash
git clone https://github.com/your-org/harpoon.git
cd harpoon
make build
sudo make install
```

## Basic Operations

### Pull Images
Download images from registries:

```bash
# Create image list
echo "nginx:latest" > images.txt
echo "alpine:3.18" >> images.txt
echo "redis:7-alpine" >> images.txt

# Pull images
hpn -a pull -f images.txt
```

### Save Images
Export images to tar files:

```bash
# Save to current directory
hpn -a save -f images.txt --save-mode 1

# Save to ./images/ directory
hpn -a save -f images.txt --save-mode 2

# Save to ./images/<project>/ directories
hpn -a save -f images.txt --save-mode 3
```

### Load Images
Import images from tar files:

```bash
# Load from current directory
hpn -a load --load-mode 1

# Load from ./images/ directory
hpn -a load --load-mode 2

# Load recursively from ./images/*/ directories
hpn -a load --load-mode 3
```

### Push Images
Upload images to registries:

```bash
# Simple push (registry/image:tag)
hpn -a push -f images.txt -r harbor.company.com --push-mode 1

# Push with project (registry/project/image:tag)
hpn -a push -f images.txt -r harbor.company.com -p production --push-mode 2
```

## Operation Modes

### Save Modes
- **Mode 1**: Save to current directory
- **Mode 2**: Save to `./images/` directory
- **Mode 3**: Save to `./images/<project>/` directories

### Load Modes
- **Mode 1**: Load from current directory
- **Mode 2**: Load from `./images/` directory
- **Mode 3**: Load recursively from `./images/*/` directories

### Push Modes
- **Mode 1**: `registry/image:tag` (simple)
- **Mode 2**: `registry/project/image:tag` (with smart project selection)

#### Smart Project Selection (Mode 2)
Priority order:
1. **Command line**: `-p myproject`
2. **Config file**: `project: myproject`
3. **Original image**: Extract from `docker.io/user/image:tag` → `user`

## Runtime Selection

### Automatic Detection
```bash
hpn -a pull -f images.txt
# Automatically detects: Docker → Podman → Nerdctl
```

### Manual Selection
```bash
# Specify runtime
hpn --runtime docker -a pull -f images.txt
hpn --runtime podman -a pull -f images.txt
hpn --runtime nerdctl -a pull -f images.txt
```

### Auto-fallback Mode
```bash
# For CI environments
hpn --auto-fallback -a pull -f images.txt
```

### Runtime Fallback Behavior
When configured runtime is unavailable:
```
Runtime 'docker' is not available
Found available runtime: podman
Use 'podman' instead of 'docker'? (y/N):
```

## Configuration

### Config File Locations
1. `--config /path/to/config.yaml` (highest priority)
2. `~/.hpn/config.yaml`
3. `/etc/hpn/config.yaml`
4. `./config.yaml` (lowest priority)

### Configuration Example
```yaml
# Registry settings
registry: harbor.company.com
project: production

# Runtime configuration
runtime:
  preferred: docker
  auto_fallback: false
  timeout: 5m
  retry:
    max_attempts: 3
    delay: 1s
    max_delay: 30s

# Proxy settings
proxy:
  http: http://proxy.company.com:8080
  https: http://proxy.company.com:8080
  enabled: true

# Default modes
modes:
  save_mode: 2
  load_mode: 2
  push_mode: 2

# Logging
logging:
  level: info
  format: text
  console: true
  colors: true

# Parallel processing
parallel:
  max_workers: 4
  auto_adjust: true
```

### Environment Variables
```bash
export HPN_REGISTRY=harbor.company.com
export HPN_PROJECT=production
export HPN_RUNTIME_PREFERRED=docker
export HPN_PROXY_HTTP=http://proxy:8080
export HPN_PROXY_HTTPS=http://proxy:8080
```

## Advanced Usage

### Parallel Processing
```bash
# Configure in config.yaml
parallel:
  max_workers: 8
  auto_adjust: true
```

### Proxy Support
```bash
# Environment variables
export http_proxy=http://proxy:8080
export https_proxy=http://proxy:8080

# Or in config file
proxy:
  http: http://proxy:8080
  https: http://proxy:8080
  enabled: true
```

### Batch Operations
```bash
# Process multiple image lists
for list in *.txt; do
  echo "Processing $list..."
  hpn -a pull -f "$list"
  hpn -a save -f "$list" --save-mode 2
done
```

### Custom Image Lists
```bash
# Generate image list from running containers
docker ps --format "table {{.Image}}" | tail -n +2 > running-images.txt

# Generate from Kubernetes
kubectl get pods -o jsonpath='{.items[*].spec.containers[*].image}' | tr ' ' '\n' | sort -u > k8s-images.txt
```

## Best Practices

### Image List Management
- Use descriptive filenames: `production-images.txt`, `dev-images.txt`
- Include version tags: `nginx:1.21` instead of `nginx:latest`
- Group related images: separate frontend, backend, database images
- Comment with `#` for documentation

### Performance Optimization
- Use appropriate `max_workers` for your system
- Enable `auto_adjust` for dynamic scaling
- Use local registry for faster operations
- Configure proxy for corporate networks

### Security
- Use specific image tags, avoid `latest`
- Verify image sources and signatures
- Use private registries for sensitive images
- Regularly update base images

### CI/CD Integration
```bash
# Use auto-fallback in CI
hpn --auto-fallback -a push -f images.txt -r registry.com -p prod --push-mode 2

# Set timeout for CI environments
hpn --config ci-config.yaml -a pull -f images.txt
```

### Troubleshooting
- Use `--runtime` to specify runtime if auto-detection fails
- Check `hpn version` for runtime information
- Enable debug logging: `HPN_LOG_LEVEL=debug`
- Verify registry authentication: `docker login registry.com`

## Error Handling

### Common Errors
```bash
# Runtime not found
Error: no container runtime found
Solution: Install Docker, Podman, or Nerdctl

# Permission denied
Error: permission denied
Solution: Add user to docker group or use sudo

# Registry authentication
Error: authentication failed
Solution: docker login registry.com

# Invalid parameters
Error: --save-mode cannot be used with push action
Solution: Use correct mode parameter for the action
```

### Debug Mode
```bash
export HPN_LOG_LEVEL=debug
hpn -a pull -f images.txt
```

## Integration Examples

### Docker Compose
```yaml
version: '3.8'
services:
  hpn:
    image: alpine:latest
    volumes:
      - ./images.txt:/images.txt
      - /var/run/docker.sock:/var/run/docker.sock
    command: hpn -a pull -f /images.txt
```

### Kubernetes Job
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: hpn-pull
spec:
  template:
    spec:
      containers:
      - name: hpn
        image: hpn:latest
        command: ["hpn", "-a", "pull", "-f", "/config/images.txt"]
        volumeMounts:
        - name: config
          mountPath: /config
      volumes:
      - name: config
        configMap:
          name: image-list
      restartPolicy: Never
```

## See Also

- [Quick Start Guide](quickstart.md)
- [Configuration Reference](configuration.md)
- [Examples](examples.md)
- [Troubleshooting](troubleshooting.md)
- [API Reference](api-reference.md)