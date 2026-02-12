package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/cobra"
	containerruntime "github.com/harpoon/hpn/internal/runtime"
)

var (
	saveImageFile string
	savePath      string
	saveWorkers   int
	saveDryRun   bool
)

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save images to tar files",
	Long:  `Save container images to tar files with checksum verification. Pass extra flags to the underlying runtime by placing them after -- (e.g. hpn save -f list.txt -- --src-tls-verify=false).`,
	RunE:  executeSave,
}

func init() {
	saveCmd.Flags().StringVarP(&saveImageFile, "file", "f", "", "Image list file (required)")
	saveCmd.Flags().StringVar(&savePath, "path", "./images", "Output directory path (default: ./images, supports multi-level paths)")
	saveCmd.Flags().BoolVar(&saveDryRun, "dry-run", false, "Only print images and output path, do not save")
	saveCmd.Flags().StringVar(&savePath, "output", "./images", "Output directory path (alias for --path)")
	saveCmd.Flags().IntVar(&saveWorkers, "workers", 0, "Number of concurrent workers (0 = use config or default 5)")
	rootCmd.AddCommand(saveCmd)
}

func executeSave(cmd *cobra.Command, args []string) error {
	if saveImageFile == "" {
		return fmt.Errorf("missing required --file parameter")
	}

	fmt.Printf("Executing save action with file: %s\n", saveImageFile)

	// Read image list from file
	images, err := readImageList(saveImageFile)
	if err != nil {
		return fmt.Errorf("failed to read image list: %v", err)
	}

	// Use path from flag or default
	saveDir := savePath
	if saveDir == "" {
		saveDir = "./images"
	}

	if saveDryRun {
		fmt.Printf("Dry-run: would save %d images to %s\n", len(images), saveDir)
		for i, img := range images {
			tarName := containerruntime.TarFilenameFromImage(img)
			fmt.Printf("  [%d] %s -> %s\n", i+1, img, filepath.Join(saveDir, tarName))
		}
		return nil
	}

	fmt.Printf("Found %d images to save\n", len(images))

	// Select container runtime
	selectedRuntime, err := selectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime selection failed: %v", err)
	}

	fmt.Printf("Using container runtime: %s\n", selectedRuntime.Name())

	// Create directory if it doesn't exist (supports multi-level paths)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return fmt.Errorf("failed to create save directory %s: %v", saveDir, err)
	}

	fmt.Printf("Saving to: %s\n", saveDir)

	// Use platform from global flag
	platformToUse := platform

	// Determine max workers: flag > config > default
	maxWorkers := 5
	if saveWorkers > 0 {
		maxWorkers = saveWorkers
	} else if cfg != nil && cfg.Parallel.MaxWorkers > 0 {
		maxWorkers = cfg.Parallel.MaxWorkers
	}
	fmt.Printf("Concurrent workers: %d\n", maxWorkers)
	fmt.Println()

	// Merge extra args: config first, then CLI (args after --)
	extraArgs := mergeExtraArgs(getConfigExtraArgsSave(), args)

	// Start timing
	startTime := time.Now()

	// Save images concurrently
	successCount, failedImages := saveImagesConcurrent(
		selectedRuntime,
		images,
		saveDir,
		platformToUse,
		maxWorkers,
		extraArgs,
	)

	// Calculate elapsed time
	elapsed := time.Since(startTime)

	// Print summary
	fmt.Printf("\nSummary: %d successful, %d failed\n", successCount, len(failedImages))
	fmt.Printf("Total time: %v\n", elapsed.Round(time.Second))

	if len(failedImages) > 0 {
		fmt.Printf("\nFailed images:\n")
		for _, img := range failedImages {
			fmt.Printf("  - %s\n", img)
		}
		return fmt.Errorf("failed to save %d images", len(failedImages))
	}

	return nil
}

// saveImageToPath saves a single image to tar file with resume support
func saveImageToPath(containerRuntime containerruntime.ContainerRuntime, image, baseDir, platformToUse string, extraArgs []string) error {
	// Parse image name to generate tar filename
	tarFilename := containerruntime.TarFilenameFromImage(image)
	tarPath := filepath.Join(baseDir, tarFilename)

	// Check for resume: verify if file exists and checksum is valid
	checksumPath := tarPath + ".sha256"
	if _, err := os.Stat(tarPath); err == nil {
		// Tar file exists, check if checksum file exists and is valid
		if _, err := os.Stat(checksumPath); err == nil {
			// Both files exist, verify checksum
			if valid, err := containerruntime.VerifyTarChecksum(tarPath); err == nil && valid {
				fmt.Printf("  ⏭️  Skipped %s (already exists and verified)\n", image)
				return nil
			} else if err != nil {
				fmt.Printf("  ⚠️  Warning: checksum verification failed for %s, will re-save\n", tarPath)
				// Delete invalid tar file and checksum file
				os.Remove(tarPath)
				os.Remove(checksumPath)
			} else {
				fmt.Printf("  ⚠️  Warning: checksum mismatch for %s, will re-save\n", tarPath)
				// Delete invalid tar file and checksum file
				os.Remove(tarPath)
				os.Remove(checksumPath)
			}
		} else {
			// Tar exists but no checksum, will re-save to generate checksum
			fmt.Printf("  ⚠️  Warning: %s exists but no checksum file, will re-save\n", tarPath)
			// Delete existing tar file (skopeo cannot modify existing docker-archive)
			os.Remove(tarPath)
		}
	}

	// Determine if multi-arch is needed
	multiArch := platformToUse == "all"

	// Show different message for Skopeo (which pulls and saves)
	runtimeName := containerRuntime.Name()
	if runtimeName == "skopeo" {
		fmt.Printf("  📥 Pulling and saving %s...\n", image)
	} else {
		fmt.Printf("  💾 Saving %s...\n", image)
	}

	// Prepare save options
	saveOptions := containerruntime.SaveOptions{
		Checksum:  true, // Always generate checksum for resume support
		MultiArch: multiArch,
		Platform:  platformToUse, // Pass platform to runtime (for Skopeo to handle platform selection)
		Debug:     IsDebug(),
		ExtraArgs: extraArgs,
	}

	// Execute save command using runtime interface
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := containerRuntime.Save(ctx, image, tarPath, saveOptions); err != nil {
		return fmt.Errorf("failed to save image: %v", err)
	}

	// Check if file was created successfully
	if _, err := os.Stat(tarPath); err != nil {
		return fmt.Errorf("tar file was not created: %v", err)
	}

	// Show completion message with different text for Skopeo
	if runtimeName == "skopeo" {
		fmt.Printf("  ✓ Pulled and saved %s to %s\n", image, tarPath)
	} else {
		fmt.Printf("  ✓ Saved %s to %s\n", image, tarPath)
	}
	return nil
}

// saveImagesConcurrent saves images concurrently with worker pool
func saveImagesConcurrent(
	runtime containerruntime.ContainerRuntime,
	images []string,
	saveDir string,
	platformToUse string,
	maxWorkers int,
	extraArgs []string,
) (int, []string) {
	type saveResult struct {
		image string
		err   error
	}

	// Create worker pool
	jobs := make(chan string, len(images))
	results := make(chan saveResult, len(images))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers && i < len(images); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for image := range jobs {
				err := saveImageToPath(runtime, image, saveDir, platformToUse, extraArgs)
				results <- saveResult{image: image, err: err}
			}
		}()
	}

	// Send jobs
	go func() {
		for _, image := range images {
			jobs <- image
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
	failedImages := []string{}
	completed := 0

	for result := range results {
		completed++
		if result.err != nil {
			fmt.Printf("[%d/%d] ❌ Failed to save %s: %v\n", completed, len(images), result.image, result.err)
			failedImages = append(failedImages, result.image)
		} else {
			fmt.Printf("[%d/%d] ✅ Successfully saved %s\n", completed, len(images), result.image)
			successCount++
		}
	}

	return successCount, failedImages
}
