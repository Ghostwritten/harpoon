package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"
	containerruntime "github.com/harpoon/hpn/internal/runtime"
)

var (
	pushImageFile string
	pushRegistry  string
	pushProject   string
	pushWorkers   int
	pushDryRun   bool
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push images to registry",
	Long:  `Push container images to a registry with flexible naming options. Pass extra flags to the underlying runtime by placing them after -- (e.g. hpn push -f list.txt -r registry.local -- --tls-verify=false).`,
	RunE:  executePush,
}

func init() {
	pushCmd.Flags().StringVarP(&pushImageFile, "file", "f", "", "Image list file (required)")
	pushCmd.Flags().StringVarP(&pushRegistry, "registry", "r", "", "Target registry (required)")
	pushCmd.Flags().StringVarP(&pushProject, "project", "p", "", "Target project namespace (optional, for unified project mode)")
	pushCmd.Flags().BoolVar(&pushDryRun, "dry-run", false, "Only print source and target images, do not push")
	pushCmd.Flags().IntVar(&pushWorkers, "workers", 0, "Number of concurrent workers (0 = use config or default 5)")
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

	// Read image list from file
	images, err := readImageList(pushImageFile)
	if err != nil {
		return fmt.Errorf("failed to read image list: %v", err)
	}

	if pushDryRun {
		fmt.Printf("Dry-run: would push %d images to registry %s\n", len(images), pushRegistry)
		if pushProject != "" {
			fmt.Printf("Unified project: %s\n", pushProject)
		}
		for i, img := range images {
			target := buildTargetImage(img, pushRegistry, pushProject)
			fmt.Printf("  [%d] %s -> %s\n", i+1, img, target)
		}
		return nil
	}

	fmt.Printf("Found %d images to push\n", len(images))
	fmt.Printf("Target registry: %s\n", pushRegistry)
	if pushProject != "" {
		fmt.Printf("Target project: %s (unified project mode)\n", pushProject)
	} else {
		fmt.Printf("Mode: preserve original paths\n")
	}

	// Select container runtime
	selectedRuntime, err := selectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime selection failed: %v", err)
	}

	fmt.Printf("Using container runtime: %s\n", selectedRuntime.Name())

	// Determine max workers: flag > config > default
	maxWorkers := 5
	if pushWorkers > 0 {
		maxWorkers = pushWorkers
	} else if cfg != nil && cfg.Parallel.MaxWorkers > 0 {
		maxWorkers = cfg.Parallel.MaxWorkers
	}
	fmt.Printf("Concurrent workers: %d\n", maxWorkers)
	fmt.Println()

	// Merge extra args: config first, then CLI (args after --)
	extraArgs := mergeExtraArgs(getConfigExtraArgsPush(), args)

	// Start timing
	startTime := time.Now()

	// Push images concurrently
	successCount, failedImages := pushImagesConcurrent(
		selectedRuntime,
		images,
		pushRegistry,
		pushProject,
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
		return fmt.Errorf("failed to push %d images", len(failedImages))
	}

	return nil
}

// pushImage pushes a single image to registry
func pushImage(containerRuntime containerruntime.ContainerRuntime, sourceImage, targetImage string, extraArgs []string) error {
	fmt.Printf("  Tag: %s -> %s\n", sourceImage, targetImage)

	// Tag the image
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := containerRuntime.Tag(ctx, sourceImage, targetImage); err != nil {
		return fmt.Errorf("failed to tag image: %v", err)
	}

	// Push the image
	pushOptions := containerruntime.PushOptions{
		Timeout:   10 * time.Minute,
		Debug:     IsDebug(),
		ExtraArgs: extraArgs,
	}

	if err := containerRuntime.Push(ctx, targetImage, pushOptions); err != nil {
		return fmt.Errorf("failed to push image: %v", err)
	}

	return nil
}

// pushImagesConcurrent pushes images concurrently with worker pool
func pushImagesConcurrent(
	runtime containerruntime.ContainerRuntime,
	images []string,
	registry string,
	project string,
	maxWorkers int,
	extraArgs []string,
) (int, []string) {
	type pushJob struct {
		sourceImage string
		targetImage string
	}

	type pushResult struct {
		sourceImage string
		targetImage string
		err         error
	}

	// Build job list
	jobs := make(chan pushJob, len(images))
	results := make(chan pushResult, len(images))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers && i < len(images); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				err := pushImage(runtime, job.sourceImage, job.targetImage, extraArgs)
				results <- pushResult{
					sourceImage: job.sourceImage,
					targetImage: job.targetImage,
					err:         err,
				}
			}
		}()
	}

	// Send jobs
	go func() {
		for _, image := range images {
			targetImage := buildTargetImage(image, registry, project)
			jobs <- pushJob{
				sourceImage: image,
				targetImage: targetImage,
			}
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
			fmt.Printf("[%d/%d] ❌ Failed to push %s: %v\n", completed, len(images), result.sourceImage, result.err)
			failedImages = append(failedImages, result.sourceImage)
		} else {
			fmt.Printf("[%d/%d] ✅ Successfully pushed %s -> %s\n", completed, len(images), result.sourceImage, result.targetImage)
			successCount++
		}
	}

	return successCount, failedImages
}
