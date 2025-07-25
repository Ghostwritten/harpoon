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
hpn -a pull -f web-images.txt

# Save to ./images/ directory
hpn -a save -f web-images.txt --save-mode 2
```

#### Load and Push Images
```bash
# Load images from ./images/
hpn -a load --load-mode 2

# Push to private registry
hpn -a push -f web-images.txt -r harbor.company.com -p production --push-mode 2
```

### Configuration-based Workflow
```bash
# Create config file
mkdir -p ~/.hpn
cat > ~/.hpn/config.yaml << EOF
registry: harbor.company.com
project: production
modes:
  save_mode: 2
  load_mode: 2
  push_mode: 2
EOF

# Use with default settings
hpn -a pull -f web-images.txt
hpn -a save -f web-images.txt
hpn -a push -f web-images.txt
```

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
hpn -a pull -f dev-images.txt

# Save for offline development
hpn -a save -f dev-images.txt --save-mode 2
```

#### Offline Development
```bash
# On machine without internet
hpn -a load --load-mode 2

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
hpn -a push -f prod-images.txt -r harbor.company.com -p production --push-mode 2

# Deploy to staging
hpn -a push -f staging-images.txt -r harbor.company.com -p staging --push-mode 2
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
        curl -L https://github.com/your-org/harpoon/releases/latest/download/hpn-linux-amd64 -o hpn
        chmod +x hpn
        sudo mv hpn /usr/local/bin/
    
    - name: Login to registry
      run: echo "${{ secrets.REGISTRY_TOKEN }}" | docker login harbor.company.com -u "${{ secrets.REGISTRY_USER }}" --password-stdin
    
    - name: Deploy images
      run: |
        hpn --auto-fallback -a push -f production-images.txt -r harbor.company.com -p production --push-mode 2
```

#### Multi-stage Pipeline
```yaml
# .github/workflows/multi-stage.yml
name: Multi-stage Deploy

on:
  push:
    branches: [main, develop]

jobs:
  deploy-staging:
    if: github.ref == 'refs/heads/develop'
    runs-on: ubuntu-latest
    steps:
    - name: Deploy to staging
      run: |
        hpn -a push -f images.txt -r harbor.company.com -p staging --push-mode 2
  
  deploy-production:
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
    - name: Deploy to production
      run: |
        hpn -a push -f images.txt -r harbor.company.com -p production --push-mode 2
```

### Jenkins Integration

#### Jenkins Pipeline
```groovy
pipeline {
    agent any
    
    environment {
        REGISTRY = 'harbor.company.com'
        PROJECT = 'production'
    }
    
    stages {
        stage('Pull Images') {
            steps {
                sh 'hpn -a pull -f production-images.txt'
            }
        }
        
        stage('Save Images') {
            steps {
                sh 'hpn -a save -f production-images.txt --save-mode 2'
                archiveArtifacts artifacts: 'images/*.tar', fingerprint: true
            }
        }
        
        stage('Deploy') {
            when {
                branch 'main'
            }
            steps {
                withCredentials([usernamePassword(credentialsId: 'harbor-creds', usernameVariable: 'USER', passwordVariable: 'PASS')]) {
                    sh 'echo $PASS | docker login $REGISTRY -u $USER --password-stdin'
                    sh 'hpn -a push -f production-images.txt -r $REGISTRY -p $PROJECT --push-mode 2'
                }
            }
        }
    }
}
```

## Container Runtime Examples

### Docker Environment
```bash
# Ensure Docker is used
hpn --runtime docker -a pull -f images.txt

# With custom Docker configuration
export DOCKER_HOST=tcp://docker.company.com:2376
hpn --runtime docker -a pull -f images.txt
```

### Podman Environment
```bash
# Use Podman (rootless)
hpn --runtime podman -a pull -f images.txt

# With Podman socket
systemctl --user start podman.socket
export DOCKER_HOST=unix:///run/user/$(id -u)/podman/podman.sock
hpn --runtime podman -a pull -f images.txt
```

### Mixed Environment
```bash
# Auto-detect and fallback
hpn --auto-fallback -a pull -f images.txt

# With configuration
cat > ~/.hpn/config.yaml << EOF
runtime:
  preferred: docker
  auto_fallback: true
EOF

hpn -a pull -f images.txt
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
hpn -a pull -f dockerhub-images.txt

# Push to private registry (preserves original project structure)
hpn -a push -f dockerhub-images.txt -r harbor.company.com --push-mode 2
```

#### Cross-registry Migration
```bash
# Migration script
#!/bin/bash
set -e

SOURCE_REGISTRY="old-registry.com"
TARGET_REGISTRY="new-registry.com"
PROJECT="migrated"

# Create image list with source registry
sed "s|^|${SOURCE_REGISTRY}/|" base-images.txt > source-images.txt

# Pull from source
hpn -a pull -f source-images.txt

# Create target image list
sed "s|${SOURCE_REGISTRY}/|${TARGET_REGISTRY}/${PROJECT}/|" source-images.txt > target-images.txt

# Push to target
hpn -a push -f target-images.txt -r $TARGET_REGISTRY -p $PROJECT --push-mode 2
```

### Kubernetes Integration

#### Image Pre-pulling
```bash
# Extract images from Kubernetes manifests
kubectl get pods -o jsonpath='{.items[*].spec.containers[*].image}' | tr ' ' '\n' | sort -u > k8s-images.txt

# Pre-pull images on nodes
hpn -a pull -f k8s-images.txt
```

#### Helm Chart Images
```bash
# Extract images from Helm chart
helm template my-app ./chart | grep -oP 'image:\s*\K[^"]*' | sort -u > helm-images.txt

# Pull and save for air-gapped deployment
hpn -a pull -f helm-images.txt
hpn -a save -f helm-images.txt --save-mode 2
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
hpn -a pull -f airgap-images.txt
hpn -a save -f airgap-images.txt --save-mode 2

# Create archive
tar -czf airgap-images.tar.gz images/
```

#### Deploy in Air-gapped Environment
```bash
# On air-gapped machine
tar -xzf airgap-images.tar.gz

# Load images
hpn -a load --load-mode 2

# Push to local registry
hpn -a push -f airgap-images.txt -r localhost:5000 -p apps --push-mode 2
```

## Automation Scripts

### Batch Processing Script
```bash
#!/bin/bash
# batch-process.sh

set -e

REGISTRIES=("harbor.company.com" "registry.company.com")
PROJECTS=("production" "staging" "development")
IMAGE_LISTS=("web-images.txt" "api-images.txt" "db-images.txt")

for registry in "${REGISTRIES[@]}"; do
    for project in "${PROJECTS[@]}"; do
        for image_list in "${IMAGE_LISTS[@]}"; do
            echo "Processing $image_list for $registry/$project"
            
            # Pull images
            hpn -a pull -f "$image_list"
            
            # Push to registry/project
            hpn -a push -f "$image_list" -r "$registry" -p "$project" --push-mode 2
            
            echo "Completed $image_list for $registry/$project"
        done
    done
done
```

### Monitoring Script
```bash
#!/bin/bash
# monitor-images.sh

IMAGE_LIST="critical-images.txt"
REGISTRY="harbor.company.com"
PROJECT="production"

# Check if images exist in registry
while IFS= read -r image; do
    if hpn -a pull -f <(echo "$image") 2>/dev/null; then
        echo "✓ $image is available"
    else
        echo "✗ $image is missing - pulling and pushing"
        
        # Pull from source
        docker pull "$image"
        
        # Push to registry
        echo "$image" | hpn -a push -f - -r "$REGISTRY" -p "$PROJECT" --push-mode 2
    fi
done < "$IMAGE_LIST"
```

### Cleanup Script
```bash
#!/bin/bash
# cleanup-old-images.sh

# Remove images older than 30 days
docker images --format "table {{.Repository}}:{{.Tag}}\t{{.CreatedAt}}" | \
    awk '$2 < "'$(date -d '30 days ago' '+%Y-%m-%d')'" {print $1}' | \
    while read image; do
        echo "Removing old image: $image"
        docker rmi "$image" 2>/dev/null || true
    done

# Clean up tar files older than 7 days
find ./images -name "*.tar" -mtime +7 -delete
```

## Performance Optimization

### Parallel Processing
```bash
# High-performance configuration
cat > ~/.hpn/config.yaml << EOF
parallel:
  max_workers: 8
  auto_adjust: true

runtime:
  timeout: 15m
  retry:
    max_attempts: 5
    delay: 2s
    max_delay: 60s
EOF

# Process large image list
hpn -a pull -f large-image-list.txt
```

### Network Optimization
```bash
# Use local registry mirror
cat > ~/.hpn/config.yaml << EOF
registry: registry-mirror.company.com
proxy:
  enabled: false  # Disable if using local mirror
EOF

# Batch operations
hpn -a pull -f batch1.txt
hpn -a pull -f batch2.txt
hpn -a pull -f batch3.txt
```

## Troubleshooting Examples

### Debug Mode
```bash
# Enable debug logging
export HPN_LOG_LEVEL=debug
hpn -a pull -f images.txt 2>&1 | tee debug.log

# Analyze debug output
grep -i error debug.log
grep -i timeout debug.log
```

### Runtime Issues
```bash
# Test different runtimes
hpn --runtime docker -a pull -f test-image.txt
hpn --runtime podman -a pull -f test-image.txt

# Check runtime availability
docker version
podman version
nerdctl version
```

### Network Issues
```bash
# Test with proxy
export http_proxy=http://proxy:8080
export https_proxy=http://proxy:8080
hpn -a pull -f images.txt

# Test without proxy
unset http_proxy https_proxy
hpn -a pull -f images.txt
```

## See Also

- [User Guide](user-guide.md) - Complete usage guide
- [Configuration](configuration.md) - Configuration reference
- [Quick Start](quickstart.md) - Getting started guide
- [Troubleshooting](troubleshooting.md) - Common issues and solutions