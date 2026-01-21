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
	if options.Platform != "" && options.Platform != "all" {
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
func (p *PodmanRuntime) Save(ctx context.Context, image string, tarPath string, options SaveOptions) error {
	cmd := exec.CommandContext(ctx, p.command, "save", "-o", tarPath, image)
	
	// Capture output for debug mode
	var stdout, stderr bytes.Buffer
	if options.Debug {
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		cmd.Stderr = &stderr
	}
	
	if err := cmd.Run(); err != nil {
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
func (p *PodmanRuntime) Tag(ctx context.Context, source, target string) error {
	cmd := exec.CommandContext(ctx, p.command, "tag", source, target)
	
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrRuntimeCommand, fmt.Sprintf("failed to tag image %s as %s", source, target))
	}

	return nil
}

// Login logs in to a container registry
func (p *PodmanRuntime) Login(ctx context.Context, registry string, options LoginOptions) error {
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

	cmd := exec.CommandContext(ctx, p.command, args...)

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

// Version returns the Podman version
func (p *PodmanRuntime) Version() (string, error) {
	cmd := exec.Command(p.command, "version", "--format", "{{.Version}}")
	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, errors.ErrRuntimeCommand, "failed to get Podman version")
	}

	return strings.TrimSpace(string(output)), nil
}