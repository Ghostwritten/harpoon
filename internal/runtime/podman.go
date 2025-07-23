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

	// Test if Podman is working
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.command, "version", "--format", "{{.Version}}")
	return cmd.Run() == nil
}

// Pull pulls an image from a registry
func (p *PodmanRuntime) Pull(ctx context.Context, image string, options PullOptions) error {
	args := []string{"pull"}

	// Add platform if specified
	if options.Platform != "" {
		args = append(args, "--platform", options.Platform)
	}

	args = append(args, image)

	cmd := exec.CommandContext(ctx, p.command, args...)

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
func (p *PodmanRuntime) Save(ctx context.Context, image string, tarPath string) error {
	cmd := exec.CommandContext(ctx, p.command, "save", "-o", tarPath, image)
	
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to save image %s to %s", image, tarPath))
	}

	return nil
}

// Load loads an image from a tar file
func (p *PodmanRuntime) Load(ctx context.Context, tarPath string) error {
	cmd := exec.CommandContext(ctx, p.command, "load", "-i", tarPath)
	
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to load image from %s", tarPath))
	}

	return nil
}

// Push pushes an image to a registry
func (p *PodmanRuntime) Push(ctx context.Context, image string, options PushOptions) error {
	cmd := exec.CommandContext(ctx, p.command, "push", image)
	
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to push image %s", image))
	}

	return nil
}

// Tag tags an image with a new name
func (p *PodmanRuntime) Tag(ctx context.Context, source, target string) error {
	cmd := exec.CommandContext(ctx, p.command, "tag", source, target)
	
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to tag image %s as %s", source, target))
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