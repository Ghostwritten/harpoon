# 测试策略文档

## 概述
本文档描述了项目的测试策略，包括不同类型的测试、测试标签和执行策略。

## 测试分类

### 1. 单元测试 (Unit Tests)
- **目的**: 测试单个函数或方法的功能
- **执行**: `go test ./...`
- **覆盖率要求**: 最低70%，目标80%
- **特点**: 快速、独立、无外部依赖

### 2. 集成测试 (Integration Tests)
- **目的**: 测试组件间的交互
- **执行**: `go test -tags=integration ./...`
- **标签**: `//go:build integration`
- **特点**: 可能需要外部服务（如Docker）

### 3. 端到端测试 (E2E Tests)
- **目的**: 测试完整的用户场景
- **执行**: `go test -tags=e2e ./...`
- **标签**: `//go:build e2e`
- **特点**: 最接近真实使用场景

### 4. 发布测试 (Release Tests)
- **目的**: 验证发布版本的质量
- **执行**: `go test -tags=release ./...`
- **标签**: `//go:build release`
- **特点**: 包含性能和兼容性测试

### 5. 性能测试 (Benchmark Tests)
- **目的**: 测试性能和资源使用
- **执行**: `go test -bench=. ./...`
- **命名**: `BenchmarkXxx`
- **特点**: 测量执行时间和内存使用

## 测试标签使用

### 在测试文件中使用标签
```go
//go:build integration
// +build integration

package main

import "testing"

func TestIntegrationExample(t *testing.T) {
    // 集成测试代码
}
```

### 标签组合
```go
//go:build integration && !race
// +build integration,!race

// 只在集成测试时运行，但不在竞态检测时运行
```

## 分支特定测试策略

### Main分支
- **测试级别**: 完整测试套件
- **包含**: 所有类型的测试
- **平台**: 多平台 (Linux, Windows, macOS)
- **Go版本**: 多版本 (1.19, 1.20, 1.21)
- **额外检查**: 安全扫描、性能测试

### Develop分支
- **测试级别**: 综合测试
- **包含**: 单元测试、集成测试
- **平台**: 多平台
- **Go版本**: 当前版本 (1.21)
- **额外检查**: 代码质量检查

### Feature/Bugfix分支
- **测试级别**: 标准测试
- **包含**: 单元测试、基础集成测试
- **平台**: Linux (主要)
- **Go版本**: 当前版本 (1.21)
- **额外检查**: 基础代码质量

### Release分支
- **测试级别**: 发布测试
- **包含**: 所有测试 + 发布特定测试
- **平台**: 所有支持平台
- **Go版本**: 所有支持版本
- **额外检查**: 完整安全扫描、性能基准

### Hotfix分支
- **测试级别**: 关键测试
- **包含**: 核心功能测试、安全测试
- **平台**: Linux (快速验证)
- **Go版本**: 当前版本
- **额外检查**: 快速安全扫描

## 测试执行命令

### 本地开发
```bash
# 快速测试
make test

# 带覆盖率的测试
make test-coverage

# 集成测试
go test -tags=integration ./...

# 性能测试
go test -bench=. ./...

# 完整CI检查
make ci-check
```

### CI/CD环境
```bash
# 基础测试
go test -short ./...

# 标准测试
go test -race ./...

# 综合测试
go test -race -coverprofile=coverage.out ./...
go test -tags=integration ./...

# 发布测试
go test -tags=release -race ./...

# 关键测试（hotfix）
go test -short -run="Test.*Critical|Test.*Security|Test.*Core" ./...
```

## 测试环境配置

### 环境变量
```bash
# 测试模式
export TEST_MODE=unit|integration|e2e|release

# 测试超时
export TEST_TIMEOUT=10m

# 测试详细输出
export TEST_VERBOSE=true

# 跳过长时间运行的测试
export TEST_SHORT=true
```

### Docker测试环境
```yaml
# docker-compose.test.yml
version: '3.8'
services:
  test-registry:
    image: registry:2
    ports:
      - "5000:5000"
  
  test-app:
    build: .
    depends_on:
      - test-registry
    environment:
      - TEST_REGISTRY_URL=test-registry:5000
```

## 测试数据管理

### 测试文件结构
```
testdata/
├── fixtures/           # 测试固定数据
│   ├── config.yaml
│   └── image-list.txt
├── golden/            # 黄金文件测试
│   ├── output.golden
│   └── error.golden
└── mocks/             # 模拟数据
    ├── registry-response.json
    └── docker-output.txt
```

### 测试辅助函数
```go
// test/helpers.go
package test

import (
    "os"
    "path/filepath"
    "testing"
)

func LoadTestData(t *testing.T, filename string) []byte {
    data, err := os.ReadFile(filepath.Join("testdata", filename))
    if err != nil {
        t.Fatalf("Failed to load test data %s: %v", filename, err)
    }
    return data
}

func SkipIfShort(t *testing.T, reason string) {
    if testing.Short() {
        t.Skipf("Skipping in short mode: %s", reason)
    }
}
```

## 测试质量指标

### 覆盖率要求
- **最低要求**: 70%
- **目标覆盖率**: 80%
- **优秀覆盖率**: 90%+

### 性能基准
- **单元测试**: < 1秒
- **集成测试**: < 30秒
- **E2E测试**: < 5分钟
- **完整测试套件**: < 15分钟

### 质量门禁
- 所有测试必须通过
- 覆盖率不能降低
- 性能不能显著退化
- 安全扫描无高危问题

## 故障排除

### 常见问题
1. **测试超时**: 增加超时时间或优化测试
2. **竞态条件**: 使用 `-race` 标志检测
3. **平台差异**: 使用构建标签处理平台特定代码
4. **依赖问题**: 使用模拟或测试容器

### 调试技巧
```bash
# 详细输出
go test -v ./...

# 运行特定测试
go test -run TestSpecificFunction ./...

# 竞态检测
go test -race ./...

# 内存分析
go test -memprofile=mem.prof ./...

# CPU分析
go test -cpuprofile=cpu.prof ./...
```

## 持续改进

### 测试指标监控
- 测试执行时间趋势
- 覆盖率变化趋势
- 失败率统计
- 性能基准对比

### 定期审查
- 每月审查测试策略
- 季度更新测试工具
- 年度评估测试架构

这个测试策略确保了不同分支类型有适当的测试覆盖，同时平衡了测试质量和执行效率。