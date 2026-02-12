package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/harpoon/hpn/pkg/errors"
	"github.com/harpoon/hpn/pkg/types"
)

// ValidateConfig validates the entire configuration
func ValidateConfig(cfg *types.Config) error {
	if err := validateRegistry(cfg.Registry); err != nil {
		return err
	}

	if err := validateProject(cfg.Project); err != nil {
		return err
	}

	if err := validateProxyConfig(&cfg.Proxy); err != nil {
		return err
	}

	if err := validateRuntimeConfig(&cfg.Runtime); err != nil {
		return err
	}

	if err := validateLoggingConfig(&cfg.Logging); err != nil {
		return err
	}

	if err := validateParallelConfig(&cfg.Parallel); err != nil {
		return err
	}

	if err := validatePathConfig(&cfg.Paths); err != nil {
		return err
	}

	return nil
}

// validateRegistry validates registry configuration
func validateRegistry(registry string) error {
	if registry == "" {
		return errors.New(errors.ErrInvalidConfig, "registry cannot be empty")
	}

	// Check if it's a valid hostname or IP
	if strings.Contains(registry, "://") {
		return errors.New(errors.ErrInvalidConfig, "registry should not include protocol (http/https)")
	}

	return nil
}

// validateProject validates project configuration
func validateProject(project string) error {
	if project == "" {
		return errors.New(errors.ErrInvalidConfig, "project cannot be empty")
	}

	// Check for invalid characters
	invalidChars := []string{":", "@", " ", "\t", "\n"}
	for _, char := range invalidChars {
		if strings.Contains(project, char) {
			return errors.New(errors.ErrInvalidConfig, fmt.Sprintf("project name contains invalid character: %s", char))
		}
	}

	return nil
}

// validateProxyConfig validates proxy configuration
func validateProxyConfig(proxy *types.ProxyConfig) error {
	if !proxy.Enabled {
		return nil
	}

	if proxy.HTTP != "" {
		if err := validateProxyURL(proxy.HTTP); err != nil {
			return errors.Wrap(err, errors.ErrInvalidConfig, "invalid HTTP proxy URL")
		}
	}

	if proxy.HTTPS != "" {
		if err := validateProxyURL(proxy.HTTPS); err != nil {
			return errors.Wrap(err, errors.ErrInvalidConfig, "invalid HTTPS proxy URL")
		}
	}

	return nil
}

// validateProxyURL validates a proxy URL
func validateProxyURL(proxyURL string) error {
	if proxyURL == "" {
		return nil
	}

	u, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("proxy URL must use http or https scheme")
	}

	if u.Host == "" {
		return fmt.Errorf("proxy URL must have a host")
	}

	return nil
}

// validateRuntimeConfig validates runtime configuration
func validateRuntimeConfig(runtime *types.RuntimeConfig) error {
	if runtime.Preferred != "" {
		validRuntimes := []string{"docker", "podman", "nerdctl", "skopeo"}
		valid := false
		for _, r := range validRuntimes {
			if runtime.Preferred == r {
				valid = true
				break
			}
		}
		if !valid {
			return errors.New(errors.ErrInvalidConfig, fmt.Sprintf("invalid preferred runtime: %s (must be one of: %s)", runtime.Preferred, strings.Join(validRuntimes, ", ")))
		}
	}

	if runtime.Timeout <= 0 {
		return errors.New(errors.ErrInvalidConfig, "runtime timeout must be positive")
	}

	if runtime.Timeout > 30*time.Minute {
		return errors.New(errors.ErrInvalidConfig, "runtime timeout cannot exceed 30 minutes")
	}

	return validateRetryConfig(&runtime.Retry)
}

// validateRetryConfig validates retry configuration
func validateRetryConfig(retry *types.RetryConfig) error {
	if retry.MaxAttempts < 1 {
		return errors.New(errors.ErrInvalidConfig, "retry max attempts must be at least 1")
	}

	if retry.MaxAttempts > 10 {
		return errors.New(errors.ErrInvalidConfig, "retry max attempts cannot exceed 10")
	}

	if retry.Delay <= 0 {
		return errors.New(errors.ErrInvalidConfig, "retry delay must be positive")
	}

	if retry.MaxDelay <= retry.Delay {
		return errors.New(errors.ErrInvalidConfig, "retry max delay must be greater than delay")
	}

	return nil
}

// validateLoggingConfig validates logging configuration
func validateLoggingConfig(logging *types.LoggingConfig) error {
	validLevels := []string{"debug", "info", "warn", "error"}
	valid := false
	for _, level := range validLevels {
		if logging.Level == level {
			valid = true
			break
		}
	}
	if !valid {
		return errors.New(errors.ErrInvalidConfig, fmt.Sprintf("invalid log level: %s (must be one of: %s)", logging.Level, strings.Join(validLevels, ", ")))
	}

	validFormats := []string{"text", "json"}
	valid = false
	for _, format := range validFormats {
		if logging.Format == format {
			valid = true
			break
		}
	}
	if !valid {
		return errors.New(errors.ErrInvalidConfig, fmt.Sprintf("invalid log format: %s (must be one of: %s)", logging.Format, strings.Join(validFormats, ", ")))
	}

	if logging.File != "" {
		// Check if the directory exists and is writable
		dir := filepath.Dir(logging.File)
		if err := validateDirectory(dir); err != nil {
			return errors.Wrap(err, errors.ErrInvalidConfig, "invalid log file directory")
		}
	}

	return nil
}

// validateParallelConfig validates parallel processing configuration
func validateParallelConfig(parallel *types.ParallelConfig) error {
	if parallel.MaxWorkers < 1 {
		return errors.New(errors.ErrInvalidConfig, "max workers must be at least 1")
	}

	if parallel.MaxWorkers > 100 {
		return errors.New(errors.ErrInvalidConfig, "max workers cannot exceed 100")
	}

	return nil
}

// validatePathConfig validates path configuration
// Note: This function does NOT create directories, it only validates that the path
// is valid and the parent directory is writable. Directories should be created
// when actually used, not during configuration validation.
func validatePathConfig(paths *types.PathConfig) error {
	if paths.SavePath != "" {
		// Only validate parent directory exists and is writable
		// Don't create the path itself - it will be created when used
		parentDir := filepath.Dir(paths.SavePath)
		if parentDir != "." && parentDir != paths.SavePath {
			if err := validateDirectoryExists(parentDir); err != nil {
				return errors.Wrap(err, errors.ErrInvalidConfig, "invalid save path parent directory")
			}
		}
	}

	if paths.LoadPath != "" {
		// Only validate parent directory exists and is writable
		// Don't create the path itself - it will be created when used
		parentDir := filepath.Dir(paths.LoadPath)
		if parentDir != "." && parentDir != paths.LoadPath {
			if err := validateDirectoryExists(parentDir); err != nil {
				return errors.Wrap(err, errors.ErrInvalidConfig, "invalid load path parent directory")
			}
		}
	}

	return nil
}

// validateDirectory checks if a directory exists and is writable
// This function will create the directory if it doesn't exist (used for log file directories)
func validateDirectory(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// Try to create the directory
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("cannot create directory: %v", err)
			}
			return nil
		}
		return fmt.Errorf("cannot access directory: %v", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory")
	}

	// Test write permission by creating a temporary file
	tempFile := filepath.Join(dir, ".hpn_write_test")
	f, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("directory is not writable: %v", err)
	}
	f.Close()
	os.Remove(tempFile)

	return nil
}

// validateDirectoryExists checks if a directory exists and is writable
// This function does NOT create the directory - it only validates existing directories
func validateDirectoryExists(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", dir)
		}
		return fmt.Errorf("cannot access directory: %v", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory")
	}

	// Test write permission by creating a temporary file
	tempFile := filepath.Join(dir, ".hpn_write_test")
	f, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("directory is not writable: %v", err)
	}
	f.Close()
	os.Remove(tempFile)

	return nil
}