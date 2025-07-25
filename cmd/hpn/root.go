package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/harpoon/hpn/internal/config"
	containerruntime "github.com/harpoon/hpn/internal/runtime"
	"github.com/harpoon/hpn/pkg/types"
)

var (
	version = "v1.1"
	commit  = "dev"
	date    = "unknown"
)

// Command line flags matching images.sh
var (
	action       string
	imageFile    string
	registry     string
	project      string
	pushMode     int
	loadMode     int
	saveMode     int
	configFile   string
	runtimeName  string
	autoFallback bool
)

// Global configuration
var (
	cfg             *types.Config
	configMgr       *config.Manager
	runtimeDetector *containerruntime.Detector
)

var rootCmd = &cobra.Command{
	Use:   "hpn",
	Short: "Manage container images (pull/save/load/push) with flexible modes",
	Long:  `Manage container images (pull/save/load/push) with flexible modes`,
	Version:       getVersionString(),
	RunE:          runCommand,
	SilenceUsage:  true, // Don't show usage on errors
	SilenceErrors: true, // Don't let Cobra print errors automatically
}

// Version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Display detailed version information including build details",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(getDetailedVersionString())
	},
}

func init() {
	// Initialize configuration manager
	configMgr = config.NewManager()
	
	// Initialize runtime detector
	runtimeDetector = containerruntime.NewDetector()
	
	// Required flags matching images.sh interface
	rootCmd.Flags().StringVarP(&action, "action", "a", "", "Action (required): pull | save | load | push")
	rootCmd.Flags().StringVarP(&imageFile, "file", "f", "", "Image list file (required for pull/save/push)")
	rootCmd.Flags().StringVarP(&registry, "registry", "r", "", "Target registry")
	rootCmd.Flags().StringVarP(&project, "project", "p", "", "Target project namespace")
	
	// Mode flags
	rootCmd.Flags().IntVar(&pushMode, "push-mode", 0, "Push mode (1|2|3)")
	rootCmd.Flags().IntVar(&loadMode, "load-mode", 0, "Load mode (1|2|3)")
	rootCmd.Flags().IntVar(&saveMode, "save-mode", 0, "Save mode (1|2|3)")
	
	// Configuration flag
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "Config file (default is $HOME/.hpn/config.yaml)")
	
	// Runtime flags
	rootCmd.Flags().StringVar(&runtimeName, "runtime", "", "Container runtime to use (docker|podman|nerdctl)")
	rootCmd.Flags().BoolVar(&autoFallback, "auto-fallback", false, "Automatically fallback to available runtime")
	
	// Version flags (in addition to --version)
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
	rootCmd.Flags().BoolP("Version", "V", false, "Show version information")
	
	// Add version subcommand
	rootCmd.AddCommand(versionCmd)
	
	// Custom usage template matching images.sh
	rootCmd.SetUsageTemplate(usageTemplate)
}

const usageTemplate = `Usage: {{.UseLine}} -a <action> -f <file> [options]

Actions:
  pull    Pull images from registry
  save    Save images to tar files  
  load    Load images from tar files
  push    Push images to registry

Options:
  -a, --action     Action: pull | save | load | push
  -f, --file       Image list file
  -r, --registry   Target registry
  -p, --project    Target project namespace
  -c, --config     Config file path
      --runtime    Container runtime: docker | podman | nerdctl
      --auto-fallback  Auto fallback to available runtime
  -v, --version    Show version
  -h, --help       Show help

Modes:
  --push-mode      1=registry/image:tag  2=registry/project/image:tag
  --save-mode      1=current dir  2=./images/  3=./images/<project>/
  --load-mode      1=current dir  2=./images/  3=recursive ./images/*/

Examples:
  hpn -a pull -f images.txt
  hpn -a save -f images.txt --save-mode 2
  hpn -a push -f images.txt -r harbor.com -p prod --push-mode 2
  hpn --runtime podman -a pull -f images.txt
`

func runCommand(cmd *cobra.Command, args []string) error {
	// Check for version flags first
	if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
		printVersionInfo()
		return nil
	}
	if versionFlag, _ := cmd.Flags().GetBool("Version"); versionFlag {
		printVersionInfo()
		return nil
	}
	
	// Load configuration
	var err error
	cfg, err = configMgr.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}
	
	// Apply configuration defaults if flags are not set
	if registry == "" {
		registry = cfg.Registry
	}
	if project == "" {
		project = cfg.Project
	}
	if pushMode == 0 {
		pushMode = int(cfg.Modes.PushMode)
	}
	if loadMode == 0 {
		loadMode = int(cfg.Modes.LoadMode)
	}
	if saveMode == 0 {
		saveMode = int(cfg.Modes.SaveMode)
	}
	
	// Validate action (skip if empty, as it might be a version-only call)
	if action != "" {
		validActions := []string{"pull", "save", "load", "push"}
		actionValid := false
		for _, validAction := range validActions {
			if action == validAction {
				actionValid = true
				break
			}
		}
		
		if !actionValid {
			return fmt.Errorf("invalid action '%s'. Valid actions: %s", action, strings.Join(validActions, ", "))
		}
	} else {
		// If no action provided and no version flags, show error
		return fmt.Errorf("missing required -a <action> parameter. Use -h for help or -v for version")
	}
	
	// Validate file parameter for actions that require it
	if action != "load" && imageFile == "" {
		return fmt.Errorf("missing required -f <image_list> parameter for action '%s'", action)
	}
	
	// Validate mode compatibility with action
	switch action {
	case "push":
		// Check for incompatible modes
		if cmd.Flags().Changed("save-mode") {
			return fmt.Errorf("--save-mode cannot be used with push action")
		}
		if cmd.Flags().Changed("load-mode") {
			return fmt.Errorf("--load-mode cannot be used with push action")
		}
		// Validate push mode range
		if pushMode < 1 || pushMode > 2 {
			return fmt.Errorf("invalid push-mode '%d'. Valid values: 1, 2", pushMode)
		}
	case "save":
		// Check for incompatible modes
		if cmd.Flags().Changed("push-mode") {
			return fmt.Errorf("--push-mode cannot be used with save action")
		}
		if cmd.Flags().Changed("load-mode") {
			return fmt.Errorf("--load-mode cannot be used with save action")
		}
		// Validate save mode range
		if saveMode < 1 || saveMode > 3 {
			return fmt.Errorf("invalid save-mode '%d'. Valid values: 1, 2, 3", saveMode)
		}
	case "load":
		// Check for incompatible modes
		if cmd.Flags().Changed("push-mode") {
			return fmt.Errorf("--push-mode cannot be used with load action")
		}
		if cmd.Flags().Changed("save-mode") {
			return fmt.Errorf("--save-mode cannot be used with load action")
		}
		// Validate load mode range
		if loadMode < 1 || loadMode > 3 {
			return fmt.Errorf("invalid load-mode '%d'. Valid values: 1, 2, 3", loadMode)
		}
	case "pull":
		// Pull doesn't use any modes, check for incompatible modes
		if cmd.Flags().Changed("push-mode") {
			return fmt.Errorf("--push-mode cannot be used with pull action")
		}
		if cmd.Flags().Changed("save-mode") {
			return fmt.Errorf("--save-mode cannot be used with pull action")
		}
		if cmd.Flags().Changed("load-mode") {
			return fmt.Errorf("--load-mode cannot be used with pull action")
		}
	}
	
	// Smart push mode adjustment: if user specifies project but uses default push mode 1,
	// automatically switch to push mode 2 to include the project
	if action == "push" && pushMode == 1 && project != "" {
		// Check if project was explicitly specified by user (not just from config default)
		projectExplicitlySet := cmd.Flags().Changed("project") || 
			(cfg != nil && cfg.Project != project) // project differs from config default
		
		if projectExplicitlySet {
			pushMode = 2
			fmt.Printf("Auto-adjusted to push mode 2 for project '%s'\n", project)
		}
	}
	
	// Execute the action
	switch action {
	case "pull":
		return executePull()
	case "save":
		return executeSave()
	case "load":
		return executeLoad()
	case "push":
		return executePush(cmd)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func executePull() error {
	fmt.Printf("Executing pull action with file: %s\n", imageFile)
	
	// Select container runtime
	selectedRuntime, err := selectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime selection failed: %v", err)
	}
	
	fmt.Printf("Using container runtime: %s\n", selectedRuntime.Name())
	
	// Read image list from file
	images, err := readImageList(imageFile)
	if err != nil {
		return fmt.Errorf("failed to read image list: %v", err)
	}
	
	fmt.Printf("Found %d images to pull\n", len(images))
	
	// Pull each image
	successCount := 0
	failedImages := []string{}
	
	for i, image := range images {
		fmt.Printf("[%d/%d] Pulling %s...\n", i+1, len(images), image)
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		pullOptions := containerruntime.PullOptions{
			Timeout: 5 * time.Minute,
		}
		
		if err := selectedRuntime.Pull(ctx, image, pullOptions); err != nil {
			fmt.Printf("❌ Failed to pull %s: %v\n", image, err)
			failedImages = append(failedImages, image)
		} else {
			fmt.Printf("✅ Successfully pulled %s\n", image)
			successCount++
		}
		cancel()
	}
	
	// Print summary
	fmt.Printf("\nSummary: %d successful, %d failed\n", successCount, len(failedImages))
	
	if len(failedImages) > 0 {
		fmt.Printf("\nFailed images:\n")
		for _, img := range failedImages {
			fmt.Printf("  - %s\n", img)
		}
		return fmt.Errorf("failed to pull %d images", len(failedImages))
	}
	
	return nil
}

func executeSave() error {
	fmt.Printf("Executing save action with file: %s, mode: %d\n", imageFile, saveMode)
	
	// Select container runtime
	selectedRuntime, err := selectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime selection failed: %v", err)
	}
	
	fmt.Printf("Using container runtime: %s\n", selectedRuntime.Name())
	
	// Read image list from file
	images, err := readImageList(imageFile)
	if err != nil {
		return fmt.Errorf("failed to read image list: %v", err)
	}
	
	fmt.Printf("Found %d images to save\n", len(images))
	
	// Determine save directory based on mode
	var saveDir string
	switch saveMode {
	case 1:
		saveDir = "." // Current directory
	case 2:
		saveDir = "./images"
		if err := os.MkdirAll(saveDir, 0755); err != nil {
			return fmt.Errorf("failed to create images directory: %v", err)
		}
	case 3:
		// Will create project-specific directories as needed
		saveDir = "./images"
		if err := os.MkdirAll(saveDir, 0755); err != nil {
			return fmt.Errorf("failed to create images directory: %v", err)
		}
	}
	
	fmt.Printf("Save mode %d: saving to %s\n", saveMode, saveDir)
	
	// Save each image
	successCount := 0
	failedImages := []string{}
	
	for i, image := range images {
		fmt.Printf("[%d/%d] Saving %s...\n", i+1, len(images), image)
		
		if err := saveImage(selectedRuntime, image, saveDir, saveMode); err != nil {
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

func executeLoad() error {
	fmt.Printf("Executing load action with mode: %d\n", loadMode)
	
	// Select container runtime
	selectedRuntime, err := selectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime selection failed: %v", err)
	}
	
	fmt.Printf("Using container runtime: %s\n", selectedRuntime.Name())
	
	// Determine load directory based on mode
	var loadDir string
	var tarFiles []string
	
	switch loadMode {
	case 1:
		// Load from current directory
		loadDir = "."
		files, err := findTarFiles(loadDir, false)
		if err != nil {
			return fmt.Errorf("failed to find tar files in current directory: %v", err)
		}
		tarFiles = files
	case 2:
		// Load from ./images/ directory
		loadDir = "./images"
		files, err := findTarFiles(loadDir, false)
		if err != nil {
			return fmt.Errorf("failed to find tar files in images directory: %v", err)
		}
		tarFiles = files
	case 3:
		// Recursively load from ./images/*/ subdirectories
		loadDir = "./images"
		files, err := findTarFiles(loadDir, true)
		if err != nil {
			return fmt.Errorf("failed to find tar files recursively: %v", err)
		}
		tarFiles = files
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

func executePush(cmd *cobra.Command) error {
	fmt.Printf("Executing push action with file: %s, mode: %d, registry: %s, project: %s\n", 
		imageFile, pushMode, registry, project)
	
	// Detect container runtime
	selectedRuntime, err := selectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime selection failed: %v", err)
	}
	
	fmt.Printf("Using container runtime: %s\n", selectedRuntime.Name())
	
	// Read image list from file
	images, err := readImageList(imageFile)
	if err != nil {
		return fmt.Errorf("failed to read image list: %v", err)
	}
	
	fmt.Printf("Found %d images to push\n", len(images))
	fmt.Printf("Target registry: %s\n", registry)
	fmt.Printf("Target project: %s\n", project)
	fmt.Printf("Push mode: %d\n", pushMode)
	
	// Push each image
	successCount := 0
	failedImages := []string{}
	
	for i, image := range images {
		fmt.Printf("[%d/%d] Pushing %s...\n", i+1, len(images), image)
		
		// Determine project name for this specific image based on push mode
		var effectiveProject string
		if pushMode == 2 {
			// For mode 2, use smart project selection
			if cmd.Flags().Changed("project") {
				// User explicitly specified project via command line
				effectiveProject = project
			} else if cfg != nil && cfg.Project != "" && cfg.Project != "library" {
				// Use config file project (if not default "library")
				effectiveProject = cfg.Project
			} else {
				// Use original image project name
				effectiveProject = extractProjectFromImage(image)
			}
		} else {
			// For mode 1, use the project as-is (though it won't be used)
			effectiveProject = project
		}
		
		if err := pushImage(selectedRuntime, image, registry, effectiveProject, pushMode); err != nil {
			fmt.Printf("❌ Failed to push %s: %v\n", image, err)
			failedImages = append(failedImages, image)
		} else {
			fmt.Printf("✅ Successfully pushed %s\n", image)
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



// saveImage saves a single image to tar file
func saveImage(containerRuntime containerruntime.ContainerRuntime, image, baseDir string, mode int) error {
	// Parse image name to generate tar filename
	tarFilename := generateTarFilename(image)
	
	var tarPath string
	
	switch mode {
	case 1, 2:
		// Mode 1: current directory, Mode 2: ./images/
		tarPath = fmt.Sprintf("%s/%s", baseDir, tarFilename)
	case 3:
		// Mode 3: ./images/<project>/
		projectDir := extractProjectFromImage(image)
		fullDir := fmt.Sprintf("%s/%s", baseDir, projectDir)
		if err := os.MkdirAll(fullDir, 0755); err != nil {
			return fmt.Errorf("failed to create project directory %s: %v", fullDir, err)
		}
		tarPath = fmt.Sprintf("%s/%s", fullDir, tarFilename)
	}
	
	// Execute save command using runtime interface
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	
	if err := containerRuntime.Save(ctx, image, tarPath); err != nil {
		return fmt.Errorf("failed to save image: %v", err)
	}
	
	// Check if file was created successfully
	if _, err := os.Stat(tarPath); err != nil {
		return fmt.Errorf("tar file was not created: %v", err)
	}
	
	fmt.Printf("  Saved: %s\n", tarPath)
	return nil
}

// generateTarFilename generates tar filename from image name
func generateTarFilename(image string) string {
	// Replace problematic characters for filename
	filename := strings.ReplaceAll(image, "/", "_")
	filename = strings.ReplaceAll(filename, ":", "_")
	
	// Add .tar extension
	return filename + ".tar"
}

// extractProjectFromImage extracts project name from image for mode 3
func extractProjectFromImage(image string) string {
	parts := strings.Split(image, "/")
	
	if len(parts) >= 3 {
		// For images like registry.k8s.io/coredns/coredns:v1.11.1
		return parts[len(parts)-2] // Return "coredns"
	} else if len(parts) == 2 {
		// For images like calico/node:v3.28.2
		return parts[0] // Return "calico"
	} else {
		// For images like nginx:latest
		return "library" // Default project name
	}
}

// findTarFiles finds all .tar files in the specified directory
func findTarFiles(dir string, recursive bool) ([]string, error) {
	var tarFiles []string
	
	if recursive {
		// Recursively find tar files in subdirectories
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".tar") {
				tarFiles = append(tarFiles, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Find tar files only in the specified directory
		files, err := filepath.Glob(filepath.Join(dir, "*.tar"))
		if err != nil {
			return nil, err
		}
		tarFiles = files
	}
	
	return tarFiles, nil
}

// loadImage loads a single tar file using the specified runtime
func loadImage(containerRuntime containerruntime.ContainerRuntime, tarFile string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	
	if err := containerRuntime.Load(ctx, tarFile); err != nil {
		return fmt.Errorf("failed to load image: %v", err)
	}
	
	return nil
}

// pushImage pushes a single image to registry with the specified mode
func pushImage(containerRuntime containerruntime.ContainerRuntime, image, targetRegistry, targetProject string, mode int) error {
	var targetImage string
	
	// Parse original image name and tag
	imageName, imageTag := parseImageNameAndTag(image)
	
	switch mode {
	case 1:
		// Mode 1: registry/image:tag (不包含项目名称)
		targetImage = fmt.Sprintf("%s/%s:%s", targetRegistry, imageName, imageTag)
	case 2:
		// Mode 2: registry/project/image:tag
		targetImage = fmt.Sprintf("%s/%s/%s:%s", targetRegistry, targetProject, imageName, imageTag)
		fmt.Printf("  Project: %s\n", targetProject)
	}
	
	fmt.Printf("  Tag: %s -> %s\n", image, targetImage)
	
	// Tag the image
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	if err := containerRuntime.Tag(ctx, image, targetImage); err != nil {
		return fmt.Errorf("failed to tag image: %v", err)
	}
	
	// Push the image
	pushOptions := containerruntime.PushOptions{
		Timeout: 10 * time.Minute,
	}
	
	if err := containerRuntime.Push(ctx, targetImage, pushOptions); err != nil {
		return fmt.Errorf("failed to push image: %v", err)
	}
	
	fmt.Printf("  Pushed: %s\n", targetImage)
	return nil
}

// parseImageNameAndTag parses image name and tag from full image string
func parseImageNameAndTag(image string) (string, string) {
	parts := strings.Split(image, "/")
	lastPart := parts[len(parts)-1]
	
	if strings.Contains(lastPart, ":") {
		tagParts := strings.Split(lastPart, ":")
		return tagParts[0], tagParts[1]
	}
	
	return lastPart, "latest"
}// getVer
// getVersionString returns formatted version information for cobra
func getVersionString() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
}

// printVersionInfo prints detailed version information consistently
func printVersionInfo() {
	fmt.Printf("Harpoon (hpn) %s\n", version)
	fmt.Printf("Commit: %s\n", commit)
	fmt.Printf("Built: %s\n", date)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

// getDetailedVersionString returns detailed version information as string
func getDetailedVersionString() string {
	return fmt.Sprintf(`Harpoon (hpn) %s
Commit: %s
Built: %s
Go version: %s
Platform: %s/%s`, version, commit, date, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}