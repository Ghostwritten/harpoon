# v2.0.0 发布指南

## ✅ 发布前检查清单

- [x] 所有功能已实现并测试
- [x] 所有平台二进制文件已构建（dist/ 目录）
- [x] Changelog 已更新
- [x] Release Notes 已准备
- [x] Git commit 已创建
- [x] Git tag v2.0.0 已创建
- [ ] 推送到 GitHub
- [ ] 创建 GitHub Release
- [ ] 上传二进制文件

## 📤 发布步骤

### 1. 推送到 GitHub

```bash
# 推送代码和标签
git push origin main
git push origin v2.0.0
```

### 2. 创建 GitHub Release

访问：https://github.com/Ghostwritten/harpoon/releases/new

**设置：**
- **Tag**: 选择 `v2.0.0`
- **Title**: `v2.0.0` 或 `Harpoon v2.0.0 - Major Release`
- **Description**: 使用下面的模板

### 3. Release 描述模板

```markdown
# Harpoon v2.0.0 - Major Release 🎉

## 🚀 What's New

### Unified Registry Authentication
- New `hpn login` command for secure registry authentication
- Supports Docker, Podman, Skopeo, and Nerdctl
- Interactive password input, stdin support, and environment variables

### Modern CLI Architecture
- Refactored to subcommand-based CLI (like docker, kubectl)
- Better command discovery and help organization
- Improved parameter scoping

### Skopeo Integration
- Added Skopeo as a first-class runtime
- Daemonless image operations
- Multi-architecture support

### Enhanced Developer Experience
- Debug mode for detailed troubleshooting
- Progress bars for long-running operations
- Better error messages

## ⚠️ Breaking Changes

**CLI Syntax:**
- `hpn -a pull` → `hpn pull`
- `hpn -a save --save-mode 2` → `hpn save --path ./images`

**Configuration:**
- `modes` section replaced with `paths` section

See [Migration Guide](https://github.com/Ghostwritten/harpoon/blob/main/docs/changelog.md#migration-guide) for details.

## 📦 Downloads

- [Linux AMD64](https://github.com/Ghostwritten/harpoon/releases/download/v2.0.0/hpn-linux-amd64)
- [Linux ARM64](https://github.com/Ghostwritten/harpoon/releases/download/v2.0.0/hpn-linux-arm64)
- [macOS Intel](https://github.com/Ghostwritten/harpoon/releases/download/v2.0.0/hpn-darwin-amd64)
- [macOS Apple Silicon](https://github.com/Ghostwritten/harpoon/releases/download/v2.0.0/hpn-darwin-arm64)
- [Windows AMD64](https://github.com/Ghostwritten/harpoon/releases/download/v2.0.0/hpn-windows-amd64.exe)

## 📚 Documentation

- [Full Changelog](https://github.com/Ghostwritten/harpoon/blob/main/docs/changelog.md#v20---2025-01-21)
- [Release Notes](https://github.com/Ghostwritten/harpoon/blob/main/docs/release-notes-v2.0.md)
- [Migration Guide](https://github.com/Ghostwritten/harpoon/blob/main/docs/changelog.md#migration-guide)

## 🔗 Links

- [GitHub Repository](https://github.com/Ghostwritten/harpoon)
- [Issues](https://github.com/Ghostwritten/harpoon/issues)
- [Discussions](https://github.com/Ghostwritten/harpoon/discussions)
```

### 4. 上传二进制文件

在 GitHub Release 页面，点击 "Attach binaries" 或拖拽以下文件：

- `dist/hpn-linux-amd64`
- `dist/hpn-linux-arm64`
- `dist/hpn-darwin-amd64`
- `dist/hpn-darwin-arm64`
- `dist/hpn-windows-amd64.exe`

### 5. 发布 Release

- 勾选 "Set as the latest release"（如果这是最新版本）
- 点击 "Publish release"

## 📋 发布后验证

发布后，验证以下内容：

1. **Release 页面**：https://github.com/Ghostwritten/harpoon/releases/tag/v2.0.0
   - 确认描述正确显示
   - 确认所有二进制文件已上传
   - 确认下载链接可用

2. **二进制文件验证**：
   ```bash
   # 下载并测试
   curl -L https://github.com/Ghostwritten/harpoon/releases/download/v2.0.0/hpn-linux-amd64 -o hpn
   chmod +x hpn
   ./hpn --version
   ./hpn login --help
   ```

3. **文档链接**：
   - 确认所有文档链接有效
   - 确认迁移指南可访问

## 🎯 发布完成

发布完成后，可以：
- 在 README 中更新版本号
- 在社交媒体或社区中宣布发布
- 监控 Issues 和 Discussions 中的反馈
