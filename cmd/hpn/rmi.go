package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	containerruntime "github.com/harpoon/hpn/internal/runtime"
	"github.com/spf13/cobra"
)

var (
	rmiImageFile string
	rmiWorkers   int
)

var rmiCmd = &cobra.Command{
	Use:   "rmi",
	Short: "Remove images from local storage",
	Long:  `Remove container images from local storage. Skopeo is not supported (it does not store images locally). Pass extra flags to the underlying runtime by placing them after -- (e.g. hpn rmi -f list.txt -- -f to force remove images in use).`,
	RunE:  executeRmi,
}

func init() {
	rmiCmd.Flags().StringVarP(&rmiImageFile, "file", "f", "", "Image list file (required; use - for stdin)")
	rmiCmd.Flags().IntVar(&rmiWorkers, "workers", 0, "Number of concurrent workers (0 = use config or default 5)")
}

func selectRmiRuntime() (containerruntime.ContainerRuntime, error) {
	r, err := selectContainerRuntime()
	if err != nil {
		return nil, err
	}
	if r.Name() == "skopeo" {
		return nil, fmt.Errorf("rmi is not supported for skopeo - Skopeo does not store images locally, use docker, podman, or nerdctl")
	}
	return r, nil
}

func executeRmi(cmd *cobra.Command, args []string) error {
	if rmiImageFile == "" {
		return usageErrorf("missing required --file parameter")
	}

	fmt.Printf("Removing images from file: %s\n", rmiImageFile)

	selectedRuntime, err := selectRmiRuntime()
	if err != nil {
		return fmt.Errorf("container runtime selection failed: %w", err)
	}

	fmt.Printf("Using container runtime: %s\n", selectedRuntime.Name())

	images, err := readImageList(rmiImageFile)
	if err != nil {
		return fmt.Errorf("failed to read image list: %w", err)
	}

	fmt.Printf("Found %d images to remove\n", len(images))

	maxWorkers := 5
	if rmiWorkers > 0 {
		maxWorkers = rmiWorkers
	} else if cfg != nil && cfg.Parallel.MaxWorkers > 0 {
		maxWorkers = cfg.Parallel.MaxWorkers
	}

	extraArgs := mergeExtraArgs(getConfigExtraArgsRmi(), args)

	startTime := time.Now()
	successCount, failedImages := rmiImagesConcurrent(selectedRuntime, images, maxWorkers, extraArgs)
	elapsed := time.Since(startTime)

	fmt.Printf("\nSummary: %d successful, %d failed\n", successCount, len(failedImages))
	fmt.Printf("Total time: %v\n", elapsed.Round(time.Second))

	if len(failedImages) > 0 {
		fmt.Printf("\nFailed images:\n")
		for _, img := range failedImages {
			fmt.Printf("  - %s\n", img)
		}
		return fmt.Errorf("failed to remove %d images", len(failedImages))
	}

	return nil
}

func rmiImagesConcurrent(
	rt containerruntime.ContainerRuntime,
	images []string,
	maxWorkers int,
	extraArgs []string,
) (int, []string) {
	type rmiResult struct {
		image string
		err   error
	}

	jobs := make(chan string, len(images))
	results := make(chan rmiResult, len(images))

	var wg sync.WaitGroup
	for i := 0; i < maxWorkers && i < len(images); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for image := range jobs {
				ctx := context.Background()
				options := containerruntime.RmiOptions{
					ExtraArgs: extraArgs,
					Debug:     IsDebug(),
				}
				err := rt.RemoveImage(ctx, image, options)
				results <- rmiResult{image: image, err: err}
			}
		}()
	}

	go func() {
		for _, image := range images {
			jobs <- image
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	successCount := 0
	var failedImages []string
	completed := 0

	for result := range results {
		completed++
		if result.err != nil {
			fmt.Printf("[%d/%d] Failed to remove %s: %v\n", completed, len(images), result.image, result.err)
			failedImages = append(failedImages, result.image)
		} else {
			fmt.Printf("[%d/%d] Successfully removed %s\n", completed, len(images), result.image)
			successCount++
		}
	}

	return successCount, failedImages
}
