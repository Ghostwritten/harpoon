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
	Save(ctx context.Context, image string, tarPath string, options SaveOptions) error
	
	// Load loads an image from a tar file
	Load(ctx context.Context, tarPath string) error
	
	// Push pushes an image to a registry
	Push(ctx context.Context, image string, options PushOptions) error
	
	// Tag tags an image with a new name
	Tag(ctx context.Context, source, target string) error
	
	// Login logs in to a container registry
	Login(ctx context.Context, registry string, options LoginOptions) error
	
	// RemoveImage removes an image from local storage (not supported by Skopeo)
	RemoveImage(ctx context.Context, image string, options RmiOptions) error
	
	// ListImages lists images in local storage (not supported by Skopeo)
	ListImages(ctx context.Context) ([]string, error)
	
	// Version returns the runtime version
	Version() (string, error)
}

// RmiOptions contains options for remove image operations
type RmiOptions struct {
	ExtraArgs []string // Passthrough args to underlying runtime (e.g. -f for force)
	Debug     bool     // Enable debug mode to show detailed output
}

// PullOptions contains options for pull operations
type PullOptions struct {
	Proxy     *ProxyConfig
	Retry     RetryConfig
	Timeout   time.Duration
	Platform  string
	MultiArch bool     // When Platform == "all", set to true
	Debug     bool     // Enable debug mode to show detailed output
	ExtraArgs []string // Passthrough args to underlying runtime (e.g. --tls-verify=false)
}

// SaveOptions contains options for save operations
type SaveOptions struct {
	Checksum  bool     // Whether to generate checksum file (default: true)
	MultiArch bool     // Whether to save all architectures (only for Skopeo, when Platform == "all")
	Platform  string   // Target platform (e.g., linux/amd64, linux/arm64). If empty, defaults to linux/amd64 for Skopeo
	Debug     bool     // Enable debug mode to show detailed output
	ExtraArgs []string // Passthrough args to underlying runtime
}

// PushOptions contains options for push operations
type PushOptions struct {
	Timeout   time.Duration
	Retry     RetryConfig
	Debug     bool     // Enable debug mode to show detailed output
	ExtraArgs []string // Passthrough args to underlying runtime
}

// LoginOptions contains options for login operations
type LoginOptions struct {
	Username      string // 用户名
	Password      string // 密码（从 stdin 或环境变量读取）
	PasswordStdin bool   // 是否从 stdin 读取密码
	Insecure      bool   // 是否允许非安全连接（HTTP 或自签名证书）
	Debug         bool   // 是否启用 debug 模式
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