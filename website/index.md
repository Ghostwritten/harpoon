---
layout: home
hero:
  name: Harpoon (hpn)
  text: Container Image Management CLI
  tagline: Modern, efficient. Pull, save, load, and push with Docker, Podman, Nerdctl, and Skopeo.
  image:
    src: /logo.png
    alt: Harpoon
  actions:
    - theme: brand
      text: Quick Start
      link: /guide/quickstart
    - theme: alt
      text: Download
      link: /download
    - theme: alt
      text: GitHub
      link: https://github.com/Ghostwritten/harpoon
features:
  - title: Multi-Runtime Support
    details: Docker, Podman, Nerdctl, Skopeo with automatic detection and fallback.
  - title: Batch Operations
    details: Efficient bulk image processing for pull, save, load, and push.
  - title: Helm Chart Support
    details: Extract image list from Helm charts for offline or CI use.
  - title: Enterprise Ready
    details: Proxy support, unified authentication, private registries.
  - title: Cross-Platform
    details: Linux, macOS, Windows. AMD64 and ARM64 binaries available.
  - title: Configuration
    details: YAML-based config with environment variables, flexible paths.
---

## Quick Install

```bash
# Linux / macOS
curl -L https://github.com/Ghostwritten/harpoon/releases/latest/download/hpn-linux-amd64 -o hpn
chmod +x hpn
sudo mv hpn /usr/local/bin/

# Verify
hpn --version
```

See [Installation](/guide/installation) for Windows, ARM, and source builds.
