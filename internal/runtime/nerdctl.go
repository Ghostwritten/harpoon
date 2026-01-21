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
	if options.Platform != "" && options.Platform != "all" {
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

	// Capture output for debug mode
	var stdout, stderr bytes.Buffer
	if options.Debug {
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		// In non-debug mode, still show progress but capture errors
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

// Save saves an image to a tar file
func (n *NerdctlRuntime) Save(ctx context.Context, image string, tarPath string, options SaveOptions) error {
	cmd := exec.CommandContext(ctx, n.command, "save", "-o", tarPath, image)
	
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to save image %s to %s", image, tarPath))
	}

	// Generate checksum file if requested (default: true)
	if options.Checksum {
		if _, err := generateChecksum(tarPath); err != nil {
			return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to generate checksum for %s", tarPath))
		}
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
	
	// Capture output for debug mode
	var stdout, stderr bytes.Buffer
	if options.Debug {
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = &stderr
	}
	
	if err := cmd.Run(); err != nil {
		errMsg := fmt.Sprintf("failed to push image %s", image)
		if options.Debug && (stdout.Len() > 0 || stderr.Len() > 0) {
			errMsg += fmt.Sprintf("\nStdout:\n%s\nStderr:\n%s", stdout.String(), stderr.String())
		} else if stderr.Len() > 0 {
			errMsg += fmt.Sprintf("\nError: %s", stderr.String())
		}
		return errors.Wrap(err, errors.ErrRuntimeCommand, errMsg)
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

// Login logs in to a container registry
func (n *NerdctlRuntime) Login(ctx context.Context, registry string, options LoginOptions) error {
	args := []string{"login"}

	if options.Username != "" {
		args = append(args, "-u", options.Username)
	}

	if options.PasswordStdin {
		args = append(args, "--password-stdin")
	} else if options.Password != "" {
		args = append(args, "-p", options.Password)
	}

	// Nerdctl 不支持 --insecure，但可以通过环境变量配置
	if options.Insecure {
		fmt.Printf("Warning: Nerdctl does not support --insecure flag directly. " +
			"Configure containerd for insecure registries.\n")
	}

	args = append(args, registry)

	cmd := exec.CommandContext(ctx, n.command, args...)

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