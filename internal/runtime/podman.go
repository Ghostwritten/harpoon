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

// PodmanRuntime implements ContainerRuntime for Podman
type PodmanRuntime struct {
	command string
}

// NewPodmanRuntime creates a new Podman runtime
func NewPodmanRuntime() *PodmanRuntime {
	return &PodmanRuntime{
		command: "podman",
	}
}

// Name returns the runtime name
func (p *PodmanRuntime) Name() string {
	return "podman"
}

// IsAvailable checks if Podman is available
func (p *PodmanRuntime) IsAvailable() bool {
	if !IsCommandAvailable(p.command) {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.command, "version", "--format", "{{.Version}}")
	return cmd.Run() == nil
}

// Pull pulls an image from a registry
func (p *PodmanRuntime) Pull(ctx context.Context, image string, options PullOptions) error {
	args := []string{"pull"}

	if options.Platform != "" && options.Platform != "all" {
		args = append(args, "--platform", options.Platform)
	}

	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	args = append(args, image)

	cmd := exec.CommandContext(ctx, p.command, args...)
	applyProxyEnv(cmd, options.Proxy)

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
			buildErrMsg(fmt.Sprintf("failed to pull image %s", image), stderr.String(), options.Debug))
	}
	return nil
}

// Save saves an image to a tar file
func (p *PodmanRuntime) Save(ctx context.Context, image string, tarPath string, options SaveOptions) error {
	args := []string{"save", "-o", tarPath}
	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	args = append(args, image)
	cmd := exec.CommandContext(ctx, p.command, args...)

	var stderr bytes.Buffer
	if options.Debug {
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		cmd.Stderr = &stderr
	}

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand,
			buildErrMsg(fmt.Sprintf("failed to save image %s to %s", image, tarPath), stderr.String(), options.Debug))
	}

	if options.Checksum {
		if _, err := generateChecksum(tarPath); err != nil {
			return errors.Wrap(err, errors.ErrRuntimeCommand,
				fmt.Sprintf("failed to generate checksum for %s", tarPath))
		}
	}
	return nil
}

// Load loads an image from a tar file.
// imageName is accepted for interface compatibility but unused: podman load reads the
// image reference directly from the tar manifest.
func (p *PodmanRuntime) Load(ctx context.Context, tarPath string, _ string) error {
	cmd := exec.CommandContext(ctx, p.command, "load", "-i", tarPath)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand,
			fmt.Sprintf("failed to load image from %s", tarPath))
	}
	return nil
}

// Push pushes an image to a registry
func (p *PodmanRuntime) Push(ctx context.Context, image string, options PushOptions) error {
	args := []string{"push"}
	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	args = append(args, image)
	cmd := exec.CommandContext(ctx, p.command, args...)

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
			buildErrMsg(fmt.Sprintf("failed to push image %s", image), stderr.String(), options.Debug))
	}
	return nil
}

// RemoveImage removes an image from local storage
func (p *PodmanRuntime) RemoveImage(ctx context.Context, image string, options RmiOptions) error {
	args := []string{"rmi"}
	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	args = append(args, image)
	cmd := exec.CommandContext(ctx, p.command, args...)

	var stderr bytes.Buffer
	if options.Debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		cmd.Stderr = &stderr
	}

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand,
			buildErrMsg(fmt.Sprintf("failed to remove image %s", image), stderr.String(), options.Debug))
	}
	return nil
}

// ListImages lists images in local storage
func (p *PodmanRuntime) ListImages(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, p.command, "images", "--format", "{{.Repository}}:{{.Tag}}")
	output, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && len(ee.Stderr) > 0 {
			return nil, errors.Wrap(err, errors.ErrRuntimeCommand,
				fmt.Sprintf("failed to list images: %s", string(ee.Stderr)))
		}
		return nil, errors.Wrap(err, errors.ErrRuntimeCommand, "failed to list images")
	}
	return parseListImagesOutput(string(output)), nil
}

// Tag tags an image with a new name
func (p *PodmanRuntime) Tag(ctx context.Context, source, target string) error {
	cmd := exec.CommandContext(ctx, p.command, "tag", source, target)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand,
			fmt.Sprintf("failed to tag image %s as %s", source, target))
	}
	return nil
}

// Login logs in to a container registry.
// The password is always delivered via stdin to prevent it appearing in the process table.
func (p *PodmanRuntime) Login(ctx context.Context, registry string, options LoginOptions) error {
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
	cmd := exec.CommandContext(ctx, p.command, args...)

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
func (p *PodmanRuntime) Logout(ctx context.Context, registry string) error {
	cmd := exec.CommandContext(ctx, p.command, "logout", registry)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand,
			fmt.Sprintf("failed to logout from registry %s", registry))
	}
	return nil
}

// Version returns the Podman version
func (p *PodmanRuntime) Version() (string, error) {
	cmd := exec.Command(p.command, "version", "--format", "{{.Version}}")
	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, errors.ErrRuntimeCommand, "failed to get Podman version")
	}
	return strings.TrimSpace(string(output)), nil
}
