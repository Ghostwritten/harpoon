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

// readImageList reads an image list from path (or os.Stdin when path is "-").
// Lines are trimmed; empty lines and lines beginning with "#" are skipped.
// Validation happens before deduplication so error messages reference the
// correct original line number in the file.
func readImageList(path string) ([]string, error) {
	var f *os.File
	if path == "-" {
		f = os.Stdin
	} else {
		var err error
		f, err = os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
	}

	var raw []string   // all non-empty, non-comment lines in file order
	var lineNums []int // original 1-based line numbers

	sc := bufio.NewScanner(f)
	lineNum := 0
	for sc.Scan() {
		lineNum++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		raw = append(raw, line)
		lineNums = append(lineNums, lineNum)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("no images found in file (empty or comments only)")
	}

	// Validate before deduplication so the reported line number is correct.
	for i, img := range raw {
		if _, parseErr := types.ParseImage(img); parseErr != nil {
			return nil, fmt.Errorf("invalid image at line %d: %q: %w", lineNums[i], img, parseErr)
		}
	}

	return dedupeImages(raw), nil
}

// validateImageList validates each image with types.ParseImage.
// Returns the 1-based position and invalid value of the first invalid image,
// or (0, "", nil) if all are valid.
func validateImageList(images []string) (lineNum int, invalid string, err error) {
	for i, img := range images {
		if _, err := types.ParseImage(img); err != nil {
			return i + 1, img, err
		}
	}
	return 0, "", nil
}

// dedupeImages removes duplicates while preserving the first-occurrence order.
func dedupeImages(images []string) []string {
	seen := make(map[string]struct{}, len(images))
	out := make([]string, 0, len(images))
	for _, img := range images {
		if _, ok := seen[img]; ok {
			continue
		}
		seen[img] = struct{}{}
		out = append(out, img)
	}
	return out
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

// IsDebug reports whether debug logging is enabled via configuration.
func IsDebug() bool {
	return cfg != nil && cfg.Logging.Level == "debug"
}

func mergeExtraArgs(configArgs, passthrough []string) []string {
	out := make([]string, 0, len(configArgs)+len(passthrough))
	out = append(out, configArgs...)
	out = append(out, passthrough...)
	return out
}

// getConfigExtraArgs returns per-operation extra args from config.
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

func getConfigExtraArgsRmi() []string {
	if cfg != nil && cfg.Runtime.ExtraArgs != nil {
		return cfg.Runtime.ExtraArgs.Rmi
	}
	return nil
}
