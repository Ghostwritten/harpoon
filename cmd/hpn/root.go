package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/harpoon/hpn/internal/config"
	"github.com/harpoon/hpn/pkg/types"
)

var (
	version = "v1.0"
	commit  = "dev"
	date    = "unknown"
)

// Command line flags matching images.sh
var (
	action     string
	imageFile  string
	registry   string
	project    string
	pushMode   int
	loadMode   int
	saveMode   int
	configFile string
)

// Global configuration
var (
	cfg        *types.Config
	configMgr  *config.Manager
)

var rootCmd = &cobra.Command{
	Use:   "hpn",
	Short: "Manage container images (pull/save/load/push) with flexible modes",
	Long: `Script Name : hpn
Description : Manage container images (pull/save/load/push) with flexible modes
Author      : zong xun
Version     : v1.0`,
	RunE: runCommand,
}

func init() {
	// Initialize configuration manager
	configMgr = config.NewManager()
	
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
	
	// Mark action as required
	rootCmd.MarkFlagRequired("action")
	
	// Custom usage template matching images.sh
	rootCmd.SetUsageTemplate(usageTemplate)
}

const usageTemplate = `Usage: {{.UseLine}} -a <action> -f <image_list> [-r <registry>] [-p <project>] [-c <config>] [--push-mode <1|2|3>] [--load-mode <1|2|3>] [--save-mode <1|2|3>]

Actions:
  pull            Pull images from external registry
  save            Save images into tar files
  load            Load images from tar files
  push            Push images to private registry

Options:
  -a, --action    Action (required): pull | save | load | push
  -f, --file      Image list file (required for pull/save/push)
  -r, --registry  Target registry (default from config or registry.k8s.local)
  -p, --project   Target project namespace (default from config or library)
  -c, --config    Config file (default: $HOME/.hpn/config.yaml)

Modes:
  --push-mode     Push mode (default from config or 1):
                    1 = Push as registry/image:tag
                    2 = Push as registry/project/image:tag
                    3 = Push preserving original project path

  --load-mode     Load mode (default from config or 1):
                    1 = Load all *.tar files from current directory
                    2 = Load all *.tar files from ./images directory
                    3 = Recursively load *.tar from subdirectories under ./images/*/

  --save-mode     Save mode (default from config or 1):
                    1 = Save tar files to current directory
                    2 = Save tar files to ./images/
                    3 = Save tar files to ./images/<project>/

Configuration:
  Config file locations (in order of precedence):
    1. File specified by --config flag
    2. $HOME/.hpn/config.yaml
    3. /etc/hpn/config.yaml
    4. ./config.yaml

  Environment variables (override config file):
    HPN_REGISTRY, HPN_PROJECT, HPN_PROXY_HTTP, HPN_PROXY_HTTPS
    http_proxy, https_proxy (standard proxy variables)

Examples:
  hpn -a pull -f images.txt
  hpn -a save -f images.txt --save-mode 2
  hpn -a push -f images.txt -r harbor.example.com -p myproject --push-mode 2
  hpn --config /path/to/config.yaml -a pull -f images.txt
`

func runCommand(cmd *cobra.Command, args []string) error {
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
	
	// Validate action
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
	
	// Validate file parameter for actions that require it
	if action != "load" && imageFile == "" {
		return fmt.Errorf("missing required -f <image_list> parameter for action '%s'", action)
	}
	
	// Validate mode ranges
	if pushMode < 1 || pushMode > 3 {
		return fmt.Errorf("invalid push-mode '%d'. Valid values: 1, 2, 3", pushMode)
	}
	if loadMode < 1 || loadMode > 3 {
		return fmt.Errorf("invalid load-mode '%d'. Valid values: 1, 2, 3", loadMode)
	}
	if saveMode < 1 || saveMode > 3 {
		return fmt.Errorf("invalid save-mode '%d'. Valid values: 1, 2, 3", saveMode)
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
		return executePush()
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func executePull() error {
	fmt.Printf("Executing pull action with file: %s\n", imageFile)
	
	// Detect container runtime
	runtime, err := detectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime detection failed: %v", err)
	}
	
	fmt.Printf("Using container runtime: %s\n", runtime)
	
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
		
		if err := pullImage(runtime, image); err != nil {
			fmt.Printf("❌ Failed to pull %s: %v\n", image, err)
			failedImages = append(failedImages, image)
		} else {
			fmt.Printf("✅ Successfully pulled %s\n", image)
			successCount++
		}
	}
	
	// Print summary
	fmt.Printf("\n📊 Pull Summary:\n")
	fmt.Printf("  ✅ Successful: %d\n", successCount)
	fmt.Printf("  ❌ Failed: %d\n", len(failedImages))
	
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
	
	// Detect container runtime
	runtime, err := detectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime detection failed: %v", err)
	}
	
	fmt.Printf("Using container runtime: %s\n", runtime)
	
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
		
		if err := saveImage(runtime, image, saveDir, saveMode); err != nil {
			fmt.Printf("❌ Failed to save %s: %v\n", image, err)
			failedImages = append(failedImages, image)
		} else {
			fmt.Printf("✅ Successfully saved %s\n", image)
			successCount++
		}
	}
	
	// Print summary
	fmt.Printf("\n📊 Save Summary:\n")
	fmt.Printf("  ✅ Successful: %d\n", successCount)
	fmt.Printf("  ❌ Failed: %d\n", len(failedImages))
	
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
	
	// Detect container runtime
	runtime, err := detectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime detection failed: %v", err)
	}
	
	fmt.Printf("Using container runtime: %s\n", runtime)
	
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
		
		if err := loadImage(runtime, tarFile); err != nil {
			fmt.Printf("❌ Failed to load %s: %v\n", tarFile, err)
			failedFiles = append(failedFiles, tarFile)
		} else {
			fmt.Printf("✅ Successfully loaded %s\n", tarFile)
			successCount++
		}
	}
	
	// Print summary
	fmt.Printf("\n📊 Load Summary:\n")
	fmt.Printf("  ✅ Successful: %d\n", successCount)
	fmt.Printf("  ❌ Failed: %d\n", len(failedFiles))
	
	if len(failedFiles) > 0 {
		fmt.Printf("\nFailed files:\n")
		for _, file := range failedFiles {
			fmt.Printf("  - %s\n", file)
		}
		return fmt.Errorf("failed to load %d files", len(failedFiles))
	}
	
	return nil
}

func executePush() error {
	fmt.Printf("Executing push action with file: %s, mode: %d, registry: %s, project: %s\n", 
		imageFile, pushMode, registry, project)
	
	// Detect container runtime
	runtime, err := detectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime detection failed: %v", err)
	}
	
	fmt.Printf("Using container runtime: %s\n", runtime)
	
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
		
		if err := pushImage(runtime, image, registry, project, pushMode); err != nil {
			fmt.Printf("❌ Failed to push %s: %v\n", image, err)
			failedImages = append(failedImages, image)
		} else {
			fmt.Printf("✅ Successfully pushed %s\n", image)
			successCount++
		}
	}
	
	// Print summary
	fmt.Printf("\n📊 Push Summary:\n")
	fmt.Printf("  ✅ Successful: %d\n", successCount)
	fmt.Printf("  ❌ Failed: %d\n", len(failedImages))
	
	if len(failedImages) > 0 {
		fmt.Printf("\nFailed images:\n")
		for _, img := range failedImages {
			fmt.Printf("  - %s\n", img)
		}
		return fmt.Errorf("failed to push %d images", len(failedImages))
	}
	
	return nil
}

// detectContainerRuntime detects available container runtime
func detectContainerRuntime() (string, error) {
	// Check for Docker
	if _, err := exec.LookPath("docker"); err == nil {
		if err := exec.Command("docker", "version").Run(); err == nil {
			return "docker", nil
		}
	}
	
	// Check for Podman
	if _, err := exec.LookPath("podman"); err == nil {
		if err := exec.Command("podman", "version").Run(); err == nil {
			return "podman", nil
		}
	}
	
	// Check for Nerdctl
	if _, err := exec.LookPath("nerdctl"); err == nil {
		if err := exec.Command("nerdctl", "version").Run(); err == nil {
			return "nerdctl", nil
		}
	}
	
	return "", fmt.Errorf("no container runtime found. Please install docker, podman, or nerdctl")
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

// pullImage pulls a single image using the specified runtime
func pullImage(runtime, image string) error {
	var cmd *exec.Cmd
	
	switch runtime {
	case "docker":
		cmd = exec.Command("docker", "pull", image)
	case "podman":
		cmd = exec.Command("podman", "pull", image)
	case "nerdctl":
		// For nerdctl, add insecure registry flag for private registries
		if strings.Contains(image, "registry.k8s.local") || 
		   strings.Contains(image, "localhost") ||
		   !strings.Contains(image, ".") {
			cmd = exec.Command("nerdctl", "--insecure-registry", "pull", image)
		} else {
			cmd = exec.Command("nerdctl", "pull", image)
		}
	default:
		return fmt.Errorf("unsupported runtime: %s", runtime)
	}
	
	// Set up environment for proxy if needed
	cmd.Env = os.Environ()
	
	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %v, output: %s", err, string(output))
	}
	
	return nil
}

// saveImage saves a single image to tar file
func saveImage(runtime, image, baseDir string, mode int) error {
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
	
	// Execute save command
	var cmd *exec.Cmd
	
	switch runtime {
	case "docker":
		cmd = exec.Command("docker", "save", "-o", tarPath, image)
	case "podman":
		cmd = exec.Command("podman", "save", "-o", tarPath, image)
	case "nerdctl":
		cmd = exec.Command("nerdctl", "save", "-o", tarPath, image)
	default:
		return fmt.Errorf("unsupported runtime: %s", runtime)
	}
	
	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %v, output: %s", err, string(output))
	}
	
	// Check if file was created successfully
	if _, err := os.Stat(tarPath); err != nil {
		return fmt.Errorf("tar file was not created: %v", err)
	}
	
	fmt.Printf("  💾 Saved to: %s\n", tarPath)
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
func loadImage(runtime, tarFile string) error {
	var cmd *exec.Cmd
	
	switch runtime {
	case "docker":
		cmd = exec.Command("docker", "load", "-i", tarFile)
	case "podman":
		cmd = exec.Command("podman", "load", "-i", tarFile)
	case "nerdctl":
		cmd = exec.Command("nerdctl", "load", "-i", tarFile)
	default:
		return fmt.Errorf("unsupported runtime: %s", runtime)
	}
	
	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %v, output: %s", err, string(output))
	}
	
	return nil
}

// pushImage pushes a single image to registry with the specified mode
func pushImage(runtime, image, targetRegistry, targetProject string, mode int) error {
	var targetImage string
	
	// Parse original image name and tag
	imageName, imageTag := parseImageNameAndTag(image)
	
	switch mode {
	case 1:
		// Mode 1: registry/image:tag
		targetImage = fmt.Sprintf("%s/%s:%s", targetRegistry, imageName, imageTag)
	case 2:
		// Mode 2: registry/project/image:tag
		targetImage = fmt.Sprintf("%s/%s/%s:%s", targetRegistry, targetProject, imageName, imageTag)
	case 3:
		// Mode 3: preserve original project path
		originalProject := extractProjectFromImage(image)
		targetImage = fmt.Sprintf("%s/%s/%s:%s", targetRegistry, originalProject, imageName, imageTag)
	}
	
	fmt.Printf("  🏷️  Tagging %s -> %s\n", image, targetImage)
	
	// Tag the image
	var tagCmd *exec.Cmd
	switch runtime {
	case "docker":
		tagCmd = exec.Command("docker", "tag", image, targetImage)
	case "podman":
		tagCmd = exec.Command("podman", "tag", image, targetImage)
	case "nerdctl":
		tagCmd = exec.Command("nerdctl", "tag", image, targetImage)
	default:
		return fmt.Errorf("unsupported runtime: %s", runtime)
	}
	
	if output, err := tagCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tag command failed: %v, output: %s", err, string(output))
	}
	
	// Push the image
	var pushCmd *exec.Cmd
	switch runtime {
	case "docker":
		pushCmd = exec.Command("docker", "push", targetImage)
	case "podman":
		pushCmd = exec.Command("podman", "push", targetImage)
	case "nerdctl":
		pushCmd = exec.Command("nerdctl", "push", targetImage)
	default:
		return fmt.Errorf("unsupported runtime: %s", runtime)
	}
	
	if output, err := pushCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("push command failed: %v, output: %s", err, string(output))
	}
	
	fmt.Printf("  📤 Pushed to: %s\n", targetImage)
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
}