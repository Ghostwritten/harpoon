# Container Runtime Support

This document details Harpoon's support for different container runtimes and how to configure them.

## Supported Runtimes

Harpoon supports three major container runtimes with automatic detection and fallback capabilities.

### Docker

**Status**: ✅ Full Support  
**Priority**: 1 (Highest)  
**Platforms**: Linux, macOS, Windows

#### Features
- Complete API compatibility
- Automatic daemon detection
- Proxy support
- Authentication integration
- Multi-architecture support

#### Installation
```bash
# Linux
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# macOS
brew install --cask docker

# Windows
# Download Docker Desktop from docker.com
```

#### Configuration
```yaml
runtime:
  preferred: docker
  timeout: 5m
```

### Podman

**Status**: ✅ Full Support  
**Priority**: 2  
**Platforms**: Linux, macOS, Windows

#### Features
- Rootless container support
- Docker CLI compatibility
- Systemd integration
- Pod management
- Enhanced security

#### Installation
```bash
# Ubuntu/Debian
sudo apt update && sudo apt install podman

# CentOS/RHEL/Fedora
sudo dnf install podman

# macOS
brew install podman
podman machine init
podman machine start
```

#### Configuration
```yaml
runtime:
  preferred: podman
  timeout: 5m
```

### Nerdctl

**Status**: ✅ Full Support  
**Priority**: 3  
**Platforms**: Linux, macOS

#### Features
- containerd backend
- Docker CLI compatibility
- Kubernetes integration
- Advanced networking
- Image encryption support

#### Installation
```bash
# Linux
curl -L https://github.com/containerd/nerdctl/releases/latest/download/nerdctl-full-linux-amd64.tar.gz -o nerdctl.tar.gz
sudo tar -xzf nerdctl.tar.gz -C /usr/local/

# macOS
brew install nerdctl
```

#### Configuration
```yaml
runtime:
  preferred: nerdctl
  timeout: 5m
```

## Runtime Detection

### Automatic Detection

Harpoon automatically detects available runtimes in priority order:

1. **Docker** - Checks for `docker` command and daemon availability
2. **Podman** - Checks for `podman` command and socket/service
3. **Nerdctl** - Checks for `nerdctl` command and containerd

```bash
# Let Harpoon auto-detect
hpn -a pull -f images.txt
```

### Manual Selection

Override automatic detection by specifying a runtime:

```bash
# Use Docker explicitly
hpn --runtime docker -a pull -f images.txt

# Use Podman explicitly
hpn --runtime podman -a pull -f images.txt

# Use Nerdctl explicitly
hpn --runtime nerdctl -a pull -f images.txt
```

### Runtime Fallback

When the preferred runtime is unavailable, Harpoon can automatically fallback:

#### Interactive Mode (Default)
```bash
$ hpn -a pull -f images.txt
Runtime 'docker' is not available
Found available runtime: podman
Use 'podman' instead of 'docker'? (y/N): y
Using 'podman' runtime
```

#### Automatic Mode (CI/CD)
```bash
# Enable auto-fallback
hpn --auto-fallback -a pull -f images.txt

# Or configure in config file
runtime:
  auto_fallback: true
```

## Runtime-Specific Features

### Docker-Specific

#### BuildKit Support
```bash
export DOCKER_BUILDKIT=1
hpn -a pull -f images.txt
```

#### Docker Desktop Integration
- Automatic credential helper integration
- Volume mount optimization
- Resource limit awareness

### Podman-Specific

#### Rootless Containers
```bash
# No sudo required
hpn -a pull -f images.txt
```

#### Systemd Integration
```bash
# Generate systemd units
podman generate systemd --new --files --name mycontainer
```

#### Pod Support
```bash
# Podman pods are handled transparently
hpn -a pull -f pod-images.txt
```

### Nerdctl-Specific

#### Containerd Namespaces
```bash
# Use specific namespace
export CONTAINERD_NAMESPACE=k8s.io
hpn -a pull -f images.txt
```

#### Image Encryption
```bash
# Encrypted images are supported
hpn -a pull -f encrypted-images.txt
```

## Configuration Examples

### Multi-Runtime Environment

```yaml
# ~/.hpn/config.yaml
runtime:
  preferred: docker      # Try Docker first
  auto_fallback: true    # Fall back to others if Docker unavailable
  timeout: 10m           # Longer timeout for slow networks
  retry:
    max_attempts: 3
    delay: 2s
    max_delay: 30s

# Runtime-specific settings
docker:
  buildkit: true
  
podman:
  rootless: true
  
nerdctl:
  namespace: default
```

### CI/CD Configuration

```yaml
# Optimized for CI environments
runtime:
  preferred: docker
  auto_fallback: true    # Essential for CI reliability
  timeout: 15m           # Longer timeout for CI
  retry:
    max_attempts: 5      # More retries in CI
    delay: 1s
    max_delay: 60s
```

### Development Configuration

```yaml
# Developer-friendly settings
runtime:
  preferred: podman      # Rootless for security
  auto_fallback: false   # Explicit control
  timeout: 5m
  
logging:
  level: debug           # Verbose logging
  console: true
```

## Runtime Comparison

| Feature | Docker | Podman | Nerdctl |
|---------|--------|--------|---------|
| **Daemon Required** | Yes | No | Yes (containerd) |
| **Root Required** | Yes* | No | Yes* |
| **Docker API Compat** | Native | High | High |
| **Kubernetes Integration** | Good | Excellent | Excellent |
| **Image Formats** | OCI, Docker | OCI, Docker | OCI, Docker |
| **Registry Support** | Full | Full | Full |
| **Windows Support** | Yes | Limited | No |
| **macOS Support** | Yes | Yes | Yes |
| **Performance** | Good | Good | Excellent |

*Can be configured for rootless operation

## Troubleshooting

### Runtime Detection Issues

**Problem**: "No container runtime found"

**Solutions**:
```bash
# Check if runtimes are installed
docker --version
podman --version
nerdctl --version

# Check if services are running
sudo systemctl status docker
sudo systemctl status podman
sudo systemctl status containerd
```

### Permission Issues

**Problem**: "Permission denied"

**Solutions**:
```bash
# Docker: Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Podman: Enable rootless mode
podman system migrate

# Nerdctl: Check containerd permissions
sudo chown $USER /run/containerd/containerd.sock
```

### Runtime Switching

**Problem**: Want to switch between runtimes

**Solutions**:
```bash
# Temporary switch
hpn --runtime podman -a pull -f images.txt

# Permanent switch
echo "runtime:\n  preferred: podman" >> ~/.hpn/config.yaml

# Environment variable
export HPN_RUNTIME_PREFERRED=podman
```

### Performance Issues

**Problem**: Slow operations

**Solutions**:
```bash
# Increase timeout
hpn --runtime docker -a pull -f images.txt
# Add to config:
runtime:
  timeout: 15m

# Use faster runtime for your use case
hpn --runtime nerdctl -a pull -f images.txt  # Often faster
```

## Best Practices

### Production Environments

1. **Explicit Runtime Selection**:
   ```bash
   hpn --runtime docker -a push -f images.txt
   ```

2. **Enable Auto-fallback**:
   ```yaml
   runtime:
     auto_fallback: true
   ```

3. **Configure Appropriate Timeouts**:
   ```yaml
   runtime:
     timeout: 15m
   ```

### Development Environments

1. **Use Rootless When Possible**:
   ```yaml
   runtime:
     preferred: podman
   ```

2. **Enable Debug Logging**:
   ```yaml
   logging:
     level: debug
   ```

3. **Quick Runtime Testing**:
   ```bash
   make dev-test  # Tests all available runtimes
   ```

### CI/CD Environments

1. **Always Use Auto-fallback**:
   ```bash
   hpn --auto-fallback -a push -f images.txt
   ```

2. **Increase Retry Attempts**:
   ```yaml
   runtime:
     retry:
       max_attempts: 5
   ```

3. **Use Appropriate Timeouts**:
   ```yaml
   runtime:
     timeout: 20m  # Longer for CI
   ```

## Future Runtime Support

### Planned Support

- **CRI-O**: Container Runtime Interface implementation
- **Firecracker**: Lightweight virtualization
- **gVisor**: Application kernel for containers

### Contributing Runtime Support

To add support for a new runtime:

1. Implement the `ContainerRuntime` interface
2. Add detection logic to `detector.go`
3. Create runtime-specific implementation
4. Add tests and documentation
5. Submit a pull request

See the [Development Guide](development.md) for details.

## Related Documentation

- [Installation Guide](installation.md) - Installing container runtimes
- [Configuration Guide](configuration.md) - Runtime configuration options
- [Troubleshooting](troubleshooting.md) - Common runtime issues
- [Architecture](architecture.md) - Runtime abstraction design