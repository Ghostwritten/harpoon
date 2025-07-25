# Architecture

This document describes the architecture and design principles of Harpoon (hpn).

## Overview

Harpoon is designed as a modern, extensible container image management tool with a clean separation of concerns and pluggable architecture.

## High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Layer     │    │  Service Layer  │    │  Runtime Layer  │
│                 │    │                 │    │                 │
│ • Command Parse │───▶│ • Image Ops     │───▶│ • Docker        │
│ • Flag Validate │    │ • Config Mgmt   │    │ • Podman        │
│ • User I/O      │    │ • Error Handle  │    │ • Nerdctl       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Config Layer   │    │  Logging Layer  │    │  Storage Layer  │
│                 │    │                 │    │                 │
│ • YAML Config   │    │ • Structured    │    │ • File System   │
│ • Env Variables │    │ • Multi-format  │    │ • Tar Archives  │
│ • Defaults      │    │ • Levels        │    │ • Temp Files    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Package Structure

### Core Packages

```
cmd/hpn/                    # Main application entry point
├── main.go                 # Application bootstrap
├── root.go                 # Root command and CLI setup
└── version.go              # Version information

internal/                   # Internal packages (not exported)
├── config/                 # Configuration management
│   ├── config.go          # Config loading and validation
│   └── validation.go      # Config validation logic
├── runtime/               # Container runtime abstraction
│   ├── interface.go       # Runtime interface definition
│   ├── detector.go        # Runtime detection logic
│   ├── docker.go          # Docker implementation
│   ├── podman.go          # Podman implementation
│   └── nerdctl.go         # Nerdctl implementation
├── service/               # Business logic services
│   ├── image.go           # Core image operations
│   ├── pull.go            # Pull service implementation
│   ├── save.go            # Save service implementation
│   ├── load.go            # Load service implementation
│   └── push.go            # Push service implementation
└── logger/                # Logging infrastructure
    ├── logger.go          # Logger interface and setup
    └── formatter.go       # Log formatting

pkg/                       # Public packages (exported)
├── types/                 # Type definitions
│   ├── config.go          # Configuration types
│   ├── image.go           # Image data structures
│   └── operation.go       # Operation result types
└── errors/                # Error handling
    └── errors.go          # Custom error types
```

## Design Principles

### 1. Separation of Concerns

Each layer has a specific responsibility:

- **CLI Layer**: User interface and command parsing
- **Service Layer**: Business logic and orchestration
- **Runtime Layer**: Container runtime abstraction
- **Config Layer**: Configuration management
- **Storage Layer**: File system operations

### 2. Interface-Based Design

All major components are defined by interfaces, enabling:

- Easy testing with mocks
- Runtime implementation swapping
- Future extensibility

```go
type ContainerRuntime interface {
    Name() string
    IsAvailable() bool
    Pull(ctx context.Context, image string, options PullOptions) error
    Save(ctx context.Context, image string, tarPath string) error
    Load(ctx context.Context, tarPath string) error
    Push(ctx context.Context, image string, options PushOptions) error
    Tag(ctx context.Context, source, target string) error
    Version() (string, error)
}
```

### 3. Configuration-Driven

The application behavior is controlled through:

- YAML configuration files
- Environment variables
- Command-line flags
- Sensible defaults

Priority order: CLI flags > Environment > Config file > Defaults

### 4. Error Handling

Structured error handling with:

- Custom error types with context
- Error codes for programmatic handling
- User-friendly error messages
- Detailed logging for debugging

## Component Interactions

### 1. Command Execution Flow

```
User Command
     │
     ▼
CLI Parsing (Cobra)
     │
     ▼
Configuration Loading
     │
     ▼
Runtime Detection
     │
     ▼
Service Layer
     │
     ▼
Runtime Operations
     │
     ▼
Result/Error Handling
```

### 2. Runtime Selection

```
Configuration Check
     │
     ├─ Explicit Runtime? ──Yes──▶ Use Specified
     │
     ▼
Auto-Detection
     │
     ├─ Docker Available? ──Yes──▶ Use Docker
     │
     ├─ Podman Available? ──Yes──▶ Use Podman
     │
     ├─ Nerdctl Available? ──Yes──▶ Use Nerdctl
     │
     ▼
Error: No Runtime Found
```

### 3. Configuration Loading

```
Default Config
     │
     ▼
Config File (/etc/hpn/config.yaml)
     │
     ▼
User Config (~/.hpn/config.yaml)
     │
     ▼
Environment Variables
     │
     ▼
Command Line Flags
     │
     ▼
Final Configuration
```

## Data Flow

### Pull Operation

```
Image List File
     │
     ▼
Parse Images
     │
     ▼
For Each Image:
├─ Runtime.Pull()
├─ Progress Reporting
└─ Error Handling
     │
     ▼
Summary Report
```

### Save Operation

```
Image List + Mode
     │
     ▼
Determine Output Path
     │
     ▼
For Each Image:
├─ Runtime.Save()
├─ File Creation
└─ Progress Reporting
     │
     ▼
Summary Report
```

### Push Operation

```
Image List + Registry + Mode
     │
     ▼
For Each Image:
├─ Determine Target Name
├─ Runtime.Tag()
├─ Runtime.Push()
└─ Progress Reporting
     │
     ▼
Summary Report
```

## Extensibility Points

### 1. New Container Runtimes

To add support for a new container runtime:

1. Implement the `ContainerRuntime` interface
2. Add detection logic to `detector.go`
3. Register in the runtime factory

### 2. New Output Formats

To add new output formats:

1. Implement the `Formatter` interface
2. Add format option to configuration
3. Register in the formatter factory

### 3. New Storage Backends

To add new storage backends:

1. Implement storage interface
2. Add backend configuration
3. Update service layer

## Performance Considerations

### 1. Parallel Processing

- Configurable worker pools
- Concurrent image operations
- Resource-aware scaling

### 2. Memory Management

- Streaming operations for large files
- Bounded memory usage
- Garbage collection optimization

### 3. Network Optimization

- Connection pooling
- Retry mechanisms
- Bandwidth throttling

## Security Architecture

### 1. Credential Management

- No credentials stored in code
- Environment variable support
- Docker config.json integration
- Secure credential passing

### 2. Runtime Security

- Minimal privilege requirements
- Secure temporary file handling
- Input validation and sanitization

### 3. Network Security

- TLS certificate validation
- Proxy support
- Insecure registry warnings

## Testing Strategy

### 1. Unit Tests

- Interface mocking
- Isolated component testing
- Configuration validation

### 2. Integration Tests

- Real runtime testing
- File system operations
- Network operations

### 3. End-to-End Tests

- Complete workflow testing
- Multi-platform validation
- Performance benchmarking

## Monitoring and Observability

### 1. Logging

- Structured logging (JSON/Text)
- Configurable log levels
- Context propagation

### 2. Metrics

- Operation counters
- Performance metrics
- Error rates

### 3. Tracing

- Operation tracing
- Performance profiling
- Debug information

## Future Considerations

### 1. Plugin System

- Runtime plugins
- Output format plugins
- Storage backend plugins

### 2. API Server

- REST API interface
- gRPC support
- Web UI

### 3. Distributed Operations

- Cluster-aware operations
- Load balancing
- Distributed caching