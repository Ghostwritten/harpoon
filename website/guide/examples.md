# Examples

Real-world examples and use cases for Harpoon (hpn).

## Basic Examples

### Simple Image Management

#### Pull and Save Images
```bash
# Create image list
cat > web-images.txt << EOF
nginx:1.21-alpine
redis:7-alpine
postgres:15-alpine
EOF

# Pull images
hpn pull -f web-images.txt

# Save to ./images/ directory
hpn save -f web-images.txt --path ./images
```

#### Load and Push Images
```bash
# Load images from ./images/
hpn load --path ./images

# Push to private registry
hpn push -f web-images.txt --registry harbor.company.com --project production
```

### Configuration-based Workflow
```bash
# Create config file
mkdir -p ~/.hpn
cat > ~/.hpn/config.yaml << EOF
registry: harbor.company.com
project: production
paths:
  save_path: ./images
  load_path: ./images
EOF

# Use with default settings
hpn pull -f web-images.txt
hpn save -f web-images.txt
hpn push -f web-images.txt
```

### Passthrough Args (pull / save / push)
Pass extra flags to the underlying runtime by placing them after `--`. You can also set `runtime.extra_args` in config.

```bash
# Insecure registry or self-signed cert (e.g. podman/skopeo)
hpn pull -f images.txt -- --tls-verify=false
hpn push -f images.txt --registry harbor.company.com -- --tls-verify=false
```

Interpretation of these args is up to the underlying tool (docker/podman/nerdctl/skopeo).

### Get Image List from Helm Chart

Extract image references from a Helm chart (remote or local) and use the list with pull/save/push. Requires [Helm](https://helm.sh) CLI to be installed.

```bash
# From a remote chart (repo/name + version)
hpn extract --chart bitnami/nginx --version 15.0.0 -o images.txt
hpn pull -f images.txt
hpn save -f images.txt --path ./images

# From a local chart directory or .tgz
hpn extract --chart ./mychart -o images.txt
hpn extract --chart ./mychart-1.0.0.tgz -f values-prod.yaml -o images.txt

# With custom values files (passed to helm template)
hpn extract --chart bitnami/nginx --version 15.0.0 -f values.yaml -f prod.yaml -o images.txt

# Write to stdout (e.g. pipe to pull)
hpn extract --chart ./mychart -o - | hpn pull -f /dev/stdin
```

Output is one image per line, compatible with `hpn pull -f` and `hpn push -f`.

## Development Workflows

### Local Development Setup

#### Development Environment
```bash
# Development images
cat > dev-images.txt << EOF
node:18-alpine
python:3.11-alpine
golang:1.21-alpine
mysql:8.0
redis:7-alpine
EOF

# Pull development images
hpn pull -f dev-images.txt

# Save for offline development
hpn save -f dev-images.txt --path ./images
```

#### Offline Development
```bash
# On machine without internet
hpn load --path ./images

# Verify images are available
docker images | grep -E "(node|python|golang|mysql|redis)"
```

### Multi-environment Deployment

#### Environment-specific Images
```bash
# Production images
cat > prod-images.txt << EOF
nginx:1.21-alpine
app:v1.2.3
database:v2.1.0
cache:v1.0.5
EOF

# Staging images
cat > staging-images.txt << EOF
nginx:1.21-alpine
app:v1.2.3-rc1
database:v2.1.0-beta
cache:v1.0.5
EOF

# Deploy to production
hpn push -f prod-images.txt --registry harbor.company.com --project production

# Deploy to staging
hpn push -f staging-images.txt --registry harbor.company.com --project staging
```

## CI/CD Integration

### GitHub Actions Integration

#### Build and Push Pipeline
```yaml
# .github/workflows/deploy.yml
name: Build and Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup hpn
      run: |
        curl -L https://github.com/Ghostwritten/harpoon/releases/latest/download/hpn-linux-amd64 -o hpn
        chmod +x hpn
        sudo mv hpn /usr/local/bin/
    
    - name: Login to registry
      run: echo "${{ secrets.REGISTRY_TOKEN }}" | docker login harbor.company.com -u "${{ secrets.REGISTRY_USER }}" --password-stdin
    
    - name: Deploy images
      run: |
        hpn --auto-fallback push -f production-images.txt --registry harbor.company.com --project production
```

## Container Runtime Examples

### Docker Environment
```bash
# Ensure Docker is used
hpn --runtime docker pull -f images.txt

# With custom Docker configuration
export DOCKER_HOST=tcp://docker.company.com:2376
hpn --runtime docker pull -f images.txt
```

### Podman Environment
```bash
# Use Podman (rootless)
hpn --runtime podman pull -f images.txt

# With Podman socket
systemctl --user start podman.socket
export DOCKER_HOST=unix:///run/user/$(id -u)/podman/podman.sock
hpn --runtime podman pull -f images.txt
```

### Mixed Environment
```bash
# Auto-detect and fallback
hpn --auto-fallback pull -f images.txt

# With configuration
cat > ~/.hpn/config.yaml << EOF
runtime:
  preferred: docker
  auto_fallback: true
EOF

hpn pull -f images.txt
```

## Advanced Use Cases

### Image Migration

#### Docker Hub to Private Registry
```bash
# Source images from Docker Hub
cat > dockerhub-images.txt << EOF
library/nginx:latest
library/redis:latest
library/postgres:latest
EOF

# Pull from Docker Hub
hpn pull -f dockerhub-images.txt

# Push to private registry (preserves original project structure)
hpn push -f dockerhub-images.txt --registry harbor.company.com --project production
```

### Air-gapped Environments

#### Prepare Images for Air-gapped Deployment
```bash
# On internet-connected machine
cat > airgap-images.txt << EOF
nginx:1.21-alpine
postgres:15-alpine
redis:7-alpine
app:v1.0.0
EOF

# Pull and save
hpn pull -f airgap-images.txt
hpn save -f airgap-images.txt --path ./images

# Create archive
tar -czf airgap-images.tar.gz images/
```

#### Deploy in Air-gapped Environment
```bash
# On air-gapped machine
tar -xzf airgap-images.tar.gz

# Load images
hpn load --path ./images

# Push to local registry
hpn push -f airgap-images.txt --registry localhost:5000 --project apps
```

## See Also

- [Quick Start](/guide/quickstart) - Getting started guide
- [Building](/guide/building) - Build from source
- [Concurrency](/guide/concurrency) - Workers and parallel operations
