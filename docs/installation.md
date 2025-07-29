# 安装

## 下载二进制文件

从 [GitHub Releases](https://github.com/your-org/harpoon/releases/latest) 下载对应平台的二进制文件：

### Linux/macOS
```bash
# 下载并安装
curl -L https://github.com/your-org/harpoon/releases/latest/download/hpn-linux-amd64 -o hpn
chmod +x hpn
sudo mv hpn /usr/local/bin/
```

### Windows
下载 `hpn-windows-amd64.exe` 并添加到 PATH。

## 从源码构建

```bash
git clone https://github.com/your-org/harpoon.git
cd harpoon
go build -o hpn ./cmd/hpn
```

## 验证安装

```bash
hpn --version
```

需要安装 Docker、Podman 或 Nerdctl 中的至少一个容器运行时。