package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/harpoon/hpn/internal/config"
	containerruntime "github.com/harpoon/hpn/internal/runtime"
	"github.com/harpoon/hpn/internal/version"
	"github.com/harpoon/hpn/pkg/types"
)

// Legacy variables for backward compatibility
// These are now managed by the version package
var (
	legacyVersion = version.GetVersion()
	legacyCommit  = version.GetShortCommit()
	legacyDate    = version.BuildDate
)

// Global command line flags
var (
	configFile   string
	runtimeName  string
	autoFallback bool
	platform     string
	debug        bool
)

// Global configuration
var (
	cfg             *types.Config
	configMgr       *config.Manager
	runtimeDetector *containerruntime.Detector
)

var rootCmd = &cobra.Command{
	Use:   "hpn",
	Short: "Manage container images (pull/save/load/push)",
	Long:  `Harpoon (hpn) is a container image management tool that supports multiple container runtimes.

It provides subcommands for pulling, saving, loading, and pushing container images with flexible configuration options.`,
	Version:       version.GetFullVersion(),
	SilenceUsage:  true, // Don't show usage on errors
	SilenceErrors: true, // Don't let Cobra print errors automatically
}

// Version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Display detailed version information including build details",
	Run: func(cmd *cobra.Command, args []string) {
		version.PrintDetailedVersion()
	},
}

func init() {
	// Initialize configuration manager
	configMgr = config.NewManager()
	
	// Initialize runtime detector
	runtimeDetector = containerruntime.NewDetector()
	
	// Global flags (available to all subcommands)
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Config file (default is $HOME/.hpn/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&runtimeName, "runtime", "", "Container runtime to use (docker|podman|nerdctl|skopeo)")
	rootCmd.PersistentFlags().BoolVar(&autoFallback, "auto-fallback", false, "Automatically fallback to available runtime")
	rootCmd.PersistentFlags().StringVar(&platform, "platform", "", "Target platform (e.g., linux/amd64, linux/arm64, all for multi-arch)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug mode to show detailed error information")
	
	// Version flags (in addition to --version)
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
	rootCmd.Flags().BoolP("Version", "V", false, "Show version information")
	
	// Add version subcommand
	rootCmd.AddCommand(versionCmd)
	
	// Load configuration when root command is executed
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
	var err error
	cfg, err = configMgr.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}
		return nil
	}
}

// IsDebug returns whether debug mode is enabled
func IsDebug() bool {
	return debug
}

// Usage template is now handled by Cobra's default template

// runCommand is no longer needed - subcommands handle their own execution

// Old execute functions removed - now handled by subcommands

// selectContainerRuntime selects the appropriate container runtime
func selectContainerRuntime() (containerruntime.ContainerRuntime, error) {
	// If runtime is explicitly specified via flag
	if runtimeName != "" {
		selectedRuntime, err := runtimeDetector.GetByName(runtimeName)
		if err != nil {
			return nil, fmt.Errorf("specified runtime '%s' is not available: %v", runtimeName, err)
		}
		return selectedRuntime, nil
	}
	
	// Check if runtime is specified in config
	var configuredRuntime string
	if cfg != nil && cfg.Runtime.Preferred != "" {
		configuredRuntime = cfg.Runtime.Preferred
	}
	
	// If configured runtime is specified, try to use it
	if configuredRuntime != "" {
		configuredRuntimeObj, err := runtimeDetector.GetByName(configuredRuntime)
		if err == nil {
			return configuredRuntimeObj, nil
		}
		
		// Configured runtime is not available, check for alternatives
		available := runtimeDetector.DetectAvailable()
		if len(available) == 0 {
			return nil, fmt.Errorf("no container runtime found. Please install docker, podman, or nerdctl")
		}
		
		// Check if auto-fallback is enabled
		if autoFallback || (cfg != nil && cfg.Runtime.AutoFallback) {
			fmt.Printf("Runtime '%s' unavailable, using '%s'\n", configuredRuntime, available[0].Name())
			return available[0], nil
		}
		
		// Ask user for confirmation
		fmt.Printf("Runtime '%s' is not available\n", configuredRuntime)
		fmt.Printf("Found available runtime: %s\n", available[0].Name())
		fmt.Printf("Use '%s' instead of '%s'? (y/N): ", available[0].Name(), configuredRuntime)
		
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		
		if response == "y" || response == "yes" {
			fmt.Printf("Using '%s' runtime\n", available[0].Name())
			return available[0], nil
		} else {
			return nil, fmt.Errorf("user declined runtime fallback. Please install '%s' or update config", configuredRuntime)
		}
	}
	
	// No specific runtime configured, use the preferred one
	preferred := runtimeDetector.GetPreferred()
	if preferred == nil {
		return nil, fmt.Errorf("no container runtime found. Please install docker, podman, or nerdctl")
	}
	
	return preferred, nil
}

// readImageList reads image list from file
func readImageList(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", filename, err)
	}
	defer file.Close()
	
	var images []string
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line != "" && !strings.HasPrefix(line, "#") {
			images = append(images, line)
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	
	if len(images) == 0 {
		return nil, fmt.Errorf("no images found in file %s", filename)
	}
	
	return images, nil
}



// saveImage removed - now handled by save.go

// verifyImageChecksum verifies the checksum of a tar file
func verifyImageChecksum(tarPath string) (bool, error) {
	checksumPath := tarPath + ".sha256"
	
	// Read expected checksum
	expectedChecksumBytes, err := os.ReadFile(checksumPath)
	if err != nil {
		return false, fmt.Errorf("failed to read checksum file: %w", err)
	}
	expectedChecksum := strings.TrimSpace(string(expectedChecksumBytes))
	
	// Calculate actual checksum
	file, err := os.Open(tarPath)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false, fmt.Errorf("failed to calculate checksum: %w", err)
	}
	
	actualChecksum := fmt.Sprintf("%x", hash.Sum(nil))
	
	return actualChecksum == expectedChecksum, nil
}

// generateTarFilename generates tar filename from image name
func generateTarFilename(image string) string {
	// Replace problematic characters for filename
	filename := strings.ReplaceAll(image, "/", "_")
	filename = strings.ReplaceAll(filename, ":", "_")
	
	// Add .tar extension
	return filename + ".tar"
}

// extractProjectFromImage removed - now in image.go

// findTarFiles, loadImage, pushImage, parseImageNameAndTag removed - now handled by subcommands

// printVersionInfo prints version information (legacy function)
func printVersionInfo() {
	version.PrintVersion()
}