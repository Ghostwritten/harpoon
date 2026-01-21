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
	loadPath      string
	loadRecursive bool
)

var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Load images from tar files",
	Long:  `Load container images from tar files.`,
	RunE:  executeLoad,
}

func init() {
	loadCmd.Flags().StringVar(&loadPath, "path", "./images", "Input directory path (default: ./images)")
	loadCmd.Flags().StringVar(&loadPath, "input", "./images", "Input directory path (alias for --path)")
	loadCmd.Flags().BoolVar(&loadRecursive, "recursive", false, "Load recursively from subdirectories")
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

	// Load each tar file
	successCount := 0
	failedFiles := []string{}

	for i, tarFile := range tarFiles {
		fmt.Printf("[%d/%d] Loading %s...\n", i+1, len(tarFiles), tarFile)

		if err := loadImage(selectedRuntime, tarFile); err != nil {
			fmt.Printf("❌ Failed to load %s: %v\n", tarFile, err)
			failedFiles = append(failedFiles, tarFile)
		} else {
			fmt.Printf("✅ Successfully loaded %s\n", tarFile)
			successCount++
		}
	}

	// Print summary
	fmt.Printf("\nSummary: %d successful, %d failed\n", successCount, len(failedFiles))

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
