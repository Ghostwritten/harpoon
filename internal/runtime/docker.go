package runtime

import (
	"context"
	"fmt"
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

	// Test if Docker daemon is running
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, d.command, "version", "--format", "{{.Server.Version}}")
	return cmd.Run() == nil
}

// Pull pulls an image from a registry
func (d *DockerRuntime) Pull(ctx context.Context, image string, options PullOptions) error {
	args := []string{"pull"}

	// Add platform if specified
	if options.Platform != "" {
		args = append(args, "--platform", options.Platform)
	}

	args = append(args, image)

	cmd := exec.CommandContext(ctx, d.command, args...)

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

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to pull image %s", image))
	}

	return nil
}

// Save saves an image to a tar file
func (d *DockerRuntime) Save(ctx context.Context, image string, tarPath string) error {
	cmd := exec.CommandContext(ctx, d.command, "save", "-o", tarPath, image)
	
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to save image %s to %s", image, tarPath))
	}

	return nil
}

// Load loads an image from a tar file
func (d *DockerRuntime) Load(ctx context.Context, tarPath string) error {
	cmd := exec.CommandContext(ctx, d.command, "load", "-i", tarPath)
	
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to load image from %s", tarPath))
	}

	return nil
}

// Push pushes an image to a registry
func (d *DockerRuntime) Push(ctx context.Context, image string, options PushOptions) error {
	cmd := exec.CommandContext(ctx, d.command, "push", image)
	
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to push image %s", image))
	}

	return nil
}

// Tag tags an image with a new name
func (d *DockerRuntime) Tag(ctx context.Context, source, target string) error {
	cmd := exec.CommandContext(ctx, d.command, "tag", source, target)
	
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to tag image %s as %s", source, target))
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