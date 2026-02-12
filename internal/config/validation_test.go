package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/harpoon/hpn/pkg/types"
)

func TestValidateConfig_DefaultConfig_Valid(t *testing.T) {
	cfg := types.DefaultConfig()
	if err := ValidateConfig(cfg); err != nil {
		t.Errorf("default config should be valid: %v", err)
	}
}

func TestValidateConfig_EmptyRegistry_Invalid(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Registry = ""
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for empty registry")
	}
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

func TestValidateConfig_RegistryWithProtocol_Invalid(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Registry = "https://registry.example.com"
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for registry with protocol")
	}
}

func TestValidateConfig_EmptyProject_Invalid(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Project = ""
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for empty project")
	}
}

func TestValidateConfig_ProjectInvalidChars_Invalid(t *testing.T) {
	cfg := types.DefaultConfig()
	for _, c := range []string{":", "@", " ", "\t"} {
		cfg.Project = "proj" + c + "ect"
		err := ValidateConfig(cfg)
		if err == nil {
			t.Errorf("expected error for project containing %q", c)
		}
	}
}

func TestValidateConfig_InvalidPreferredRuntime_Invalid(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Runtime.Preferred = "invalid"
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for invalid preferred runtime")
	}
}

func TestValidateConfig_ValidRuntimes_Valid(t *testing.T) {
	for _, r := range []string{"docker", "podman", "nerdctl", "skopeo"} {
		cfg := types.DefaultConfig()
		cfg.Runtime.Preferred = r
		if err := ValidateConfig(cfg); err != nil {
			t.Errorf("runtime %q should be valid: %v", r, err)
		}
	}
}

func TestValidateConfig_ZeroTimeout_Invalid(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Runtime.Timeout = 0
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for zero timeout")
	}
}

func TestValidateConfig_TimeoutExceeds30Min_Invalid(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Runtime.Timeout = 31 * time.Minute
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for timeout > 30 minutes")
	}
}

func TestValidateConfig_RetryMaxAttemptsZero_Invalid(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Runtime.Retry.MaxAttempts = 0
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for max_attempts < 1")
	}
}

func TestValidateConfig_RetryMaxAttemptsOver10_Invalid(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Runtime.Retry.MaxAttempts = 11
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for max_attempts > 10")
	}
}

func TestValidateConfig_InvalidLogLevel_Invalid(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Logging.Level = "invalid"
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for invalid log level")
	}
}

func TestValidateConfig_InvalidLogFormat_Invalid(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Logging.Format = "xml"
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for invalid log format")
	}
}

func TestValidateConfig_MaxWorkersZero_Invalid(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Parallel.MaxWorkers = 0
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for max_workers < 1")
	}
}

func TestValidateConfig_MaxWorkersOver100_Invalid(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Parallel.MaxWorkers = 101
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for max_workers > 100")
	}
}

func TestValidateConfig_SavePathParentDirNotExist_Invalid(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Paths.SavePath = "/nonexistent/parent/dir/images"
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error when save path parent does not exist")
	}
}

func TestValidateConfig_LoadPathParentDirNotExist_Invalid(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Paths.LoadPath = "/nonexistent/parent/dir/images"
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error when load path parent does not exist")
	}
}

func TestValidateConfig_RelativePaths_Valid(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Paths.SavePath = "./images"
	cfg.Paths.LoadPath = "./images"
	if err := ValidateConfig(cfg); err != nil {
		t.Errorf("relative paths should be valid: %v", err)
	}
}

func TestValidateConfig_ExistingParentDir_Valid(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := types.DefaultConfig()
	cfg.Paths.SavePath = filepath.Join(subDir, "images")
	cfg.Paths.LoadPath = filepath.Join(subDir, "images")
	if err := ValidateConfig(cfg); err != nil {
		t.Errorf("existing parent dir should be valid: %v", err)
	}
}

func TestValidateConfig_ProxyDisabled_SkipsURLCheck(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Proxy.Enabled = false
	cfg.Proxy.HTTP = "not-a-valid-url"
	if err := ValidateConfig(cfg); err != nil {
		t.Errorf("disabled proxy with bad URL should still pass: %v", err)
	}
}

func TestValidateConfig_ProxyEnabled_InvalidHTTPURL_Invalid(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Proxy.Enabled = true
	cfg.Proxy.HTTP = "not-a-valid-url"
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for invalid HTTP proxy URL when proxy enabled")
	}
}

func TestValidateDirectory_CreatesDirIfNotExist(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "newdir")
	if err := validateDirectory(dir); err != nil {
		t.Fatalf("validateDirectory should create dir: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Error("path should be a directory")
	}
}

func TestValidateDirectoryExists_NotExist_ReturnsError(t *testing.T) {
	err := validateDirectoryExists(filepath.Join(t.TempDir(), "nonexistent"))
	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
}
