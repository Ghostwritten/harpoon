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
- [Release Notes](https://github.com/Ghostwritten/harpoon/blob/main/docs/release/v2.0.0.md)
- [Migration Guide](https://github.com/Ghostwritten/harpoon/blob/main/docs/changelog.md#migration-guide)

## 🔗 Links

- [GitHub Repository](https://github.com/Ghostwritten/harpoon)
- [Issues](https://github.com/Ghostwritten/harpoon/issues)
- [Discussions](https://github.com/Ghostwritten/harpoon/discussions)
