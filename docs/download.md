# 下载和安装指南

## 📥 下载二进制文件

### 方式1: GitHub Releases 页面下载
1. 访问 [Releases 页面](https://github.com/你的用户名/harpoon/releases)
2. 选择最新版本
3. 在 "Assets" 部分下载适合你系统的二进制文件：
   - **Linux 64位**: `hpn-linux-amd64`
   - **Linux ARM64**: `hpn-linux-arm64`
   - **macOS Intel**: `hpn-darwin-amd64`
   - **macOS Apple Silicon**: `hpn-darwin-arm64`
   - **Windows 64位**: `hpn-windows-amd64.exe`

### 方式2: 命令行下载

#### Linux/macOS
```bash
# 下载最新版本 (替换为实际版本号)
VERSION="v1.0.0"

# Linux AMD64
curl -L -o hpn "https://github.com/你的用户名/harpoon/releases/download/${VERSION}/hpn-linux-amd64"

# macOS Intel
curl -L -o hpn "https://github.com/你的用户名/harpoon/releases/download/${VERSION}/hpn-darwin-amd64"

# macOS Apple Silicon
curl -L -o hpn "https://github.com/你的用户名/harpoon/releases/download/${VERSION}/hpn-darwin-arm64"

# 添加执行权限
chmod +x hpn

# 移动到系统路径 (可选)
sudo mv hpn /usr/local/bin/
```

#### Windows (PowerShell)
```powershell
# 下载
$VERSION = "v1.0.0"
Invoke-WebRequest -Uri "https://github.com/你的用户名/harpoon/releases/download/$VERSION/hpn-windows-amd64.exe" -OutFile "hpn.exe"

# 添加到 PATH (可选)
# 将 hpn.exe 移动到 PATH 中的目录
```

### 方式3: 使用 GitHub CLI
```bash
# 安装 GitHub CLI 后
gh release download v1.0.0 --repo 你的用户名/harpoon

# 或下载最新版本
gh release download --repo 你的用户名/harpoon
```

## 🚀 安装验证

下载后验证安装：
```bash
# 检查版本
./hpn --version

# 查看帮助
./hpn --help
```

## 🔄 自动安装脚本

你也可以创建一个自动安装脚本供用户使用：

```bash
# 一键安装 (Linux/macOS)
# Download directly from GitHub Releases (recommended)
# See the download section above for direct download commands
```

## 📦 包管理器支持 (未来)

未来可以考虑支持：
- **Homebrew** (macOS/Linux): `brew install hpn`
- **Chocolatey** (Windows): `choco install hpn`
- **Snap** (Linux): `snap install hpn`
- **Go Install**: `go install github.com/你的用户名/harpoon/cmd/hpn@latest`