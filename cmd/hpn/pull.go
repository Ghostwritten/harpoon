package main

import (
	"context"
	"fmt"
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
	Long: `Pull container images from a registry to local storage.
Pass extra flags to the underlying runtime (docker/podman/nerdctl/skopeo) by
placing them after -- (e.g. hpn pull -f list.txt -- --tls-verify=false).`,
	RunE: executePull,
}

func init() {
	pullCmd.Flags().StringVarP(&pullImageFile, "file", "f", "", "Image list file (required; use - for stdin)")
	pullCmd.Flags().StringVar(&pullPlatform, "platform", "", "Target platform (e.g. linux/amd64, linux/arm64, all for multi-arch)")
	pullCmd.Flags().IntVar(&pullWorkers, "workers", 0, "Number of concurrent workers (0 = use config or default 5)")
}

func executePull(cmd *cobra.Command, args []string) error {
	if pullImageFile == "" {
		return usageErrorf("missing required --file parameter")
	}

	fmt.Printf("Executing pull action with file: %s\n", pullImageFile)

	selectedRuntime, err := selectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime selection failed: %v", err)
	}
	fmt.Printf("Using container runtime: %s\n", selectedRuntime.Name())

	images, err := readImageList(pullImageFile)
	if err != nil {
		return fmt.Errorf("failed to read image list: %v", err)
	}
	fmt.Printf("Found %d images to pull\n", len(images))

	// Resolve configuration with fallback defaults
	timeout := 5 * time.Minute
	if cfg != nil && cfg.Runtime.Timeout > 0 {
		timeout = cfg.Runtime.Timeout
	}
	retryCfg := containerruntime.RetryConfig{MaxAttempts: 3, Delay: time.Second, MaxDelay: 30 * time.Second}
	if cfg != nil {
		retryCfg = cfg.Runtime.Retry.ToRuntimeRetryConfig()
	}
	maxWorkers := 5
	if pullWorkers > 0 {
		maxWorkers = pullWorkers
	} else if cfg != nil && cfg.Parallel.MaxWorkers > 0 {
		maxWorkers = cfg.Parallel.MaxWorkers
	}

	platformToUse := pullPlatform
	if platformToUse == "" {
		platformToUse = platform
	}
	if platformToUse != "" {
		fmt.Printf("Target platform: %s\n", platformToUse)
	}
	fmt.Printf("Timeout: %v, Max retries: %d, Concurrent workers: %d\n",
		timeout, retryCfg.MaxAttempts, maxWorkers)
	fmt.Println()

	var proxyConfig *containerruntime.ProxyConfig
	if cfg != nil && cfg.Proxy.Enabled {
		proxyConfig = cfg.Proxy.ToRuntimeProxyConfig()
	}
	extraArgs := mergeExtraArgs(getConfigExtraArgsPull(), args)

	ctx := cmd.Context()
	startTime := time.Now()

	results := runWorkerPool(ctx, images, maxWorkers, func(image string) error {
		return retryWithBackoff(ctx, func() error {
			opCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			return selectedRuntime.Pull(opCtx, image, containerruntime.PullOptions{
				Timeout:   timeout,
				Retry:     retryCfg,
				Platform:  platformToUse,
				MultiArch: platformToUse == "all",
				Proxy:     proxyConfig,
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
			fmt.Printf("[%d/%d] ❌ Failed to pull %s: %v\n", i+1, len(results), r.Job, r.Err)
			failedImages = append(failedImages, r.Job)
		} else {
			fmt.Printf("[%d/%d] ✅ Successfully pulled %s\n", i+1, len(results), r.Job)
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
		return fmt.Errorf("failed to pull %d images", len(failedImages))
	}
	return nil
}
