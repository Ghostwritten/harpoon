# API Reference

Complete command-line interface reference for Harpoon (hpn).

## Command Syntax

```
hpn -a <action> -f <file> [options]
```

## Global Options

### Required Options

#### `-a, --action`
- **Type**: string
- **Required**: Yes
- **Values**: `pull`, `save`, `load`, `push`
- **Description**: Specifies the operation to perform

#### `-f, --file`
- **Type**: string
- **Required**: Yes (except for load action)
- **Description**: Path to image list file
- **Example**: `-f images.txt`

### Optional Options

#### `-r, --registry`
- **Type**: string
- **Default**: From config or `registry.k8s.local`
- **Description**: Target registry for push operations
- **Example**: `-r harbor.company.com`

#### `-p, --project`
- **Type**: string
- **Default**: From config or `library`
- **Description**: Target project namespace
- **Example**: `-p production`

#### `-c, --config`
- **Type**: string
- **Default**: `~/.hpn/config.yaml`
- **Description**: Path to configuration file
- **Example**: `-c /path/to/config.yaml`

#### `--runtime`
- **Type**: string
- **Values**: `docker`, `podman`, `nerdctl`
- **Description**: Container runtime to use
- **Example**: `--runtime podman`

#### `--auto-fallback`
- **Type**: boolean
- **Default**: `false`
- **Description**: Automatically fallback to available runtime
- **Example**: `--auto-fallback`

#### `-v, --version`
- **Type**: boolean
- **Description**: Show version information
- **Example**: `-v`

#### `-h, --help`
- **Type**: boolean
- **Description**: Show help message
- **Example**: `-h`

## Mode Options

### Push Mode

#### `--push-mode`
- **Type**: integer
- **Default**: From config or `1`
- **Values**: `1`, `2`
- **Description**: Push operation mode

**Mode 1**: `registry/image:tag`
- Simple push without project namespace
- Example: `harbor.com/nginx:latest`

**Mode 2**: `registry/project/image:tag`
- Push with smart project selection
- Priority: command line > config file > original image project
- Example: `harbor.com/production/nginx:latest`

### Save Mode

#### `--save-mode`
- **Type**: integer
- **Default**: From config or `1`
- **Values**: `1`, `2`, `3`
- **Description**: Save operation mode

**Mode 1**: Save to current directory
- Files: `./nginx_latest.tar`

**Mode 2**: Save to `./images/` directory
- Files: `./images/nginx_latest.tar`

**Mode 3**: Save to `./images/<project>/` directories
- Files: `./images/library/nginx_latest.tar`

### Load Mode

#### `--load-mode`
- **Type**: integer
- **Default**: From config or `1`
- **Values**: `1`, `2`, `3`
- **Description**: Load operation mode

**Mode 1**: Load from current directory
- Loads all `*.tar` files from current directory

**Mode 2**: Load from `./images/` directory
- Loads all `*.tar` files from `./images/`

**Mode 3**: Load recursively from `./images/*/` directories
- Recursively loads `*.tar` files from subdirectories

## Actions

### Pull Action

Pull images from registries.

**Syntax**:
```bash
hpn -a pull -f <image-list> [options]
```

**Required Parameters**:
- `-f, --file`: Image list file

**Optional Parameters**:
- `--runtime`: Container runtime
- `--auto-fallback`: Auto-fallback mode
- `-c, --config`: Configuration file

**Example**:
```bash
hpn -a pull -f images.txt --runtime docker
```

**Image List Format**:
```
nginx:latest
alpine:3.18
redis:7-alpine
```

### Save Action

Save images to tar files.

**Syntax**:
```bash
hpn -a save -f <image-list> [--save-mode <1|2|3>] [options]
```

**Required Parameters**:
- `-f, --file`: Image list file

**Optional Parameters**:
- `--save-mode`: Save mode (1, 2, or 3)
- `--runtime`: Container runtime
- `-c, --config`: Configuration file

**Examples**:
```bash
# Save to current directory
hpn -a save -f images.txt --save-mode 1

# Save to ./images/ directory
hpn -a save -f images.txt --save-mode 2

# Save to ./images/<project>/ directories
hpn -a save -f images.txt --save-mode 3
```

### Load Action

Load images from tar files.

**Syntax**:
```bash
hpn -a load [--load-mode <1|2|3>] [options]
```

**Required Parameters**:
- None (image list file not required)

**Optional Parameters**:
- `--load-mode`: Load mode (1, 2, or 3)
- `--runtime`: Container runtime
- `-c, --config`: Configuration file

**Examples**:
```bash
# Load from current directory
hpn -a load --load-mode 1

# Load from ./images/ directory
hpn -a load --load-mode 2

# Load recursively from ./images/*/ directories
hpn -a load --load-mode 3
```

### Push Action

Push images to registries.

**Syntax**:
```bash
hpn -a push -f <image-list> -r <registry> [-p <project>] [--push-mode <1|2>] [options]
```

**Required Parameters**:
- `-f, --file`: Image list file
- `-r, --registry`: Target registry

**Optional Parameters**:
- `-p, --project`: Target project
- `--push-mode`: Push mode (1 or 2)
- `--runtime`: Container runtime
- `-c, --config`: Configuration file

**Examples**:
```bash
# Simple push (mode 1)
hpn -a push -f images.txt -r harbor.company.com --push-mode 1

# Push with project (mode 2)
hpn -a push -f images.txt -r harbor.company.com -p production --push-mode 2

# Smart project selection (mode 2, no -p specified)
hpn -a push -f images.txt -r harbor.company.com --push-mode 2
```

## Version Command

Show detailed version information.

**Syntax**:
```bash
hpn version
```

**Output**:
```
Harpoon (hpn) v1.1
Commit: abc123
Built: 2024-12-19T10:30:00Z
Go version: go1.21.0
Platform: linux/amd64
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | Configuration error |
| 4 | Runtime error |
| 5 | Network error |
| 6 | File system error |

## Environment Variables

All configuration options can be set via environment variables with `HPN_` prefix:

### Registry Settings
```bash
export HPN_REGISTRY=harbor.company.com
export HPN_PROJECT=production
```

### Runtime Settings
```bash
export HPN_RUNTIME_PREFERRED=docker
export HPN_RUNTIME_AUTO_FALLBACK=true
export HPN_RUNTIME_TIMEOUT=10m
```

### Proxy Settings
```bash
export HPN_PROXY_HTTP=http://proxy:8080
export HPN_PROXY_HTTPS=http://proxy:8080
export HPN_PROXY_ENABLED=true

# Standard proxy variables
export http_proxy=http://proxy:8080
export https_proxy=http://proxy:8080
```

### Logging Settings
```bash
export HPN_LOG_LEVEL=debug
export HPN_LOG_FORMAT=json
export HPN_LOG_FILE=./hpn.log
```

## Configuration File Reference

### File Locations
1. `--config /path/to/config.yaml` (highest priority)
2. `~/.hpn/config.yaml`
3. `/etc/hpn/config.yaml`
4. `./config.yaml` (lowest priority)

### Configuration Schema
```yaml
# Registry settings
registry: string
project: string

# Runtime configuration
runtime:
  preferred: string        # docker|podman|nerdctl
  auto_fallback: boolean
  timeout: duration
  retry:
    max_attempts: integer  # 1-10
    delay: duration
    max_delay: duration

# Proxy settings
proxy:
  http: string
  https: string
  enabled: boolean

# Default modes
modes:
  save_mode: integer      # 1-3
  load_mode: integer      # 1-3
  push_mode: integer      # 1-2

# Logging
logging:
  level: string           # debug|info|warn|error
  format: string          # text|json
  file: string
  console: boolean
  timestamp: boolean
  colors: boolean

# Parallel processing
parallel:
  max_workers: integer    # 1-100
  auto_adjust: boolean
```

## Error Messages

### Common Error Patterns

#### Parameter Validation
```bash
Error: --save-mode cannot be used with push action
Error: invalid push-mode '3'. Valid values: 1, 2
Error: missing required -a <action> parameter
```

#### Runtime Errors
```bash
Error: no container runtime found
Error: specified runtime 'docker' is not available
Error: container runtime detection failed
```

#### Network Errors
```bash
Error: failed to pull image: connection refused
Error: authentication failed for registry
Error: context deadline exceeded
```

#### File System Errors
```bash
Error: failed to read image list: no such file or directory
Error: permission denied: ./images/
Error: no space left on device
```

## Usage Examples

### Basic Operations
```bash
# Pull images
hpn -a pull -f images.txt

# Save images
hpn -a save -f images.txt --save-mode 2

# Load images
hpn -a load --load-mode 2

# Push images
hpn -a push -f images.txt -r harbor.com -p prod --push-mode 2
```

### Runtime Selection
```bash
# Auto-detect runtime
hpn -a pull -f images.txt

# Specify runtime
hpn --runtime podman -a pull -f images.txt

# Auto-fallback mode
hpn --auto-fallback -a pull -f images.txt
```

### Configuration
```bash
# Use custom config
hpn --config ./custom-config.yaml -a pull -f images.txt

# Override with environment variables
HPN_REGISTRY=harbor.com hpn -a push -f images.txt
```

### Advanced Usage
```bash
# Complex push operation
hpn -a push -f images.txt \
  -r harbor.company.com \
  -p production \
  --push-mode 2 \
  --runtime docker \
  --config ~/.hpn/prod-config.yaml
```

## Integration Examples

### Shell Scripts
```bash
#!/bin/bash
set -e

# Configuration
REGISTRY="harbor.company.com"
PROJECT="production"
IMAGE_LIST="production-images.txt"

# Pull and push
hpn -a pull -f "$IMAGE_LIST"
hpn -a push -f "$IMAGE_LIST" -r "$REGISTRY" -p "$PROJECT" --push-mode 2
```

### Makefile
```makefile
.PHONY: deploy
deploy:
	hpn -a pull -f images.txt
	hpn -a push -f images.txt -r $(REGISTRY) -p $(PROJECT) --push-mode 2
```

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

## See Also

- [User Guide](user-guide.md) - Complete usage guide
- [Configuration](configuration.md) - Configuration reference
- [Examples](examples.md) - Real-world examples
- [Troubleshooting](troubleshooting.md) - Common issues