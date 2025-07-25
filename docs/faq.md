# Frequently Asked Questions (FAQ)

## General Questions

### What is Harpoon?

Harpoon (hpn) is a modern container image management CLI tool written in Go. It provides efficient operations for pulling, saving, loading, and pushing container images with support for multiple container runtimes.

### Why use Harpoon instead of Docker CLI directly?

Harpoon offers several advantages:

- **Multi-runtime support**: Works with Docker, Podman, and Nerdctl
- **Batch operations**: Process multiple images efficiently
- **Flexible modes**: Different save/load/push strategies
- **Configuration management**: YAML-based configuration
- **Smart runtime detection**: Automatic fallback between runtimes
- **Enterprise features**: Proxy support, authentication handling

### What container runtimes are supported?

- **Docker**: Full support with automatic detection
- **Podman**: Full support with rootless containers
- **Nerdctl**: Full support with containerd backend

## Installation & Setup

### How do I install Harpoon?

See the [Installation Guide](installation.md) for detailed instructions. The quickest method is:

```bash
curl -L https://github.com/your-org/harpoon/releases/latest/download/hpn-linux-amd64 -o hpn
chmod +x hpn
sudo mv hpn /usr/local/bin/
```

### Do I need to install all container runtimes?

No, you only need at least one container runtime installed. Harpoon will automatically detect and use the available runtime.

### How do I configure Harpoon?

Create a configuration file at `~/.hpn/config.yaml`:

```yaml
registry: harbor.company.com
project: production
runtime:
  preferred: docker
  auto_fallback: true
```

See the [Configuration Guide](configuration.md) for all options.

## Usage Questions

### How do I pull multiple images at once?

Create a text file with image names (one per line):

```bash
echo "nginx:latest" > images.txt
echo "alpine:3.18" >> images.txt
hpn -a pull -f images.txt
```

### What are the different save modes?

- **Mode 1**: Save to current directory
- **Mode 2**: Save to `./images/` directory
- **Mode 3**: Save to `./images/<project>/` directory

```bash
hpn -a save -f images.txt --save-mode 2
```

### What are the different push modes?

- **Mode 1**: `registry/image:tag` (simple)
- **Mode 2**: `registry/project/image:tag` (with smart project selection)

```bash
hpn -a push -f images.txt -r harbor.com --push-mode 2
```

### How does smart project selection work in Push Mode 2?

Priority order:
1. Command line `-p` parameter
2. Configuration file `project` setting
3. Original image project name

### Can I use Harpoon in CI/CD pipelines?

Yes! Use the `--auto-fallback` flag for automatic runtime selection:

```bash
hpn --auto-fallback -a push -f images.txt -r registry.com -p production --push-mode 2
```

## Troubleshooting

### "No container runtime found" error

Install at least one container runtime:

```bash
# Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Or Podman
sudo apt install podman
```

### "Permission denied" error

Add your user to the docker group:

```bash
sudo usermod -aG docker $USER
newgrp docker
```

Or use sudo:

```bash
sudo hpn -a pull -f images.txt
```

### "Authentication failed" error

Login to your registry first:

```bash
docker login harbor.company.com
# or
podman login harbor.company.com
```

### Images not found after load

Check the load mode and directory:

```bash
# List tar files
ls -la *.tar
ls -la images/*.tar

# Use correct load mode
hpn -a load --load-mode 1  # current directory
hpn -a load --load-mode 2  # ./images/ directory
```

### Runtime detection issues

Check available runtimes:

```bash
docker --version
podman --version
nerdctl --version
```

Force a specific runtime:

```bash
hpn --runtime docker -a pull -f images.txt
```

## Configuration Questions

### Where should I put the configuration file?

Harpoon looks for configuration in this order:

1. File specified by `--config` flag
2. `~/.hpn/config.yaml`
3. `/etc/hpn/config.yaml`
4. `./config.yaml`

### Can I use environment variables?

Yes, environment variables override config file settings:

```bash
export HPN_REGISTRY=harbor.company.com
export HPN_PROJECT=production
export HPN_RUNTIME_PREFERRED=podman
```

### How do I configure proxy settings?

In your config file:

```yaml
proxy:
  http: http://proxy.company.com:8080
  https: http://proxy.company.com:8080
  enabled: true
```

Or use standard environment variables:

```bash
export http_proxy=http://proxy.company.com:8080
export https_proxy=http://proxy.company.com:8080
```

## Performance Questions

### How can I speed up operations?

1. **Use parallel processing**:
   ```yaml
   parallel:
     max_workers: 8
     auto_adjust: true
   ```

2. **Use local registry** for faster transfers

3. **Configure appropriate timeouts**:
   ```yaml
   runtime:
     timeout: 10m
   ```

### How many images can I process at once?

There's no hard limit, but consider:

- Available disk space for save operations
- Network bandwidth for pull/push operations
- System memory and CPU resources

Start with 10-20 images and adjust based on performance.

## Advanced Usage

### Can I use custom registries?

Yes, specify the registry with `-r` flag:

```bash
hpn -a push -f images.txt -r my-registry.com -p myproject --push-mode 2
```

### How do I handle insecure registries?

For development environments, some runtimes support insecure registries. Configure this in your container runtime, not in Harpoon.

### Can I script Harpoon operations?

Yes, Harpoon is designed for scripting:

```bash
#!/bin/bash
set -e

# Pull images
hpn -a pull -f production-images.txt

# Save for backup
hpn -a save -f production-images.txt --save-mode 2

# Push to staging
hpn -a push -f production-images.txt -r staging-registry.com -p staging --push-mode 2
```

### How do I migrate images between registries?

```bash
# Pull from source
hpn -a pull -f images.txt

# Push to destination with project mapping
hpn -a push -f images.txt -r target-registry.com -p newproject --push-mode 2
```

## Development Questions

### How do I contribute to Harpoon?

See the [Development Guide](development.md) for:

- Setting up development environment
- Building from source
- Running tests
- Submitting pull requests

### How do I report bugs?

1. Check existing [GitHub Issues](https://github.com/your-org/harpoon/issues)
2. Create a new issue with:
   - Harpoon version (`hpn --version`)
   - Operating system and architecture
   - Container runtime version
   - Complete error message
   - Steps to reproduce

### How do I request features?

1. Check [GitHub Discussions](https://github.com/your-org/harpoon/discussions)
2. Create a feature request with:
   - Use case description
   - Expected behavior
   - Alternative solutions considered

## Compatibility Questions

### What Go version is required?

- **Runtime**: No Go installation required for pre-built binaries
- **Building**: Go 1.21+ required for building from source

### What platforms are supported?

- **Linux**: AMD64, ARM64
- **macOS**: AMD64 (Intel), ARM64 (Apple Silicon)
- **Windows**: AMD64

### Is Harpoon compatible with Docker Compose?

Harpoon operates on individual images, not Docker Compose services. You can extract image names from `docker-compose.yml` and use them with Harpoon.

### Can I use Harpoon with Kubernetes?

Yes, Harpoon can help with:

- Pre-pulling images to nodes
- Migrating images between registries
- Backing up critical images

## Security Questions

### How does Harpoon handle credentials?

Harpoon uses the same credential sources as your container runtime:

- Docker: `~/.docker/config.json`
- Podman: `~/.config/containers/auth.json`
- Environment variables
- Interactive login prompts

### Is it safe to use in production?

Yes, Harpoon is designed for production use with:

- No credential storage in code
- Secure temporary file handling
- Input validation and sanitization
- Comprehensive error handling

### How do I secure registry communications?

- Use HTTPS registries (default)
- Configure proper TLS certificates
- Use authentication tokens with minimal permissions
- Regularly rotate credentials

## Still Have Questions?

- Check the [User Guide](user-guide.md) for detailed usage
- Review [Examples](examples.md) for real-world scenarios
- Visit [GitHub Discussions](https://github.com/your-org/harpoon/discussions)
- Create an [Issue](https://github.com/your-org/harpoon/issues) for bugs