package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	containerruntime "github.com/harpoon/hpn/internal/runtime"
)

var (
	pushImageFile string
	pushRegistry  string
	pushProject   string
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push images to registry",
	Long:  `Push container images to a registry with flexible naming options.`,
	RunE:  executePush,
}

func init() {
	pushCmd.Flags().StringVarP(&pushImageFile, "file", "f", "", "Image list file (required)")
	pushCmd.Flags().StringVarP(&pushRegistry, "registry", "r", "", "Target registry (required)")
	pushCmd.Flags().StringVarP(&pushProject, "project", "p", "", "Target project namespace (optional, for unified project mode)")
	rootCmd.AddCommand(pushCmd)
}

func executePush(cmd *cobra.Command, args []string) error {
	if pushImageFile == "" {
		return fmt.Errorf("missing required --file parameter")
	}
	if pushRegistry == "" {
		return fmt.Errorf("missing required --registry parameter")
	}

	fmt.Printf("Executing push action with file: %s, registry: %s\n", pushImageFile, pushRegistry)

	// Select container runtime
	selectedRuntime, err := selectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime selection failed: %v", err)
	}

	fmt.Printf("Using container runtime: %s\n", selectedRuntime.Name())

	// Read image list from file
	images, err := readImageList(pushImageFile)
	if err != nil {
		return fmt.Errorf("failed to read image list: %v", err)
	}

	fmt.Printf("Found %d images to push\n", len(images))
	fmt.Printf("Target registry: %s\n", pushRegistry)
	if pushProject != "" {
		fmt.Printf("Target project: %s (unified project mode)\n", pushProject)
	} else {
		fmt.Printf("Mode: preserve original paths\n")
	}

	// Push each image
	successCount := 0
	failedImages := []string{}

	for i, image := range images {
		fmt.Printf("[%d/%d] Pushing %s...\n", i+1, len(images), image)

		// Build target image name using the new logic
		targetImage := buildTargetImage(image, pushRegistry, pushProject)

		if err := pushImage(selectedRuntime, image, targetImage); err != nil {
			fmt.Printf("❌ Failed to push %s: %v\n", image, err)
			failedImages = append(failedImages, image)
		} else {
			fmt.Printf("✅ Successfully pushed %s -> %s\n", image, targetImage)
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
		return fmt.Errorf("failed to push %d images", len(failedImages))
	}

	return nil
}

// pushImage pushes a single image to registry
func pushImage(containerRuntime containerruntime.ContainerRuntime, sourceImage, targetImage string) error {
	fmt.Printf("  Tag: %s -> %s\n", sourceImage, targetImage)

	// Tag the image
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := containerRuntime.Tag(ctx, sourceImage, targetImage); err != nil {
		return fmt.Errorf("failed to tag image: %v", err)
	}

	// Push the image
	pushOptions := containerruntime.PushOptions{
		Timeout: 10 * time.Minute,
		Debug:   IsDebug(),
	}

	if err := containerRuntime.Push(ctx, targetImage, pushOptions); err != nil {
		return fmt.Errorf("failed to push image: %v", err)
	}

	return nil
}
