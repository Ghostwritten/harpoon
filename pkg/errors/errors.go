package errors

import (
	"fmt"
)

// HarpoonError represents a custom error with additional context
type HarpoonError struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Cause   error                  `json:"cause,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (e *HarpoonError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *HarpoonError) Unwrap() error {
	return e.Cause
}

// WithContext adds context to the error
func (e *HarpoonError) WithContext(key string, value interface{}) *HarpoonError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// ErrorCode represents different types of errors
type ErrorCode int

const (
	// Runtime errors
	ErrRuntimeNotFound ErrorCode = iota + 1000
	ErrRuntimeUnavailable
	ErrRuntimeCommand

	// Image errors
	ErrImageNotFound
	ErrImageInvalid
	ErrImageParsing

	// Registry errors
	ErrRegistryAuth
	ErrRegistryConnection
	ErrRegistryTimeout

	// Network errors
	ErrNetworkTimeout
	ErrNetworkConnection
	ErrProxyConnection

	// File system errors
	ErrInsufficientSpace
	ErrFileNotFound
	ErrFilePermission
	ErrFileOperation

	// Configuration errors
	ErrInvalidConfig
	ErrConfigNotFound
	ErrConfigParsing
)

// String returns the string representation of the error code
func (e ErrorCode) String() string {
	switch e {
	case ErrRuntimeNotFound:
		return "RUNTIME_NOT_FOUND"
	case ErrRuntimeUnavailable:
		return "RUNTIME_UNAVAILABLE"
	case ErrRuntimeCommand:
		return "RUNTIME_COMMAND"
	case ErrImageNotFound:
		return "IMAGE_NOT_FOUND"
	case ErrImageInvalid:
		return "IMAGE_INVALID"
	case ErrImageParsing:
		return "IMAGE_PARSING"
	case ErrRegistryAuth:
		return "REGISTRY_AUTH"
	case ErrRegistryConnection:
		return "REGISTRY_CONNECTION"
	case ErrRegistryTimeout:
		return "REGISTRY_TIMEOUT"
	case ErrNetworkTimeout:
		return "NETWORK_TIMEOUT"
	case ErrNetworkConnection:
		return "NETWORK_CONNECTION"
	case ErrProxyConnection:
		return "PROXY_CONNECTION"
	case ErrInsufficientSpace:
		return "INSUFFICIENT_SPACE"
	case ErrFileNotFound:
		return "FILE_NOT_FOUND"
	case ErrFilePermission:
		return "FILE_PERMISSION"
	case ErrFileOperation:
		return "FILE_OPERATION"
	case ErrInvalidConfig:
		return "INVALID_CONFIG"
	case ErrConfigNotFound:
		return "CONFIG_NOT_FOUND"
	case ErrConfigParsing:
		return "CONFIG_PARSING"
	default:
		return "UNKNOWN"
	}
}

// New creates a new HarpoonError
func New(code ErrorCode, message string) *HarpoonError {
	return &HarpoonError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code ErrorCode, message string) *HarpoonError {
	return &HarpoonError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// Common error constructors
func NewRuntimeNotFound(runtime string) *HarpoonError {
	return New(ErrRuntimeNotFound, fmt.Sprintf("container runtime '%s' not found", runtime)).
		WithContext("runtime", runtime)
}

func NewImageNotFound(image string) *HarpoonError {
	return New(ErrImageNotFound, fmt.Sprintf("image '%s' not found", image)).
		WithContext("image", image)
}

func NewRegistryAuthError(registry string) *HarpoonError {
	return New(ErrRegistryAuth, fmt.Sprintf("authentication failed for registry '%s'", registry)).
		WithContext("registry", registry)
}

func NewInsufficientSpace(required, available int64) *HarpoonError {
	return New(ErrInsufficientSpace, fmt.Sprintf("insufficient disk space: required %d bytes, available %d bytes", required, available)).
		WithContext("required", required).
		WithContext("available", available)
}