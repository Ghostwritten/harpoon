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

// DockerRuntime implements ContainerRuntime for Docker
type DockerRuntime struct {
	command string
}

// NewDockerRuntime creates a new Docker runtime
func NewDockerRuntime() *DockerRuntime {
	return &DockerRuntime{
		command: "docker",
	}
}

// Name returns the runtime name
func (d *DockerRuntime) Name() string {
	return "docker"
}

// IsAvailable checks if Docker is available
func (d *DockerRuntime) IsAvailable() bool {
	if !IsCommandAvailable(d.command) {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, d.command, "version", "--format", "{{.Server.Version}}")
	return cmd.Run() == nil
}

// Pull pulls an image from a registry
func (d *DockerRuntime) Pull(ctx context.Context, image string, options PullOptions) error {
	args := []string{"pull"}

	if options.Platform != "" && options.Platform != "all" {
		args = append(args, "--platform", options.Platform)
	}

	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	args = append(args, image)

	cmd := exec.CommandContext(ctx, d.command, args...)
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
func (d *DockerRuntime) Save(ctx context.Context, image string, tarPath string, options SaveOptions) error {
	args := []string{"save", "-o", tarPath}
	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	args = append(args, image)
	cmd := exec.CommandContext(ctx, d.command, args...)

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
// imageName is accepted for interface compatibility but unused: docker load reads the
// image reference directly from the tar manifest.
func (d *DockerRuntime) Load(ctx context.Context, tarPath string, _ string) error {
	cmd := exec.CommandContext(ctx, d.command, "load", "-i", tarPath)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand,
			fmt.Sprintf("failed to load image from %s", tarPath))
	}
	return nil
}

// Push pushes an image to a registry
func (d *DockerRuntime) Push(ctx context.Context, image string, options PushOptions) error {
	args := []string{"push"}
	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	args = append(args, image)
	cmd := exec.CommandContext(ctx, d.command, args...)

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
func (d *DockerRuntime) RemoveImage(ctx context.Context, image string, options RmiOptions) error {
	args := []string{"rmi"}
	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	args = append(args, image)
	cmd := exec.CommandContext(ctx, d.command, args...)

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
func (d *DockerRuntime) ListImages(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, d.command, "images", "--format", "{{.Repository}}:{{.Tag}}")
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
func (d *DockerRuntime) Tag(ctx context.Context, source, target string) error {
	cmd := exec.CommandContext(ctx, d.command, "tag", source, target)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand,
			fmt.Sprintf("failed to tag image %s as %s", source, target))
	}
	return nil
}

// Login logs in to a container registry.
// The password is always delivered via stdin to prevent it appearing in the
// process table (which would happen if passed as a -p flag argument).
func (d *DockerRuntime) Login(ctx context.Context, registry string, options LoginOptions) error {
	args := []string{"login"}

	if options.Username != "" {
		args = append(args, "-u", options.Username)
	}

	if options.Insecure {
		// Docker does not support --insecure directly; provide guidance.
		fmt.Printf("Warning: Docker does not support --insecure flag directly. " +
			"For HTTP registries, configure daemon.json insecureRegistries. " +
			"For self-signed certificates, add certs to /etc/docker/certs.d/\n")
	}

	// Always use --password-stdin to keep the credential out of the process table.
	if options.Password != "" || options.PasswordStdin {
		args = append(args, "--password-stdin")
	}

	args = append(args, registry)
	cmd := exec.CommandContext(ctx, d.command, args...)

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
func (d *DockerRuntime) Logout(ctx context.Context, registry string) error {
	cmd := exec.CommandContext(ctx, d.command, "logout", registry)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand,
			fmt.Sprintf("failed to logout from registry %s", registry))
	}
	return nil
}

// Version returns the Docker version
func (d *DockerRuntime) Version() (string, error) {
	cmd := exec.Command(d.command, "version", "--format", "{{.Client.Version}}")
	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, errors.ErrRuntimeCommand, "failed to get Docker version")
	}
	return strings.TrimSpace(string(output)), nil
}

// applyProxyEnv sets HTTP/HTTPS proxy environment variables on cmd when proxy is enabled.
func applyProxyEnv(cmd *exec.Cmd, proxy *ProxyConfig) {
	if proxy == nil || !proxy.Enabled {
		return
	}
	env := os.Environ()
	if proxy.HTTP != "" {
		env = append(env, fmt.Sprintf("http_proxy=%s", proxy.HTTP))
	}
	if proxy.HTTPS != "" {
		env = append(env, fmt.Sprintf("https_proxy=%s", proxy.HTTPS))
	}
	cmd.Env = env
}

// buildErrMsg constructs an error message, appending stderr content when present.
func buildErrMsg(base, stderr string, debug bool) string {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return base
	}
	if debug {
		return fmt.Sprintf("%s\nStderr:\n%s", base, stderr)
	}
	return fmt.Sprintf("%s\nError: %s", base, stderr)
}
