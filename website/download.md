# Download and Installation

## Download Binaries

### Option 1: GitHub Releases

1. Go to [Releases](https://github.com/Ghostwritten/harpoon/releases)
2. Choose the latest version (e.g. v2.0.2)
3. Under "Assets", download the binary for your platform:
   - **Linux AMD64**: `hpn-linux-amd64`
   - **Linux ARM64**: `hpn-linux-arm64`
   - **macOS Intel**: `hpn-darwin-amd64`
   - **macOS Apple Silicon**: `hpn-darwin-arm64`
   - **Windows AMD64**: `hpn-windows-amd64.exe`

### Option 2: Command line

#### Linux / macOS

```bash
# Set version (use a specific tag like v2.0.2 or "latest" for the latest release)
VERSION="v2.0.2"

# Linux AMD64
curl -L -o hpn "https://github.com/Ghostwritten/harpoon/releases/download/${VERSION}/hpn-linux-amd64"

# Linux ARM64
curl -L -o hpn "https://github.com/Ghostwritten/harpoon/releases/download/${VERSION}/hpn-linux-arm64"

# macOS Intel
curl -L -o hpn "https://github.com/Ghostwritten/harpoon/releases/download/${VERSION}/hpn-darwin-amd64"

# macOS Apple Silicon
curl -L -o hpn "https://github.com/Ghostwritten/harpoon/releases/download/${VERSION}/hpn-darwin-arm64"

chmod +x hpn
sudo mv hpn /usr/local/bin/
```

#### Windows (PowerShell)

```powershell
$VERSION = "v2.0.2"
Invoke-WebRequest -Uri "https://github.com/Ghostwritten/harpoon/releases/download/$VERSION/hpn-windows-amd64.exe" -OutFile "hpn.exe"
# Add hpn.exe to your PATH if desired
```

### Option 3: GitHub CLI

```bash
gh release download v2.0.2 --repo Ghostwritten/harpoon
# Or latest:
gh release download --repo Ghostwritten/harpoon
```

## Verify installation

```bash
./hpn --version
./hpn --help
```

## Package managers (future)

Possible future support:

- **Homebrew** (macOS/Linux): `brew install hpn`
- **Chocolatey** (Windows): `choco install hpn`
- **Snap** (Linux): `snap install hpn`
- **Go install**: `go install github.com/Ghostwritten/harpoon/cmd/hpn@latest`
