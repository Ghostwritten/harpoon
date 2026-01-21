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
)

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save images to tar files",
	Long:  `Save container images to tar files with checksum verification.`,
	RunE:  executeSave,
}

func init() {
	saveCmd.Flags().StringVarP(&saveImageFile, "file", "f", "", "Image list file (required)")
	saveCmd.Flags().StringVar(&savePath, "path", "./images", "Output directory path (default: ./images, supports multi-level paths)")
	saveCmd.Flags().StringVar(&savePath, "output", "./images", "Output directory path (alias for --path)")
	rootCmd.AddCommand(saveCmd)
}

func executeSave(cmd *cobra.Command, args []string) error {
	if saveImageFile == "" {
		return fmt.Errorf("missing required --file parameter")
	}

	fmt.Printf("Executing save action with file: %s\n", saveImageFile)

	// Select container runtime
	selectedRuntime, err := selectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime selection failed: %v", err)
	}

	fmt.Printf("Using container runtime: %s\n", selectedRuntime.Name())

	// Read image list from file
	images, err := readImageList(saveImageFile)
	if err != nil {
		return fmt.Errorf("failed to read image list: %v", err)
	}

	fmt.Printf("Found %d images to save\n", len(images))

	// Use path from flag or default
	saveDir := savePath
	if saveDir == "" {
		saveDir = "./images"
	}

	// Create directory if it doesn't exist (supports multi-level paths)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return fmt.Errorf("failed to create save directory %s: %v", saveDir, err)
	}

	fmt.Printf("Saving to: %s\n", saveDir)

	// Use platform from global flag
	platformToUse := platform

	// Save each image
	successCount := 0
	failedImages := []string{}

	for i, image := range images {
		fmt.Printf("[%d/%d] Saving %s...\n", i+1, len(images), image)

		if err := saveImageToPath(selectedRuntime, image, saveDir, platformToUse); err != nil {
			fmt.Printf("❌ Failed to save %s: %v\n", image, err)
			failedImages = append(failedImages, image)
		} else {
			fmt.Printf("✅ Successfully saved %s\n", image)
			successCount++
		}
	}

	// Print summary
	fmt.Printf("\nSummary: %d successful, %d failed\n", successCount, len(failedImages))

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
func saveImageToPath(containerRuntime containerruntime.ContainerRuntime, image, baseDir, platformToUse string) error {
	// Parse image name to generate tar filename
	tarFilename := generateTarFilename(image)
	tarPath := filepath.Join(baseDir, tarFilename)

	// Check for resume: verify if file exists and checksum is valid
	checksumPath := tarPath + ".sha256"
	if _, err := os.Stat(tarPath); err == nil {
		// Tar file exists, check if checksum file exists and is valid
		if _, err := os.Stat(checksumPath); err == nil {
			// Both files exist, verify checksum
			if valid, err := verifyImageChecksum(tarPath); err == nil && valid {
				fmt.Printf("  ⏭️  Skipped %s (already exists and verified)\n", image)
				return nil
			} else if err != nil {
				fmt.Printf("  ⚠️  Warning: checksum verification failed for %s, will re-save\n", tarPath)
			} else {
				fmt.Printf("  ⚠️  Warning: checksum mismatch for %s, will re-save\n", tarPath)
			}
		} else {
			// Tar exists but no checksum, will re-save to generate checksum
			fmt.Printf("  ⚠️  Warning: %s exists but no checksum file, will re-save\n", tarPath)
		}
	}

	// Determine if multi-arch is needed
	multiArch := platformToUse == "all"

	// Prepare save options
	saveOptions := containerruntime.SaveOptions{
		Checksum:  true, // Always generate checksum for resume support
		MultiArch: multiArch,
		Platform:  platformToUse, // Pass platform to runtime (for Skopeo to handle platform selection)
		Debug:     IsDebug(),
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

	fmt.Printf("  ✓ Saved %s to %s\n", image, tarPath)
	return nil
}
