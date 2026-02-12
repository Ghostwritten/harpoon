package chart

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Options holds chart fetch options.
type Options struct {
	// Chart is either "repo/name" (e.g. bitnami/nginx) or a local path (e.g. ./nginx.tgz or ./chart-dir).
	Chart string
	// Version is required when Chart is "repo/name"; ignored for local paths.
	Version string
}

// FetchResult holds the chart directory and a cleanup function.
type FetchResult struct {
	Dir     string
	Cleanup func()
}

// FetchChart fetches and prepares a chart directory. For remote charts (repo/name) it runs helm pull.
// For local paths it uses the path as-is or extracts a .tgz to a temp dir. Caller must call Cleanup when done.
func FetchChart(ctx context.Context, opts Options) (*FetchResult, error) {
	chart := strings.TrimSpace(opts.Chart)
	if chart == "" {
		return nil, fmt.Errorf("chart is required")
	}

	// Check if chart is a local path: exists as file/dir, or is absolute, or starts with . or ..
	if isLocalChart(chart) {
		abs, err := filepath.Abs(chart)
		if err != nil {
			return nil, fmt.Errorf("chart path: %w", err)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("chart path %s: %w", chart, err)
		}
		if info.IsDir() {
			return &FetchResult{Dir: abs, Cleanup: func() {}}, nil
		}
		// .tgz: extract to temp dir
		dir, err := os.MkdirTemp("", "hpn-chart-*")
		if err != nil {
			return nil, fmt.Errorf("create temp dir: %w", err)
		}
		if err := extractTgz(abs, dir); err != nil {
			os.RemoveAll(dir)
			return nil, fmt.Errorf("extract chart: %w", err)
		}
		// tgz often contains a single top-level dir (e.g. nginx/)
		entries, _ := os.ReadDir(dir)
		chartDir := dir
		if len(entries) == 1 && entries[0].IsDir() {
			chartDir = filepath.Join(dir, entries[0].Name())
		}
		return &FetchResult{
			Dir:     chartDir,
			Cleanup: func() { os.RemoveAll(dir) },
		}, nil
	}

	// Remote: repo/name
	if opts.Version == "" {
		return nil, fmt.Errorf("version is required when using remote chart %s", chart)
	}
	if err := ensureHelm(ctx); err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp("", "hpn-chart-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	cmd := exec.CommandContext(ctx, "helm", "pull", chart, "--version", opts.Version, "--untar", "--untardir", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.RemoveAll(dir)
		return nil, fmt.Errorf("helm pull %s --version %s: %w", chart, opts.Version, err)
	}
	// helm pull --untar --untardir D extracts to D/<chart-name>/
	entries, _ := os.ReadDir(dir)
	var chartDir string
	for _, e := range entries {
		if e.IsDir() {
			chartDir = filepath.Join(dir, e.Name())
			break
		}
	}
	if chartDir == "" {
		os.RemoveAll(dir)
		return nil, fmt.Errorf("helm pull did not create a chart directory under %s", dir)
	}
	return &FetchResult{
		Dir:     chartDir,
		Cleanup: func() { os.RemoveAll(dir) },
	}, nil
}

func isLocalChart(s string) bool {
	if s == "" {
		return false
	}
	if filepath.IsAbs(s) {
		return true
	}
	if s == "." || s == ".." || strings.HasPrefix(s, "./") || strings.HasPrefix(s, "../") {
		return true
	}
	// If it exists as a file or directory, treat as local path (e.g. path/to/chart.tgz or ./chart)
	if _, err := os.Stat(s); err == nil {
		return true
	}
	// "repo/name" (e.g. bitnami/nginx) does not exist as path -> remote
	return false
}

func extractTgz(tgzPath, destDir string) error {
	f, err := os.Open(tgzPath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		target := filepath.Join(destDir, h.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			return fmt.Errorf("tar entry outside destination: %s", h.Name)
		}
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			w, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(w, tr); err != nil {
				w.Close()
				return err
			}
			w.Close()
		}
	}
	return nil
}

func ensureHelm(ctx context.Context) error {
	if _, err := exec.LookPath("helm"); err != nil {
		return fmt.Errorf("helm CLI is required but not found in PATH: %w. Install Helm from https://helm.sh/docs/intro/install/", err)
	}
	cmd := exec.CommandContext(ctx, "helm", "version", "--short")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("helm version check failed: %w", err)
	}
	return nil
}
