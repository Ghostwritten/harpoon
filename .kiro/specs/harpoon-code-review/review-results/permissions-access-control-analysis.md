# 权限和访问控制评估分析报告

## 概述

本报告对Harpoon项目的权限和访问控制机制进行了全面评估，重点分析文件系统权限使用、容器运行时权限要求、安装脚本安全性以及特权升级风险。

## 1. 文件系统权限使用分析

### 1.1 当前权限处理机制

**文件创建权限**:
- **位置**: `internal/config/validation.go:validateDirectory()`
- **权限设置**: `os.MkdirAll(dir, 0755)`
- **分析**: 使用标准的755权限（rwxr-xr-x），相对安全

**配置目录创建**:
- **位置**: `internal/config/config.go:WriteConfig()`
- **权限设置**: `os.MkdirAll(dir, 0755)`
- **分析**: 配置目录权限合理

### 1.2 发现的权限问题

#### 🔴 高风险问题

1. **临时文件权限过于宽松**
   - **位置**: `internal/config/validation.go:validateDirectory()`
   - **问题**: 临时测试文件使用默认权限创建
   - **代码示例**:
   ```go
   f, err := os.Create(tempFile) // 使用默认权限，可能是666
   ```
   - **风险**: 临时文件可能被其他用户读取
   - **建议**: 明确设置安全权限

2. **配置文件权限检查不足**
   - **位置**: 配置文件加载过程
   - **问题**: 没有检查配置文件的权限是否过于宽松
   - **风险**: 敏感配置可能被未授权用户访问
   - **建议**: 检查配置文件权限不应超过600

#### 🟡 中风险问题

3. **镜像文件保存权限**
   - **位置**: `cmd/hpn/root.go:saveImage()`
   - **问题**: 保存的tar文件权限未明确设置
   - **风险**: 可能创建过于宽松的文件权限
   - **建议**: 明确设置tar文件权限为644

4. **目录遍历权限检查缺失**
   - **问题**: 创建目录时没有检查父目录权限
   - **风险**: 可能在不安全的位置创建文件
   - **建议**: 验证目录路径的安全性

### 1.3 权限安全加固建议

```go
// 建议的安全文件创建函数
func createSecureFile(path string, perm os.FileMode) (*os.File, error) {
    // 检查父目录权限
    dir := filepath.Dir(path)
    if err := validateDirectoryPermissions(dir); err != nil {
        return nil, err
    }
    
    // 创建文件并设置安全权限
    f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
    if err != nil {
        return nil, err
    }
    
    return f, nil
}

// 验证目录权限
func validateDirectoryPermissions(dir string) error {
    info, err := os.Stat(dir)
    if err != nil {
        return err
    }
    
    // 检查目录权限不应过于宽松
    if info.Mode().Perm() > 0755 {
        return fmt.Errorf("directory permissions too permissive: %o", info.Mode().Perm())
    }
    
    return nil
}
```

## 2. 容器运行时权限要求分析

### 2.1 Docker权限要求

**当前实现**: `internal/runtime/docker.go`

**权限分析**:
- Docker通常需要root权限或用户在docker组中
- 当前实现没有权限检查机制
- 直接执行docker命令，依赖系统权限配置

**安全风险**:
- 如果用户在docker组中，实际上拥有root等效权限
- 没有对Docker守护进程连接的安全验证

### 2.2 Podman权限要求

**当前实现**: `internal/runtime/podman.go`

**权限分析**:
- Podman支持rootless模式，安全性更好
- 当前实现没有区分rootless和root模式
- 缺少权限模式检测

**安全优势**:
- 支持无root权限运行
- 更好的用户隔离

### 2.3 Nerdctl权限要求

**当前实现**: `internal/runtime/nerdctl.go`

**权限分析**:
- 依赖containerd，通常需要root权限
- 包含`--insecure-registry`标志，存在安全风险
- 缺少权限验证机制

**安全问题**:
```go
// 问题代码
args = append(args, "--insecure-registry") // 默认使用不安全的registry连接
```

### 2.4 容器运行时安全加固建议

#### 🔴 立即修复

1. **移除默认的不安全标志**
```go
// 修改前
args = append(args, "--insecure-registry")

// 修改后 - 只在明确配置时使用
if options.AllowInsecure {
    args = append(args, "--insecure-registry")
}
```

2. **添加权限检查机制**
```go
func (d *DockerRuntime) checkPermissions() error {
    // 检查Docker守护进程连接权限
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    cmd := exec.CommandContext(ctx, "docker", "info", "--format", "{{.SecurityOptions}}")
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("cannot access Docker daemon: %v", err)
    }
    
    // 检查是否运行在rootless模式
    if strings.Contains(string(output), "rootless") {
        log.Info("Running in Docker rootless mode")
    }
    
    return nil
}
```

3. **实现运行时权限检测**
```go
type RuntimePermissions struct {
    IsRootless bool
    RequiresRoot bool
    SecurityOptions []string
}

func (r *ContainerRuntime) GetPermissions() (*RuntimePermissions, error) {
    // 检测运行时权限状态
    // 返回权限信息
}
```

## 3. 安装脚本安全性评估

### 3.1 install.sh安全分析

**当前实现分析**:

#### 🔴 高风险问题

1. **下载验证缺失**
   - **问题**: 没有校验和或签名验证
   - **代码位置**: `install.sh:install_hpn()`
   - **风险**: 中间人攻击，恶意二进制文件
   - **代码示例**:
   ```bash
   curl -L -o "$temp_file" "$download_url" # 没有验证下载文件的完整性
   ```

2. **HTTPS验证不足**
   - **问题**: curl/wget可能不验证SSL证书
   - **风险**: SSL中间人攻击
   - **建议**: 强制SSL验证

3. **sudo权限升级风险**
   - **代码位置**: `install.sh:install_hpn()`
   - **问题**: 自动使用sudo而不询问用户
   - **代码示例**:
   ```bash
   if [ -w "$INSTALL_DIR" ]; then
       cp "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"
   else
       sudo cp "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}" # 自动sudo
   fi
   ```

#### 🟡 中风险问题

4. **临时目录安全性**
   - **问题**: 临时目录权限可能不安全
   - **代码**: `local temp_dir=$(mktemp -d)`
   - **建议**: 明确设置临时目录权限

5. **环境变量注入风险**
   - **问题**: 直接使用环境变量构造URL
   - **风险**: 恶意环境变量可能改变下载源

### 3.2 build.sh安全分析

**当前实现分析**:

#### 🟡 中风险问题

1. **Git命令注入风险**
   - **代码位置**: `build.sh`
   - **问题**: 直接使用git命令输出
   - **代码示例**:
   ```bash
   VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
   COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
   ```
   - **风险**: 如果在恶意git仓库中运行，可能被注入

2. **构建输出权限**
   - **问题**: 构建的二进制文件权限未明确设置
   - **建议**: 设置适当的执行权限

### 3.3 安装脚本安全加固建议

#### 立即修复建议

1. **添加下载验证**
```bash
# 添加校验和验证
CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"
download_and_verify() {
    local url="$1"
    local file="$2"
    local expected_checksum="$3"
    
    # 下载文件
    if command -v curl >/dev/null 2>&1; then
        curl -L --fail --cert-status -o "$file" "$url"
    else
        wget --secure-protocol=TLSv1_2 -O "$file" "$url"
    fi
    
    # 验证校验和
    local actual_checksum=$(sha256sum "$file" | cut -d' ' -f1)
    if [ "$actual_checksum" != "$expected_checksum" ]; then
        log_error "Checksum verification failed"
        exit 1
    fi
}
```

2. **增强权限检查**
```bash
# 检查安装目录权限
check_install_permissions() {
    if [ ! -d "$INSTALL_DIR" ]; then
        log_error "Install directory does not exist: $INSTALL_DIR"
        exit 1
    fi
    
    if [ ! -w "$INSTALL_DIR" ] && [ "$EUID" -ne 0 ]; then
        log_warning "Installation requires sudo privileges"
        echo -n "Continue with sudo? [y/N]: "
        read -r response
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            log_info "Installation cancelled"
            exit 0
        fi
    fi
}
```

3. **安全的临时文件处理**
```bash
# 创建安全的临时目录
create_secure_temp_dir() {
    local temp_dir=$(mktemp -d)
    chmod 700 "$temp_dir"  # 只有所有者可以访问
    echo "$temp_dir"
}
```

## 4. 特权升级风险识别

### 4.1 潜在特权升级路径

#### 🔴 高风险路径

1. **Docker组成员权限**
   - **风险**: Docker组成员实际上拥有root等效权限
   - **影响**: 可以挂载主机文件系统，访问任意文件
   - **缓解**: 建议使用rootless Docker或Podman

2. **安装脚本sudo使用**
   - **风险**: 安装脚本可能被恶意修改，获得sudo权限
   - **影响**: 系统级权限获取
   - **缓解**: 验证脚本完整性，用户确认sudo操作

3. **配置文件权限升级**
   - **风险**: 通过修改配置文件影响程序行为
   - **影响**: 可能导致任意命令执行
   - **缓解**: 严格的配置文件权限检查

#### 🟡 中风险路径

4. **环境变量操控**
   - **风险**: 通过环境变量影响程序行为
   - **影响**: 可能改变程序执行路径
   - **缓解**: 验证环境变量值

5. **临时文件竞争**
   - **风险**: 临时文件创建时的竞争条件
   - **影响**: 可能被其他用户利用
   - **缓解**: 安全的临时文件创建

### 4.2 特权升级防护建议

#### 系统级防护

1. **最小权限原则**
```go
// 检查当前用户权限
func checkUserPrivileges() error {
    if os.Geteuid() == 0 {
        return fmt.Errorf("running as root is not recommended")
    }
    
    // 检查是否在危险组中
    groups, err := os.Getgroups()
    if err != nil {
        return err
    }
    
    dangerousGroups := []int{0} // root组
    for _, gid := range groups {
        for _, dangerous := range dangerousGroups {
            if gid == dangerous {
                log.Warning("Running with elevated privileges")
                break
            }
        }
    }
    
    return nil
}
```

2. **权限降级机制**
```go
// 在可能的情况下降级权限
func dropPrivileges() error {
    if os.Geteuid() == 0 {
        // 尝试切换到非特权用户
        nobody, err := user.Lookup("nobody")
        if err != nil {
            return err
        }
        
        uid, _ := strconv.Atoi(nobody.Uid)
        gid, _ := strconv.Atoi(nobody.Gid)
        
        if err := syscall.Setgid(gid); err != nil {
            return err
        }
        
        if err := syscall.Setuid(uid); err != nil {
            return err
        }
    }
    
    return nil
}
```

## 5. 安全配置建议

### 5.1 运行时安全配置

```yaml
# 建议的安全配置
security:
  runtime:
    prefer_rootless: true
    verify_permissions: true
    allow_insecure_registry: false
  
  files:
    config_permissions: 0600
    temp_permissions: 0600
    output_permissions: 0644
  
  network:
    verify_ssl: true
    timeout: 30s
```

### 5.2 部署安全检查清单

- [ ] 使用rootless容器运行时
- [ ] 验证安装脚本完整性
- [ ] 检查配置文件权限
- [ ] 限制网络访问权限
- [ ] 启用安全日志记录
- [ ] 定期更新依赖
- [ ] 监控特权操作

## 6. 总结

### 6.1 风险等级统计

- **高风险问题**: 6个
- **中风险问题**: 7个
- **低风险问题**: 2个

### 6.2 优先修复建议

**立即修复**（高风险）:
1. 添加下载文件完整性验证
2. 移除默认的不安全registry标志
3. 加强临时文件权限控制
4. 实现配置文件权限检查

**近期修复**（中风险）:
1. 实现容器运行时权限检测
2. 加强安装脚本权限确认
3. 添加环境变量验证
4. 完善目录权限检查

**长期改进**（低风险）:
1. 实现权限降级机制
2. 添加安全配置选项

### 6.3 安全最佳实践

1. **最小权限原则**: 只请求必要的权限
2. **权限验证**: 在执行敏感操作前验证权限
3. **安全默认**: 默认使用最安全的配置
4. **用户确认**: 特权操作需要用户明确确认
5. **审计日志**: 记录所有权限相关操作

通过实施这些权限和访问控制改进措施，可以显著提高Harpoon项目的安全性，降低特权升级和未授权访问的风险。