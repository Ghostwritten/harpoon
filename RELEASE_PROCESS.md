# Release Process

This document describes the end-to-end release process for Harpoon (`hpn`).
Follow these steps every time a new version is published.

---

## Overview

```
Write release notes  →  Update CHANGELOG  →  Push tag  →  CI builds & publishes
```

All binary builds and GitHub Release creation are automated via
`.github/workflows/release.yml`. A tag push is the sole trigger.

---

## Step-by-step

### 1. Decide the version number

Follow [Semantic Versioning](https://semver.org):

| Change type | Example |
|-------------|---------|
| Breaking interface change | `v2.1.0` → `v3.0.0` |
| New feature (backwards-compatible) | `v2.1.0` → `v2.2.0` |
| Bug fix / security fix only | `v2.1.0` → `v2.1.1` |

### 2. Write the release notes

Create `docs/release/vX.Y.Z.md` using the existing files as a template
(e.g. `docs/release/v2.0.3.md`).

Required sections (omit if empty):

```markdown
# Harpoon vX.Y.Z — Release Notes

> Released: YYYY-MM-DD

## New Features
## Bug Fixes
## Security Fixes
## Interface Changes   ← include if any Go interfaces changed
## Refactoring
## CI/CD Improvements
## Upgrade Notes       ← include if users must change anything
## Files Changed
## Test Results
```

### 3. Update the changelog

Append a summary entry to `docs/changelog.md` following the existing format.

### 4. Commit and tag

```bash
# Stage only the intended files
git add docs/release/vX.Y.Z.md docs/changelog.md

git commit -m "Release vX.Y.Z: <one-line summary>"

# Create annotated tag
git tag -a vX.Y.Z -m "Release vX.Y.Z"

# Push commit + tag together
git push origin main
git push origin vX.Y.Z
```

> **Important:** the tag push triggers `.github/workflows/release.yml`.
> Do not push the tag until the release notes file is committed.

### 5. Monitor the CI run

Go to **Actions → Release** in the GitHub UI and verify:

1. Tests pass (`go test -race ./...`)
2. All five binaries build successfully
3. SHA256 checksums are generated
4. `softprops/action-gh-release` creates the Release and attaches the assets

If the workflow fails, fix the issue on `main`, delete the tag locally and
remotely, then re-tag:

```bash
git tag -d vX.Y.Z
git push origin :refs/tags/vX.Y.Z
# fix, commit, re-tag
```

### 6. Verify the GitHub Release

- All five binaries (`linux-amd64`, `linux-arm64`, `darwin-amd64`,
  `darwin-arm64`, `windows-amd64.exe`) are attached.
- `checksums.txt` is attached.
- The release body matches `docs/release/vX.Y.Z.md`.
- The release is **not** marked as a pre-release or draft.

### 7. Post-release

- Confirm `gh-pages` updated: push to `main` (with `website/**` changes if
  needed) triggers `.github/workflows/deploy-docs.yml`.
- Announce in relevant channels if applicable.

---

## CI workflow reference

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `test.yml` | push/PR to `main` | Unit tests, race detector, vet, golangci-lint |
| `release.yml` | push of `v*` tag | Cross-platform build + GitHub Release |
| `deploy-docs.yml` | push to `main` touching `website/**`, `docs/**`, `README.md` | VitePress build → gh-pages |

---

## Linter note

The project uses `golangci-lint`. The configuration is in `.golangci.yml`.
Run `make lint` locally before tagging to avoid a failing release CI run.

---

## Supported platforms

| Binary | Build flags |
|--------|-------------|
| `hpn-linux-amd64` | `CGO_ENABLED=0 -tags netgo,osusergo` (static) |
| `hpn-linux-arm64` | `CGO_ENABLED=0 -tags netgo,osusergo` (static) |
| `hpn-darwin-amd64` | standard |
| `hpn-darwin-arm64` | standard |
| `hpn-windows-amd64.exe` | standard |
