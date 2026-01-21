package types

import (
	"time"

	"github.com/harpoon/hpn/internal/runtime"
)

// Config represents the application configuration
type Config struct {
	Registry string         `yaml:"registry" json:"registry" mapstructure:"registry"`
	Project  string         `yaml:"project" json:"project" mapstructure:"project"`
	Proxy    ProxyConfig    `yaml:"proxy" json:"proxy" mapstructure:"proxy"`
	Runtime  RuntimeConfig  `yaml:"runtime" json:"runtime" mapstructure:"runtime"`
	Logging  LoggingConfig  `yaml:"logging" json:"logging" mapstructure:"logging"`
	Parallel ParallelConfig `yaml:"parallel" json:"parallel" mapstructure:"parallel"`
	Paths    PathConfig     `yaml:"paths" json:"paths" mapstructure:"paths"`
}

// ProxyConfig contains proxy settings
type ProxyConfig struct {
	HTTP    string `yaml:"http" json:"http" mapstructure:"http"`
	HTTPS   string `yaml:"https" json:"https" mapstructure:"https"`
	Enabled bool   `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
}

// RuntimeConfig contains container runtime settings
type RuntimeConfig struct {
	Preferred    string        `yaml:"preferred" json:"preferred" mapstructure:"preferred"`
	Timeout      time.Duration `yaml:"timeout" json:"timeout" mapstructure:"timeout"`
	Retry        RetryConfig   `yaml:"retry" json:"retry" mapstructure:"retry"`
	AutoFallback bool          `yaml:"auto_fallback" json:"auto_fallback" mapstructure:"auto_fallback"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level     string `yaml:"level" json:"level" mapstructure:"level"`
	Format    string `yaml:"format" json:"format" mapstructure:"format"` // "text" or "json"
	File      string `yaml:"file" json:"file" mapstructure:"file"`
	Console   bool   `yaml:"console" json:"console" mapstructure:"console"`
	Timestamp bool   `yaml:"timestamp" json:"timestamp" mapstructure:"timestamp"`
	Colors    bool   `yaml:"colors" json:"colors" mapstructure:"colors"`
}

// ParallelConfig contains parallel processing settings
type ParallelConfig struct {
	MaxWorkers int  `yaml:"max_workers" json:"max_workers" mapstructure:"max_workers"`
	AutoAdjust bool `yaml:"auto_adjust" json:"auto_adjust" mapstructure:"auto_adjust"`
}

// PathConfig contains default paths for operations
type PathConfig struct {
	SavePath string `yaml:"save_path" json:"save_path" mapstructure:"save_path"`
	LoadPath string `yaml:"load_path" json:"load_path" mapstructure:"load_path"`
}

// RetryConfig contains retry settings
type RetryConfig struct {
	MaxAttempts int           `yaml:"max_attempts" json:"max_attempts" mapstructure:"max_attempts"`
	Delay       time.Duration `yaml:"delay" json:"delay" mapstructure:"delay"`
	MaxDelay    time.Duration `yaml:"max_delay" json:"max_delay" mapstructure:"max_delay"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		Registry: "registry.k8s.local",
		Project:  "library",
		Proxy: ProxyConfig{
			Enabled: false,
		},
		Runtime: RuntimeConfig{
			Preferred:    "",
			Timeout:      5 * time.Minute,
			AutoFallback: false,
			Retry: RetryConfig{
				MaxAttempts: 3,
				Delay:       time.Second,
				MaxDelay:    30 * time.Second,
			},
		},
		Logging: LoggingConfig{
			Level:     "info",
			Format:    "text",
			Console:   true,
			Timestamp: true,
			Colors:    true,
		},
		Parallel: ParallelConfig{
			MaxWorkers: 4,
			AutoAdjust: true,
		},
		Paths: PathConfig{
			SavePath: "./images",
			LoadPath: "./images",
		},
	}
}

// ToRuntimeProxyConfig converts ProxyConfig to runtime.ProxyConfig
func (p *ProxyConfig) ToRuntimeProxyConfig() *runtime.ProxyConfig {
	return &runtime.ProxyConfig{
		HTTP:    p.HTTP,
		HTTPS:   p.HTTPS,
		Enabled: p.Enabled,
	}
}

// ToRuntimeRetryConfig converts RetryConfig to runtime.RetryConfig
func (r *RetryConfig) ToRuntimeRetryConfig() runtime.RetryConfig {
	return runtime.RetryConfig{
		MaxAttempts: r.MaxAttempts,
		Delay:       r.Delay,
		MaxDelay:    r.MaxDelay,
	}
}