package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/harpoon/hpn/pkg/errors"
	"github.com/harpoon/hpn/pkg/types"
)

// Manager handles configuration loading and management
type Manager struct {
	config *types.Config
	viper  *viper.Viper
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		viper: viper.New(),
	}
}

// Load loads configuration from various sources with proper priority
func (m *Manager) Load(configFile string) (*types.Config, error) {
	// Start with default configuration
	m.config = types.DefaultConfig()

	// Set up viper
	m.setupViper()

	// Load configuration file if specified or found
	if err := m.loadConfigFile(configFile); err != nil {
		return nil, err
	}

	// Load environment variables
	m.loadEnvironmentVariables()

	// Unmarshal into config struct
	if err := m.viper.Unmarshal(m.config); err != nil {
		return nil, errors.Wrap(err, errors.ErrConfigParsing, "failed to parse configuration")
	}

	// Validate configuration
	if err := ValidateConfig(m.config); err != nil {
		return nil, err
	}

	return m.config, nil
}

// setupViper configures viper settings
func (m *Manager) setupViper() {
	m.viper.SetConfigName("config")
	m.viper.SetConfigType("yaml")
	
	// Add config search paths
	m.viper.AddConfigPath(".")
	m.viper.AddConfigPath("$HOME/.hpn")
	m.viper.AddConfigPath("/etc/hpn")

	// Set environment variable prefix
	m.viper.SetEnvPrefix("HPN")
	m.viper.AutomaticEnv()
	
	// Replace dots and dashes with underscores for env vars
	m.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
}

// loadConfigFile loads configuration from file
func (m *Manager) loadConfigFile(configFile string) error {
	if configFile != "" {
		// Use specified config file
		m.viper.SetConfigFile(configFile)
		if err := m.viper.ReadInConfig(); err != nil {
			if os.IsNotExist(err) {
				return errors.New(errors.ErrConfigNotFound, fmt.Sprintf("config file not found: %s", configFile))
			}
			return errors.Wrap(err, errors.ErrConfigParsing, fmt.Sprintf("failed to read config file: %s", configFile))
		}
	} else {
		// Try to find config file in search paths
		if err := m.viper.ReadInConfig(); err != nil {
			// It's okay if no config file is found, we'll use defaults
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return errors.Wrap(err, errors.ErrConfigParsing, "failed to read config file")
			}
		}
	}

	return nil
}

// loadEnvironmentVariables sets up environment variable mappings
func (m *Manager) loadEnvironmentVariables() {
	// Map environment variables to config keys
	envMappings := map[string]string{
		"HPN_REGISTRY":           "registry",
		"HPN_PROJECT":            "project",
		"HPN_PROXY_HTTP":         "proxy.http",
		"HPN_PROXY_HTTPS":        "proxy.https",
		"HPN_PROXY_ENABLED":      "proxy.enabled",
		"HPN_RUNTIME_PREFERRED":  "runtime.preferred",
		"HPN_RUNTIME_TIMEOUT":    "runtime.timeout",
		"HPN_LOG_LEVEL":          "logging.level",
		"HPN_LOG_FORMAT":         "logging.format",
		"HPN_LOG_FILE":           "logging.file",
		"HPN_LOG_CONSOLE":        "logging.console",
		"HPN_PARALLEL_MAX":       "parallel.max_workers",
		"HPN_PARALLEL_AUTO":      "parallel.auto_adjust",
	}

	for envVar, configKey := range envMappings {
		if value := os.Getenv(envVar); value != "" {
			m.viper.Set(configKey, value)
		}
	}

	// Handle proxy environment variables (standard names)
	if httpProxy := os.Getenv("http_proxy"); httpProxy != "" {
		m.viper.Set("proxy.http", httpProxy)
		m.viper.Set("proxy.enabled", true)
	}
	if httpsProxy := os.Getenv("https_proxy"); httpsProxy != "" {
		m.viper.Set("proxy.https", httpsProxy)
		m.viper.Set("proxy.enabled", true)
	}
}

// GetConfigPath returns the path of the loaded config file
func (m *Manager) GetConfigPath() string {
	return m.viper.ConfigFileUsed()
}

// WriteConfig writes the current configuration to a file
func (m *Manager) WriteConfig(filename string) error {
	if filename == "" {
		return errors.New(errors.ErrInvalidConfig, "config filename cannot be empty")
	}

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(err, errors.ErrFileOperation, "failed to create config directory")
	}

	// Write config file
	m.viper.SetConfigFile(filename)
	if err := m.viper.WriteConfig(); err != nil {
		return errors.Wrap(err, errors.ErrFileOperation, "failed to write config file")
	}

	return nil
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *types.Config {
	return m.config
}

// SetConfig sets the configuration
func (m *Manager) SetConfig(config *types.Config) {
	m.config = config
}