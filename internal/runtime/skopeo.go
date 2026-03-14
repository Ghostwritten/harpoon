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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.command, "--version")
	return cmd.Run() == nil
}

// Pull pulls an image from a registry.
// Skopeo has no local image store, so this copies the image into the first
// available daemon (Docker or Podman). If neither is available, an error is
// returned — use `hpn save` instead, which combines pull and save in one step.
func (s *SkopeoRuntime) Pull(ctx context.Context, image string, options PullOptions) error {
	// Build source transport string
	src := fmt.Sprintf("docker://%s", image)

	// Determine platform override args
	var platformArgs []string
	if options.MultiArch || options.Platform == "all" {
		platformArgs = append(platformArgs, "--all")
	} else if options.Platform != "" {
		platformArgs = append(platformArgs,
			"--override-arch", extractArch(options.Platform),
			"--override-os", extractOS(options.Platform))
	}

	// Try Docker daemon first
	if IsCommandAvailable("docker") {
		dst := fmt.Sprintf("docker-daemon:%s", image)
		if err := s.runCopy(ctx, src, dst, platformArgs, options.ExtraArgs, options.Proxy, options.Debug); err == nil {
			return nil
		}
	}

	// Try Podman storage
	if IsCommandAvailable("podman") {
		dst := fmt.Sprintf("containers-storage:%s", image)
		if err := s.runCopy(ctx, src, dst, platformArgs, options.ExtraArgs, options.Proxy, options.Debug); err == nil {
			return nil
		}
	}

	return errors.New(errors.ErrRuntimeCommand,
		"skopeo pull requires docker or podman: Skopeo has no local image store. "+
			"Use 'hpn save' to pull and archive in one step.")
}

// Save saves an image to a tar file using docker-archive format.
// Falls back to OCI layout if the registry returns a manifest type that
// docker-archive does not support.
func (s *SkopeoRuntime) Save(ctx context.Context, image string, tarPath string, options SaveOptions) error {
	err := s.saveWithDockerArchive(ctx, image, tarPath, options)
	if err == nil {
		return nil
	}

	// Fallback when the registry uses an OCI or schema-1 manifest
	errStr := err.Error()
	if strings.Contains(errStr, "Unsupported manifest type") ||
		strings.Contains(errStr, "need a Docker schema 2 manifest") {
		return s.saveWithOCIDir(ctx, image, tarPath, options)
	}
	return err
}

// saveWithDockerArchive saves the image in docker-archive format (most compatible)
func (s *SkopeoRuntime) saveWithDockerArchive(ctx context.Context, image string, tarPath string, options SaveOptions) error {
	var platformArgs []string
	if options.MultiArch {
		platformArgs = append(platformArgs, "--all")
	} else {
		targetPlatform := options.Platform
		if targetPlatform == "" {
			targetPlatform = "linux/amd64"
		}
		if targetPlatform != "all" {
			platformArgs = append(platformArgs,
				"--override-arch", extractArch(targetPlatform),
				"--override-os", extractOS(targetPlatform))
		}
	}

	src := fmt.Sprintf("docker://%s", image)
	dst := fmt.Sprintf("docker-archive:%s", tarPath)
	if err := s.runCopy(ctx, src, dst, platformArgs, options.ExtraArgs, nil, options.Debug); err != nil {
		return err
	}

	if options.Checksum {
		if _, err := generateChecksum(tarPath); err != nil {
			return errors.Wrap(err, errors.ErrRuntimeCommand,
				fmt.Sprintf("failed to generate checksum for %s", tarPath))
		}
	}
	return nil
}

// saveWithOCIDir saves the image as an OCI layout directory then tars it.
func (s *SkopeoRuntime) saveWithOCIDir(ctx context.Context, image string, tarPath string, options SaveOptions) error {
	// Use os.MkdirTemp to avoid predictable paths (TOCTOU / symlink attacks)
	tempDir, err := os.MkdirTemp("", "hpn-skopeo-oci-*")
	if err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand, "failed to create temp dir for OCI layout")
	}
	defer os.RemoveAll(tempDir)

	var platformArgs []string
	if options.MultiArch {
		platformArgs = append(platformArgs, "--all")
	} else {
		targetPlatform := options.Platform
		if targetPlatform == "" {
			targetPlatform = "linux/amd64"
		}
		if targetPlatform != "all" {
			platformArgs = append(platformArgs,
				"--override-arch", extractArch(targetPlatform),
				"--override-os", extractOS(targetPlatform))
		}
	}

	src := fmt.Sprintf("docker://%s", image)
	dst := fmt.Sprintf("oci:%s", tempDir)
	if err := s.runCopy(ctx, src, dst, platformArgs, options.ExtraArgs, nil, options.Debug); err != nil {
		return err
	}

	fmt.Printf("    📦 Packing OCI layout to tar file...\n")
	tarCmd := exec.CommandContext(ctx, "tar", "-czf", tarPath, "-C", tempDir, ".")

	var tarStderr bytes.Buffer
	if options.Debug {
		tarCmd.Stderr = io.MultiWriter(os.Stderr, &tarStderr)
	} else {
		tarCmd.Stderr = &tarStderr
	}

	if err := tarCmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand,
			fmt.Sprintf("failed to pack OCI dir to tar: %s", strings.TrimSpace(tarStderr.String())))
	}

	if options.Checksum {
		if _, err := generateChecksum(tarPath); err != nil {
			return errors.Wrap(err, errors.ErrRuntimeCommand,
				fmt.Sprintf("failed to generate checksum for %s", tarPath))
		}
	}
	return nil
}

// Load loads an image from a tar file using skopeo copy.
// imageName is used as the destination reference when copying to a daemon.
// Falls back to docker load / podman load when skopeo copy is not applicable.
func (s *SkopeoRuntime) Load(ctx context.Context, tarPath string, imageName string) error {
	src := fmt.Sprintf("docker-archive:%s", tarPath)

	// Try skopeo copy to Docker daemon
	if IsCommandAvailable("docker") {
		dst := fmt.Sprintf("docker-daemon:%s", imageName)
		args := []string{"copy", "--preserve-digests", src, dst}
		cmd := exec.CommandContext(ctx, s.command, args...)
		if err := cmd.Run(); err == nil {
			return nil
		}
		// Fallback: plain docker load (works even without imageName)
		loadCmd := exec.CommandContext(ctx, "docker", "load", "-i", tarPath)
		if err := loadCmd.Run(); err == nil {
			return nil
		}
	}

	// Try skopeo copy to Podman storage
	if IsCommandAvailable("podman") {
		dst := fmt.Sprintf("containers-storage:%s", imageName)
		args := []string{"copy", "--preserve-digests", src, dst}
		cmd := exec.CommandContext(ctx, s.command, args...)
		if err := cmd.Run(); err == nil {
			return nil
		}
		loadCmd := exec.CommandContext(ctx, "podman", "load", "-i", tarPath)
		if err := loadCmd.Run(); err == nil {
			return nil
		}
	}

	return errors.New(errors.ErrRuntimeCommand,
		fmt.Sprintf("cannot load image from %s: Skopeo requires docker or podman daemon", tarPath))
}

// Push pushes an image to a registry from the local Docker daemon.
func (s *SkopeoRuntime) Push(ctx context.Context, image string, options PushOptions) error {
	args := []string{"copy", "--preserve-digests"}
	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	args = append(args, fmt.Sprintf("docker-daemon:%s", image))
	args = append(args, fmt.Sprintf("docker://%s", image))

	cmd := exec.CommandContext(ctx, s.command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand,
			fmt.Sprintf("failed to push image %s: %s", image, strings.TrimSpace(string(output))))
	}
	return nil
}

// RemoveImage is not supported by Skopeo, which has no local image store.
func (s *SkopeoRuntime) RemoveImage(_ context.Context, _ string, _ RmiOptions) error {
	return errors.New(errors.ErrRuntimeCommand,
		"rmi is not supported for skopeo — Skopeo has no local image store; use docker, podman, or nerdctl")
}

// ListImages returns an empty slice. Skopeo has no local image store.
func (s *SkopeoRuntime) ListImages(_ context.Context) ([]string, error) {
	return []string{}, nil
}

// Tag copies the image from one docker-daemon reference to another.
func (s *SkopeoRuntime) Tag(ctx context.Context, source, target string) error {
	args := []string{"copy", "--preserve-digests",
		fmt.Sprintf("docker-daemon:%s", source),
		fmt.Sprintf("docker-daemon:%s", target),
	}
	cmd := exec.CommandContext(ctx, s.command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand,
			fmt.Sprintf("failed to tag image %s as %s: %s", source, target, strings.TrimSpace(string(output))))
	}
	return nil
}

// Login logs in to a container registry.
// The password is always delivered via stdin to prevent it appearing in the process table.
func (s *SkopeoRuntime) Login(ctx context.Context, registry string, options LoginOptions) error {
	args := []string{"login"}

	if options.Username != "" {
		args = append(args, "-u", options.Username)
	}

	if options.Insecure {
		args = append(args, "--tls-verify=false")
	}

	// Always use --password-stdin to keep the credential out of the process table.
	if options.Password != "" || options.PasswordStdin {
		args = append(args, "--password-stdin")
	}

	args = append(args, registry)
	cmd := exec.CommandContext(ctx, s.command, args...)

	if options.Password != "" {
		cmd.Stdin = strings.NewReader(options.Password)
	}

	var stderr bytes.Buffer
	if options.Debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = &stderr
	}

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand,
			buildErrMsg(fmt.Sprintf("failed to login to registry %s", registry), stderr.String(), options.Debug))
	}
	return nil
}

// Logout logs out from a container registry
func (s *SkopeoRuntime) Logout(ctx context.Context, registry string) error {
	cmd := exec.CommandContext(ctx, s.command, "logout", registry)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand,
			fmt.Sprintf("failed to logout from registry %s", registry))
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
	versionStr := strings.TrimSpace(string(output))
	parts := strings.Fields(versionStr)
	if len(parts) >= 3 {
		return parts[2], nil
	}
	return versionStr, nil
}

// runCopy executes a skopeo copy command from src to dst.
func (s *SkopeoRuntime) runCopy(ctx context.Context, src, dst string, platformArgs, extraArgs []string, proxy *ProxyConfig, debug bool) error {
	args := []string{"copy", "--preserve-digests"}
	args = append(args, platformArgs...)
	if len(extraArgs) > 0 {
		args = append(args, extraArgs...)
	}
	args = append(args, src, dst)

	cmd := exec.CommandContext(ctx, s.command, args...)
	applyProxyEnv(cmd, proxy)

	var stderr bytes.Buffer
	if debug {
		cmd.Stdout = io.MultiWriter(os.Stdout, io.Discard)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = &stderr
	}

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand,
			buildErrMsg(fmt.Sprintf("skopeo copy %s → %s failed", src, dst), stderr.String(), debug))
	}
	return nil
}

// extractArch extracts architecture from platform string (e.g. "linux/amd64" -> "amd64")
func extractArch(platform string) string {
	parts := strings.SplitN(platform, "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return "amd64"
}

// extractOS extracts OS from platform string (e.g. "linux/amd64" -> "linux")
func extractOS(platform string) string {
	parts := strings.SplitN(platform, "/", 2)
	if len(parts) >= 1 && parts[0] != "" {
		return parts[0]
	}
	return "linux"
}

// sanitizeImageName sanitizes image name for use in file paths (kept for any external callers)
func sanitizeImageName(image string) string {
	replacer := strings.NewReplacer("/", "_", ":", "_", "@", "_")
	return replacer.Replace(image)
}
