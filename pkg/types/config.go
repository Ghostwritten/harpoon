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
	Modes    ModeConfig     `yaml:"modes" json:"modes" mapstructure:"modes"`
}

// ProxyConfig contains proxy settings
type ProxyConfig struct {
	HTTP    string `yaml:"http" json:"http" mapstructure:"http"`
	HTTPS   string `yaml:"https" json:"https" mapstructure:"https"`
	Enabled bool   `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
}

// RuntimeConfig contains container runtime settings
type RuntimeConfig struct {
	Preferred string        `yaml:"preferred" json:"preferred" mapstructure:"preferred"`
	Timeout   time.Duration `yaml:"timeout" json:"timeout" mapstructure:"timeout"`
	Retry     RetryConfig   `yaml:"retry" json:"retry" mapstructure:"retry"`
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

// ModeConfig contains default operation modes
type ModeConfig struct {
	SaveMode SaveMode `yaml:"save_mode" json:"save_mode" mapstructure:"save_mode"`
	LoadMode LoadMode `yaml:"load_mode" json:"load_mode" mapstructure:"load_mode"`
	PushMode PushMode `yaml:"push_mode" json:"push_mode" mapstructure:"push_mode"`
}

// RetryConfig contains retry settings
type RetryConfig struct {
	MaxAttempts int           `yaml:"max_attempts" json:"max_attempts" mapstructure:"max_attempts"`
	Delay       time.Duration `yaml:"delay" json:"delay" mapstructure:"delay"`
	MaxDelay    time.Duration `yaml:"max_delay" json:"max_delay" mapstructure:"max_delay"`
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

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		Registry: "registry.k8s.local",
		Project:  "library",
		Proxy: ProxyConfig{
			Enabled: false,
		},
		Runtime: RuntimeConfig{
			Preferred: "",
			Timeout:   5 * time.Minute,
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
		Modes: ModeConfig{
			SaveMode: SaveModeCurrentDir,
			LoadMode: LoadModeCurrentDir,
			PushMode: PushModeSimple,
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