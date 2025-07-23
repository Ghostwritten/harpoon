# Harpoon (hpn) - Container Image Management Tool

Harpoon (`hpn`) 是一个强大的容器镜像管理工具，用于云原生环境中的镜像拉取、保存、加载和推送操作。它支持多种容器运行时，并提供灵活的操作模式。

## 目录

- [安装](#安装)
- [快速开始](#快速开始)
- [命令参考](#命令参考)
- [操作模式](#操作模式)
- [配置](#配置)
- [示例](#示例)

## 安装

### 从源码编译

```bash
git clone <repository-url>
cd harpoon-go
go build -o hpn cmd/hpn/main.go
```

### 二进制下载

从 [Releases](releases) 页面下载适合您系统的预编译二进制文件。

## 快速开始

1. 创建镜像列表文件 `images.txt`：
```
nginx:latest
redis:alpine
mysql:8.0
```

2. 拉取镜像：
```bash
hpn -a pull -f images.txt
```

3. 保存镜像到 tar 文件：
```bash
hpn -a save -f images.txt --save-mode 2
```

4. 加载镜像：
```bash
hpn -a load --load-mode 2
```

5. 推送到私有仓库：
```bash
hpn -a push -f images.txt -r registry.example.com -p myproject --push-mode 2
```

## 命令参考

### 基本语法

```bash
hpn -a <action> -f <image_list> [-r <registry>] [-p <project>] [--push-mode <1|2|3>] [--load-mode <1|2|3>] [--save-mode <1|2|3>]
```

### 必需参数

- `-a, --action`: 操作类型 (必需)
  - `pull`: 从外部仓库拉取镜像
  - `save`: 将镜像保存为 tar 文件
  - `load`: 从 tar 文件加载镜像
  - `push`: 推送镜像到私有仓库

### 可选参数

- `-f, --file`: 镜像列表文件 (pull/save/push 操作必需)
- `-r, --registry`: 目标仓库地址 (默认: registry.k8s.local)
- `-p, --project`: 目标项目命名空间 (默认: library)

### 全局选项

- `--config`: 配置文件路径
- `--log-level`: 日志级别 (debug, info, warn, error)
- `--log-file`: 日志文件路径
- `--quiet`: 静默模式
- `--output`: 输出格式 (text, json)
- `--runtime`: 指定容器运行时 (docker, podman, nerdctl)

## 操作模式

### Push 模式 (--push-mode)

- **模式 1** (默认): `registry/image:tag`
  - 推送格式：`registry.example.com/nginx:latest`
  
- **模式 2**: `registry/project/image:tag`
  - 推送格式：`registry.example.com/myproject/nginx:latest`
  
- **模式 3**: 保留原始项目路径
  - 推送格式：`registry.example.com/original-project/nginx:latest`

### Load 模式 (--load-mode)

- **模式 1** (默认): 从当前目录加载所有 `*.tar` 文件
- **模式 2**: 从 `./images/` 目录加载所有 `*.tar` 文件
- **模式 3**: 递归从 `./images/*/` 子目录加载 `*.tar` 文件

### Save 模式 (--save-mode)

- **模式 1** (默认): 保存 tar 文件到当前目录
- **模式 2**: 保存 tar 文件到 `./images/` 目录
- **模式 3**: 保存 tar 文件到 `./images/<project>/` 目录

## 配置

### 配置文件

默认配置文件位置：
- `~/.hpn/config.yaml`
- `/etc/hpn/config.yaml`

示例配置：

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

### 环境变量

所有配置选项都可以通过环境变量设置，使用 `HPN_` 前缀：

```bash
export HPN_REGISTRY=registry.example.com
export HPN_PROJECT=myproject
export HPN_PROXY_HTTP=http://proxy.example.com:8080
export HPN_LOG_LEVEL=debug
```

## 示例

### 基本操作示例

#### 1. 拉取镜像列表

创建 `k8s-images.txt`：
```
k8s.gcr.io/kube-apiserver:v1.28.0
k8s.gcr.io/kube-controller-manager:v1.28.0
k8s.gcr.io/kube-scheduler:v1.28.0
k8s.gcr.io/kube-proxy:v1.28.0
```

拉取镜像：
```bash
hpn -a pull -f k8s-images.txt
```

#### 2. 保存镜像到不同位置

保存到当前目录：
```bash
hpn -a save -f k8s-images.txt --save-mode 1
```

保存到 images 目录：
```bash
hpn -a save -f k8s-images.txt --save-mode 2
```

按项目分类保存：
```bash
hpn -a save -f k8s-images.txt --save-mode 3
```

#### 3. 加载镜像

从当前目录加载：
```bash
hpn -a load --load-mode 1
```

从 images 目录加载：
```bash
hpn -a load --load-mode 2
```

递归加载：
```bash
hpn -a load --load-mode 3
```

#### 4. 推送到私有仓库

简单推送：
```bash
hpn -a push -f k8s-images.txt -r harbor.example.com --push-mode 1
```

带项目命名空间推送：
```bash
hpn -a push -f k8s-images.txt -r harbor.example.com -p kubernetes --push-mode 2
```

保留原始路径推送：
```bash
hpn -a push -f k8s-images.txt -r harbor.example.com --push-mode 3
```

### 高级用法示例

#### 使用代理拉取镜像

```bash
HPN_PROXY_HTTP=http://proxy.example.com:8080 \
HPN_PROXY_HTTPS=http://proxy.example.com:8080 \
hpn -a pull -f images.txt
```

#### 指定容器运行时

```bash
hpn --runtime podman -a pull -f images.txt
```

#### JSON 输出格式

```bash
hpn --output json -a pull -f images.txt
```

#### 静默模式

```bash
hpn --quiet -a save -f images.txt --save-mode 2
```

### 批处理脚本示例

#### 完整的镜像迁移流程

```bash
#!/bin/bash

# 1. 拉取镜像
echo "拉取镜像..."
hpn -a pull -f production-images.txt

# 2. 保存镜像
echo "保存镜像..."
hpn -a save -f production-images.txt --save-mode 2

# 3. 传输到目标环境 (示例)
echo "传输镜像文件..."
rsync -av images/ target-server:/tmp/images/

# 4. 在目标环境加载镜像
echo "在目标环境加载镜像..."
ssh target-server "cd /tmp && hpn -a load --load-mode 2"

# 5. 推送到私有仓库
echo "推送到私有仓库..."
ssh target-server "hpn -a push -f /tmp/production-images.txt -r harbor.internal.com -p production --push-mode 2"
```

## 故障排除

### 常见问题

1. **容器运行时未找到**
   ```
   Error: No container runtime found. Please install docker, podman, or nerdctl.
   ```
   解决方案：安装至少一个支持的容器运行时。

2. **镜像拉取失败**
   ```
   Error: Failed to pull image nginx:latest: network timeout
   ```
   解决方案：检查网络连接或配置代理。

3. **磁盘空间不足**
   ```
   Error: Insufficient disk space for save operation
   ```
   解决方案：清理磁盘空间或选择其他保存位置。

4. **仓库认证失败**
   ```
   Error: Authentication failed for registry
   ```
   解决方案：配置正确的认证信息或使用 docker login。

### 调试模式

启用详细日志输出：
```bash
hpn --log-level debug -a pull -f images.txt
```

### 日志文件

查看详细日志：
```bash
hpn --log-file hpn.log -a pull -f images.txt
tail -f hpn.log
```

## 性能优化

### 并行处理

配置并行工作线程数：
```yaml
parallel:
  max_workers: 10
  adaptive: true
```

### 网络优化

配置代理和超时：
```yaml
proxy:
  http: http://proxy.example.com:8080
  https: http://proxy.example.com:8080
  enabled: true
runtime:
  timeout: 600s
```

## 贡献

欢迎贡献代码！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详细信息。

## 许可证

本项目采用 [MIT License](LICENSE) 许可证。