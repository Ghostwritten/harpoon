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
	pullImageFile string
	pullPlatform  string
	pullWorkers   int
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull images from registry",
	Long:  `Pull container images from a registry to local. Pass extra flags to the underlying runtime (docker/podman/nerdctl/skopeo) by placing them after -- (e.g. hpn pull -f list.txt -- --tls-verify=false).`,
	RunE:  executePull,
}

func init() {
	pullCmd.Flags().StringVarP(&pullImageFile, "file", "f", "", "Image list file (required)")
	pullCmd.Flags().StringVar(&pullPlatform, "platform", "", "Target platform (e.g., linux/amd64, linux/arm64, all for multi-arch)")
	pullCmd.Flags().IntVar(&pullWorkers, "workers", 0, "Number of concurrent workers (0 = use config or default 5)")
	rootCmd.AddCommand(pullCmd)
}

func executePull(cmd *cobra.Command, args []string) error {
	if pullImageFile == "" {
		return fmt.Errorf("missing required --file parameter")
	}

	fmt.Printf("Executing pull action with file: %s\n", pullImageFile)

	// Select container runtime
	selectedRuntime, err := selectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime selection failed: %v", err)
	}

	fmt.Printf("Using container runtime: %s\n", selectedRuntime.Name())

	// Read image list from file
	images, err := readImageList(pullImageFile)
	if err != nil {
		return fmt.Errorf("failed to read image list: %v", err)
	}

	fmt.Printf("Found %d images to pull\n", len(images))

	// Start timing
	startTime := time.Now()

	// Get configuration values with defaults
	timeout := 5 * time.Minute
	if cfg != nil && cfg.Runtime.Timeout > 0 {
		timeout = cfg.Runtime.Timeout
	}

	retryConfig := containerruntime.RetryConfig{
		MaxAttempts: 3,
		Delay:       time.Second,
		MaxDelay:    30 * time.Second,
	}
	if cfg != nil {
		retryConfig = cfg.Runtime.Retry.ToRuntimeRetryConfig()
	}

	// Determine max workers: flag > config > default
	maxWorkers := 5
	if pullWorkers > 0 {
		maxWorkers = pullWorkers
	} else if cfg != nil && cfg.Parallel.MaxWorkers > 0 {
		maxWorkers = cfg.Parallel.MaxWorkers
	}

	// Use platform from flag or global flag
	platformToUse := pullPlatform
	if platformToUse == "" {
		platformToUse = platform // Fallback to global platform flag
	}

	// Show configuration
	if platformToUse != "" {
		fmt.Printf("Target platform: %s\n", platformToUse)
	}
	fmt.Printf("Timeout: %v, Max retries: %d, Concurrent workers: %d\n", timeout, retryConfig.MaxAttempts, maxWorkers)
	fmt.Println()

	// Prepare proxy config
	var proxyConfig *containerruntime.ProxyConfig
	if cfg != nil && cfg.Proxy.Enabled {
		proxyConfig = cfg.Proxy.ToRuntimeProxyConfig()
	}

	// Merge extra args: config first, then CLI (args after --)
	extraArgs := mergeExtraArgs(getConfigExtraArgsPull(), args)

	// Pull images with concurrency and retry
	successCount, failedImages := pullImagesConcurrent(
		selectedRuntime,
		images,
		platformToUse,
		timeout,
		retryConfig,
		proxyConfig,
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
		return fmt.Errorf("failed to pull %d images", len(failedImages))
	}

	return nil
}

// pullImagesConcurrent pulls images concurrently with retry mechanism
func pullImagesConcurrent(
	runtime containerruntime.ContainerRuntime,
	images []string,
	platform string,
	timeout time.Duration,
	retryConfig containerruntime.RetryConfig,
	proxyConfig *containerruntime.ProxyConfig,
	maxWorkers int,
	extraArgs []string,
) (int, []string) {
	type pullResult struct {
		image string
		err   error
	}

	// Create worker pool
	jobs := make(chan string, len(images))
	results := make(chan pullResult, len(images))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers && i < len(images); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for image := range jobs {
				err := pullImageWithRetry(runtime, image, platform, timeout, retryConfig, proxyConfig, extraArgs)
				results <- pullResult{image: image, err: err}
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
			fmt.Printf("[%d/%d] ❌ Failed to pull %s: %v\n", completed, len(images), result.image, result.err)
			failedImages = append(failedImages, result.image)
		} else {
			fmt.Printf("[%d/%d] ✅ Successfully pulled %s\n", completed, len(images), result.image)
			successCount++
		}
	}

	return successCount, failedImages
}

// pullImageWithRetry pulls a single image with retry mechanism
func pullImageWithRetry(
	runtime containerruntime.ContainerRuntime,
	image string,
	platform string,
	timeout time.Duration,
	retryConfig containerruntime.RetryConfig,
	proxyConfig *containerruntime.ProxyConfig,
	extraArgs []string,
) error {
	var lastErr error
	delay := retryConfig.Delay

	for attempt := 1; attempt <= retryConfig.MaxAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		// Set MultiArch if platform is "all"
		multiArch := platform == "all"

		pullOptions := containerruntime.PullOptions{
			Timeout:   timeout,
			Retry:     retryConfig,
			Platform:  platform,
			MultiArch: multiArch,
			Proxy:     proxyConfig,
			Debug:     IsDebug(),
			ExtraArgs: extraArgs,
		}

		err := runtime.Pull(ctx, image, pullOptions)
		cancel()

		if err == nil {
			if attempt > 1 {
				fmt.Printf("  ✓ Pulled %s after %d attempts\n", image, attempt)
			}
			return nil
		}

		lastErr = err

		// Don't retry on last attempt
		if attempt < retryConfig.MaxAttempts {
			// Exponential backoff with max delay
			if delay > retryConfig.MaxDelay {
				delay = retryConfig.MaxDelay
			}
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * 1.5) // Increase delay by 50%
			if delay > retryConfig.MaxDelay {
				delay = retryConfig.MaxDelay
			}
		}
	}

	return fmt.Errorf("failed after %d attempts: %v", retryConfig.MaxAttempts, lastErr)
}
