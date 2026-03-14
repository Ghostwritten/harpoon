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

// NerdctlRuntime implements ContainerRuntime for Nerdctl
type NerdctlRuntime struct {
	command string
}

// NewNerdctlRuntime creates a new Nerdctl runtime
func NewNerdctlRuntime() *NerdctlRuntime {
	return &NerdctlRuntime{
		command: "nerdctl",
	}
}

// Name returns the runtime name
func (n *NerdctlRuntime) Name() string {
	return "nerdctl"
}

// IsAvailable checks if Nerdctl is available
func (n *NerdctlRuntime) IsAvailable() bool {
	if !IsCommandAvailable(n.command) {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, n.command, "version")
	return cmd.Run() == nil
}

// Pull pulls an image from a registry.
// Pass --insecure-registry via ExtraArgs when pulling from an HTTP registry.
func (n *NerdctlRuntime) Pull(ctx context.Context, image string, options PullOptions) error {
	args := []string{"pull"}

	if options.Platform != "" && options.Platform != "all" {
		args = append(args, "--platform", options.Platform)
	}

	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	args = append(args, image)

	cmd := exec.CommandContext(ctx, n.command, args...)
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
func (n *NerdctlRuntime) Save(ctx context.Context, image string, tarPath string, options SaveOptions) error {
	args := []string{"save", "-o", tarPath}
	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	args = append(args, image)
	cmd := exec.CommandContext(ctx, n.command, args...)

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
// imageName is accepted for interface compatibility but unused: nerdctl load reads the
// image reference directly from the tar manifest.
func (n *NerdctlRuntime) Load(ctx context.Context, tarPath string, _ string) error {
	cmd := exec.CommandContext(ctx, n.command, "load", "-i", tarPath)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand,
			fmt.Sprintf("failed to load image from %s", tarPath))
	}
	return nil
}

// Push pushes an image to a registry.
// Pass --insecure-registry via ExtraArgs when pushing to an HTTP registry.
func (n *NerdctlRuntime) Push(ctx context.Context, image string, options PushOptions) error {
	args := []string{"push"}
	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	args = append(args, image)
	cmd := exec.CommandContext(ctx, n.command, args...)

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
func (n *NerdctlRuntime) RemoveImage(ctx context.Context, image string, options RmiOptions) error {
	args := []string{"rmi"}
	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	args = append(args, image)
	cmd := exec.CommandContext(ctx, n.command, args...)

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
func (n *NerdctlRuntime) ListImages(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, n.command, "images", "--format", "{{.Repository}}:{{.Tag}}")
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
func (n *NerdctlRuntime) Tag(ctx context.Context, source, target string) error {
	cmd := exec.CommandContext(ctx, n.command, "tag", source, target)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand,
			fmt.Sprintf("failed to tag image %s as %s", source, target))
	}
	return nil
}

// Login logs in to a container registry.
// The password is always delivered via stdin to prevent it appearing in the process table.
func (n *NerdctlRuntime) Login(ctx context.Context, registry string, options LoginOptions) error {
	args := []string{"login"}

	if options.Username != "" {
		args = append(args, "-u", options.Username)
	}

	if options.Insecure {
		// Nerdctl does not support --insecure at login time.
		fmt.Printf("Warning: Nerdctl does not support --insecure flag directly. " +
			"Configure containerd for insecure registries in /etc/containerd/config.toml\n")
	}

	// Always use --password-stdin to keep the credential out of the process table.
	if options.Password != "" || options.PasswordStdin {
		args = append(args, "--password-stdin")
	}

	args = append(args, registry)
	cmd := exec.CommandContext(ctx, n.command, args...)

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
func (n *NerdctlRuntime) Logout(ctx context.Context, registry string) error {
	cmd := exec.CommandContext(ctx, n.command, "logout", registry)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand,
			fmt.Sprintf("failed to logout from registry %s", registry))
	}
	return nil
}

// Version returns the Nerdctl version
func (n *NerdctlRuntime) Version() (string, error) {
	cmd := exec.Command(n.command, "version", "--format", "{{.Client.Version}}")
	output, err := cmd.Output()
	if err != nil {
		// Fallback: plain version output
		cmd = exec.Command(n.command, "version")
		output, err = cmd.Output()
		if err != nil {
			return "", errors.Wrap(err, errors.ErrRuntimeCommand, "failed to get Nerdctl version")
		}
		for _, line := range strings.Split(string(output), "\n") {
			if strings.Contains(line, "Version:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1]), nil
				}
			}
		}
		return "unknown", nil
	}
	return strings.TrimSpace(string(output)), nil
}
