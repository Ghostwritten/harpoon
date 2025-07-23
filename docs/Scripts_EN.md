# Harpoon 🎯

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-v1.0-green.svg)](releases)
[![Shell](https://img.shields.io/badge/shell-bash-orange.svg)](README.md)

**Harpoon** is a powerful cloud-native container image management tool designed specifically for Kubernetes environments. It provides flexible image pulling, saving, loading, and pushing capabilities with multiple operation modes to accommodate different deployment scenarios.

> 🚀 **Future Roadmap**: Harpoon will be rewritten in Go, providing the `hpn` CLI tool to bring enhanced container image management capabilities to the cloud-native ecosystem.

## 🌟 Features

- **Multi-Container Runtime Support**: Compatible with Docker, Podman, and Nerdctl
- **Flexible Operation Modes**: Each operation supports multiple modes for different scenarios
- **Proxy Support**: Built-in HTTP/HTTPS proxy configuration
- **Comprehensive Logging**: Colorized log output with file logging support
- **Batch Operations**: Support for bulk image processing
- **Private Registry Support**: Complete private image registry push functionality

## 📦 Installation

```bash
# Clone the repository
git clone https://github.com/ghostwritten/harpoon.git
cd harpoon

# Grant execute permissions
chmod +x images.sh
```

## 🚀 Quick Start

### Basic Usage

```bash
./images.sh -a <action> -f <image_list> [options]
```

### Parameters

| Parameter | Description | Required |
|-----------|-------------|----------|
| `-a, --action` | Action type: pull/save/load/push | ✅ |
| `-f, --file` | Image list file | ✅ (for pull/save/push) |
| `-r, --registry` | Target registry address (default: registry.k8s.local) | ❌ |
| `-p, --project` | Target project namespace (default: library) | ❌ |
| `--push-mode` | Push mode (1-3, default: 1) | ❌ |
| `--load-mode` | Load mode (1-3, default: 1) | ❌ |
| `--save-mode` | Save mode (1-3, default: 1) | ❌ |

## 📋 Detailed Usage Scenarios

### Scenario 1: Basic Image Pull and Save

**Use Case**: Preparing base container images for offline environments

```bash
# 1. Create image list file
cat > base-images.txt << EOF
nginx:1.21
redis:7.0
mysql:8.0
node:18-alpine
python:3.11-slim
EOF

# 2. Pull images
./images.sh -a pull -f base-images.txt

# 3. Save to current directory
./images.sh -a save -f base-images.txt --save-mode 1

# 4. Check generated files
ls -la *.tar
# Expected output:
# docker.io_nginx_1.21.tar
# docker.io_redis_7.0.tar
# docker.io_mysql_8.0.tar
# ...
```

### Scenario 2: Kubernetes Cluster Image Management

**Use Case**: Preparing system images for Kubernetes clusters

```bash
# 1. Create k8s component image list
cat > k8s-images.txt << EOF
k8s.gcr.io/kube-apiserver:v1.25.0
k8s.gcr.io/kube-controller-manager:v1.25.0
k8s.gcr.io/kube-scheduler:v1.25.0
k8s.gcr.io/kube-proxy:v1.25.0
k8s.gcr.io/pause:3.8
k8s.gcr.io/etcd:3.5.4-0
k8s.gcr.io/coredns/coredns:v1.9.3
EOF

# 2. Pull images (using proxy)
export http_proxy=http://192.168.21.101:7890
export https_proxy=http://192.168.21.101:7890
./images.sh -a pull -f k8s-images.txt

# 3. Save by project (mode 3)
./images.sh -a save -f k8s-images.txt --save-mode 3

# 4. Directory structure
tree images/
# images/
# ├── kube-apiserver/
# │   └── k8s.gcr.io_kube-apiserver_kube-apiserver_v1.25.0.tar
# ├── kube-controller-manager/
# │   └── k8s.gcr.io_kube-controller-manager_kube-controller-manager_v1.25.0.tar
# └── ...
```

### Scenario 3: Air-Gapped Environment Deployment

**Use Case**: Deploying applications in environments without network access

```bash
# Prepare in networked environment
# 1. Application image list
cat > app-images.txt << EOF
nginx:1.21
postgresql:13
redis:7.0-alpine
busybox:1.35
EOF

# 2. Pull and save to images directory
./images.sh -a pull -f app-images.txt
./images.sh -a save -f app-images.txt --save-mode 2

# 3. Package for transfer to air-gapped environment
tar -czf offline-images.tar.gz images/

# In air-gapped environment
# 4. Extract and load
tar -xzf offline-images.tar.gz
./images.sh -a load --load-mode 2

# 5. Verify image loading
docker images | grep -E "(nginx|postgresql|redis|busybox)"
```

### Scenario 4: Private Registry Push

**Use Case**: Pushing images to enterprise private registry

```bash
# 1. Prepare enterprise application images
cat > enterprise-images.txt << EOF
nginx:1.21
redis:7.0
mysql:8.0
java:openjdk-17
EOF

# 2. Pull public images
./images.sh -a pull -f enterprise-images.txt

# 3. Push to private registry - Mode 1 (flat structure)
./images.sh -a push -f enterprise-images.txt \
  -r harbor.company.com \
  --push-mode 1

# Result: harbor.company.com/nginx:1.21
#         harbor.company.com/redis:7.0

# 4. Push to private registry - Mode 2 (project namespace)
./images.sh -a push -f enterprise-images.txt \
  -r harbor.company.com \
  -p production \
  --push-mode 2

# Result: harbor.company.com/production/nginx:1.21
#         harbor.company.com/production/redis:7.0

# 5. Push to private registry - Mode 3 (preserve original project path)
./images.sh -a push -f enterprise-images.txt \
  -r harbor.company.com \
  --push-mode 3

# Result: harbor.company.com/library/nginx:1.21 (if original was docker.io/library/nginx:1.21)
```

### Scenario 5: CI/CD Pipeline Integration

**Use Case**: Integration with GitLab CI/CD

```yaml
# .gitlab-ci.yml
stages:
  - prepare
  - deploy

prepare-images:
  stage: prepare
  script:
    - chmod +x images.sh
    - ./images.sh -a pull -f deployment/images.txt
    - ./images.sh -a save -f deployment/images.txt --save-mode 2
    - tar -czf images-${CI_COMMIT_SHA}.tar.gz images/
  artifacts:
    paths:
      - images-${CI_COMMIT_SHA}.tar.gz
    expire_in: 1 day

deploy-to-k8s:
  stage: deploy
  script:
    - tar -xzf images-${CI_COMMIT_SHA}.tar.gz
    - ./images.sh -a load --load-mode 2
    - ./images.sh -a push -f deployment/images.txt -r ${HARBOR_REGISTRY} -p ${PROJECT_NAME} --push-mode 2
    - kubectl apply -f k8s/
```

### Scenario 6: Multi-Architecture Image Handling

**Use Case**: Handling ARM64 and AMD64 architecture images

```bash
# 1. Multi-architecture image list
cat > multi-arch-images.txt << EOF
nginx:1.21
redis:7.0-alpine
node:18-alpine
python:3.11-slim
EOF

# 2. Pull current architecture images
./images.sh -a pull -f multi-arch-images.txt

# 3. Save by architecture
mkdir -p images/amd64 images/arm64
./images.sh -a save -f multi-arch-images.txt --save-mode 2

# 4. Push to multi-architecture private registry
./images.sh -a push -f multi-arch-images.txt \
  -r registry.internal.com \
  -p multi-arch \
  --push-mode 2
```

### Scenario 7: Disaster Recovery

**Use Case**: Rapid recovery of critical service images

```bash
# 1. Create critical service image manifest
cat > critical-images.txt << EOF
nginx:1.21
postgresql:13
redis:7.0
rabbitmq:3.11-management
elasticsearch:8.5.0
EOF

# 2. Regular image backup
./images.sh -a pull -f critical-images.txt
./images.sh -a save -f critical-images.txt --save-mode 2
cp -r images/ /backup/container-images-$(date +%Y%m%d)/

# 3. Quick loading during disaster recovery
./images.sh -a load --load-mode 2
docker images | grep -E "(nginx|postgresql|redis|rabbitmq|elasticsearch)"

# 4. Rapid deployment to new environment
./images.sh -a push -f critical-images.txt \
  -r disaster-recovery-registry.com \
  -p emergency \
  --push-mode 2
```

## 🔧 Advanced Configuration

### Proxy Settings

```bash
# Set in script or environment variables
export http_proxy=http://proxy.company.com:8080
export https_proxy=http://proxy.company.com:8080

# Or modify default values in script
http_proxy=${http_proxy:-"http://your-proxy:port"}
https_proxy=${https_proxy:-"http://your-proxy:port"}
```

### Container Runtime Configuration

The script automatically detects available container runtime with priority order:
1. Docker
2. Podman  
3. Nerdctl

For Nerdctl, the script automatically adds the `--insecure-registry` parameter.

## 📊 Operation Modes Explained

### Save Modes
- **Mode 1**: Save to current directory (default)
- **Mode 2**: Save to `./images/` directory
- **Mode 3**: Save to `./images/<project>/` organized by project

### Load Modes
- **Mode 1**: Load all `.tar` files from current directory (default)
- **Mode 2**: Load all `.tar` files from `./images/` directory
- **Mode 3**: Recursively load `.tar` files from `./images/*/` subdirectories

### Push Modes
- **Mode 1**: Push as `registry/image:tag` (default)
- **Mode 2**: Push as `registry/project/image:tag`
- **Mode 3**: Push preserving original project path

## 🐛 Troubleshooting

### Common Issues

1. **Permission Issues**
   ```bash
   sudo chmod +x images.sh
   # Or modify docker user group permissions
   sudo usermod -aG docker $USER
   ```

2. **Proxy Connection Issues**
   ```bash
   # Check proxy connection
   curl -I --proxy http://192.168.21.101:7890 https://docker.io
   ```

3. **Insufficient Disk Space**
   ```bash
   # Clean unused images
   docker system prune -a
   ```

4. **Image Pull Timeout**
   ```bash
   # Increase Docker daemon timeout configuration
   # /etc/docker/daemon.json
   {
     "max-concurrent-downloads": 3,
     "max-download-attempts": 5
   }
   ```

## 🤝 Contributing

Issues and Pull Requests are welcome!

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📋 TODO

- [ ] Rewrite in Go (`hpn` CLI tool)
- [ ] Support image signature verification
- [ ] Add image scanning functionality
- [ ] Support OCI format
- [ ] Add configuration file support
- [ ] Implement parallel processing
- [ ] Add progress bar display
- [ ] Support incremental image synchronization

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

Thanks to all developers and projects contributing to the container ecosystem.

---

**Harpoon** - Precision targeting for container image management challenges 🎯
