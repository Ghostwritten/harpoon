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
	listImagesChart       string
	listImagesVersion     string
	listImagesValues      []string
	listImagesOutput      string
	listImagesReleaseName string
)

var listImagesCmd = &cobra.Command{
	Use:   "list-images",
	Short: "Extract container image list from a Helm chart",
	Long: `Extract image references from a Helm chart by running 'helm template' and parsing the rendered manifests.
Chart can be specified as repo/name (with --version) or as a local path (directory or .tgz file).
Output is one image per line, compatible with 'hpn pull -f' and 'hpn push -f'.

Examples:
  hpn list-images --chart bitnami/nginx --version 15.0.0 -o images.txt
  hpn list-images --chart ./mychart.tgz -f values.yaml
  hpn list-images --chart ./mychart -o -   # write to stdout`,
	RunE: runListImages,
}

func init() {
	listImagesCmd.Flags().StringVar(&listImagesChart, "chart", "", "Chart: repo/name (e.g. bitnami/nginx) or path to local dir or .tgz")
	listImagesCmd.Flags().StringVar(&listImagesVersion, "version", "", "Chart version (required when using repo/name)")
	listImagesCmd.Flags().StringSliceVarP(&listImagesValues, "values", "f", nil, "Values files to pass to helm template (can be repeated)")
	listImagesCmd.Flags().StringVarP(&listImagesOutput, "output", "o", "", "Output file (default: stdout)")
	listImagesCmd.Flags().StringVar(&listImagesReleaseName, "release-name", "release", "Helm release name for template")
}

func runListImages(cmd *cobra.Command, args []string) error {
	chartRef := strings.TrimSpace(listImagesChart)
	if chartRef == "" {
		return fmt.Errorf("--chart is required")
	}
	// Version required only for remote charts (repo/name). We rely on chart package to validate.
	opts := chart.GetImagesOptions{
		Chart:        chartRef,
		Version:      listImagesVersion,
		ReleaseName: listImagesReleaseName,
		ValuesFiles:  listImagesValues,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	images, err := chart.GetImages(ctx, opts)
	if err != nil {
		return err
	}
	out := os.Stdout
	if listImagesOutput != "" && listImagesOutput != "-" {
		f, err := os.Create(listImagesOutput)
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
