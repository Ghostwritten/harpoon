# Configuration Reference

Complete reference for Harpoon (hpn) configuration options.

## Configuration Sources

Configuration is loaded in the following priority order (highest to lowest):

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Configuration file**
4. **Default values** (lowest priority)

## Configuration File Locations

Harpoon searches for configuration files in this order:

1. `--config /path/to/config.yaml` (specified via flag)
2. `~/.hpn/config.yaml` (user config)
3. `/etc/hpn/config.yaml` (system config)
4. `./config.yaml` (local config)

## Configuration File Format

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

# Default operation modes
modes:
  save_mode: 2
  load_mode: 2
  push_mode: 2

# Logging configuration
logging:
  level: info
  format: text
  file: ./hpn.log
  console: true
  timestamp: true
  colors: true

# Parallel processing
parallel:
  max_workers: 4
  auto_adjust: true
```

## Configuration Options

### Registry Settings

#### `registry`
- **Type**: string
- **Default**: `registry.k8s.local`
- **Description**: Default target registry for push operations
- **Example**: `harbor.company.com`

#### `project`
- **Type**: string
- **Default**: `library`
- **Description**: Default project namespace for push operations
- **Example**: `production`

### Runtime Configuration

#### `runtime.preferred`
- **Type**: string
- **Default**: `""` (auto-detect)
- **Values**: `docker`, `podman`, `nerdctl`
- **Description**: Preferred container runtime
- **Example**: `docker`

#### `runtime.auto_fallback`
- **Type**: boolean
- **Default**: `false`
- **Description**: Automatically fallback to available runtime when preferred is unavailable
- **Example**: `true`

#### `runtime.timeout`
- **Type**: duration
- **Default**: `5m`
- **Description**: Timeout for runtime operations
- **Example**: `10m`

#### `runtime.retry.max_attempts`
- **Type**: integer
- **Default**: `3`
- **Range**: 1-10
- **Description**: Maximum retry attempts for failed operations
- **Example**: `5`

#### `runtime.retry.delay`
- **Type**: duration
- **Default**: `1s`
- **Description**: Initial delay between retries
- **Example**: `2s`

#### `runtime.retry.max_delay`
- **Type**: duration
- **Default**: `30s`
- **Description**: Maximum delay between retries
- **Example**: `60s`

### Proxy Settings

#### `proxy.http`
- **Type**: string
- **Default**: `""`
- **Description**: HTTP proxy URL
- **Example**: `http://proxy.company.com:8080`

#### `proxy.https`
- **Type**: string
- **Default**: `""`
- **Description**: HTTPS proxy URL
- **Example**: `http://proxy.company.com:8080`

#### `proxy.enabled`
- **Type**: boolean
- **Default**: `false`
- **Description**: Enable proxy support
- **Example**: `true`

### Operation Modes

#### `modes.save_mode`
- **Type**: integer
- **Default**: `1`
- **Values**: `1`, `2`, `3`
- **Description**: Default save mode
  - `1`: Save to current directory
  - `2`: Save to `./images/`
  - `3`: Save to `./images/<project>/`

#### `modes.load_mode`
- **Type**: integer
- **Default**: `1`
- **Values**: `1`, `2`, `3`
- **Description**: Default load mode
  - `1`: Load from current directory
  - `2`: Load from `./images/`
  - `3`: Load recursively from `./images/*/`

#### `modes.push_mode`
- **Type**: integer
- **Default**: `1`
- **Values**: `1`, `2`
- **Description**: Default push mode
  - `1`: `registry/image:tag`
  - `2`: `registry/project/image:tag` (smart project selection)

### Logging Configuration

#### `logging.level`
- **Type**: string
- **Default**: `info`
- **Values**: `debug`, `info`, `warn`, `error`
- **Description**: Log level
- **Example**: `debug`

#### `logging.format`
- **Type**: string
- **Default**: `text`
- **Values**: `text`, `json`
- **Description**: Log output format
- **Example**: `json`

#### `logging.file`
- **Type**: string
- **Default**: `""`
- **Description**: Log file path (empty = no file logging)
- **Example**: `./hpn.log`

#### `logging.console`
- **Type**: boolean
- **Default**: `true`
- **Description**: Enable console logging
- **Example**: `false`

#### `logging.timestamp`
- **Type**: boolean
- **Default**: `true`
- **Description**: Include timestamps in logs
- **Example**: `false`

#### `logging.colors`
- **Type**: boolean
- **Default**: `true`
- **Description**: Enable colored output
- **Example**: `false`

### Parallel Processing

#### `parallel.max_workers`
- **Type**: integer
- **Default**: `4`
- **Range**: 1-100
- **Description**: Maximum number of parallel workers
- **Example**: `8`

#### `parallel.auto_adjust`
- **Type**: boolean
- **Default**: `true`
- **Description**: Automatically adjust workers based on system resources
- **Example**: `false`

## Environment Variables

All configuration options can be overridden using environment variables with the `HPN_` prefix:

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

# Standard proxy variables (also supported)
export http_proxy=http://proxy:8080
export https_proxy=http://proxy:8080
```

### Logging Settings
```bash
export HPN_LOG_LEVEL=debug
export HPN_LOG_FORMAT=json
export HPN_LOG_FILE=./hpn.log
export HPN_LOG_CONSOLE=true
```

### Parallel Processing
```bash
export HPN_PARALLEL_MAX=8
export HPN_PARALLEL_AUTO=false
```

## Command-line Flags

All configuration options can be overridden using command-line flags:

### Basic Options
```bash
hpn -a pull -f images.txt \
  -r harbor.company.com \
  -p production \
  -c ~/.hpn/config.yaml
```

### Runtime Options
```bash
hpn --runtime docker \
  --auto-fallback \
  -a pull -f images.txt
```

### Mode Options
```bash
hpn -a save -f images.txt --save-mode 2
hpn -a load --load-mode 2
hpn -a push -f images.txt --push-mode 2
```

## Configuration Examples

### Development Environment
```yaml
# ~/.hpn/config.yaml
registry: localhost:5000
project: dev
runtime:
  preferred: docker
  auto_fallback: true
modes:
  save_mode: 1
  load_mode: 1
  push_mode: 1
logging:
  level: debug
  console: true
  colors: true
```

### Production Environment
```yaml
# /etc/hpn/config.yaml
registry: harbor.company.com
project: production
runtime:
  preferred: docker
  auto_fallback: true
  timeout: 10m
  retry:
    max_attempts: 5
    delay: 2s
    max_delay: 60s
proxy:
  http: http://proxy.company.com:8080
  https: http://proxy.company.com:8080
  enabled: true
modes:
  save_mode: 2
  load_mode: 2
  push_mode: 2
logging:
  level: info
  format: json
  file: /var/log/hpn.log
  console: false
parallel:
  max_workers: 8
  auto_adjust: true
```

### CI/CD Environment
```yaml
# ci-config.yaml
registry: registry.company.com
project: ci-builds
runtime:
  auto_fallback: true
  timeout: 15m
modes:
  push_mode: 2
logging:
  level: info
  format: json
  console: true
  colors: false
parallel:
  max_workers: 6
```

## Configuration Validation

Harpoon validates configuration on startup and will report errors for:

- Invalid values (e.g., negative timeouts)
- Out-of-range values (e.g., max_workers > 100)
- Invalid formats (e.g., malformed URLs)
- Missing required dependencies

### Validation Examples
```bash
# Check configuration
hpn --config ~/.hpn/config.yaml -a pull -f /dev/null

# Common validation errors
Error: invalid log level 'verbose'. Valid values: debug, info, warn, error
Error: max_workers must be between 1 and 100
Error: invalid proxy URL format
```

## Configuration Best Practices

### Security
- Store sensitive configuration in user config (`~/.hpn/config.yaml`)
- Use environment variables for secrets in CI/CD
- Avoid storing credentials in configuration files
- Use specific registry URLs, avoid wildcards

### Performance
- Set appropriate `max_workers` for your system
- Enable `auto_adjust` for dynamic workloads
- Use reasonable timeout values
- Configure proxy settings for corporate networks

### Maintainability
- Use descriptive project names
- Document custom configurations
- Use consistent naming conventions
- Version control your configuration files

### Environment-specific
- Use different configs for dev/staging/prod
- Override sensitive values with environment variables
- Use CI-specific configurations for automated builds
- Test configuration changes in non-production first

## See Also

- [User Guide](user-guide.md)
- [Quick Start](quickstart.md)
- [Examples](examples.md)
- [Troubleshooting](troubleshooting.md)