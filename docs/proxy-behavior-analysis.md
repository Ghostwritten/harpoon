# 代理环境变量行为分析

## 当前行为总结

### 1. 默认行为（未在配置文件中启用代理）

**代码位置**: `cmd/hpn/root.go:328-332`
```go
var proxyConfig *containerruntime.ProxyConfig
if cfg != nil && cfg.Proxy.Enabled {
    proxyConfig = cfg.Proxy.ToRuntimeProxyConfig()
}
```

**Runtime 实现**: `internal/runtime/docker.go:58-68` (类似逻辑适用于所有 runtime)
```go
// Set proxy environment if configured
if options.Proxy != nil && options.Proxy.Enabled {
    env := os.Environ()
    if options.Proxy.HTTP != "" {
        env = append(env, fmt.Sprintf("http_proxy=%s", options.Proxy.HTTP))
    }
    if options.Proxy.HTTPS != "" {
        env = append(env, fmt.Sprintf("https_proxy=%s", options.Proxy.HTTPS))
    }
    cmd.Env = env
}
```

### 2. 关键发现

**当 `options.Proxy == nil` 或 `options.Proxy.Enabled == false` 时**：
- 代码**不会设置 `cmd.Env`**
- Go 的 `exec.Command` 会**自动继承父进程的所有环境变量**
- 这意味着如果环境中存在 `http_proxy`、`https_proxy`、`HTTP_PROXY`、`HTTPS_PROXY` 等变量，子进程（docker/podman/nerdctl/skopeo）会**自动使用这些代理**

**结论**: ✅ **是的，默认情况下，如果环境中存在代理变量，拉取镜像会走代理**

### 3. 特殊情况：Skopeo Save 方法

**代码位置**: `internal/runtime/skopeo.go:113-114`
```go
env := os.Environ()
cmd.Env = env
```

**问题**: Skopeo 的 Save 方法总是设置 `cmd.Env`，即使没有配置代理。这实际上不会造成问题，因为 `os.Environ()` 会返回所有环境变量，所以行为是一致的。

## 环境变量继承机制

Go 的 `exec.Command` 环境变量继承规则：
1. **如果 `cmd.Env == nil`**：子进程继承父进程的所有环境变量
2. **如果 `cmd.Env != nil`**：子进程只使用 `cmd.Env` 中的环境变量

## 当前实现的行为

| 场景 | 配置文件代理状态 | 环境变量存在 | 实际行为 |
|------|----------------|------------|---------|
| 场景1 | `Enabled: false` 或未配置 | ✅ 存在 | ✅ **使用环境变量中的代理** |
| 场景2 | `Enabled: false` 或未配置 | ❌ 不存在 | ❌ 不使用代理 |
| 场景3 | `Enabled: true` | ✅ 存在 | ✅ 使用配置文件中的代理（会覆盖环境变量） |
| 场景4 | `Enabled: true` | ❌ 不存在 | ✅ 使用配置文件中的代理 |

## 建议

当前实现是合理的：
- ✅ 默认情况下会使用环境变量中的代理（符合用户期望）
- ✅ 配置文件可以覆盖环境变量（提供更精确的控制）
- ✅ 如果环境变量不存在，不会出错

## 测试验证

可以通过以下方式验证：

```bash
# 设置代理环境变量
export http_proxy=http://proxy.example.com:8080
export https_proxy=http://proxy.example.com:8080

# 拉取镜像（应该会使用代理）
./hpn pull -f images.txt

# 检查是否使用了代理（可以通过网络监控或代理日志验证）
```
