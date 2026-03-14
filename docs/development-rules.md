# Harpoon Development Rules

Conventions and practices established while building Harpoon with AI assistance. Use as reference for contributors and AI collaboration.

## 1. Code and Structure

| Rule | Description |
|------|-------------|
| **Comments in English** | All Go source and comments are in English. |
| **Centralized subcommand registration** | Register subcommands in `cmd/hpn/root.go` `init()`. Do not call `rootCmd.AddCommand` in individual command files. |
| **Exit codes** | `0` success; `1` runtime/operational error; `2` usage error (missing/invalid args). Scripts can rely on exit code 2. |
| **Usage errors** | Use `usageErrorf(format, args...)` from `main.go`. Do not append `: usage` to user-facing messages. |
| **Package doc comments** | Add package-level doc comments for `go doc` visibility (e.g. `pkg/errors`, `internal/runtime`). |

## 2. Release and Versioning

| Rule | Description |
|------|-------------|
| **Push before tag** | Commit changes, then `git push origin main`, then `git tag vX.Y.Z`, then `git push origin vX.Y.Z`. |
| **Release notes** | Create `docs/release/vX.Y.Z.md`. The release workflow uses it as the GitHub Release body. |
| **Version sync** | Update README version badge, download doc examples, and changelog `[vX.Y.Z]` section. |
| **Asset check** | Ensure logo and images are not ignored by `.gitignore`; add and commit them if needed. |

## 3. CI and Quality

| Rule | Description |
|------|-------------|
| **go mod** | Run `go mod verify` and `go mod tidy && git diff --exit-code go.mod go.sum` in CI. |
| **go vet** | Run `go vet ./...` in CI. |
| **Tests** | Add table-driven tests for shared logic (e.g. `readImageList`, `mergeExtraArgs`). |

## 4. Documentation and Writing

| Rule | Description |
|------|-------------|
| **Blog style** | Humorous, conversational; start with context/motivation before technical content. |
| **Real examples** | Prefer real command output (e.g. `hpn ls`, `hpn ls --path`) over hypothetical snippets. |
| **Scenarios** | Document concrete scenarios: breakpoint resume, concurrency (`--workers`), CI/CD integration. |
| **Progression** | Structure docs from basics to advanced. |

## 5. AI Collaboration

| Rule | Description |
|------|-------------|
| **Plan vs execute** | Use Plan mode for design; do not edit the plan file when executing. |
| **Todos** | Align todos with the plan; mark progress; avoid creating duplicate todos. |
| **Ask mode** | Guidance only; no edits. Switch to Agent mode for changes. |
| **Context** | Provide clear file paths and dependencies when requesting multi-file changes. |

## 6. Feature and Product Design

| Rule | Description |
|------|-------------|
| **Defaults** | Sensible defaults (e.g. breakpoint resume on save, checksums enabled). |
| **Runtime abstraction** | Support Docker, Podman, Nerdctl, Skopeo; use `--auto-fallback` for CI. |
| **CI-friendly** | Exit codes, `--password-stdin`, `--workers` for scripts and pipelines. |
| **Config optional** | CLI flags override config; config provides defaults, not requirements. |
