package runtime

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/harpoon/hpn/pkg/errors"
)

// SkopeoRuntime implements ContainerRuntime for Skopeo
type SkopeoRuntime struct {
	command string
}

// NewSkopeoRuntime creates a new Skopeo runtime
func NewSkopeoRuntime() *SkopeoRuntime {
	return &SkopeoRuntime{
		command: "skopeo",
	}
}

// Name returns the runtime name
func (s *SkopeoRuntime) Name() string {
	return "skopeo"
}

// IsAvailable checks if Skopeo is available
func (s *SkopeoRuntime) IsAvailable() bool {
	if !IsCommandAvailable(s.command) {
		return false
	}

	// Test if Skopeo is working
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.command, "--version")
	return cmd.Run() == nil
}

// Pull pulls an image from a registry
func (s *SkopeoRuntime) Pull(ctx context.Context, image string, options PullOptions) error {
	args := []string{"copy", "--preserve-digests"}

	// Add --all flag for multi-arch support
	if options.MultiArch || options.Platform == "all" {
		args = append(args, "--all")
	}

	// Add platform if specified (and not "all")
	if options.Platform != "" && options.Platform != "all" {
		args = append(args, "--override-arch", extractArch(options.Platform))
		args = append(args, "--override-os", extractOS(options.Platform))
	}

	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	// Source: docker://image
	args = append(args, fmt.Sprintf("docker://%s", image))
	
	// Destination: dir:path (temporary directory for pull)
	// Note: Skopeo pull typically saves to a directory, but we'll use a temp dir
	// For consistency with other runtimes, we'll save to a temp location
	tempDir := fmt.Sprintf("/tmp/skopeo-pull-%s", sanitizeImageName(image))
	defer os.RemoveAll(tempDir)
	
	args = append(args, fmt.Sprintf("dir:%s", tempDir))

	cmd := exec.CommandContext(ctx, s.command, args...)

	// Set proxy environment if configured
	if options.Proxy != nil && options.Proxy.Enabled {
		env := os.Environ()
		if options.Proxy.HTTP != "" {
			env = append(env, fmt.Sprintf("http_proxy=%s", options.Proxy.HTTP))
		}
		if options.Proxy.HTTPS != "" {
			env = append(env, fmt.Sprintf("https_proxy=%s", options.Proxy.HTTPS))
		}
		cmd.Env = env
	}

	// Capture output for debug mode
	var stdout, stderr bytes.Buffer
	if options.Debug {
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		// In non-debug mode, capture stderr for error messages
		cmd.Stdout = os.Stdout
		cmd.Stderr = &stderr
	}

	err := cmd.Run()
	if err != nil {
		errMsg := fmt.Sprintf("failed to pull image %s", image)
		if options.Debug && (stdout.Len() > 0 || stderr.Len() > 0) {
			errMsg += fmt.Sprintf("\nStdout:\n%s\nStderr:\n%s", stdout.String(), stderr.String())
		} else if stderr.Len() > 0 {
			errMsg += fmt.Sprintf("\nError: %s", stderr.String())
		}
		return errors.Wrap(err, errors.ErrRuntimeCommand, errMsg)
	}

	return nil
}

// Save saves an image to a tar file using docker-archive format
func (s *SkopeoRuntime) Save(ctx context.Context, image string, tarPath string, options SaveOptions) error {
	// First, try docker-archive format (default, most compatible)
	err := s.saveWithDockerArchive(ctx, image, tarPath, options)
	if err == nil {
		return nil
	}

	// Check if error is due to OCI manifest type
	errStr := err.Error()
	if strings.Contains(errStr, "Unsupported manifest type") || 
	   strings.Contains(errStr, "need a Docker schema 2 manifest") {
		// Fallback: use oci format (dir), then pack to tar
		return s.saveWithOCIDir(ctx, image, tarPath, options)
	}

	// Return original error if not OCI manifest related
	return err
}

// saveWithDockerArchive saves image using docker-archive format
func (s *SkopeoRuntime) saveWithDockerArchive(ctx context.Context, image string, tarPath string, options SaveOptions) error {
	args := []string{"copy", "--preserve-digests"}

	// Add --all flag for multi-arch support
	if options.MultiArch {
		args = append(args, "--all")
	} else {
		// If not multi-arch, specify platform
		// Default to linux/amd64 if platform is not specified (most container images are Linux)
		targetPlatform := options.Platform
		if targetPlatform == "" {
			targetPlatform = "linux/amd64"
		}
		
		// Only add platform override if it's not "all" and not empty
		if targetPlatform != "" && targetPlatform != "all" {
			args = append(args, "--override-arch", extractArch(targetPlatform))
			args = append(args, "--override-os", extractOS(targetPlatform))
		}
	}

	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	// Source: docker://image
	args = append(args, fmt.Sprintf("docker://%s", image))
	
	// Destination: docker-archive:tarPath (default format, most compatible)
	args = append(args, fmt.Sprintf("docker-archive:%s", tarPath))

	cmd := exec.CommandContext(ctx, s.command, args...)

	// Set proxy environment if configured
	env := os.Environ()
	cmd.Env = env

	// For Skopeo, show progress output in non-debug mode too
	// Skopeo copy shows useful progress information (copying layers, etc.)
	var stdout, stderr bytes.Buffer
	if options.Debug {
		// Debug mode: show all output
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		// Non-debug mode: show stdout (progress) but capture stderr (errors)
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
		cmd.Stderr = &stderr
	}

	err := cmd.Run()
	if err != nil {
		errMsg := fmt.Sprintf("failed to save image %s to %s", image, tarPath)
		if options.Debug && (stdout.Len() > 0 || stderr.Len() > 0) {
			errMsg += fmt.Sprintf("\nStdout:\n%s\nStderr:\n%s", stdout.String(), stderr.String())
		} else if stderr.Len() > 0 {
			errMsg += fmt.Sprintf("\nError: %s", stderr.String())
		}
		return errors.Wrap(err, errors.ErrRuntimeCommand, errMsg)
	}

	// Generate checksum file if requested (default: true)
	if options.Checksum {
		if _, err := generateChecksum(tarPath); err != nil {
			return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to generate checksum for %s", tarPath))
		}
	}

	return nil
}

// saveWithOCIDir saves image using oci format (dir), then packs to tar
func (s *SkopeoRuntime) saveWithOCIDir(ctx context.Context, image string, tarPath string, options SaveOptions) error {
	// Create temporary directory for OCI layout
	tempDir := fmt.Sprintf("/tmp/skopeo-oci-%s-%d", sanitizeImageName(image), time.Now().Unix())
	defer os.RemoveAll(tempDir)

	args := []string{"copy", "--preserve-digests"}

	// Add --all flag for multi-arch support
	if options.MultiArch {
		args = append(args, "--all")
	} else {
		// If not multi-arch, specify platform
		targetPlatform := options.Platform
		if targetPlatform == "" {
			targetPlatform = "linux/amd64"
		}
		
		if targetPlatform != "" && targetPlatform != "all" {
			args = append(args, "--override-arch", extractArch(targetPlatform))
			args = append(args, "--override-os", extractOS(targetPlatform))
		}
	}

	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	// Source: docker://image
	args = append(args, fmt.Sprintf("docker://%s", image))
	
	// Destination: oci:tempDir (OCI layout format)
	args = append(args, fmt.Sprintf("oci:%s", tempDir))

	cmd := exec.CommandContext(ctx, s.command, args...)
	env := os.Environ()
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	if options.Debug {
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		cmd.Stderr = &stderr
	}

	err := cmd.Run()
	if err != nil {
		errMsg := fmt.Sprintf("failed to save image %s to OCI dir: %s", image, stderr.String())
		return errors.Wrap(err, errors.ErrRuntimeCommand, errMsg)
	}

	// Pack OCI directory to tar file
	// Use tar command to create tar from directory
	// Pack OCI directory to tar file
	fmt.Printf("    📦 Packing OCI layout to tar file...\n")
	tarCmd := exec.CommandContext(ctx, "tar", "-czf", tarPath, "-C", tempDir, ".")
	tarCmd.Env = env
	
	var tarStderr bytes.Buffer
	if options.Debug {
		tarCmd.Stderr = io.MultiWriter(os.Stderr, &tarStderr)
	} else {
		tarCmd.Stderr = &tarStderr
	}

	err = tarCmd.Run()
	if err != nil {
		errMsg := fmt.Sprintf("failed to pack OCI dir to tar: %s", tarStderr.String())
		return errors.Wrap(err, errors.ErrRuntimeCommand, errMsg)
	}

	// Generate checksum file if requested
	if options.Checksum {
		if _, err := generateChecksum(tarPath); err != nil {
			return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to generate checksum for %s", tarPath))
		}
	}

	return nil
}

// Load loads an image from a tar file using oci-archive or docker-archive format
func (s *SkopeoRuntime) Load(ctx context.Context, tarPath string) error {
	// Skopeo doesn't have a direct "load" command like Docker
	// Instead, we copy from oci-archive/docker-archive to docker-daemon
	// But this requires Docker daemon to be running
	
	// For now, we'll use a workaround: copy to a temp directory first
	// Then use docker load or similar
	
	// Actually, Skopeo can copy from docker-archive to docker-daemon
	// But we need to know the image name, which we don't have here
	// So we'll need to inspect the archive first or use a different approach
	
	// For compatibility, we'll try to use docker load as fallback
	// But first check if we can use skopeo copy docker-archive:file.tar docker-daemon:image:tag
	
	// Since we don't have the image name in Load, we'll need to inspect the archive
	// For now, let's use a simpler approach: copy to docker-daemon with a generic name
	// This is a limitation - we might need to change the interface
	
	// Actually, let's check if docker is available and use it as fallback
	if IsCommandAvailable("docker") {
		cmd := exec.CommandContext(ctx, "docker", "load", "-i", tarPath)
		if err := cmd.Run(); err != nil {
			return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to load image from %s", tarPath))
		}
		return nil
	}
	
	// If docker is not available, we can't load using Skopeo alone
	// Skopeo copy requires knowing the destination image name
	return errors.New(errors.ErrRuntimeCommand, fmt.Sprintf("cannot load image from %s: Skopeo requires Docker daemon or explicit image name", tarPath))
}

// Push pushes an image to a registry
func (s *SkopeoRuntime) Push(ctx context.Context, image string, options PushOptions) error {
	// Skopeo push requires source and destination
	// We need to get the source from the image name
	// For now, we'll assume the image is already in the local Docker daemon
	// and we'll copy from docker-daemon to docker://
	
	args := []string{"copy", "--preserve-digests"}
	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	args = append(args, fmt.Sprintf("docker-daemon:%s", image))
	args = append(args, fmt.Sprintf("docker://%s", image))

	cmd := exec.CommandContext(ctx, s.command, args...)
	
	// Capture output for better error messages
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to push image %s: %s", image, string(output)))
	}

	return nil
}

// RemoveImage removes an image from local storage.
// Skopeo does not store images locally; this operation is not supported.
func (s *SkopeoRuntime) RemoveImage(ctx context.Context, image string, options RmiOptions) error {
	_ = image
	_ = options
	return errors.New(errors.ErrRuntimeCommand, "rmi is not supported for skopeo - Skopeo does not store images locally, use docker, podman, or nerdctl")
}

// ListImages lists images in local storage. Skopeo does not store images locally; returns empty slice.
func (s *SkopeoRuntime) ListImages(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

// Tag tags an image with a new name
func (s *SkopeoRuntime) Tag(ctx context.Context, source, target string) error {
	// Skopeo doesn't have a direct tag command
	// We'll use copy from docker-daemon:source to docker-daemon:target
	args := []string{"copy", "--preserve-digests"}
	args = append(args, fmt.Sprintf("docker-daemon:%s", source))
	args = append(args, fmt.Sprintf("docker-daemon:%s", target))

	cmd := exec.CommandContext(ctx, s.command, args...)
	
	// Capture output for better error messages
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to tag image %s as %s: %s", source, target, string(output)))
	}

	return nil
}

// Login logs in to a container registry
func (s *SkopeoRuntime) Login(ctx context.Context, registry string, options LoginOptions) error {
	args := []string{"login"}

	if options.Username != "" {
		args = append(args, "-u", options.Username)
	}

	if options.PasswordStdin {
		args = append(args, "--password-stdin")
	} else if options.Password != "" {
		args = append(args, "-p", options.Password)
	}

	if options.Insecure {
		args = append(args, "--tls-verify=false")
	}

	args = append(args, registry)

	cmd := exec.CommandContext(ctx, s.command, args...)

	// 如果使用 --password-stdin，从 stdin 读取密码
	if options.PasswordStdin && options.Password != "" {
		cmd.Stdin = strings.NewReader(options.Password)
	}

	// Capture output for debug mode
	var stdout, stderr bytes.Buffer
	if options.Debug {
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		// In non-debug mode, show output but capture errors
		cmd.Stdout = os.Stdout
		cmd.Stderr = &stderr
	}

	err := cmd.Run()
	if err != nil {
		errMsg := fmt.Sprintf("failed to login to registry %s", registry)
		if options.Debug && (stdout.Len() > 0 || stderr.Len() > 0) {
			errMsg += fmt.Sprintf("\nStdout:\n%s\nStderr:\n%s", stdout.String(), stderr.String())
		} else if stderr.Len() > 0 {
			errMsg += fmt.Sprintf("\nError: %s", stderr.String())
		}
		return errors.Wrap(err, errors.ErrRuntimeCommand, errMsg)
	}

	return nil
}

// Version returns the Skopeo version
func (s *SkopeoRuntime) Version() (string, error) {
	cmd := exec.Command(s.command, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, errors.ErrRuntimeCommand, "failed to get Skopeo version")
	}

	// Parse version from output (format: "skopeo version X.Y.Z")
	versionStr := strings.TrimSpace(string(output))
	parts := strings.Fields(versionStr)
	if len(parts) >= 3 {
		return parts[2], nil
	}

	return versionStr, nil
}

// Helper functions

// extractArch extracts architecture from platform string (e.g., "linux/amd64" -> "amd64")
func extractArch(platform string) string {
	parts := strings.Split(platform, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return "amd64" // default
}

// extractOS extracts OS from platform string (e.g., "linux/amd64" -> "linux")
func extractOS(platform string) string {
	parts := strings.Split(platform, "/")
	if len(parts) >= 1 {
		return parts[0]
	}
	return "linux" // default
}

// sanitizeImageName sanitizes image name for use in file paths
func sanitizeImageName(image string) string {
	// Replace special characters with underscores
	replacer := strings.NewReplacer(
		"/", "_",
		":", "_",
		"@", "_",
	)
	return replacer.Replace(image)
}
