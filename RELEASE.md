# 简化发布流程

双分支模型的Go命令行工具发布流程。

## 🌿 分支说明
- **main** - 稳定发布分支
- **develop** - 开发分支

## 🔄 开发和发布

### 日常开发
```bash
# 在 develop 分支开发
git checkout develop
git pull origin develop

# 开发功能
# ... 编码 ...
go test ./...
./build.sh current

# 提交到 develop（包括所有更改）
git add -A  # 添加所有更改，包括删除的文件
git commit -m "feat: your feature description"
git push origin develop
```

### 发布新版本
```bash
# 将 develop 合并到 main
git checkout main
git pull origin main
git merge develop
git push origin main

# 创建版本标签
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions 会自动构建和发布。

## 📋 版本规范
- `v1.0.0` - 主要版本（破坏性变更）
- `v1.1.0` - 次要版本（新功能）
- `v1.0.1` - 补丁版本（bug修复）

## 🛠 自动化
- **develop** 分支：推送时自动运行测试
- **main** 分支：推送时自动运行测试
- **标签推送**：自动构建多平台二进制文件并创建 GitHub Release