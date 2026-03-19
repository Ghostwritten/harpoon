package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/harpoon/hpn/internal/chart"
	"github.com/spf13/cobra"
)

var (
	extractChart       string
	extractVersion     string
	extractValues      []string
	extractOutput      string
	extractReleaseName string
)

var extractCmd = &cobra.Command{
	Use:     "extract",
	Aliases: []string{"list-images"},
	Short:   "Extract image list from a Helm chart",
	Long: `Extract image references from a Helm chart by running 'helm template' and parsing the rendered manifests.
Chart can be specified as repo/name (with --version) or as a local path (directory or .tgz file).
Output is one image per line, compatible with 'hpn pull -f' and 'hpn push -f'.

Examples:
  hpn extract --chart bitnami/nginx --version 15.0.0 -o images.txt
  hpn extract --chart ./mychart.tgz -f values.yaml
  hpn extract --chart ./mychart -o -   # write to stdout`,
	RunE: runExtract,
}

func init() {
	extractCmd.Flags().StringVar(&extractChart, "chart", "", "Chart: repo/name (e.g. bitnami/nginx) or path to local dir or .tgz")
	extractCmd.Flags().StringVar(&extractVersion, "version", "", "Chart version (required when using repo/name)")
	extractCmd.Flags().StringSliceVarP(&extractValues, "values", "f", nil, "Values files to pass to helm template (can be repeated)")
	extractCmd.Flags().StringVarP(&extractOutput, "output", "o", "", "Output file (default: stdout)")
	extractCmd.Flags().StringVar(&extractReleaseName, "release-name", "release", "Helm release name for template")
}

func runExtract(cmd *cobra.Command, args []string) error {
	chartRef := strings.TrimSpace(extractChart)
	if chartRef == "" {
		return usageErrorf("--chart is required")
	}
	// Version required only for remote charts (repo/name). We rely on chart package to validate.
	opts := chart.GetImagesOptions{
		Chart:        chartRef,
		Version:      extractVersion,
		ReleaseName: extractReleaseName,
		ValuesFiles:  extractValues,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	images, err := chart.GetImages(ctx, opts)
	if err != nil {
		return err
	}
	out := os.Stdout
	if extractOutput != "" && extractOutput != "-" {
		f, err := os.Create(extractOutput)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer f.Close()
		out = f
	}
	for _, img := range images {
		fmt.Fprintln(out, img)
	}
	return nil
}
