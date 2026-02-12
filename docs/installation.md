# Installation

## Download binaries

Download the binary for your platform from [GitHub Releases](https://github.com/Ghostwritten/harpoon/releases/latest):

### Linux / macOS

```bash
curl -L https://github.com/Ghostwritten/harpoon/releases/latest/download/hpn-linux-amd64 -o hpn
chmod +x hpn
sudo mv hpn /usr/local/bin/
```

### Windows

Download `hpn-windows-amd64.exe` and add it to your PATH.

## Build from source

```bash
git clone https://github.com/Ghostwritten/harpoon.git
cd harpoon
go build -o hpn ./cmd/hpn
```

## Verify installation

```bash
hpn --version
```

At least one container runtime (Docker, Podman, or Nerdctl) must be installed. For `list-images`, [Helm](https://helm.sh) is also required.
