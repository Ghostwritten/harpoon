package runtime

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/harpoon/hpn/pkg/errors"
)

// Detector implements RuntimeDetector interface
type Detector struct {
	runtimes map[string]ContainerRuntime
}

// NewDetector creates a new runtime detector
func NewDetector() *Detector {
	return &Detector{
		runtimes: make(map[string]ContainerRuntime),
	}
}

// DetectAvailable detects all available container runtimes
func (d *Detector) DetectAvailable() []ContainerRuntime {
	var available []ContainerRuntime

	// Initialize runtime implementations
	runtimes := []ContainerRuntime{
		NewDockerRuntime(),
		NewPodmanRuntime(),
		NewNerdctlRuntime(),
		NewSkopeoRuntime(),
	}

	// Check availability and store
	for _, runtime := range runtimes {
		if runtime.IsAvailable() {
			available = append(available, runtime)
			d.runtimes[runtime.Name()] = runtime
		}
	}

	// Sort by priority (Docker > Podman > Nerdctl > Skopeo)
	sort.Slice(available, func(i, j int) bool {
		priority := map[string]int{
			"docker":  1,
			"podman":  2,
			"nerdctl": 3,
			"skopeo":  4,
		}
		return priority[available[i].Name()] < priority[available[j].Name()]
	})

	return available
}

// GetPreferred returns the preferred runtime based on priority
func (d *Detector) GetPreferred() ContainerRuntime {
	available := d.DetectAvailable()
	if len(available) == 0 {
		return nil
	}
	return available[0]
}

// GetByName returns a runtime by name
func (d *Detector) GetByName(name string) (ContainerRuntime, error) {
	// Ensure detection has been run
	if len(d.runtimes) == 0 {
		d.DetectAvailable()
	}

	runtime, exists := d.runtimes[name]
	if !exists {
		return nil, errors.NewRuntimeNotFound(name)
	}

	if !runtime.IsAvailable() {
		return nil, errors.New(errors.ErrRuntimeUnavailable, fmt.Sprintf("runtime '%s' is not available", name))
	}

	return runtime, nil
}

// IsCommandAvailable checks if a command is available in PATH
func IsCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// parseListImagesOutput parses docker/podman/nerdctl images output, filters out <none>:<none> and empty lines
func parseListImagesOutput(output string) []string {
	var result []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "<none>") {
			continue
		}
		result = append(result, line)
	}
	return result
}