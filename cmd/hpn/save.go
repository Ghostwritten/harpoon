package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	containerruntime "github.com/harpoon/hpn/internal/runtime"
)

var (
	saveImageFile string
	savePath      string
	saveWorkers   int
	saveDryRun    bool
)

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save images to tar files",
	Long: `Save container images to tar files with checksum verification.
Pass extra flags to the underlying runtime by placing them after --
(e.g. hpn save -f list.txt -- --src-tls-verify=false).`,
	RunE: executeSave,
}

func init() {
	saveCmd.Flags().StringVarP(&saveImageFile, "file", "f", "", "Image list file (required; use - for stdin)")
	// --path is the canonical flag; --output is a deprecated alias kept for backwards compatibility.
	saveCmd.Flags().StringVarP(&savePath, "path", "p", "./images", "Output directory path")
	saveCmd.Flags().StringVar(&savePath, "output", "./images", "Output directory path (deprecated alias for --path)")
	if err := saveCmd.Flags().MarkDeprecated("output", "use --path instead"); err != nil {
		// MarkDeprecated only fails if the flag doesn't exist; safe to ignore here.
		_ = err
	}
	saveCmd.Flags().BoolVar(&saveDryRun, "dry-run", false, "Only print images and output path, do not save")
	saveCmd.Flags().IntVar(&saveWorkers, "workers", 0, "Number of concurrent workers (0 = use config or default 5)")
}

func executeSave(cmd *cobra.Command, args []string) error {
	if saveImageFile == "" {
		return usageErrorf("missing required --file parameter")
	}

	fmt.Printf("Executing save action with file: %s\n", saveImageFile)

	images, err := readImageList(saveImageFile)
	if err != nil {
		return fmt.Errorf("failed to read image list: %v", err)
	}

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

	selectedRuntime, err := selectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime selection failed: %v", err)
	}
	fmt.Printf("Using container runtime: %s\n", selectedRuntime.Name())

	// 0700: only the owner can access the directory; image archives may be sensitive.
	if err := os.MkdirAll(saveDir, 0700); err != nil {
		return fmt.Errorf("failed to create save directory %s: %v", saveDir, err)
	}
	fmt.Printf("Saving to: %s\n", saveDir)

	platformToUse := platform

	maxWorkers := 5
	if saveWorkers > 0 {
		maxWorkers = saveWorkers
	} else if cfg != nil && cfg.Parallel.MaxWorkers > 0 {
		maxWorkers = cfg.Parallel.MaxWorkers
	}
	fmt.Printf("Concurrent workers: %d\n", maxWorkers)
	fmt.Println()

	extraArgs := mergeExtraArgs(getConfigExtraArgsSave(), args)
	ctx := cmd.Context()
	startTime := time.Now()

	results := runWorkerPool(ctx, images, maxWorkers, func(image string) error {
		return saveImageToPath(ctx, selectedRuntime, image, saveDir, platformToUse, extraArgs)
	})

	elapsed := time.Since(startTime)

	successCount := 0
	var failedImages []string
	for i, r := range results {
		if r.Err != nil {
			fmt.Printf("[%d/%d] ❌ Failed to save %s: %v\n", i+1, len(results), r.Job, r.Err)
			failedImages = append(failedImages, r.Job)
		} else {
			fmt.Printf("[%d/%d] ✅ Successfully saved %s\n", i+1, len(results), r.Job)
			successCount++
		}
	}

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

// saveImageToPath saves a single image to a tar file with resume support.
// If a tar and valid checksum file already exist, the image is skipped.
func saveImageToPath(ctx context.Context, rt containerruntime.ContainerRuntime, image, baseDir, platformToUse string, extraArgs []string) error {
	tarFilename := containerruntime.TarFilenameFromImage(image)
	tarPath := filepath.Join(baseDir, tarFilename)
	checksumPath := tarPath + ".sha256"

	// Resume check: skip if tar + valid checksum already present
	if _, err := os.Stat(tarPath); err == nil {
		if _, err := os.Stat(checksumPath); err == nil {
			if valid, err := containerruntime.VerifyTarChecksum(tarPath); err == nil && valid {
				fmt.Printf("  ⏭️  Skipped %s (already exists and verified)\n", image)
				return nil
			}
		}
		// Corrupt or missing checksum — re-save
		fmt.Printf("  ⚠️  Re-saving %s (checksum missing or invalid)\n", tarPath)
		os.Remove(tarPath)
		os.Remove(checksumPath)
	}

	multiArch := platformToUse == "all"
	if rt.Name() == "skopeo" {
		fmt.Printf("  📥 Pulling and saving %s...\n", image)
	} else {
		fmt.Printf("  💾 Saving %s...\n", image)
	}

	opCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	if err := rt.Save(opCtx, image, tarPath, containerruntime.SaveOptions{
		Checksum:  true,
		MultiArch: multiArch,
		Platform:  platformToUse,
		Debug:     IsDebug(),
		ExtraArgs: extraArgs,
	}); err != nil {
		return fmt.Errorf("failed to save image: %v", err)
	}

	if _, err := os.Stat(tarPath); err != nil {
		return fmt.Errorf("tar file was not created: %v", err)
	}

	if rt.Name() == "skopeo" {
		fmt.Printf("  ✓ Pulled and saved %s to %s\n", image, tarPath)
	} else {
		fmt.Printf("  ✓ Saved %s to %s\n", image, tarPath)
	}
	return nil
}
