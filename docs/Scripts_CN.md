# Harpoon 🎯

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-v1.0-green.svg)](releases)
[![Shell](https://img.shields.io/badge/shell-bash-orange.svg)](README.md)

**Harpoon** 是一个强大的云原生容器镜像管理工具，专为 Kubernetes 环境设计。它提供了灵活的镜像拉取、保存、加载和推送功能，支持多种操作模式以适应不同的部署场景。

> 🚀 **未来规划**: Harpoon 将使用 Go 语言重写，提供 `hpn` 命令行工具，为云原生生态系统带来更强大的镜像管理能力。

## 🌟 特性

- **多容器运行时支持**: 兼容 Docker、Podman 和 Nerdctl
- **灵活的操作模式**: 每种操作都支持多种模式以适应不同场景
- **代理支持**: 内置 HTTP/HTTPS 代理配置
- **详细日志记录**: 彩色日志输出，支持文件记录
- **批量操作**: 支持批量镜像处理
- **私有仓库支持**: 完整的私有镜像仓库推送功能

## 📦 安装

```bash
# 克隆仓库
git clone https://github.com/yourusername/harpoon.git
cd harpoon

# 赋予执行权限
chmod +x images.sh
```

## 🚀 快速开始

### 基本用法

```bash
./images.sh -a <action> -f <image_list> [options]
```

### 参数说明

| 参数 | 描述 | 必需 |
|------|------|------|
| `-a, --action` | 操作类型: pull/save/load/push | ✅ |
| `-f, --file` | 镜像列表文件 | ✅ (pull/save/push) |
| `-r, --registry` | 目标仓库地址 (默认: registry.k8s.local) | ❌ |
| `-p, --project` | 目标项目命名空间 (默认: library) | ❌ |
| `--push-mode` | 推送模式 (1-3, 默认: 1) | ❌ |
| `--load-mode` | 加载模式 (1-3, 默认: 1) | ❌ |
| `--save-mode` | 保存模式 (1-3, 默认: 1) | ❌ |

## 📋 详细使用场景

### 场景 1: 基础镜像拉取和保存

**用例**: 为离线环境准备基础容器镜像

```bash
# 1. 创建镜像列表文件
cat > base-images.txt << EOF
nginx:1.21
redis:7.0
mysql:8.0
node:18-alpine
python:3.11-slim
EOF

# 2. 拉取镜像
./images.sh -a pull -f base-images.txt

# 3. 保存到当前目录
./images.sh -a save -f base-images.txt --save-mode 1

# 4. 查看生成的文件
ls -la *.tar
# 预期输出:
# docker.io_nginx_1.21.tar
# docker.io_redis_7.0.tar
# docker.io_mysql_8.0.tar
# ...
```

### 场景 2: Kubernetes 集群镜像管理

**用例**: 为 Kubernetes 集群准备系统镜像

```bash
# 1. 创建 k8s 组件镜像列表
cat > k8s-images.txt << EOF
k8s.gcr.io/kube-apiserver:v1.25.0
k8s.gcr.io/kube-controller-manager:v1.25.0
k8s.gcr.io/kube-scheduler:v1.25.0
k8s.gcr.io/kube-proxy:v1.25.0
k8s.gcr.io/pause:3.8
k8s.gcr.io/etcd:3.5.4-0
k8s.gcr.io/coredns/coredns:v1.9.3
EOF

# 2. 拉取镜像 (使用代理)
export http_proxy=http://192.168.21.101:7890
export https_proxy=http://192.168.21.101:7890
./images.sh -a pull -f k8s-images.txt

# 3. 按项目保存 (mode 3)
./images.sh -a save -f k8s-images.txt --save-mode 3

# 4. 目录结构
tree images/
# images/
# ├── kube-apiserver/
# │   └── k8s.gcr.io_kube-apiserver_kube-apiserver_v1.25.0.tar
# ├── kube-controller-manager/
# │   └── k8s.gcr.io_kube-controller-manager_kube-controller-manager_v1.25.0.tar
# └── ...
```

### 场景 3: 离线环境部署

**用例**: 在无网络环境中部署应用

```bash
# 在有网络的环境中准备
# 1. 应用镜像列表
cat > app-images.txt << EOF
nginx:1.21
postgresql:13
redis:7.0-alpine
busybox:1.35
EOF

# 2. 拉取并保存到 images 目录
./images.sh -a pull -f app-images.txt
./images.sh -a save -f app-images.txt --save-mode 2

# 3. 打包传输到离线环境
tar -czf offline-images.tar.gz images/

# 在离线环境中
# 4. 解压并加载
tar -xzf offline-images.tar.gz
./images.sh -a load --load-mode 2

# 5. 验证镜像加载
docker images | grep -E "(nginx|postgresql|redis|busybox)"
```

### 场景 4: 私有仓库推送

**用例**: 将镜像推送到企业私有仓库

```bash
# 1. 准备企业应用镜像
cat > enterprise-images.txt << EOF
nginx:1.21
redis:7.0
mysql:8.0
java:openjdk-17
EOF

# 2. 拉取公共镜像
./images.sh -a pull -f enterprise-images.txt

# 3. 推送到私有仓库 - 模式 1 (扁平结构)
./images.sh -a push -f enterprise-images.txt \
  -r harbor.company.com \
  --push-mode 1

# 结果: harbor.company.com/nginx:1.21
#       harbor.company.com/redis:7.0

# 4. 推送到私有仓库 - 模式 2 (项目命名空间)
./images.sh -a push -f enterprise-images.txt \
  -r harbor.company.com \
  -p production \
  --push-mode 2

# 结果: harbor.company.com/production/nginx:1.21
#       harbor.company.com/production/redis:7.0

# 5. 推送到私有仓库 - 模式 3 (保持原始项目路径)
./images.sh -a push -f enterprise-images.txt \
  -r harbor.company.com \
  --push-mode 3

# 结果: harbor.company.com/library/nginx:1.21 (如果原始是 docker.io/library/nginx:1.21)
```

### 场景 5: CI/CD 流水线集成

**用例**: 在 GitLab CI/CD 中使用

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

### 场景 6: 多架构镜像处理

**用例**: 处理 ARM64 和 AMD64 架构镜像

```bash
# 1. 多架构镜像列表
cat > multi-arch-images.txt << EOF
nginx:1.21
redis:7.0-alpine
node:18-alpine
python:3.11-slim
EOF

# 2. 拉取当前架构镜像
./images.sh -a pull -f multi-arch-images.txt

# 3. 按架构保存
mkdir -p images/amd64 images/arm64
./images.sh -a save -f multi-arch-images.txt --save-mode 2

# 4. 推送到支持多架构的私有仓库
./images.sh -a push -f multi-arch-images.txt \
  -r registry.internal.com \
  -p multi-arch \
  --push-mode 2
```

### 场景 7: 灾难恢复场景

**用例**: 快速恢复关键服务镜像

```bash
# 1. 创建关键服务镜像清单
cat > critical-images.txt << EOF
nginx:1.21
postgresql:13
redis:7.0
rabbitmq:3.11-management
elasticsearch:8.5.0
EOF

# 2. 定期备份镜像
./images.sh -a pull -f critical-images.txt
./images.sh -a save -f critical-images.txt --save-mode 2
cp -r images/ /backup/container-images-$(date +%Y%m%d)/

# 3. 灾难恢复时快速加载
./images.sh -a load --load-mode 2
docker images | grep -E "(nginx|postgresql|redis|rabbitmq|elasticsearch)"

# 4. 快速部署到新环境
./images.sh -a push -f critical-images.txt \
  -r disaster-recovery-registry.com \
  -p emergency \
  --push-mode 2
```

## 🔧 高级配置

### 代理设置

```bash
# 在脚本中或环境变量中设置
export http_proxy=http://proxy.company.com:8080
export https_proxy=http://proxy.company.com:8080

# 或者直接修改脚本中的默认值
http_proxy=${http_proxy:-"http://your-proxy:port"}
https_proxy=${https_proxy:-"http://your-proxy:port"}
```

### 容器运行时配置

脚本自动检测可用的容器运行时，优先级顺序：
1. Docker
2. Podman  
3. Nerdctl

对于 Nerdctl，脚本会自动添加 `--insecure-registry` 参数。

## 📊 操作模式详解

### Save 模式
- **模式 1**: 保存到当前目录（默认）
- **模式 2**: 保存到 `./images/` 目录
- **模式 3**: 保存到 `./images/<project>/` 按项目分类

### Load 模式
- **模式 1**: 从当前目录加载所有 `.tar` 文件（默认）
- **模式 2**: 从 `./images/` 目录加载所有 `.tar` 文件
- **模式 3**: 递归从 `./images/*/` 子目录加载 `.tar` 文件

### Push 模式
- **模式 1**: 推送为 `registry/image:tag`（默认）
- **模式 2**: 推送为 `registry/project/image:tag`
- **模式 3**: 保持原始项目路径推送

## 🐛 故障排除

### 常见问题

1. **权限问题**
   ```bash
   sudo chmod +x images.sh
   # 或者修改 docker 用户组权限
   sudo usermod -aG docker $USER
   ```

2. **代理连接问题**
   ```bash
   # 检查代理连接
   curl -I --proxy http://192.168.21.101:7890 https://docker.io
   ```

3. **磁盘空间不足**
   ```bash
   # 清理无用镜像
   docker system prune -a
   ```

4. **镜像拉取超时**
   ```bash
   # 增加 Docker 守护进程超时配置
   # /etc/docker/daemon.json
   {
     "max-concurrent-downloads": 3,
     "max-download-attempts": 5
   }
   ```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📋 待办事项

- [ ] Go 语言重写 (`hpn` 命令行工具)
- [ ] 支持镜像签名验证
- [ ] 添加镜像扫描功能
- [ ] 支持 OCI 格式
- [ ] 添加配置文件支持
- [ ] 实现并行处理
- [ ] 添加进度条显示
- [ ] 支持镜像增量同步

## 📜 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

感谢所有为容器生态系统做出贡献的开发者和项目。

---

**Harpoon** - 精准打击容器镜像管理难题 🎯
