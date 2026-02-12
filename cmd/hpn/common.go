package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/harpoon/hpn/internal/config"
	containerruntime "github.com/harpoon/hpn/internal/runtime"
	"github.com/harpoon/hpn/pkg/types"
)

var (
	cfg             *types.Config
	configMgr       *config.Manager
	runtimeDetector *containerruntime.Detector
	configFile      string
	runtimeName     string
	autoFallback    bool
)

func init() {
	configMgr = config.NewManager()
	runtimeDetector = containerruntime.NewDetector()
}

// readImageList reads image list from file: one image per line, skip empty lines and # comments.
func readImageList(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var images []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		images = append(images, line)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("no images found in file (empty or comments only)")
	}
	return images, nil
}

func loadConfigIfNeeded() error {
	if cfg != nil {
		return nil
	}
	var err error
	cfg, err = configMgr.Load(configFile)
	return err
}

func selectContainerRuntime() (containerruntime.ContainerRuntime, error) {
	if err := loadConfigIfNeeded(); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if runtimeName != "" {
		return runtimeDetector.GetByName(runtimeName)
	}
	if cfg != nil && cfg.Runtime.Preferred != "" {
		r, err := runtimeDetector.GetByName(cfg.Runtime.Preferred)
		if err == nil {
			return r, nil
		}
		if !autoFallback && !cfg.Runtime.AutoFallback {
			return nil, err
		}
	}
	r := runtimeDetector.GetPreferred()
	if r == nil {
		return nil, fmt.Errorf("no container runtime available (docker, podman, nerdctl, or skopeo)")
	}
	return r, nil
}

func IsDebug() bool {
	if cfg != nil && cfg.Logging.Level == "debug" {
		return true
	}
	return false
}

func mergeExtraArgs(configArgs, passthrough []string) []string {
	out := make([]string, 0, len(configArgs)+len(passthrough))
	out = append(out, configArgs...)
	out = append(out, passthrough...)
	return out
}

func getConfigExtraArgsPull() []string {
	if cfg != nil && cfg.Runtime.ExtraArgs != nil {
		return cfg.Runtime.ExtraArgs.Pull
	}
	return nil
}

func getConfigExtraArgsSave() []string {
	if cfg != nil && cfg.Runtime.ExtraArgs != nil {
		return cfg.Runtime.ExtraArgs.Save
	}
	return nil
}

func getConfigExtraArgsPush() []string {
	if cfg != nil && cfg.Runtime.ExtraArgs != nil {
		return cfg.Runtime.ExtraArgs.Push
	}
	return nil
}
