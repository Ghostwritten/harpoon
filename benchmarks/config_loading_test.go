package benchmarks

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/harpoon/hpn/internal/config"
	"github.com/harpoon/hpn/pkg/types"
)

// BenchmarkConfigLoading tests configuration loading performance
func BenchmarkConfigLoading(b *testing.B) {
	// Create a temporary config file
	configContent := `
registry: "harbor.example.com"
project: "production"
proxy:
  enabled: true
  http: "http://proxy.example.com:8080"
  https: "https://proxy.example.com:8080"
runtime:
  preferred: "docker"
  timeout: "5m"
  auto_fallback: true
  retry:
    max_attempts: 3
    delay: "1s"
    max_delay: "30s"
logging:
  level: "info"
  format: "json"
  console: true
  timestamp: true
  colors: false
parallel:
  max_workers: 8
  auto_adjust: true
modes:
  save_mode: 2
  load_mode: 2
  push_mode: 2
`

	tmpFile, err := ioutil.TempFile("", "hpn-config-*.yaml")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		b.Fatal(err)
	}
	tmpFile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := config.NewManager()
		_, err := manager.Load(tmpFile.Name())
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConfigLoadingMemory tests memory allocation during config loading
func BenchmarkConfigLoadingMemory(b *testing.B) {
	// Create a temporary config file
	configContent := `
registry: "harbor.example.com"
project: "production"
runtime:
  preferred: "docker"
  timeout: "5m"
logging:
  level: "info"
  format: "json"
parallel:
  max_workers: 8
`

	tmpFile, err := ioutil.TempFile("", "hpn-config-*.yaml")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		b.Fatal(err)
	}
	tmpFile.Close()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := config.NewManager()
		_, err := manager.Load(tmpFile.Name())
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDefaultConfigCreation tests default config creation performance
func BenchmarkDefaultConfigCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = types.DefaultConfig()
	}
}

// BenchmarkConfigValidation tests config validation performance
func BenchmarkConfigValidation(b *testing.B) {
	cfg := types.DefaultConfig()
	cfg.Registry = "harbor.example.com"
	cfg.Project = "production"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := config.ValidateConfig(cfg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEnvironmentVariableLoading tests env var loading performance
func BenchmarkEnvironmentVariableLoading(b *testing.B) {
	// Set some environment variables
	os.Setenv("HPN_REGISTRY", "harbor.example.com")
	os.Setenv("HPN_PROJECT", "production")
	os.Setenv("HPN_RUNTIME_PREFERRED", "docker")
	os.Setenv("HPN_LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("HPN_REGISTRY")
		os.Unsetenv("HPN_PROJECT")
		os.Unsetenv("HPN_RUNTIME_PREFERRED")
		os.Unsetenv("HPN_LOG_LEVEL")
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := config.NewManager()
		_, err := manager.Load("")
		if err != nil {
			b.Fatal(err)
		}
	}
}