# GitHub Actions测试配置说明

## Docker Hub认证配置

### GitHub Secrets配置

在GitHub仓库的 Settings > Secrets and variables > Actions 中添加以下secrets：

```
DOCKER_HUB_USERNAME: ghostwritten
DOCKER_HUB_TOKEN: [Your Docker Hub Token - Configure in GitHub Secrets]
```

### 使用说明

1. **DOCKER_HUB_USERNAME**: 你的Docker Hub用户名 `ghostwritten`
2. **DOCKER_HUB_TOKEN**: 你的Docker Hub访问token，用于推送测试镜像
3. **目标仓库**: `docker.io/ghostwritten` - 测试镜像将推送到这个命名空间下

### 测试镜像命名规范

测试镜像将使用以下命名格式：
- `docker.io/ghostwritten/hpn-test-hello:latest`
- `docker.io/ghostwritten/hpn-test-alpine:latest`
- `docker.io/ghostwritten/hpn-test-{image-name}:{tag}`

所有测试镜像都会在测试完成后自动清理。

## Runtime配置优化

### 智能Runtime回退机制

当用户在配置文件 `~/.hpn/config.yaml` 中指定了特定的runtime，但该runtime不可用时：

```yaml
# 配置文件示例
runtime:
  preferred: docker  # 用户指定使用docker
  timeout: 5m
  auto_fallback: false  # 是否自动回退到其他runtime
```

### 用户交互流程

1. **检测配置的runtime不可用**
   ```
   ⚠️  配置的runtime 'docker' 不可用
   🔍 检测到可用的runtime: podman
   
   ❓ 是否使用 'podman' 替代 'docker'? (y/N): 
   ```

2. **用户选择**
   - 输入 `y` 或 `yes`: 使用podman继续执行
   - 输入 `n` 或 `no`: 退出并提示用户安装docker或修改配置
   - 直接回车: 默认为 `no`

3. **自动回退模式**（用于CI环境）
   ```bash
   hpn --auto-fallback -a pull -f images.txt
   ```
   或在配置文件中设置：
   ```yaml
   runtime:
     auto_fallback: true
   ```

### 实现要点

1. **用户友好的提示信息**
   - 清楚说明当前情况
   - 提供可用的替代方案
   - 给出明确的操作指导

2. **CI/CD环境支持**
   - 支持 `--auto-fallback` 参数
   - 支持环境变量 `HPN_AUTO_FALLBACK=true`
   - 在非交互环境中自动选择最佳可用runtime

3. **错误处理**
   - 如果没有任何可用runtime，提供安装指导
   - 记录runtime选择决策到日志
   - 提供详细的错误信息和解决建议

## 测试环境配置

### Ubuntu环境
- Docker: 预装
- Podman: 需要安装 `sudo apt-get install -y podman`
- Nerdctl: 需要手动安装

### macOS环境  
- Docker: 通过Docker Desktop
- Podman: 通过Homebrew `brew install podman`
- Nerdctl: 通过Homebrew `brew install nerdctl`

### Windows环境
- Docker: 通过Docker Desktop
- Podman: 通过官方安装包
- Nerdctl: 通过官方发布包

## 安全注意事项

1. **Token安全**
   - 仅在GitHub Secrets中存储token
   - 不要在代码或日志中暴露token
   - 定期轮换token

2. **测试镜像清理**
   - 所有测试镜像都会自动清理
   - 避免在公共仓库中留下测试垃圾
   - 使用明确的测试标识前缀

3. **权限控制**
   - Token仅具有推送权限
   - 限制在指定的命名空间内操作
   - 监控异常的推送活动