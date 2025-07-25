# GitHub分支保护规则配置

## 概述
本文档描述了项目的分支保护规则配置，需要在GitHub仓库设置中手动配置。

## Main分支保护规则

### 基本设置
- **分支名称模式**: `main`
- **限制推送**: ✅ 启用
- **要求Pull Request**: ✅ 启用

### Pull Request要求
- **必需的审查数量**: 2
- **解除过时的审查**: ✅ 启用
- **要求代码所有者审查**: ✅ 启用
- **限制可以解除审查的用户**: 仓库管理员

### 状态检查要求
必须通过以下检查才能合并：
- `test-suite` - 完整测试套件
- `build-check` - 构建检查
- `security-scan` - 安全扫描
- `lint-check` - 代码规范检查
- `coverage-check` - 测试覆盖率检查

### 其他限制
- **要求分支是最新的**: ✅ 启用
- **要求对话解决**: ✅ 启用
- **限制推送**: ✅ 启用（仅管理员可直接推送）
- **允许强制推送**: ❌ 禁用
- **允许删除**: ❌ 禁用

## Develop分支保护规则

### 基本设置
- **分支名称模式**: `develop`
- **限制推送**: ✅ 启用
- **要求Pull Request**: ✅ 启用

### Pull Request要求
- **必需的审查数量**: 1
- **解除过时的审查**: ✅ 启用
- **要求代码所有者审查**: ❌ 禁用

### 状态检查要求
必须通过以下检查才能合并：
- `test-suite` - 基础测试套件
- `lint-check` - 代码规范检查
- `build-check` - 构建检查

### 其他限制
- **要求分支是最新的**: ✅ 启用
- **要求对话解决**: ✅ 启用
- **限制推送**: ✅ 启用
- **允许强制推送**: ❌ 禁用
- **允许删除**: ❌ 禁用

## Release分支保护规则

### 基本设置
- **分支名称模式**: `release/*`
- **限制推送**: ✅ 启用
- **要求Pull Request**: ✅ 启用

### Pull Request要求
- **必需的审查数量**: 1
- **解除过时的审查**: ✅ 启用
- **要求代码所有者审查**: ✅ 启用

### 状态检查要求
必须通过以下检查才能合并：
- `release-test-suite` - 发布测试套件
- `build-all-platforms` - 全平台构建
- `security-scan` - 安全扫描
- `performance-test` - 性能测试

## 配置步骤

### 1. 访问分支保护设置
1. 进入GitHub仓库
2. 点击 `Settings` 标签
3. 在左侧菜单中选择 `Branches`
4. 点击 `Add rule` 添加新规则

### 2. 配置Main分支
1. 在 `Branch name pattern` 中输入 `main`
2. 勾选 `Restrict pushes that create files larger than 100 MB`
3. 勾选 `Require a pull request before merging`
   - 勾选 `Require approvals` 并设置为 2
   - 勾选 `Dismiss stale pull request approvals when new commits are pushed`
   - 勾选 `Require review from code owners`
4. 勾选 `Require status checks to pass before merging`
   - 勾选 `Require branches to be up to date before merging`
   - 添加必需的状态检查（见上述列表）
5. 勾选 `Require conversation resolution before merging`
6. 勾选 `Restrict pushes that create files larger than 100 MB`
7. 不勾选 `Allow force pushes`
8. 不勾选 `Allow deletions`

### 3. 配置Develop分支
按照类似步骤配置develop分支，但审查要求较宽松。

### 4. 配置Release分支
使用通配符模式 `release/*` 配置release分支保护。

## 验证配置
配置完成后，可以通过以下方式验证：

1. 尝试直接推送到protected分支（应该被拒绝）
2. 创建Pull Request并验证状态检查要求
3. 验证审查要求是否正确执行

## 注意事项
- 分支保护规则只能通过GitHub Web界面配置，无法通过代码自动化
- 管理员默认可以绕过某些限制，可以在设置中进一步限制
- 状态检查名称必须与CI/CD工作流中的job名称完全匹配
- 建议定期审查和更新分支保护规则