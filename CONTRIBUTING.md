# Contributing to Harpoon (hpn)

Thank you for your interest in contributing. This document covers how to get started and submit changes.

## Getting started

1. **Clone the repository**
   ```bash
   git clone https://github.com/Ghostwritten/harpoon.git
   cd harpoon
   ```

2. **Build the CLI**
   ```bash
   go build ./cmd/hpn
   ```
   The binary will be created in the current directory (or use `-o dist/hpn` to place it in `dist/`).

3. **Run tests**
   ```bash
   go test ./...
   ```
   To run tests for a specific package, e.g. the CLI:
   ```bash
   go test ./cmd/hpn/...
   ```

## Code style and quality

- **Formatting:** Format Go code with `gofmt` or `go fmt ./...`. Code and comments in the codebase are in **English**.
- **Linting:** CI enforces `golangci-lint`. Run locally before opening a PR:
  ```bash
  go vet ./...
  golangci-lint run --timeout=5m
  # or via Makefile:
  make lint
  ```
  The configuration is in `.golangci.yml` at the repo root.

## Submitting changes

- Use a descriptive branch name (e.g. `fix/login-flag`, `feat/add-xyz`).
- Link to an issue in the PR description when applicable.
- Ensure `go test ./...` and `go vet ./...` pass.

**Exit codes:** The CLI uses exit code 0 for success, 1 for runtime or operational errors (e.g. runtime not found, pull/push failed), and 2 for usage errors (missing required flag, invalid arguments). See `cmd/hpn/main.go` for the implementation.

**Logging:** Current output uses `fmt.Printf`/`Fprintf`. Structured or leveled logging (e.g. `log/slog`) may be introduced later for `--debug` or consistency; no change is required for this phase.

For design or architecture details, see the [docs/](docs/) directory.
