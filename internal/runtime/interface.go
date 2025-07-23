package runtime

import (
	"context"
	"time"
)

// ContainerRuntime defines the interface for container runtime operations
type ContainerRuntime interface {
	// Name returns the name of the runtime (docker, podman, nerdctl)
	Name() string
	
	// IsAvailable checks if the runtime is available on the system
	IsAvailable() bool
	
	// Pull pulls an image from a registry
	Pull(ctx context.Context, image string, options PullOptions) error
	
	// Save saves an image to a tar file
	Save(ctx context.Context, image string, tarPath string) error
	
	// Load loads an image from a tar file
	Load(ctx context.Context, tarPath string) error
	
	// Push pushes an image to a registry
	Push(ctx context.Context, image string, options PushOptions) error
	
	// Tag tags an image with a new name
	Tag(ctx context.Context, source, target string) error
	
	// Version returns the runtime version
	Version() (string, error)
}

// PullOptions contains options for pull operations
type PullOptions struct {
	Proxy     *ProxyConfig
	Retry     RetryConfig
	Timeout   time.Duration
	Platform  string
}

// PushOptions contains options for push operations
type PushOptions struct {
	Timeout time.Duration
	Retry   RetryConfig
}

// ProxyConfig contains proxy configuration
type ProxyConfig struct {
	HTTP    string
	HTTPS   string
	Enabled bool
}

// RetryConfig contains retry configuration
type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
	MaxDelay    time.Duration
}

// RuntimeDetector detects and manages available container runtimes
type RuntimeDetector interface {
	// DetectAvailable returns all available runtimes
	DetectAvailable() []ContainerRuntime
	
	// GetPreferred returns the preferred runtime based on priority
	GetPreferred() ContainerRuntime
	
	// GetByName returns a runtime by name
	GetByName(name string) (ContainerRuntime, error)
}