# 分支管理策略

## 概述
本项目采用GitFlow工作流程，确保代码质量和发布稳定性。

## 分支类型

### 主要分支

#### main分支
- **用途**: 生产就绪代码，每个commit对应一个发布版本
- **保护级别**: 最高
- **合并来源**: release分支、hotfix分支
- **直接提交**: 禁止
- **命名**: `main`

#### develop分支  
- **用途**: 集成分支，包含下一个版本的最新开发代码
- **保护级别**: 高
- **合并来源**: feature分支、release分支、hotfix分支
- **直接提交**: 禁止（除紧急情况）
- **命名**: `develop`

### 支持分支

#### Feature分支
- **用途**: 新功能开发
- **生命周期**: 临时
- **创建来源**: develop分支
- **合并目标**: develop分支
- **命名规范**: `feature/功能描述`
- **示例**: 
  - `feature/add-runtime-detection`
  - `feature/improve-error-handling`
  - `feature/support-podman`

#### Release分支
- **用途**: 发布准备，bug修复，版本号更新
- **生命周期**: 临时
- **创建来源**: develop分支
- **合并目标**: main分支和develop分支
- **命名规范**: `release/版本号`
- **示例**:
  - `release/v1.1.0`
  - `release/v1.2.0-beta.1`

#### Hotfix分支
- **用途**: 紧急修复生产环境问题
- **生命周期**: 临时
- **创建来源**: main分支
- **合并目标**: main分支和develop分支
- **命名规范**: `hotfix/问题描述`
- **示例**:
  - `hotfix/critical-security-fix`
  - `hotfix/memory-leak-fix`

#### Bugfix分支
- **用途**: 修复非紧急bug
- **生命周期**: 临时
- **创建来源**: develop分支
- **合并目标**: develop分支
- **命名规范**: `bugfix/问题描述`
- **示例**:
  - `bugfix/fix-config-parsing`
  - `bugfix/handle-empty-image-list`

## 工作流程

### 功能开发流程
```bash
# 1. 从develop创建feature分支
git checkout develop
git pull origin develop
git checkout -b feature/new-awesome-feature

# 2. 开发功能
# ... 编码、测试 ...

# 3. 提交代码
git add .
git commit -m "feat: add awesome new feature"

# 4. 推送分支
git push -u origin feature/new-awesome-feature

# 5. 创建Pull Request到develop分支
# 6. 代码审查通过后合并
# 7. 删除feature分支
```

### 发布流程
```bash
# 1. 从develop创建release分支
git checkout develop
git pull origin develop
git checkout -b release/v1.1.0

# 2. 更新版本号和文档
# 编辑version文件、changelog等

# 3. 提交发布准备
git add .
git commit -m "chore: prepare release v1.1.0"

# 4. 推送release分支
git push -u origin release/v1.1.0

# 5. 运行发布前测试
# 6. 创建PR到main分支
# 7. 合并到main并创建tag
# 8. 合并回develop分支
```

### Hotfix流程
```bash
# 1. 从main创建hotfix分支
git checkout main
git pull origin main
git checkout -b hotfix/critical-bug-fix

# 2. 修复问题
# ... 编码、测试 ...

# 3. 提交修复
git add .
git commit -m "fix: resolve critical security issue"

# 4. 推送分支
git push -u origin hotfix/critical-bug-fix

# 5. 创建PR到main分支
# 6. 紧急审查和合并
# 7. 创建hotfix版本tag
# 8. 合并回develop分支
```

## Commit消息规范

### 格式
```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type类型
- `feat`: 新功能
- `fix`: bug修复
- `docs`: 文档更新
- `style`: 代码格式化
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动
- `perf`: 性能优化
- `ci`: CI/CD相关

### 示例
```bash
feat(runtime): add automatic runtime detection

Add support for automatic container runtime detection with fallback mechanism.
This allows the tool to work seamlessly across different environments.

Closes #123
```

## 版本号规范

### 语义化版本
采用 [Semantic Versioning](https://semver.org/) 规范：
- `MAJOR.MINOR.PATCH`
- `v1.2.3`

### 版本类型
- **Major**: 不兼容的API变更
- **Minor**: 向后兼容的功能性新增
- **Patch**: 向后兼容的问题修正

### 预发布版本
- `v1.2.0-alpha.1`: Alpha版本
- `v1.2.0-beta.1`: Beta版本
- `v1.2.0-rc.1`: Release Candidate

## 分支清理

### 自动清理
- Feature分支合并后自动删除
- Release分支合并后保留一段时间再删除
- Hotfix分支合并后自动删除

### 手动清理
```bash
# 查看已合并的分支
git branch --merged develop

# 删除本地已合并分支
git branch -d feature/old-feature

# 删除远程分支
git push origin --delete feature/old-feature
```

## 最佳实践

### 分支命名
- 使用小写字母和连字符
- 描述性强，简洁明了
- 避免使用特殊字符

### 提交频率
- 小而频繁的提交
- 每个提交都应该是可工作的状态
- 相关变更放在同一个提交中

### Pull Request
- 提供清晰的描述
- 包含相关的测试
- 及时响应审查意见
- 保持PR的大小合理

### 代码审查
- 至少一人审查
- 关注代码质量、安全性、性能
- 提供建设性的反馈
- 及时完成审查

这个分支策略确保了代码质量，支持并行开发，并提供了清晰的发布流程。