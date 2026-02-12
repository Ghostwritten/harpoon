package chart

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// TemplateOptions holds options for helm template.
type TemplateOptions struct {
	ReleaseName string
	ValuesFiles []string
}

// RunTemplate runs helm template on the chart directory and returns the rendered YAML.
func RunTemplate(ctx context.Context, chartDir, releaseName string, valuesFiles []string) ([]byte, error) {
	if err := ensureHelm(ctx); err != nil {
		return nil, err
	}
	if releaseName == "" {
		releaseName = "release"
	}
	args := []string{"template", releaseName, chartDir}
	for _, f := range valuesFiles {
		if f != "" {
			args = append(args, "-f", f)
		}
	}
	cmd := exec.CommandContext(ctx, "helm", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("helm template: %w\nstderr: %s", err, stderr.String())
	}
	return stdout.Bytes(), nil
}
