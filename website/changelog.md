# Changelog

All notable changes to Harpoon (hpn) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Renamed

- **list-images → extract**: The subcommand that extracts image list from Helm charts has been renamed to `extract`. The name `list-images` remains available as an alias for backward compatibility.

## [v2.0.2] - 2025-02-13

### Added

- **Project logo**: New `logo.png` at repo root, used in README.
- **CONTRIBUTING.md**: Build, test, vet, code style (gofmt, English comments), PR process, exit code and logging notes.
- **Package doc comments**: `pkg/errors` and `internal/runtime` for `go doc` visibility.
- **Cmd tests**: Table-driven tests for `readImageList` and `mergeExtraArgs` in `cmd/hpn/common_test.go`.
- **CI**: `go mod verify`, `go mod tidy` check, and `go vet ./...` in the test workflow.

### Changed

- **Subcommand registration**: All subcommands (pull, save, load, push, login, ls, rmi) now registered centrally in `root.go`.
- **Comments**: All Go comments in runtime packages (interface, docker, podman, nerdctl, skopeo) translated to English.

### Fixed

- **Exit codes**: CLI now uses 0 (success), 1 (runtime/operational error), 2 (usage error). Scripts can rely on exit code 2 for missing required flags or invalid arguments.
- **Usage error messages**: User-facing messages for missing `-f`, `--chart`, or `--registry` no longer append ": usage".

### Other

- **.gitignore**: Added `images/` (default save/load output directory).

## [v2.0.1] - 2025-02-13

### Added

- **list-images**: New `hpn list-images` subcommand (now `extract`) to extract container image lists from Helm charts (remote or local tgz/dir).
- **ls**: New `hpn ls` subcommand (alias `list`) with three modes: runtime, path, or check file.
- **rmi**: New `hpn rmi` subcommand to remove images from the local runtime.
- **internal/chart**: New package for chart fetch, `helm template`, and YAML image extraction.

### Improved

- **list-images filtering**: Image extraction now filters out non-image strings.

## [v2.0] - 2025-01-21

### Major Release Highlights

- **Login Command**: Unified registry authentication for Docker, Podman, Skopeo, Nerdctl.
- **Skopeo Integration**: Full support for pull, save, load, and push.
- **CLI Refactoring**: Subcommand-based architecture (hpn pull, hpn save, etc.).
- **Configuration**: New `paths` section, removed `modes`.

## [v1.1] - 2024-12-19

- `--runtime` and `--auto-fallback` parameters.
- Smart runtime detection with fallback mechanism.

## [v1.0] - 2024-12-01

- Initial release of Harpoon Go rewrite.
- Support for pull, save, load, push with Docker, Podman, Nerdctl.

---

For the complete changelog, see the [full changelog on GitHub](https://github.com/Ghostwritten/harpoon/blob/main/docs/changelog.md).
