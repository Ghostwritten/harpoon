# 合并策略指南

## 概述
本文档定义了不同类型Pull Request的合并策略，确保代码历史的清晰性和可追溯性。

## 合并策略类型

### 1. Merge Commit (合并提交)
**使用场景**: Feature分支合并到develop分支
**优点**: 
- 保留完整的分支历史
- 清晰显示功能开发过程
- 便于回滚整个功能

**操作方式**:
```bash
git checkout develop
git merge --no-ff feature/new-feature
```

### 2. Squash and Merge (压缩合并)
**使用场景**: 
- 小型bug修复
- 文档更新
- 配置变更
- 多个小提交的功能

**优点**:
- 保持主分支历史简洁
- 将相关变更合并为单个提交
- 便于生成changelog

**操作方式**:
通过GitHub界面选择"Squash and merge"

### 3. Rebase and Merge (变基合并)
**使用场景**: 
- 单个提交的变更
- 已经整理好的提交历史
- 不需要保留分支信息的情况

**优点**:
- 线性的提交历史
- 没有额外的合并提交
- 清晰的时间线

**操作方式**:
通过GitHub界面选择"Rebase and merge"

## 分支特定策略

### Main分支
- **来源**: Release分支、Hotfix分支
- **策略**: Merge Commit
- **原因**: 保留发布历史，便于版本追踪

```bash
# Release分支合并到main
git checkout main
git merge --no-ff release/v1.1.0
git tag v1.1.0
```

### Develop分支
- **来源**: Feature分支、Bugfix分支、Release分支、Hotfix分支
- **策略**: 根据变更类型选择

#### Feature分支 → Develop
- **策略**: Merge Commit
- **原因**: 保留功能开发历史

#### Bugfix分支 → Develop  
- **策略**: Squash and Merge
- **原因**: 简化bug修复历史

#### 文档/配置更新 → Develop
- **策略**: Squash and Merge
- **原因**: 避免琐碎提交污染历史

## GitHub设置配置

### 仓库合并设置
在GitHub仓库设置中配置：

1. **Settings** → **General** → **Pull Requests**
2. 启用以下选项：
   - ✅ Allow merge commits
   - ✅ Allow squash merging
   - ✅ Allow rebase merging
3. 设置默认合并策略：
   - **Default to merge commits** (推荐)

### 分支保护规则
为不同分支设置不同的合并要求：

#### Main分支
- 要求Pull Request
- 要求状态检查通过
- 要求代码审查
- 限制合并方式为Merge Commit

#### Develop分支
- 要求Pull Request
- 要求基础状态检查通过
- 允许所有合并方式

## 合并检查清单

### 合并前检查
- [ ] 所有CI检查通过
- [ ] 代码审查完成
- [ ] 冲突已解决
- [ ] 提交消息符合规范
- [ ] 相关文档已更新

### 选择合并策略
根据以下决策树选择合并策略：

```
是否为功能分支？
├─ 是 → 使用 Merge Commit
└─ 否 → 是否为多个小提交？
    ├─ 是 → 使用 Squash and Merge
    └─ 否 → 使用 Rebase and Merge
```

## 提交消息规范

### 合并提交消息格式
```
Merge pull request #123 from feature/new-awesome-feature

feat: add awesome new feature

- Add runtime detection
- Improve error handling
- Update documentation
```

### Squash提交消息格式
```
feat: add awesome new feature (#123)

- Add runtime detection
- Improve error handling  
- Update documentation

Co-authored-by: Developer Name <email@example.com>
```

## 最佳实践

### 1. 保持分支整洁
- 合并前整理提交历史
- 移除调试提交
- 合并相关的小提交

### 2. 编写清晰的合并消息
- 描述变更的目的
- 列出主要变更点
- 包含相关Issue链接

### 3. 及时删除已合并分支
```bash
# 删除本地分支
git branch -d feature/completed-feature

# 删除远程分支
git push origin --delete feature/completed-feature
```

### 4. 处理合并冲突
```bash
# 更新目标分支
git checkout develop
git pull origin develop

# 合并并解决冲突
git checkout feature/my-feature
git merge develop

# 解决冲突后提交
git add .
git commit -m "resolve merge conflicts"
```

## 回滚策略

### 回滚Merge Commit
```bash
# 回滚到合并前状态
git revert -m 1 <merge-commit-hash>
```

### 回滚Squash Commit
```bash
# 直接回滚提交
git revert <commit-hash>
```

### 紧急回滚
```bash
# 硬重置到指定提交（谨慎使用）
git reset --hard <commit-hash>
git push --force-with-lease origin main
```

## 自动化工具

### GitHub Actions检查
创建自动检查合并策略的工作流：

```yaml
name: Merge Strategy Check
on:
  pull_request:
    types: [opened, synchronize]

jobs:
  check-merge-strategy:
    runs-on: ubuntu-latest
    steps:
    - name: Check merge strategy
      run: |
        # 根据分支类型建议合并策略
        echo "Suggested merge strategy based on branch type"
```

### 合并后清理
```yaml
name: Post Merge Cleanup
on:
  pull_request:
    types: [closed]

jobs:
  cleanup:
    if: github.event.pull_request.merged == true
    runs-on: ubuntu-latest
    steps:
    - name: Delete merged branch
      run: |
        # 自动删除已合并的feature分支
```

这个合并策略确保了代码历史的清晰性，同时为不同类型的变更提供了合适的合并方式。