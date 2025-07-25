# Development Guide

Complete guide for developers contributing to Harpoon (hpn).

## Quick Start

### Prerequisites
- Go 1.21+
- Git
- Make
- One of: Docker, Podman, or Nerdctl (for testing)

### Setup Development Environment
```bash
# Clone repository
git clone https://github.com/your-org/harpoon.git
cd harpoon

# Install dependencies
make deps

# Build and test
make build
make dev-test
```

## Development Workflow

### Standard Workflow
```bash
# 1. Create feature branch
git checkout -b feature/new-feature

# 2. Make changes
vim cmd/hpn/root.go

# 3. Test locally
make dev-test

# 4. Run full tests
git push origin feature/new-feature

# 5. Create pull request
gh pr create --title "Add new feature" --body "Description"
```

### Quick Iteration
```bash
# Fast development cycle
make build && ./hpn -v

# Test specific functionality
make build && ./hpn -a pull -f test-images.txt
```

## Project Structure

```
.
├── cmd/hpn/           # Main application entry point
│   ├── main.go        # Application bootstrap
│   └── root.go        # CLI command definitions
├── internal/          # Internal packages (not importable)
│   ├── config/        # Configuration management
│   │   ├── config.go  # Config loading and parsing
│   │   └── validation.go # Config validation
│   ├── runtime/       # Container runtime interfaces
│   │   ├── interface.go # Runtime interface definition
│   │   ├── detector.go  # Runtime detection
│   │   ├── docker.go    # Docker implementation
│   │   ├── podman.go    # Podman implementation
│   │   └── nerdctl.go   # Nerdctl implementation
│   └── service/       # Business logic services
├── pkg/               # Public packages (importable)
│   ├── types/         # Type definitions
│   │   ├── config.go  # Configuration types
│   │   └── image.go   # Image-related types
│   └── errors/        # Error handling
│       └── errors.go  # Custom error types
├── docs/              # Documentation
├── .github/workflows/ # CI/CD configurations
├── Makefile          # Build automation
└── go.mod            # Go module definition
```

## Build System

### Make Targets
```bash
# Essential commands
make build          # Build for current platform
make test           # Run unit tests
make dev-test       # Development test suite
make clean          # Clean build artifacts
make fmt            # Format code

# Advanced commands
make build-all      # Build for all platforms
make test-coverage  # Test with coverage report
make lint           # Lint code
make package        # Create release packages
make install        # Install to system
```

### Build Configuration
The build system uses the following variables:
- `VERSION`: Version string (from Makefile)
- `COMMIT`: Git commit hash (auto-detected)
- `BUILD_DATE`: Build timestamp (auto-generated)

## Testing

### Local Testing
```bash
# Unit tests
go test -v ./...

# Integration tests
make dev-test

# Specific package tests
go test -v ./internal/runtime/

# Test with coverage
make test-coverage
```

### CI/CD Testing
GitHub Actions provides comprehensive testing:

- **Basic Tests**: Cross-platform functionality
- **Build Tests**: Multi-architecture builds
- **Runtime Tests**: Docker/Podman/Nerdctl compatibility
- **Integration Tests**: Full workflow testing
- **E2E Tests**: Real registry operations

### Test Categories

#### Unit Tests
- Test individual functions and methods
- Mock external dependencies
- Fast execution (< 1 second per test)

#### Integration Tests
- Test component interactions
- Use real container runtimes
- Moderate execution time (< 30 seconds)

#### End-to-End Tests
- Test complete workflows
- Use real registries (in CI only)
- Longer execution time (< 5 minutes)

## Code Standards

### Go Conventions
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Use `golangci-lint` for linting
- Write meaningful variable and function names
- Add comments for exported functions

### Error Handling
```go
// Use custom error types
func validateConfig(cfg *Config) error {
    if cfg.Registry == "" {
        return errors.New(errors.ErrInvalidConfig, "registry cannot be empty")
    }
    return nil
}

// Wrap errors with context
func pullImage(runtime Runtime, image string) error {
    if err := runtime.Pull(image); err != nil {
        return errors.Wrap(err, errors.ErrRuntimeCommand, 
            fmt.Sprintf("failed to pull image %s", image))
    }
    return nil
}
```

### Logging
```go
// Use structured logging
log.Info("pulling image", 
    log.String("image", image),
    log.String("runtime", runtime.Name()))

// Log errors with context
log.Error("pull failed",
    log.String("image", image),
    log.Error(err))
```

## Debugging

### Debug Mode
```bash
# Enable debug logging
export HPN_LOG_LEVEL=debug
./hpn -a pull -f images.txt

# Verbose Go testing
go test -v -run TestSpecificFunction ./...
```

### Common Issues

#### Build Issues
```bash
# Clean and rebuild
make clean && make build

# Check Go version
go version

# Update dependencies
go mod tidy
```

#### Runtime Issues
```bash
# Test specific runtime
./hpn --runtime docker -a pull -f test-images.txt

# Check runtime availability
docker version
podman version
nerdctl version
```

#### Test Failures
```bash
# Run specific test
go test -v ./internal/runtime/ -run TestDockerRuntime

# Run with race detection
go test -race ./...

# Generate test coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Performance Profiling

### CPU Profiling
```bash
# Generate CPU profile
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Profile specific operation
go test -bench=BenchmarkPull -cpuprofile=cpu.prof ./...
```

### Memory Profiling
```bash
# Generate memory profile
go test -memprofile=mem.prof ./...
go tool pprof mem.prof

# Check for memory leaks
go test -memprofile=mem.prof -memprofilerate=1 ./...
```

## Release Process

### Version Management
1. Update version in `Makefile`
2. Update version in `cmd/hpn/root.go`
3. Update `docs/changelog.md`
4. Create git tag

### Release Steps
```bash
# 1. Update version
sed -i 's/VERSION=v1.1/VERSION=v1.2/' Makefile
sed -i 's/version = "v1.1"/version = "v1.2"/' cmd/hpn/root.go

# 2. Update changelog
vim docs/changelog.md

# 3. Build and test
make build-all
make test

# 4. Create packages
make package

# 5. Commit and tag
git add .
git commit -m "Release v1.2"
git tag v1.2
git push origin main --tags

# 6. Create GitHub release
gh release create v1.2 dist/* --title "Release v1.2" --notes-file RELEASE_NOTES.md
```

## Contributing Guidelines

### Pull Request Process
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Update documentation
6. Submit pull request

### Commit Messages
Use conventional commit format:
```
feat: add new runtime detection feature
fix: resolve memory leak in image processing
docs: update configuration guide
test: add integration tests for push operation
```

### Code Review Checklist
- [ ] Code follows Go conventions
- [ ] Tests added for new functionality
- [ ] Documentation updated
- [ ] No breaking changes (or properly documented)
- [ ] Performance impact considered
- [ ] Security implications reviewed

## IDE Setup

### VS Code
```json
{
    "go.useLanguageServer": true,
    "go.formatTool": "gofmt",
    "go.lintTool": "golangci-lint",
    "go.testFlags": ["-v"],
    "editor.formatOnSave": true
}
```

### Vim/Neovim
```vim
" Install vim-go plugin
Plugin 'fatih/vim-go'

" Configure Go settings
let g:go_fmt_command = "gofmt"
let g:go_metalinter_enabled = ['vet', 'golint', 'errcheck']
let g:go_auto_type_info = 1
```

## Troubleshooting

### Common Development Issues

#### Import Path Issues
```bash
# Fix import paths
go mod tidy
go clean -modcache
```

#### Build Cache Issues
```bash
# Clear build cache
go clean -cache
go clean -testcache
```

#### Dependency Issues
```bash
# Update dependencies
go get -u ./...
go mod tidy
```

## Resources

### Documentation
- [Go Documentation](https://golang.org/doc/)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Viper Configuration](https://github.com/spf13/viper)

### Container Runtimes
- [Docker Engine API](https://docs.docker.com/engine/api/)
- [Podman API](https://docs.podman.io/en/latest/markdown/podman-system-service.1.html)
- [Nerdctl](https://github.com/containerd/nerdctl)

### Testing
- [Go Testing](https://golang.org/pkg/testing/)
- [Testify](https://github.com/stretchr/testify)
- [GitHub Actions](https://docs.github.com/en/actions)

### Tools
- [golangci-lint](https://golangci-lint.run/)
- [GoReleaser](https://goreleaser.com/)
- [GitHub CLI](https://cli.github.com/)