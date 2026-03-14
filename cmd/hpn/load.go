package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	containerruntime "github.com/harpoon/hpn/internal/runtime"
)

var (
	loadPath       string
	loadRecursive  bool
	loadWorkers    int
	loadSkipVerify bool
)

var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Load images from tar files",
	Long: `Load container images from tar files. Use --path (-p) to specify the directory.
If a .sha256 file exists next to a tar file, checksum verification is performed before load.
Use --skip-verify to bypass verification.`,
	RunE: executeLoad,
}

func init() {
	loadCmd.Flags().StringVarP(&loadPath, "path", "p", "./images", "Input directory containing tar files (default: ./images)")
	loadCmd.Flags().BoolVar(&loadRecursive, "recursive", false, "Load recursively from subdirectories")
	loadCmd.Flags().BoolVar(&loadSkipVerify, "skip-verify", false, "Skip checksum verification for tar files")
	loadCmd.Flags().IntVar(&loadWorkers, "workers", 0, "Number of concurrent workers (0 = use config or default 5)")
}

func executeLoad(cmd *cobra.Command, args []string) error {
	loadDir := loadPath
	if loadDir == "" {
		loadDir = "./images"
	}

	fmt.Printf("Executing load action from: %s\n", loadDir)

	selectedRuntime, err := selectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime selection failed: %v", err)
	}
	fmt.Printf("Using container runtime: %s\n", selectedRuntime.Name())

	tarFiles, err := findTarFiles(loadDir, loadRecursive)
	if err != nil {
		return fmt.Errorf("failed to find tar files: %v", err)
	}
	if len(tarFiles) == 0 {
		return fmt.Errorf("no tar files found in %s", loadDir)
	}
	fmt.Printf("Found %d tar files to load\n", len(tarFiles))

	maxWorkers := 5
	if loadWorkers > 0 {
		maxWorkers = loadWorkers
	} else if cfg != nil && cfg.Parallel.MaxWorkers > 0 {
		maxWorkers = cfg.Parallel.MaxWorkers
	}
	fmt.Printf("Concurrent workers: %d\n", maxWorkers)
	fmt.Println()

	ctx := cmd.Context()
	startTime := time.Now()
	verify := !loadSkipVerify

	results := runWorkerPool(ctx, tarFiles, maxWorkers, func(tarFile string) error {
		return loadImage(ctx, selectedRuntime, tarFile, verify)
	})

	elapsed := time.Since(startTime)

	successCount := 0
	var failedFiles []string
	for i, r := range results {
		if r.Err != nil {
			fmt.Printf("[%d/%d] ❌ Failed to load %s: %v\n", i+1, len(results), r.Job, r.Err)
			failedFiles = append(failedFiles, r.Job)
		} else {
			fmt.Printf("[%d/%d] ✅ Successfully loaded %s\n", i+1, len(results), r.Job)
			successCount++
		}
	}

	fmt.Printf("\nSummary: %d successful, %d failed\n", successCount, len(failedFiles))
	fmt.Printf("Total time: %v\n", elapsed.Round(time.Second))

	if len(failedFiles) > 0 {
		fmt.Printf("\nFailed files:\n")
		for _, f := range failedFiles {
			fmt.Printf("  - %s\n", f)
		}
		return fmt.Errorf("failed to load %d files", len(failedFiles))
	}
	return nil
}

// loadImage loads a single tar file. When verifyChecksum is true and a .sha256
// companion file exists, it is verified before the load is attempted.
// The image name is derived from the filename and passed to the runtime so that
// runtimes without a tar-manifest parser (e.g. Skopeo) can name the image correctly.
func loadImage(ctx context.Context, rt containerruntime.ContainerRuntime, tarFile string, verifyChecksum bool) error {
	if verifyChecksum {
		checksumPath := tarFile + ".sha256"
		if _, err := os.Stat(checksumPath); err == nil {
			valid, err := containerruntime.VerifyTarChecksum(tarFile)
			if err != nil {
				return fmt.Errorf("checksum verification failed: %w", err)
			}
			if !valid {
				return fmt.Errorf("checksum mismatch: tar file does not match %s", checksumPath)
			}
		}
	}

	// Derive the expected image name from the filename (best-effort).
	// Docker/Podman/Nerdctl ignore this and read from the tar manifest;
	// Skopeo uses it as the copy destination.
	imageName := containerruntime.ImageNameFromTarFilename(filepath.Base(tarFile))

	opCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	if err := rt.Load(opCtx, tarFile, imageName); err != nil {
		return fmt.Errorf("failed to load image: %v", err)
	}
	return nil
}

// findTarFiles returns .tar files in dir (optionally recursive).
func findTarFiles(dir string, recursive bool) ([]string, error) {
	var tarFiles []string

	if recursive {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".tar") {
				tarFiles = append(tarFiles, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tar") {
				tarFiles = append(tarFiles, filepath.Join(dir, entry.Name()))
			}
		}
	}

	return tarFiles, nil
}
