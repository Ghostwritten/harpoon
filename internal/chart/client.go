package chart

import (
	"context"
	"fmt"
)

// GetImagesOptions holds options for getting image list from a chart.
type GetImagesOptions struct {
	Chart        string
	Version      string
	ReleaseName  string
	ValuesFiles  []string
}

// GetImages fetches the chart (if remote), runs helm template, extracts image references, and returns a deduplicated list.
func GetImages(ctx context.Context, opts GetImagesOptions) ([]string, error) {
	fetchResult, err := FetchChart(ctx, Options{Chart: opts.Chart, Version: opts.Version})
	if err != nil {
		return nil, fmt.Errorf("fetch chart: %w", err)
	}
	defer fetchResult.Cleanup()

	yamlOut, err := RunTemplate(ctx, fetchResult.Dir, opts.ReleaseName, opts.ValuesFiles)
	if err != nil {
		return nil, fmt.Errorf("helm template: %w", err)
	}

	images, err := ExtractImagesFromYAML(yamlOut)
	if err != nil {
		return nil, fmt.Errorf("extract images: %w", err)
	}
	return images, nil
}
