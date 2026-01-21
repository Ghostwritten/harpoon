---
name: CLI 架构重构：子命令化与参数优化
overview: 将 hpn 从单一命令+action 模式重构为子命令模式（hpn pull/save/load/push），移除 mode 参数，改用灵活的路径和镜像命名参数，将操作特定参数移到对应子命令中，符合业界 CLI 设计标准。
status: completed
date: 2025-01-20
todos:
  - id: create_subcommands
    content: 创建子命令结构：pull, save, load, push 子命令
    status: completed
  - id: create_image_utils
    content: 创建镜像解析工具函数：parseImage（支持多级路径），buildTargetImage（实现三种命名模式）
    status: completed
  - id: refactor_pull_cmd
    content: 重构 Pull 子命令：移除 -a pull，创建 pull 子命令
    status: completed
    dependencies:
      - create_subcommands
  - id: refactor_save_cmd
    content: 重构 Save 子命令：移除 -a save 和 --save-mode，使用 --path 参数（支持多级路径）
    status: completed
    dependencies:
      - create_subcommands
  - id: refactor_load_cmd
    content: 重构 Load 子命令：移除 -a load 和 --load-mode，使用 --path 和 --recursive
    status: completed
    dependencies:
      - create_subcommands
  - id: refactor_push_cmd
    content: 重构 Push 子命令：移除 -a push 和 --push-mode，实现三种镜像命名模式（默认保持、追加路径、统一项目）
    status: completed
    dependencies:
      - create_subcommands
      - create_image_utils
  - id: update_root_cmd
    content: 更新 Root 命令：移除 action 和 mode 参数，移除 -r/-p（移到 push），保留全局参数
    status: completed
    dependencies:
      - refactor_pull_cmd
      - refactor_save_cmd
      - refactor_load_cmd
      - refactor_push_cmd
  - id: update_config_types
    content: 更新配置类型：移除 Mode 类型，添加路径字段
    status: completed
  - id: update_documentation
    content: 更新 usage template 和帮助信息，反映新的子命令结构和 Push 命名逻辑
    status: completed
    dependencies:
      - update_root_cmd
  - id: test_cli_refactor
    content: 测试新的 CLI 结构：子命令、参数作用域、多级路径、Push 三种命名模式
    status: completed
    dependencies:
      - update_documentation
---

# CLI 架构重构：子命令化与参数优化

## 概述

将 hpn 从单一命令+action 模式重构为标准的子命令模式（类似 docker、kubectl），移除复杂的 mode 参数，改用直观的路径和镜像命名参数，将操作特定参数移到对应子命令中，提升用户体验和可维护性。

## 核心改进点

### 1. 子命令化（移除 -a 参数）

**当前方式**：

```bash
hpn -a pull -f images.txt
hpn -a save -f images.txt --save-mode 2
```

**改进后**：

```bash
hpn pull -f images.txt
hpn save -f images.txt --path ./images
```

**业界标准**：

- Docker: `docker pull`, `docker push`, `docker save`, `docker load`
- Kubectl: `kubectl get`, `kubectl apply`, `kubectl delete`
- Skopeo: `skopeo copy`, `skopeo inspect`

### 2. 移除 Mode 参数，改用路径参数

**Save/Load 操作**：

- 移除 `--save-mode` 和 `--load-mode`
- 添加 `--path` 或 `--output` 参数（save）
- 添加 `--path` 或 `--input` 参数（load）
- **默认路径**：`./images`
- **多级路径支持**：`--path` 支持多级路径，如 `--path ./output/images/prod`，自动创建所有必要的目录层级

**Push 操作**：

- 移除 `--push-mode`
- 改用灵活的镜像命名参数

### 3. Push 操作的镜像命名灵活性

**当前问题**：

- Mode 1: `registry/image:tag`（不包含项目名）
- Mode 2: `registry/project/image:tag`（包含项目名）
- 无法处理复杂场景，特别是 images.txt 中包含不同项目结构的镜像

**改进方案**：基于源镜像路径结构的智能命名

#### 核心原则

1. **默认行为（只指定 --registry）**：只替换镜像仓库，保持原有路径结构完全不变
2. **追加路径（--registry 包含路径）**：在原有路径前追加新路径
3. **统一项目（--registry + --project）**：所有镜像统一推送到指定项目下

#### 场景分析

| 场景 | 源镜像 | 命令 | 目标镜像 | 说明 |

|------|--------|------|---------|------|

| **场景1：默认保持路径** | `old-registry.com/project/nginx:latest` | `--registry new-registry.com` | `new-registry.com/project/nginx:latest` | 只替换仓库，路径不变 |

| **场景2：默认保持路径（无项目）** | `old-registry.com/nginx:latest` | `--registry new-registry.com` | `new-registry.com/nginx:latest` | 只替换仓库，无项目名保持不变 |

| **场景3：追加路径** | `old-registry.com/project/nginx:latest` | `--registry new-registry.com/path/xx` | `new-registry.com/path/xx/project/nginx:latest` | 在原有路径前追加新路径 |

| **场景4：统一项目** | `old-registry.com/project1/nginx:latest`<br>`old-registry.com/project2/redis:latest`<br>`old-registry.com/nginx:latest` | `--registry new-registry.com --project newproject` | `new-registry.com/newproject/nginx:latest`<br>`new-registry.com/newproject/redis:latest`<br>`new-registry.com/newproject/nginx:latest` | 所有镜像统一推送到 newproject 下 |

| **场景5：混合场景** | `registry.com/path/to/nginx:latest` | `--registry new-registry.com/base` | `new-registry.com/base/path/to/nginx:latest` | 追加路径，保持原有完整路径 |

#### 参数设计

**基础参数**：

- `--registry <registry>`: 目标仓库（必需）
  - 如果包含路径（如 `new-registry.com/path/xx`），则追加到源镜像路径前
  - 如果只是仓库（如 `new-registry.com`），则只替换仓库，保持源路径
- `--project <project>`: 统一项目名（可选）
  - 与 `--registry` 同时使用时，所有镜像统一推送到 `registry/project/image:tag`
  - 忽略源镜像的原有项目/路径结构

**高级参数**（可选，用于特殊场景）：

- `--target <image:tag>`: 完整目标镜像名（覆盖所有其他参数，用于单个镜像的特殊处理）

#### 命名逻辑实现

```go
// ImageParts 表示解析后的镜像组件
type ImageParts struct {
    Registry string // 仓库地址（如 registry.com）
    Path     string // 路径部分（如 project 或 path/to/project，可能为空）
    Name     string // 镜像名（如 nginx）
    Tag      string // 标签（如 latest）
}

// parseImage 解析镜像名为组件，支持多级路径
func parseImage(image string) ImageParts {
    // 示例：
    // old-registry.com/project/nginx:latest -> {Registry: "old-registry.com", Path: "project", Name: "nginx", Tag: "latest"}
    // old-registry.com/path/to/nginx:latest -> {Registry: "old-registry.com", Path: "path/to", Name: "nginx", Tag: "latest"}
    // nginx:latest -> {Registry: "", Path: "", Name: "nginx", Tag: "latest"}
}

// buildTargetImage 构建目标镜像名
func buildTargetImage(sourceImage, registry, project string) string {
    // 解析源镜像
    sourceParts := parseImage(sourceImage)
    
    // 场景1: 指定了 --project，统一替换项目
    if project != "" {
        // 提取 registry 的基础部分（如果 registry 包含路径，只取基础部分）
        baseRegistry := extractRegistryBase(registry)
        return fmt.Sprintf("%s/%s/%s:%s", 
            baseRegistry,  // new-registry.com
            project,       // newproject
            sourceParts.Name,
            sourceParts.Tag)
    }
    
    // 场景2: --registry 包含路径（如 new-registry.com/path/xx）
    if hasPath(registry) {
        // 追加路径模式
        return fmt.Sprintf("%s/%s/%s:%s",
            registry,           // new-registry.com/path/xx
            sourceParts.Path,   // project (原有路径)
            sourceParts.Name,
            sourceParts.Tag)
    }
    
    // 场景3: 默认模式，只替换仓库，保持路径
    // 如果源镜像没有路径，则直接 registry/name:tag
    if sourceParts.Path == "" {
        return fmt.Sprintf("%s/%s:%s",
            registry,
            sourceParts.Name,
            sourceParts.Tag)
    }
    
    // 如果源镜像有路径，保持路径
    return fmt.Sprintf("%s/%s/%s:%s",
        registry,           // new-registry.com
        sourceParts.Path,   // project (保持原有路径)
        sourceParts.Name,
        sourceParts.Tag)
}

// extractRegistryBase 从 registry 中提取基础仓库名
// 例如：new-registry.com/path/xx -> new-registry.com
func extractRegistryBase(registry string) string {
    parts := strings.Split(registry, "/")
    return parts[0]
}

// hasPath 检查 registry 是否包含路径
func hasPath(registry string) bool {
    return strings.Contains(registry, "/")
}
```

### 4. 参数作用域优化

**当前问题**：

- `-r, --registry` 和 `-p, --project` 在 root 命令中定义，但只在 push 时使用
- 用户执行 `hpn -h` 时看到所有参数，容易混淆

**改进方案**：

- 将操作特定参数移到对应子命令中
- Root 命令只保留全局参数（如 `--config`, `--runtime`, `--debug`, `--platform`）

**参数分类**：

| 参数类型 | 参数 | 作用域 | 说明 |

|---------|------|--------|------|

| **全局参数** | `--config`, `--runtime`, `--platform`, `--debug`, `--auto-fallback` | Root | 所有子命令可用 |

| **Pull 参数** | `--file`, `--all-platforms` | pull | 仅 pull 命令 |

| **Save 参数** | `--file`, `--path`, `--output` | save | 仅 save 命令（支持多级路径） |

| **Load 参数** | `--path`, `--input`, `--recursive` | load | 仅 load 命令 |

| **Push 参数** | `--file`, `--registry`, `--project` | push | 仅 push 命令 |

## 实施计划

### 1. 创建子命令结构

**文件**: `cmd/hpn/root.go`

创建子命令：

- `pullCmd`: `hpn pull`
- `saveCmd`: `hpn save`
- `loadCmd`: `hpn load`
- `pushCmd`: `hpn push`

### 2. 创建镜像解析工具函数

**文件**: `cmd/hpn/image.go` (新建)

实现镜像解析和构建函数：

- `parseImage(image string) ImageParts`: 解析镜像名为组件，支持多级路径
- `buildTargetImage(sourceImage, registry, project string) string`: 实现三种命名模式
- `extractRegistryBase(registry string) string`: 提取基础仓库名
- `hasPath(registry string) bool`: 检查 registry 是否包含路径

### 3. 重构 Pull 子命令

**文件**: `cmd/hpn/pull.go` (新建)

- 移除 `-a pull` 逻辑
- 将 `executePull()` 移到 `pullCmd.RunE`
- 参数：`--file` (必需), `--platform`, `--all-platforms`

### 4. 重构 Save 子命令

**文件**: `cmd/hpn/save.go` (新建)

- 移除 `-a save` 逻辑
- 将 `executeSave()` 移到 `saveCmd.RunE`
- 参数：`--file` (必需), `--path` (默认: `./images`, 支持多级路径), `--output` (别名)
- 确保 `--path` 支持多级路径，如 `--path ./output/images/prod`，自动创建目录

### 5. 重构 Load 子命令

**文件**: `cmd/hpn/load.go` (新建)

- 移除 `-a load` 逻辑
- 将 `executeLoad()` 移到 `loadCmd.RunE`
- 参数：`--path` (默认: `./images`), `--input` (别名), `--recursive`

### 6. 重构 Push 子命令（重点）

**文件**: `cmd/hpn/push.go` (新建)

- 移除 `-a push` 逻辑
- 将 `executePush()` 移到 `pushCmd.RunE`
- 实现三种镜像命名模式：

  1. **默认保持路径**：`--registry new-registry.com` → 只替换仓库，保持路径
  2. **追加路径**：`--registry new-registry.com/path/xx` → 在原有路径前追加
  3. **统一项目**：`--registry new-registry.com --project newproject` → 所有镜像统一推送到指定项目

### 7. 更新配置结构

**文件**: `pkg/types/config.go`

- 移除 `SaveMode`, `LoadMode`, `PushMode` 类型
- 添加 `SavePath`, `LoadPath` 字符串字段
- 移除 mode 相关的默认配置

### 8. 更新 Root 命令

**文件**: `cmd/hpn/root.go`

- 移除 `action` 参数和相关逻辑
- 移除 mode 参数
- 移除 `-r, --registry` 和 `-p, --project`（移到 push 子命令）
- 保留全局参数：`--config`, `--runtime`, `--platform`, `--debug`, `--auto-fallback`
- 更新 usage template

## 技术细节

### 子命令结构示例

```go
var pullCmd = &cobra.Command{
    Use:   "pull",
    Short: "Pull images from registry",
    Long:  `Pull container images from a registry to local.`,
    RunE:  executePull,
}

func init() {
    pullCmd.Flags().StringVarP(&imageFile, "file", "f", "", "Image list file (required)")
    pullCmd.Flags().StringVar(&platform, "platform", "", "Target platform (e.g., linux/amd64, all)")
    rootCmd.AddCommand(pullCmd)
}
```

### 参数验证

- Pull: 验证 `--file` 必需
- Save: 验证 `--file` 必需，`--path` 可选（默认 `./images`，支持多级路径）
- Load: 验证 `--path` 可选（默认 `./images`），支持 `--recursive`
- Push: 验证 `--file` 必需，`--registry` 必需

## 文件清单

### 新建文件

- `cmd/hpn/pull.go` - Pull 子命令
- `cmd/hpn/save.go` - Save 子命令
- `cmd/hpn/load.go` - Load 子命令
- `cmd/hpn/push.go` - Push 子命令
- `cmd/hpn/image.go` - 镜像解析工具函数
- `docs/plan/cli-refactor-plan.md` - 实施计划文档

### 修改文件

- `cmd/hpn/root.go` - 重构为子命令结构，移除 action 和 mode 参数
- `pkg/types/config.go` - 移除 Mode 类型，添加路径字段

## 迁移示例

### 旧命令 → 新命令

```bash
# Pull
hpn -a pull -f images.txt
→ hpn pull -f images.txt

hpn -a pull -f images.txt --platform linux/amd64
→ hpn pull -f images.txt --platform linux/amd64

# Save
hpn -a save -f images.txt --save-mode 2
→ hpn save -f images.txt
→ hpn save -f images.txt --path ./images

hpn -a save -f images.txt --save-mode 1
→ hpn save -f images.txt --path .

# Save 多级路径
hpn save -f images.txt --path ./output/images/prod
→ 自动创建 ./output/images/prod 目录

# Load
hpn -a load --load-mode 2
→ hpn load
→ hpn load --path ./images

hpn -a load --load-mode 3
→ hpn load --recursive
→ hpn load --path ./images --recursive

# Push - 场景1：默认保持路径
hpn -a push -f images.txt -r new-registry.com --push-mode 1
→ hpn push -f images.txt --registry new-registry.com
# old-registry.com/project/nginx:latest → new-registry.com/project/nginx:latest

# Push - 场景2：追加路径
hpn push -f images.txt --registry new-registry.com/path/xx
# old-registry.com/project/nginx:latest → new-registry.com/path/xx/project/nginx:latest

# Push - 场景3：统一项目
hpn -a push -f images.txt -r new-registry.com -p newproject --push-mode 2
→ hpn push -f images.txt --registry new-registry.com --project newproject
# 所有镜像统一推送到 new-registry.com/newproject/...
```

## 测试要点

1. **子命令测试**：验证所有子命令正常工作
2. **参数作用域测试**：验证 push 特定参数不在 pull 命令中显示
3. **默认值测试**：验证默认路径为 `./images`
4. **多级路径测试**：验证 Save 操作支持多级路径，自动创建目录
5. **Push 命名测试**：验证三种镜像命名模式

   - 默认保持路径模式
   - 追加路径模式
   - 统一项目模式

6. **向后兼容测试**：验证旧命令不再工作（符合预期）

## 注意事项

1. **向后兼容性**：这是一个破坏性变更，需要更新文档和用户指南
2. **参数冲突**：确保全局参数和子命令参数不冲突
3. **帮助信息**：确保 `hpn -h` 显示子命令列表，`hpn <subcommand> -h` 显示子命令特定参数
4. **配置迁移**：配置文件中的 mode 设置需要迁移到新格式
5. **路径处理**：确保多级路径的创建和解析正确，支持相对路径和绝对路径