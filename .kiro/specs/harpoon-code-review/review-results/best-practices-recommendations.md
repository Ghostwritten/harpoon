# Harpoon项目最佳实践建议

## 概述

基于对Harpoon项目的全面代码审查结果，本文档制定了针对性的最佳实践建议，涵盖编码规范、架构改进、测试策略和开发流程优化四个核心领域。这些建议旨在提升代码质量、增强系统可维护性，并建立可持续的开发实践。

## 1. 编码规范建议

### 1.1 Go语言编码标准

#### 代码格式化规范

**强制要求**:
```bash
# 所有代码必须通过gofmt格式化
gofmt -s -w .

# 使用goimports管理导入语句
goimports -w .

# 导入顺序规范
import (
    // 标准库
    "context"
    "fmt"
    "os"
    
    // 第三方库
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    
    // 本地包
    "github.com/harpoon/hpn/internal/config"
    "github.com/harpoon/hpn/pkg/types"
)
```

**自动化工具配置**:
```yaml
# .golangci.yml
linters-settings:
  gofmt:
    simplify: true
  goimports:
    local-prefixes: github.com/harpoon/hpn
  
linters:
  enable:
    - gofmt
    - goimports
    - govet
    - ineffassign
    - misspell
```

#### 命名规范

**包命名**:
```go
// ✅ 推荐：简洁、小写、单数
package config
package runtime
package errors

// ❌ 避免：复数、下划线、大写
package configs
package container_runtime
package ConfigManager
```

**函数和方法命名**:
```go
// ✅ 推荐：动词开头，驼峰命名
func LoadConfig() (*Config, error)
func ValidateImage(image string) error
func (r *DockerRuntime) IsAvailable() bool

// ❌ 避免：缩写、下划线
func load_cfg() (*Config, error)
func chk_img(img string) error
```

**常量和变量命名**:
```go
// ✅ 推荐：常量使用驼峰或全大写
const (
    DefaultTimeout = 5 * time.Minute
    MAX_WORKERS    = 10
)

// ✅ 推荐：变量使用驼峰
var (
    configManager *config.Manager
    runtimeDetector *runtime.Detector
)

// ❌ 避免：魔法数字
if pushMode < 1 || pushMode > 2 {  // 应该定义常量
```

#### 注释规范

**包级别注释**:
```go
// Package config provides configuration management functionality
// for the Harpoon container image management tool.
//
// It supports loading configuration from multiple sources including
// files, environment variables, and command-line flags.
package config
```

**公共API注释**:
```go
// ContainerRuntime defines the interface for container runtime operations.
// Implementations must be thread-safe and support context cancellation.
type ContainerRuntime interface {
    // Name returns the human-readable name of the runtime.
    Name() string
    
    // IsAvailable checks if the runtime is installed and accessible.
    IsAvailable() bool
    
    // Pull downloads the specified image with the given options.
    // It returns an error if the pull operation fails.
    Pull(ctx context.Context, image string, options PullOptions) error
}
```

**复杂逻辑注释**:
```go
func generateTarFilename(image string) string {
    // Replace filesystem-unsafe characters with underscores
    // Docker image names can contain '/' and ':' which are not safe for filenames
    filename := strings.ReplaceAll(image, "/", "_")
    filename = strings.ReplaceAll(filename, ":", "_")
    
    // Add .tar extension for clarity
    return filename + ".tar"
}
```

### 1.2 错误处理规范

#### 统一错误处理模式

**自定义错误类型使用**:
```go
// ✅ 推荐：使用自定义错误类型
func LoadConfig(path string) (*Config, error) {
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return nil, errors.New(errors.ErrConfigNotFound, 
            fmt.Sprintf("config file not found: %s", path))
    }
    
    // 处理逻辑...
    
    if err := validateConfig(config); err != nil {
        return nil, errors.Wrap(err, errors.ErrConfigValidation, 
            "config validation failed")
    }
    
    return config, nil
}

// ❌ 避免：直接使用fmt.Errorf
func LoadConfig(path string) (*Config, error) {
    return nil, fmt.Errorf("failed to load config: %v", err)
}
```

**错误上下文丰富化**:
```go
// ✅ 推荐：提供丰富的错误上下文
func (d *DockerRuntime) Pull(ctx context.Context, image string, options PullOptions) error {
    cmd := exec.CommandContext(ctx, d.command, "pull", image)
    
    if err := cmd.Run(); err != nil {
        return errors.Wrap(err, errors.ErrRuntimeCommand, 
            "failed to pull image").
            WithContext("image", image).
            WithContext("runtime", d.Name()).
            WithContext("command", strings.Join(cmd.Args, " "))
    }
    
    return nil
}
```

#### 错误处理最佳实践

**错误检查模式**:
```go
// ✅ 推荐：立即检查错误
config, err := LoadConfig(configPath)
if err != nil {
    return fmt.Errorf("failed to load configuration: %w", err)
}

// ✅ 推荐：使用errors.Is和errors.As
if errors.Is(err, errors.ErrConfigNotFound) {
    // 处理配置文件不存在的情况
}

var harpoonErr *errors.HarpoonError
if errors.As(err, &harpoonErr) {
    // 处理自定义错误类型
    log.Printf("Error code: %d, Context: %v", harpoonErr.Code, harpoonErr.Context)
}
```

### 1.3 并发编程规范

#### Context使用规范

**Context传递**:
```go
// ✅ 推荐：Context作为第一个参数
func ProcessImages(ctx context.Context, images []string, options ProcessOptions) error {
    for _, image := range images {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            if err := processImage(ctx, image, options); err != nil {
                return err
            }
        }
    }
    return nil
}

// ✅ 推荐：使用context.WithTimeout
func (d *DockerRuntime) Pull(ctx context.Context, image string, options PullOptions) error {
    ctx, cancel := context.WithTimeout(ctx, options.Timeout)
    defer cancel()
    
    // 执行pull操作
    return d.executePull(ctx, image, options)
}
```

#### 并发安全模式

**工作池模式**:
```go
// 推荐的并发处理模式
type WorkerPool struct {
    maxWorkers int
    jobs       chan Job
    results    chan Result
    wg         sync.WaitGroup
}

func (wp *WorkerPool) Start(ctx context.Context) {
    for i := 0; i < wp.maxWorkers; i++ {
        wp.wg.Add(1)
        go wp.worker(ctx)
    }
}

func (wp *WorkerPool) worker(ctx context.Context) {
    defer wp.wg.Done()
    
    for {
        select {
        case <-ctx.Done():
            return
        case job, ok := <-wp.jobs:
            if !ok {
                return
            }
            result := job.Process()
            wp.results <- result
        }
    }
}
```

## 2. 架构改进建议

### 2.1 分层架构优化

#### 清晰的依赖方向

**推荐的架构分层**:
```
┌─────────────────────────────────────┐
│           Presentation Layer        │  cmd/
├─────────────────────────────────────┤
│           Application Layer         │  internal/service/
├─────────────────────────────────────┤
│            Domain Layer             │  internal/domain/
├─────────────────────────────────────┤
│         Infrastructure Layer        │  internal/runtime/, internal/config/
├─────────────────────────────────────┤
│            Common Layer             │  pkg/
└─────────────────────────────────────┘
```

**依赖注入模式**:
```go
// 推荐的依赖注入结构
type Application struct {
    configManager   config.Manager
    runtimeDetector runtime.Detector
    imageService    service.ImageService
    logger          logger.Logger
}

func NewApplication(deps Dependencies) *Application {
    return &Application{
        configManager:   deps.ConfigManager,
        runtimeDetector: deps.RuntimeDetector,
        imageService:    deps.ImageService,
        logger:          deps.Logger,
    }
}

// 依赖接口定义
type Dependencies struct {
    ConfigManager   config.Manager
    RuntimeDetector runtime.Detector
    ImageService    service.ImageService
    Logger          logger.Logger
}
```

### 2.2 接口设计原则

#### 接口隔离原则

**细粒度接口设计**:
```go
// ✅ 推荐：职责单一的接口
type ImagePuller interface {
    Pull(ctx context.Context, image string, options PullOptions) error
}

type ImageSaver interface {
    Save(ctx context.Context, image string, tarPath string) error
}

type ImageLoader interface {
    Load(ctx context.Context, tarPath string) error
}

// 组合接口
type ContainerRuntime interface {
    ImagePuller
    ImageSaver
    ImageLoader
    RuntimeInfo
}

// ❌ 避免：过大的接口
type ContainerRuntime interface {
    Pull(ctx context.Context, image string, options PullOptions) error
    Save(ctx context.Context, image string, tarPath string) error
    Load(ctx context.Context, tarPath string) error
    Push(ctx context.Context, image string, options PushOptions) error
    Tag(ctx context.Context, source, target string) error
    Version() (string, error)
    Name() string
    IsAvailable() bool
    // ... 更多方法
}
```

#### 接口版本管理

**向后兼容的接口演进**:
```go
// V1接口
type ContainerRuntimeV1 interface {
    Pull(ctx context.Context, image string) error
    Save(ctx context.Context, image string, tarPath string) error
}

// V2接口 - 扩展而不破坏
type ContainerRuntimeV2 interface {
    ContainerRuntimeV1
    PullWithOptions(ctx context.Context, image string, options PullOptions) error
    SaveWithOptions(ctx context.Context, image string, tarPath string, options SaveOptions) error
}
```

### 2.3 配置管理架构

#### 分层配置模式

**配置优先级**:
```go
type ConfigManager struct {
    sources []ConfigSource
}

type ConfigSource interface {
    Load() (*Config, error)
    Priority() int
}

// 配置源优先级（数字越大优先级越高）
const (
    PriorityDefault     = 0  // 默认配置
    PriorityConfigFile  = 10 // 配置文件
    PriorityEnvironment = 20 // 环境变量
    PriorityCommandLine = 30 // 命令行参数
)

func (cm *ConfigManager) LoadConfig() (*Config, error) {
    // 按优先级合并配置
    config := &Config{}
    
    // 排序配置源
    sort.Slice(cm.sources, func(i, j int) bool {
        return cm.sources[i].Priority() < cm.sources[j].Priority()
    })
    
    // 依次应用配置
    for _, source := range cm.sources {
        sourceConfig, err := source.Load()
        if err != nil {
            continue // 忽略加载失败的配置源
        }
        config = mergeConfig(config, sourceConfig)
    }
    
    return config, nil
}
```

#### 配置验证架构

**分层验证模式**:
```go
type ConfigValidator interface {
    Validate(config *Config) error
}

type ValidationChain struct {
    validators []ConfigValidator
}

func (vc *ValidationChain) Validate(config *Config) error {
    for _, validator := range vc.validators {
        if err := validator.Validate(config); err != nil {
            return err
        }
    }
    return nil
}

// 具体验证器
type RegistryValidator struct{}
func (rv *RegistryValidator) Validate(config *Config) error {
    if config.Registry == "" {
        return errors.New(errors.ErrConfigValidation, "registry cannot be empty")
    }
    return nil
}

type ProxyValidator struct{}
func (pv *ProxyValidator) Validate(config *Config) error {
    if config.Proxy.HTTP != "" {
        if _, err := url.Parse(config.Proxy.HTTP); err != nil {
            return errors.Wrap(err, errors.ErrConfigValidation, "invalid proxy URL")
        }
    }
    return nil
}
```

## 3. 测试策略和标准

### 3.1 测试分层策略

#### 测试金字塔

**测试分层比例**:
```
        /\
       /  \
      / E2E \     10% - 端到端测试
     /______\
    /        \
   /Integration\ 20% - 集成测试
  /____________\
 /              \
/   Unit Tests   \ 70% - 单元测试
/________________\
```

#### 单元测试标准

**测试文件组织**:
```go
// internal/config/config_test.go
package config

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestConfigManager_LoadConfig(t *testing.T) {
    tests := []struct {
        name        string
        configPath  string
        expectError bool
        expected    *Config
    }{
        {
            name:       "valid config file",
            configPath: "testdata/valid-config.yaml",
            expected: &Config{
                Registry: "registry.example.com",
                Project:  "test-project",
            },
        },
        {
            name:        "non-existent config file",
            configPath:  "testdata/non-existent.yaml",
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manager := NewManager()
            config, err := manager.LoadConfig(tt.configPath)
            
            if tt.expectError {
                require.Error(t, err)
                return
            }
            
            require.NoError(t, err)
            assert.Equal(t, tt.expected, config)
        })
    }
}
```

**测试覆盖率要求**:
```bash
# 最低覆盖率要求
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total

# 目标覆盖率
# - 单元测试: 80%+
# - 集成测试: 60%+
# - 总体覆盖率: 75%+
```

#### 集成测试标准

**容器运行时集成测试**:
```go
// internal/runtime/integration_test.go
// +build integration

package runtime

import (
    "context"
    "testing"
    "time"
)

func TestDockerRuntime_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    
    runtime := NewDockerRuntime()
    if !runtime.IsAvailable() {
        t.Skip("Docker runtime not available")
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    testImage := "alpine:latest"
    
    t.Run("pull image", func(t *testing.T) {
        err := runtime.Pull(ctx, testImage, PullOptions{})
        require.NoError(t, err)
    })
    
    t.Run("save image", func(t *testing.T) {
        tarPath := filepath.Join(t.TempDir(), "test.tar")
        err := runtime.Save(ctx, testImage, tarPath)
        require.NoError(t, err)
        
        // 验证文件存在
        _, err = os.Stat(tarPath)
        require.NoError(t, err)
    })
}
```

#### 端到端测试标准

**CLI端到端测试**:
```go
// e2e/cli_test.go
// +build e2e

package e2e

import (
    "os/exec"
    "testing"
)

func TestCLI_PullCommand(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        wantCode int
    }{
        {
            name:     "pull single image",
            args:     []string{"pull", "alpine:latest"},
            wantCode: 0,
        },
        {
            name:     "pull with invalid image",
            args:     []string{"pull", "invalid-image-name"},
            wantCode: 1,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := exec.Command("./hpn", tt.args...)
            err := cmd.Run()
            
            if tt.wantCode == 0 {
                require.NoError(t, err)
            } else {
                require.Error(t, err)
                if exitError, ok := err.(*exec.ExitError); ok {
                    assert.Equal(t, tt.wantCode, exitError.ExitCode())
                }
            }
        })
    }
}
```

### 3.2 测试工具和框架

#### 推荐的测试库

**基础测试库**:
```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
)
```

**Mock生成工具**:
```bash
# 安装mockery
go install github.com/vektra/mockery/v2@latest

# 生成mock
//go:generate mockery --name=ContainerRuntime --output=mocks
type ContainerRuntime interface {
    Pull(ctx context.Context, image string, options PullOptions) error
}
```

**基准测试标准**:
```go
// benchmarks/performance_test.go
func BenchmarkImageProcessing(b *testing.B) {
    images := generateTestImages(10)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        processImages(images)
    }
}

func BenchmarkParallelProcessing(b *testing.B) {
    images := generateTestImages(10)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        processImagesParallel(images, 4)
    }
}
```

### 3.3 测试数据管理

#### 测试数据组织

**目录结构**:
```
testdata/
├── configs/
│   ├── valid-config.yaml
│   ├── invalid-config.yaml
│   └── minimal-config.yaml
├── images/
│   ├── test-images.txt
│   └── large-images.txt
└── fixtures/
    ├── docker-output.json
    └── error-responses.json
```

**测试数据生成**:
```go
// testutil/generators.go
package testutil

func GenerateTestConfig() *types.Config {
    return &types.Config{
        Registry: "test-registry.com",
        Project:  "test-project",
        Runtime: types.RuntimeConfig{
            Preferred: "docker",
            Timeout:   300,
        },
    }
}

func GenerateTestImages(count int) []string {
    images := make([]string, count)
    for i := 0; i < count; i++ {
        images[i] = fmt.Sprintf("test/image-%d:latest", i)
    }
    return images
}
```

## 4. 开发流程优化

### 4.1 Git工作流规范

#### 分支管理策略

**Git Flow模式**:
```
main (生产分支)
├── develop (开发分支)
│   ├── feature/add-parallel-processing
│   ├── feature/improve-error-handling
│   └── feature/add-tests
├── release/v1.2.0 (发布分支)
├── hotfix/security-fix (热修复分支)
└── bugfix/fix-config-loading (缺陷修复分支)
```

**提交消息规范**:
```
<type>(<scope>): <subject>

<body>

<footer>

# 类型 (type)
feat:     新功能
fix:      缺陷修复
docs:     文档更新
style:    代码格式化
refactor: 重构
test:     测试相关
chore:    构建工具或辅助工具的变动

# 示例
feat(runtime): add parallel image processing support

Implement worker pool pattern to process multiple images concurrently.
This improves performance by 3-5x for batch operations.

- Add WorkerPool struct with configurable worker count
- Implement semaphore-based concurrency control
- Add progress tracking and cancellation support

Closes #123
```

#### 代码审查流程

**Pull Request模板**:
```markdown
## 变更描述
简要描述本次变更的内容和目的。

## 变更类型
- [ ] 新功能 (feature)
- [ ] 缺陷修复 (bugfix)
- [ ] 重构 (refactor)
- [ ] 文档更新 (docs)
- [ ] 测试改进 (test)

## 测试
- [ ] 添加了新的测试用例
- [ ] 所有现有测试通过
- [ ] 手动测试通过

## 检查清单
- [ ] 代码遵循项目编码规范
- [ ] 添加了必要的文档和注释
- [ ] 更新了相关的README或文档
- [ ] 考虑了向后兼容性
- [ ] 进行了安全性考虑

## 相关Issue
Closes #123
Related to #456
```

### 4.2 CI/CD流程优化

#### 完整的CI流水线

```yaml
# .github/workflows/ci.yml
name: Continuous Integration

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.20, 1.21]
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Cache dependencies
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    
    - name: Download dependencies
      run: go mod download
    
    - name: Verify dependencies
      run: go mod verify
    
    - name: Format check
      run: |
        gofmt -l . | tee /tmp/gofmt.out
        test ! -s /tmp/gofmt.out
    
    - name: Vet check
      run: go vet ./...
    
    - name: Lint check
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
    
    - name: Run unit tests
      run: go test -race -coverprofile=coverage.out ./...
    
    - name: Check test coverage
      run: |
        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        echo "Test coverage: $COVERAGE%"
        if (( $(echo "$COVERAGE < 75" | bc -l) )); then
          echo "Test coverage is below 75%"
          exit 1
        fi
    
    - name: Run integration tests
      run: go test -tags=integration ./...
    
    - name: Security scan
      run: |
        go install golang.org/x/vuln/cmd/govulncheck@latest
        govulncheck ./...
    
    - name: Build
      run: go build -v ./cmd/hpn

  e2e-test:
    runs-on: ubuntu-latest
    needs: test
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Build binary
      run: go build -o hpn ./cmd/hpn
    
    - name: Run E2E tests
      run: go test -tags=e2e ./e2e/...
```

#### 发布流水线

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      with:
        fetch-depth: 0
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run tests
      run: go test -v ./...
    
    - name: Build multi-platform binaries
      run: |
        GOOS=linux GOARCH=amd64 go build -o dist/hpn-linux-amd64 ./cmd/hpn
        GOOS=linux GOARCH=arm64 go build -o dist/hpn-linux-arm64 ./cmd/hpn
        GOOS=darwin GOARCH=amd64 go build -o dist/hpn-darwin-amd64 ./cmd/hpn
        GOOS=darwin GOARCH=arm64 go build -o dist/hpn-darwin-arm64 ./cmd/hpn
        GOOS=windows GOARCH=amd64 go build -o dist/hpn-windows-amd64.exe ./cmd/hpn
    
    - name: Generate checksums
      run: |
        cd dist
        sha256sum * > checksums.txt
    
    - name: Create release
      uses: softprops/action-gh-release@v1
      with:
        files: dist/*
        generate_release_notes: true
```

### 4.3 开发环境标准化

#### 开发工具配置

**VS Code配置** (`.vscode/settings.json`):
```json
{
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.testFlags": ["-v", "-race"],
    "go.coverOnSave": true,
    "go.coverageDecorator": {
        "type": "gutter"
    },
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
        "source.organizeImports": true
    }
}
```

**Pre-commit hooks** (`.pre-commit-config.yaml`):
```yaml
repos:
  - repo: local
    hooks:
      - id: go-fmt
        name: go fmt
        entry: gofmt
        args: [-w, -s]
        language: system
        files: \.go$
      
      - id: go-imports
        name: go imports
        entry: goimports
        args: [-w]
        language: system
        files: \.go$
      
      - id: go-vet
        name: go vet
        entry: go vet
        language: system
        files: \.go$
        pass_filenames: false
      
      - id: go-test
        name: go test
        entry: go test
        args: [-v, ./...]
        language: system
        files: \.go$
        pass_filenames: false
```

#### Makefile标准化

```makefile
# Makefile
.PHONY: build test lint fmt clean install

# 变量定义
BINARY_NAME=hpn
BUILD_DIR=dist
MAIN_PATH=./cmd/hpn

# 默认目标
all: fmt lint test build

# 构建
build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# 多平台构建
build-all:
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# 测试
test:
	go test -v -race -coverprofile=coverage.out ./...

# 集成测试
test-integration:
	go test -v -tags=integration ./...

# 端到端测试
test-e2e:
	go test -v -tags=e2e ./e2e/...

# 测试覆盖率
coverage: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 代码检查
lint:
	golangci-lint run

# 格式化
fmt:
	gofmt -s -w .
	goimports -w .

# 安全扫描
security:
	govulncheck ./...

# 清理
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# 安装
install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# 开发环境设置
dev-setup:
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	pre-commit install

# 依赖更新
deps-update:
	go get -u ./...
	go mod tidy

# 依赖审计
deps-audit:
	go list -json -m all | nancy sleuth
```

## 5. 质量保障体系

### 5.1 代码质量门禁

#### 质量指标要求

**代码质量指标**:
```yaml
# quality-gates.yml
quality_gates:
  test_coverage:
    minimum: 75%
    target: 85%
  
  code_complexity:
    cyclomatic_complexity: 10
    cognitive_complexity: 15
  
  code_duplication:
    maximum: 3%
  
  security:
    vulnerabilities: 0
    security_hotspots: 0
  
  maintainability:
    technical_debt_ratio: 5%
    code_smells: 0
```

#### 自动化质量检查

**SonarQube配置** (`sonar-project.properties`):
```properties
sonar.projectKey=harpoon-hpn
sonar.projectName=Harpoon
sonar.projectVersion=1.0

sonar.sources=.
sonar.exclusions=**/*_test.go,**/vendor/**,**/testdata/**

sonar.tests=.
sonar.test.inclusions=**/*_test.go
sonar.test.exclusions=**/vendor/**

sonar.go.coverage.reportPaths=coverage.out
```

### 5.2 性能监控标准

#### 性能基准

**性能指标定义**:
```go
// benchmarks/performance_standards.go
package benchmarks

// 性能基准标准
const (
    // 单个镜像操作时间限制
    MaxPullTime = 5 * time.Minute
    MaxSaveTime = 2 * time.Minute
    MaxLoadTime = 1 * time.Minute
    
    // 批量操作性能要求
    MinThroughput = 2 // 每分钟处理的镜像数
    MaxMemoryUsage = 500 * 1024 * 1024 // 500MB
    
    // 并发性能要求
    MinConcurrency = 4 // 最小并发数
    MaxConcurrency = 16 // 最大并发数
)

func BenchmarkPerformanceStandards(b *testing.B) {
    // 性能基准测试
}
```

#### 监控和告警

**性能监控配置**:
```yaml
# monitoring.yml
monitoring:
  metrics:
    - name: operation_duration
      type: histogram
      help: "Duration of image operations"
      labels: [operation, runtime, status]
    
    - name: memory_usage
      type: gauge
      help: "Memory usage during operations"
    
    - name: concurrent_operations
      type: gauge
      help: "Number of concurrent operations"
  
  alerts:
    - name: high_operation_duration
      condition: operation_duration > 300s
      severity: warning
    
    - name: high_memory_usage
      condition: memory_usage > 500MB
      severity: critical
```

## 6. 总结

### 6.1 实施优先级

**立即实施** (高优先级):
1. 代码格式化和基础规范
2. 错误处理标准化
3. 基础测试框架建立
4. CI/CD流水线配置

**近期实施** (中优先级):
1. 完整测试套件开发
2. 架构重构和优化
3. 性能监控体系
4. 开发环境标准化

**长期规划** (低优先级):
1. 高级架构模式应用
2. 完整的质量保障体系
3. 性能优化和调优
4. 开发流程持续改进

### 6.2 预期收益

**短期收益**:
- 代码质量显著提升
- 开发效率提高30%+
- 缺陷率降低50%+

**长期收益**:
- 建立可持续的开发实践
- 团队协作效率提升
- 系统可维护性大幅改善
- 为项目长期发展奠定基础

### 6.3 成功关键因素

1. **团队承诺**: 全团队对最佳实践的认同和执行
2. **工具支持**: 完善的开发工具和自动化流程
3. **持续改进**: 定期回顾和优化开发实践
4. **知识分享**: 团队内部的知识传递和培训

通过实施这些最佳实践建议，Harpoon项目可以建立起现代化的、可持续的开发实践体系，为项目的长期成功奠定坚实基础。