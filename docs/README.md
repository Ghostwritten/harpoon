# Harpoon (hpn) - Container Image Management Tool

Harpoon (`hpn`) is a powerful container image management tool designed for cloud-native environments. It provides flexible image pulling, saving, loading, and pushing capabilities with support for multiple container runtimes and flexible operation modes.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Command Reference](#command-reference)
- [Operation Modes](#operation-modes)
- [Configuration](#configuration)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)

## Installation

### Build from Source

```bash
git clone <repository-url>
cd harpoon-go
go build -o hpn cmd/hpn/main.go
```

### Binary Download

Download pre-compiled binaries for your system from the [Releases](releases) page.

## Quick Start

1. Create an image list file `images.txt`:
```
nginx:latest
redis:alpine
mysql:8.0
```

2. Pull images:
```bash
hpn -a pull -f images.txt
```

3. Save images to tar files:
```bash
hpn -a save -f images.txt --save-mode 2
```

4. Load images:
```bash
hpn -a load --load-mode 2
```

5. Push to private registry:
```bash
hpn -a push -f images.txt -r registry.example.com -p myproject --push-mode 2
```

## Command Reference

### Basic Syntax

```bash
hpn -a <action> -f <image_list> [-r <registry>] [-p <project>] [--push-mode <1|2|3>] [--load-mode <1|2|3>] [--save-mode <1|2|3>]
```

### Required Parameters

- `-a, --action`: Action type (required)
  - `pull`: Pull images from external registry
  - `save`: Save images into tar files
  - `load`: Load images from tar files
  - `push`: Push images to private registry

### Optional Parameters

- `-f, --file`: Image list file (required for pull/save/push)
- `-r, --registry`: Target registry address (default: registry.k8s.local)
- `-p, --project`: Target project namespace (default: library)

### Global Options

- `--config`: Configuration file path
- `--log-level`: Log level (debug, info, warn, error)
- `--log-file`: Log file path
- `--quiet`: Quiet mode
- `--output`: Output format (text, json)
- `--runtime`: Specify container runtime (docker, podman, nerdctl)

## Operation Modes

### Push Modes (--push-mode)

- **Mode 1** (default): `registry/image:tag`
  - Push format: `registry.example.com/nginx:latest`
  
- **Mode 2**: `registry/project/image:tag`
  - Push format: `registry.example.com/myproject/nginx:latest`
  
- **Mode 3**: Preserve original project path
  - Push format: `registry.example.com/original-project/nginx:latest`

### Load Modes (--load-mode)

- **Mode 1** (default): Load all `*.tar` files from current directory
- **Mode 2**: Load all `*.tar` files from `./images/` directory
- **Mode 3**: Recursively load `*.tar` files from `./images/*/` subdirectories

### Save Modes (--save-mode)

- **Mode 1** (default): Save tar files to current directory
- **Mode 2**: Save tar files to `./images/` directory
- **Mode 3**: Save tar files to `./images/<project>/` directory

## Configuration

### Configuration File

Default configuration file locations:
- `~/.hpn/config.yaml`
- `/etc/hpn/config.yaml`

Example configuration:

```yaml
registry: registry.k8s.local
project: library
proxy:
  http: http://192.168.21.101:7890
  https: http://192.168.21.101:7890
  enabled: true
runtime:
  preferred: docker
  timeout: 300s
logging:
  level: info
  format: text
  console: true
  file: ./hpn.log
parallel:
  max_workers: 5
  adaptive: true
```

### Environment Variables

All configuration options can be set via environment variables using the `HPN_` prefix:

```bash
export HPN_REGISTRY=registry.example.com
export HPN_PROJECT=myproject
export HPN_PROXY_HTTP=http://proxy.example.com:8080
export HPN_LOG_LEVEL=debug
```

## Examples

### Basic Operations

#### 1. Pull Image List

Create `k8s-images.txt`:
```
k8s.gcr.io/kube-apiserver:v1.28.0
k8s.gcr.io/kube-controller-manager:v1.28.0
k8s.gcr.io/kube-scheduler:v1.28.0
k8s.gcr.io/kube-proxy:v1.28.0
```

Pull images:
```bash
hpn -a pull -f k8s-images.txt
```

#### 2. Save Images to Different Locations

Save to current directory:
```bash
hpn -a save -f k8s-images.txt --save-mode 1
```

Save to images directory:
```bash
hpn -a save -f k8s-images.txt --save-mode 2
```

Save organized by project:
```bash
hpn -a save -f k8s-images.txt --save-mode 3
```

#### 3. Load Images

Load from current directory:
```bash
hpn -a load --load-mode 1
```

Load from images directory:
```bash
hpn -a load --load-mode 2
```

Recursive load:
```bash
hpn -a load --load-mode 3
```

#### 4. Push to Private Registry

Simple push:
```bash
hpn -a push -f k8s-images.txt -r harbor.example.com --push-mode 1
```

Push with project namespace:
```bash
hpn -a push -f k8s-images.txt -r harbor.example.com -p kubernetes --push-mode 2
```

Push preserving original path:
```bash
hpn -a push -f k8s-images.txt -r harbor.example.com --push-mode 3
```

### Advanced Usage

#### Pull Images with Proxy

```bash
HPN_PROXY_HTTP=http://proxy.example.com:8080 \
HPN_PROXY_HTTPS=http://proxy.example.com:8080 \
hpn -a pull -f images.txt
```

#### Specify Container Runtime

```bash
hpn --runtime podman -a pull -f images.txt
```

#### JSON Output Format

```bash
hpn --output json -a pull -f images.txt
```

#### Quiet Mode

```bash
hpn --quiet -a save -f images.txt --save-mode 2
```

### Batch Processing Example

#### Complete Image Migration Workflow

```bash
#!/bin/bash

# 1. Pull images
echo "Pulling images..."
hpn -a pull -f production-images.txt

# 2. Save images
echo "Saving images..."
hpn -a save -f production-images.txt --save-mode 2

# 3. Transfer to target environment (example)
echo "Transferring image files..."
rsync -av images/ target-server:/tmp/images/

# 4. Load images in target environment
echo "Loading images in target environment..."
ssh target-server "cd /tmp && hpn -a load --load-mode 2"

# 5. Push to private registry
echo "Pushing to private registry..."
ssh target-server "hpn -a push -f /tmp/production-images.txt -r harbor.internal.com -p production --push-mode 2"
```

## Troubleshooting

### Common Issues

1. **Container runtime not found**
   ```
   Error: No container runtime found. Please install docker, podman, or nerdctl.
   ```
   Solution: Install at least one supported container runtime.

2. **Image pull failed**
   ```
   Error: Failed to pull image nginx:latest: network timeout
   ```
   Solution: Check network connection or configure proxy.

3. **Insufficient disk space**
   ```
   Error: Insufficient disk space for save operation
   ```
   Solution: Clean up disk space or choose another save location.

4. **Registry authentication failed**
   ```
   Error: Authentication failed for registry
   ```
   Solution: Configure correct authentication or use docker login.

### Debug Mode

Enable verbose logging:
```bash
hpn --log-level debug -a pull -f images.txt
```

### Log Files

View detailed logs:
```bash
hpn --log-file hpn.log -a pull -f images.txt
tail -f hpn.log
```

## Performance Optimization

### Parallel Processing

Configure parallel worker threads:
```yaml
parallel:
  max_workers: 10
  adaptive: true
```

### Network Optimization

Configure proxy and timeout:
```yaml
proxy:
  http: http://proxy.example.com:8080
  https: http://proxy.example.com:8080
  enabled: true
runtime:
  timeout: 600s
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is licensed under the [MIT License](LICENSE).