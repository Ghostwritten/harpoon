package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	hpnerrors "github.com/harpoon/hpn/pkg/errors"
	"github.com/harpoon/hpn/pkg/types"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
	if m.config != nil {
		t.Error("new manager should have nil config until Load")
	}
}

func TestLoad_NoConfigFile_UsesDefaults(t *testing.T) {
	m := NewManager()
	cfg, err := m.Load("")
	if err != nil {
		t.Fatalf("Load with no file: %v", err)
	}
	if cfg == nil {
		t.Fatal("config is nil")
	}
	// When no config file is found, defaults come from types.DefaultConfig().
	// If a config file was found (e.g. ~/.hpn/config.yaml), skip default checks.
	if m.GetConfigPath() == "" {
		if cfg.Registry != "registry.k8s.local" {
			t.Errorf("default registry: got %q", cfg.Registry)
		}
		if cfg.Project != "library" {
			t.Errorf("default project: got %q", cfg.Project)
		}
		if cfg.Parallel.MaxWorkers != 5 {
			t.Errorf("default max_workers: got %d", cfg.Parallel.MaxWorkers)
		}
		if cfg.Paths.SavePath != "./images" {
			t.Errorf("default save_path: got %q", cfg.Paths.SavePath)
		}
	}
}

func TestLoad_NonExistentFile_ReturnsError(t *testing.T) {
	m := NewManager()
	_, err := m.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for non-existent config file")
	}
	var he *hpnerrors.HarpoonError
	if !errors.As(err, &he) {
		t.Errorf("expected HarpoonError, got %T", err)
	}
	if he.Code != hpnerrors.ErrConfigNotFound {
		t.Errorf("expected ErrConfigNotFound, got %v", he.Code)
	}
}

func TestLoad_ValidConfigFile_MergesWithDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	const yaml = `
registry: my.registry.local
project: myproject
parallel:
  max_workers: 3
`
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	m := NewManager()
	cfg, err := m.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Registry != "my.registry.local" {
		t.Errorf("registry: got %q", cfg.Registry)
	}
	if cfg.Project != "myproject" {
		t.Errorf("project: got %q", cfg.Project)
	}
	if cfg.Parallel.MaxWorkers != 3 {
		t.Errorf("max_workers: got %d", cfg.Parallel.MaxWorkers)
	}
}

func TestGetConfigPath(t *testing.T) {
	m := NewManager()
	_, _ = m.Load("")
	// When no config file is in search paths, GetConfigPath() is empty.
	// If user has ~/.hpn/config.yaml, GetConfigPath() will be non-empty; we only test explicit file next.
	if m.GetConfigPath() != "" {
		t.Logf("Config file found at %s (e.g. ~/.hpn/config.yaml), testing explicit file load", m.GetConfigPath())
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	_ = os.WriteFile(path, []byte("registry: r\nproject: p\n"), 0644)
	_, _ = m.Load(path)
	if m.GetConfigPath() != path {
		t.Errorf("GetConfigPath after Load(%q): got %q", path, m.GetConfigPath())
	}
}

func TestGetConfig_SetConfig(t *testing.T) {
	m := NewManager()
	_, _ = m.Load("")
	cfg := m.GetConfig()
	if cfg == nil {
		t.Fatal("GetConfig returned nil")
	}

	newCfg := types.DefaultConfig()
	newCfg.Registry = "other.registry"
	m.SetConfig(newCfg)
	got := m.GetConfig()
	if got.Registry != "other.registry" {
		t.Errorf("SetConfig/GetConfig: got registry %q", got.Registry)
	}
}

func TestWriteConfig_EmptyFilename_ReturnsError(t *testing.T) {
	m := NewManager()
	_, _ = m.Load("")
	err := m.WriteConfig("")
	if err == nil {
		t.Fatal("expected error for empty filename")
	}
	var he *hpnerrors.HarpoonError
	if !errors.As(err, &he) || he.Code != hpnerrors.ErrInvalidConfig {
		t.Errorf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestWriteConfig_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "config.yaml")
	m := NewManager()
	_, _ = m.Load("")
	m.config = types.DefaultConfig()
	err := m.WriteConfig(path)
	if err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("config file was not created: %s", path)
	}
}
