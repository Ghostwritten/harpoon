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
	pushWorkers   int
	pushDryRun    bool
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push images to registry",
	Long: `Push container images to a registry with flexible naming options.
Pass extra flags to the underlying runtime by placing them after --
(e.g. hpn push -f list.txt -r registry.local -- --tls-verify=false).`,
	RunE: executePush,
}

func init() {
	pushCmd.Flags().StringVarP(&pushImageFile, "file", "f", "", "Image list file (required; use - for stdin)")
	pushCmd.Flags().StringVarP(&pushRegistry, "registry", "r", "", "Target registry (required)")
	pushCmd.Flags().StringVarP(&pushProject, "project", "p", "", "Target project namespace (optional, for unified project mode)")
	pushCmd.Flags().BoolVar(&pushDryRun, "dry-run", false, "Only print source and target images, do not push")
	pushCmd.Flags().IntVar(&pushWorkers, "workers", 0, "Number of concurrent workers (0 = use config or default 5)")
}

// pushJob pairs a source image reference with its computed target reference.
type pushJob struct {
	source string
	target string
}

func executePush(cmd *cobra.Command, args []string) error {
	if pushImageFile == "" {
		return usageErrorf("missing required --file parameter")
	}
	if pushRegistry == "" {
		return usageErrorf("missing required --registry parameter")
	}

	fmt.Printf("Executing push action with file: %s, registry: %s\n", pushImageFile, pushRegistry)

	images, err := readImageList(pushImageFile)
	if err != nil {
		return fmt.Errorf("failed to read image list: %v", err)
	}

	// Build job list (source → target mapping) upfront
	jobs := make([]pushJob, len(images))
	for i, img := range images {
		jobs[i] = pushJob{source: img, target: buildTargetImage(img, pushRegistry, pushProject)}
	}

	if pushDryRun {
		fmt.Printf("Dry-run: would push %d images to registry %s\n", len(images), pushRegistry)
		if pushProject != "" {
			fmt.Printf("Unified project: %s\n", pushProject)
		}
		for i, j := range jobs {
			fmt.Printf("  [%d] %s -> %s\n", i+1, j.source, j.target)
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

	selectedRuntime, err := selectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime selection failed: %v", err)
	}
	fmt.Printf("Using container runtime: %s\n", selectedRuntime.Name())

	timeout := 5 * time.Minute
	if cfg != nil && cfg.Runtime.Timeout > 0 {
		timeout = cfg.Runtime.Timeout
	}
	retryCfg := containerruntime.RetryConfig{MaxAttempts: 3, Delay: time.Second, MaxDelay: 30 * time.Second}
	if cfg != nil {
		retryCfg = cfg.Runtime.Retry.ToRuntimeRetryConfig()
	}
	maxWorkers := 5
	if pushWorkers > 0 {
		maxWorkers = pushWorkers
	} else if cfg != nil && cfg.Parallel.MaxWorkers > 0 {
		maxWorkers = cfg.Parallel.MaxWorkers
	}

	fmt.Printf("Timeout: %v, Max retries: %d, Concurrent workers: %d\n",
		timeout, retryCfg.MaxAttempts, maxWorkers)
	fmt.Println()

	extraArgs := mergeExtraArgs(getConfigExtraArgsPush(), args)
	ctx := cmd.Context()
	startTime := time.Now()

	results := runWorkerPool(ctx, jobs, maxWorkers, func(job pushJob) error {
		return retryWithBackoff(ctx, func() error {
			opCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			fmt.Printf("  Tag: %s -> %s\n", job.source, job.target)
			if err := selectedRuntime.Tag(opCtx, job.source, job.target); err != nil {
				return fmt.Errorf("failed to tag image: %v", err)
			}
			return selectedRuntime.Push(opCtx, job.target, containerruntime.PushOptions{
				Timeout:   timeout,
				Debug:     IsDebug(),
				ExtraArgs: extraArgs,
			})
		}, retryCfg)
	})

	elapsed := time.Since(startTime)

	successCount := 0
	var failedImages []string
	for i, r := range results {
		if r.Err != nil {
			fmt.Printf("[%d/%d] ❌ Failed to push %s: %v\n", i+1, len(results), r.Job.source, r.Err)
			failedImages = append(failedImages, r.Job.source)
		} else {
			fmt.Printf("[%d/%d] ✅ Successfully pushed %s -> %s\n", i+1, len(results), r.Job.source, r.Job.target)
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
		return fmt.Errorf("failed to push %d images", len(failedImages))
	}
	return nil
}
