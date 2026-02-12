package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	containerruntime "github.com/harpoon/hpn/internal/runtime"
)

var (
	loadPath      string
	loadRecursive bool
	loadWorkers   int
)

var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Load images from tar files",
	Long:  `Load container images from tar files. Use --path (-p) to specify the directory.`,
	RunE:  executeLoad,
}

func init() {
	loadCmd.Flags().StringVarP(&loadPath, "path", "p", "./images", "Input directory containing tar files (default: ./images)")
	loadCmd.Flags().BoolVar(&loadRecursive, "recursive", false, "Load recursively from subdirectories")
	loadCmd.Flags().IntVar(&loadWorkers, "workers", 0, "Number of concurrent workers (0 = use config or default 5)")
	rootCmd.AddCommand(loadCmd)
}

func executeLoad(cmd *cobra.Command, args []string) error {
	// Use path from flag or default
	loadDir := loadPath
	if loadDir == "" {
		loadDir = "./images"
	}

	fmt.Printf("Executing load action from: %s\n", loadDir)

	// Select container runtime
	selectedRuntime, err := selectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime selection failed: %v", err)
	}

	fmt.Printf("Using container runtime: %s\n", selectedRuntime.Name())

	// Find tar files
	var tarFiles []string
	if loadRecursive {
		files, err := findTarFiles(loadDir, true)
		if err != nil {
			return fmt.Errorf("failed to find tar files recursively: %v", err)
		}
		tarFiles = files
	} else {
		files, err := findTarFiles(loadDir, false)
		if err != nil {
			return fmt.Errorf("failed to find tar files in directory: %v", err)
		}
		tarFiles = files
	}

	if len(tarFiles) == 0 {
		return fmt.Errorf("no tar files found in %s", loadDir)
	}

	fmt.Printf("Found %d tar files to load\n", len(tarFiles))

	// Determine max workers: flag > config > default
	maxWorkers := 5
	if loadWorkers > 0 {
		maxWorkers = loadWorkers
	} else if cfg != nil && cfg.Parallel.MaxWorkers > 0 {
		maxWorkers = cfg.Parallel.MaxWorkers
	}
	fmt.Printf("Concurrent workers: %d\n", maxWorkers)
	fmt.Println()

	// Start timing
	startTime := time.Now()

	// Load tar files concurrently
	successCount, failedFiles := loadImagesConcurrent(
		selectedRuntime,
		tarFiles,
		maxWorkers,
	)

	// Calculate elapsed time
	elapsed := time.Since(startTime)

	// Print summary
	fmt.Printf("\nSummary: %d successful, %d failed\n", successCount, len(failedFiles))
	fmt.Printf("Total time: %v\n", elapsed.Round(time.Second))

	if len(failedFiles) > 0 {
		fmt.Printf("\nFailed files:\n")
		for _, file := range failedFiles {
			fmt.Printf("  - %s\n", file)
		}
		return fmt.Errorf("failed to load %d files", len(failedFiles))
	}

	return nil
}

// loadImage loads a single image from tar file
func loadImage(containerRuntime containerruntime.ContainerRuntime, tarFile string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := containerRuntime.Load(ctx, tarFile); err != nil {
		return fmt.Errorf("failed to load image: %v", err)
	}

	return nil
}

// findTarFiles finds tar files in the specified directory
func findTarFiles(dir string, recursive bool) ([]string, error) {
	var tarFiles []string

	if recursive {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".tar") && !strings.HasSuffix(path, ".sha256") {
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
			if !entry.IsDir() {
				name := entry.Name()
				if strings.HasSuffix(name, ".tar") && !strings.HasSuffix(name, ".sha256") {
					tarFiles = append(tarFiles, filepath.Join(dir, name))
				}
			}
		}
	}

	return tarFiles, nil
}

// loadImagesConcurrent loads images concurrently with worker pool
func loadImagesConcurrent(
	runtime containerruntime.ContainerRuntime,
	tarFiles []string,
	maxWorkers int,
) (int, []string) {
	type loadResult struct {
		tarFile string
		err     error
	}

	// Create worker pool
	jobs := make(chan string, len(tarFiles))
	results := make(chan loadResult, len(tarFiles))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers && i < len(tarFiles); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for tarFile := range jobs {
				err := loadImage(runtime, tarFile)
				results <- loadResult{tarFile: tarFile, err: err}
			}
		}()
	}

	// Send jobs
	go func() {
		for _, tarFile := range tarFiles {
			jobs <- tarFile
		}
		close(jobs)
	}()

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	successCount := 0
	failedFiles := []string{}
	completed := 0

	for result := range results {
		completed++
		if result.err != nil {
			fmt.Printf("[%d/%d] ❌ Failed to load %s: %v\n", completed, len(tarFiles), result.tarFile, result.err)
			failedFiles = append(failedFiles, result.tarFile)
		} else {
			fmt.Printf("[%d/%d] ✅ Successfully loaded %s\n", completed, len(tarFiles), result.tarFile)
			successCount++
		}
	}

	return successCount, failedFiles
}
