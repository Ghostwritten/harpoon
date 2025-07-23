package service

import (
	"context"
	"time"

	"github.com/harpoon/hpn/internal/runtime"
)

// ImageService defines the interface for image operations
type ImageService interface {
	Pull(ctx context.Context, req PullRequest) (*OperationResult, error)
	Save(ctx context.Context, req SaveRequest) (*OperationResult, error)
	Load(ctx context.Context, req LoadRequest) (*OperationResult, error)
	Push(ctx context.Context, req PushRequest) (*OperationResult, error)
}

// PullRequest contains parameters for pull operations
type PullRequest struct {
	Images      []string
	Parallel    int
	ProxyConfig *runtime.ProxyConfig
	Retry       runtime.RetryConfig
	Timeout     time.Duration
}

// SaveRequest contains parameters for save operations
type SaveRequest struct {
	Images   []string
	Mode     SaveMode
	Parallel int
	BaseDir  string
}

// LoadRequest contains parameters for load operations
type LoadRequest struct {
	Mode     LoadMode
	Parallel int
	BaseDir  string
	Pattern  string
}

// PushRequest contains parameters for push operations
type PushRequest struct {
	Images   []string
	Registry string
	Project  string
	Mode     PushMode
	Parallel int
	Timeout  time.Duration
}

// OperationResult contains the result of an operation
type OperationResult struct {
	Success  []string          `json:"success"`
	Failed   []FailedOperation `json:"failed"`
	Duration time.Duration     `json:"duration"`
	Summary  string            `json:"summary"`
}

// FailedOperation represents a failed operation
type FailedOperation struct {
	Item  string `json:"item"`
	Error string `json:"error"`
}

// SaveMode defines how images are saved
type SaveMode int

const (
	SaveModeCurrentDir SaveMode = iota + 1 // Save to current directory
	SaveModeImagesDir                      // Save to ./images/
	SaveModeProjectDir                     // Save to ./images/<project>/
)

// LoadMode defines how images are loaded
type LoadMode int

const (
	LoadModeCurrentDir LoadMode = iota + 1 // Load from current directory
	LoadModeImagesDir                      // Load from ./images/
	LoadModeRecursive                      // Load recursively from ./images/*/
)

// PushMode defines how images are pushed
type PushMode int

const (
	PushModeSimple   PushMode = iota + 1 // registry/image:tag
	PushModeProject                      // registry/project/image:tag
	PushModePreserve                     // Preserve original project path
)