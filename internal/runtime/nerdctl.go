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

	// Test if Nerdctl is working
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, n.command, "version")
	return cmd.Run() == nil
}

// Pull pulls an image from a registry
func (n *NerdctlRuntime) Pull(ctx context.Context, image string, options PullOptions) error {
	args := []string{"pull"}

	// Add insecure registry flag for private registries
	args = append(args, "--insecure-registry")

	// Add platform if specified
	if options.Platform != "" {
		args = append(args, "--platform", options.Platform)
	}

	args = append(args, image)

	cmd := exec.CommandContext(ctx, n.command, args...)

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
func (n *NerdctlRuntime) Save(ctx context.Context, image string, tarPath string) error {
	cmd := exec.CommandContext(ctx, n.command, "save", "-o", tarPath, image)
	
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to save image %s to %s", image, tarPath))
	}

	return nil
}

// Load loads an image from a tar file
func (n *NerdctlRuntime) Load(ctx context.Context, tarPath string) error {
	cmd := exec.CommandContext(ctx, n.command, "load", "-i", tarPath)
	
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to load image from %s", tarPath))
	}

	return nil
}

// Push pushes an image to a registry
func (n *NerdctlRuntime) Push(ctx context.Context, image string, options PushOptions) error {
	args := []string{"push"}
	
	// Add insecure registry flag for private registries
	args = append(args, "--insecure-registry")
	args = append(args, image)

	cmd := exec.CommandContext(ctx, n.command, args...)
	
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to push image %s", image))
	}

	return nil
}

// Tag tags an image with a new name
func (n *NerdctlRuntime) Tag(ctx context.Context, source, target string) error {
	cmd := exec.CommandContext(ctx, n.command, "tag", source, target)
	
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to tag image %s as %s", source, target))
	}

	return nil
}

// Version returns the Nerdctl version
func (n *NerdctlRuntime) Version() (string, error) {
	cmd := exec.Command(n.command, "version", "--format", "{{.Client.Version}}")
	output, err := cmd.Output()
	if err != nil {
		// Try alternative format
		cmd = exec.Command(n.command, "version")
		output, err = cmd.Output()
		if err != nil {
			return "", errors.Wrap(err, errors.ErrRuntimeCommand, "failed to get Nerdctl version")
		}
		
		// Parse version from output
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Version:") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					return strings.TrimSpace(parts[1]), nil
				}
			}
		}
		return "unknown", nil
	}

	return strings.TrimSpace(string(output)), nil
}