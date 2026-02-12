package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	containerruntime "github.com/harpoon/hpn/internal/runtime"
	"github.com/spf13/cobra"
)

var (
	lsPath  string
	lsFile  string
	lsOutput string
)

type lsRow struct {
	image    string
	size     int64
	checksum string
}

type lsCheckRow struct {
	image  string
	status string
}

var lsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List images (from runtime, path, or check file)",
	Long: `List container images. Three modes:
  1. hpn ls                    - List images in local runtime (docker/podman/nerdctl)
  2. hpn ls --path ./images    - List saved tar files in path with SIZE and CHECKSUM
  3. hpn ls -f images.txt      - Check images.txt against runtime (or --path if set)`,
	RunE: executeLs,
}

func init() {
	lsCmd.Flags().StringVar(&lsPath, "path", "", "Directory path containing saved tar files (for list-path or check mode)")
	lsCmd.Flags().StringVarP(&lsFile, "file", "f", "", "Image list file (for check mode: verify each image exists)")
	lsCmd.Flags().StringVarP(&lsOutput, "output", "o", "table", "Output format: table, plain")
	rootCmd.AddCommand(lsCmd)
}

func selectLsRuntime() (containerruntime.ContainerRuntime, error) {
	r, err := selectContainerRuntime()
	if err != nil {
		return nil, err
	}
	if r.Name() == "skopeo" {
		return nil, fmt.Errorf("ls (runtime mode) is not supported for skopeo - Skopeo does not store images locally, use docker, podman, or nerdctl")
	}
	return r, nil
}

func executeLs(cmd *cobra.Command, args []string) error {
	_ = loadConfigIfNeeded()

	path := lsPath
	hasFile := lsFile != ""
	hasPath := path != ""

	if hasFile {
		return executeLsCheck(lsFile, path, hasPath, cmd)
	}
	if hasPath {
		return executeLsPath(path, cmd)
	}
	return executeLsRuntime(cmd)
}

func executeLsRuntime(cmd *cobra.Command) error {
	rt, err := selectLsRuntime()
	if err != nil {
		return fmt.Errorf("container runtime: %w", err)
	}
	fmt.Printf("Listing images from %s\n", rt.Name())

	ctx := context.Background()
	images, err := rt.ListImages(ctx)
	if err != nil {
		return fmt.Errorf("list images: %w", err)
	}
	sort.Strings(images)

	if len(images) == 0 {
		fmt.Println("No images found")
		return nil
	}

	format := strings.ToLower(lsOutput)
	if format == "plain" {
		for _, img := range images {
			fmt.Println(img)
		}
		return nil
	}

	fmt.Printf("\n%-50s\n", "IMAGE")
	fmt.Println(strings.Repeat("-", 50))
	for _, img := range images {
		fmt.Println(img)
	}
	fmt.Printf("\n%d images\n", len(images))
	return nil
}

func executeLsPath(path string, cmd *cobra.Command) error {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return fmt.Errorf("path %s: %w", path, err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("read directory %s: %w", path, err)
	}

	var tarFiles []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".tar") && !strings.HasSuffix(name, ".tar.sha256") {
			tarFiles = append(tarFiles, name)
		}
	}

	if len(tarFiles) == 0 {
		fmt.Println("No images found in path")
		return nil
	}

	var rows []lsRow
	for _, tarFile := range tarFiles {
		tarPath := filepath.Join(path, tarFile)
		image := containerruntime.ImageNameFromTarFilename(tarFile)
		info, err := os.Stat(tarPath)
		if err != nil {
			continue
		}
		checksumPath := tarPath + ".sha256"
		var cs string
		if _, err := os.Stat(checksumPath); os.IsNotExist(err) {
			cs = "missing"
		} else {
			valid, err := containerruntime.VerifyTarChecksum(tarPath)
			if err != nil {
				cs = "missing"
			} else if valid {
				cs = "ok"
			} else {
				cs = "mismatch"
			}
		}
		rows = append(rows, lsRow{image: image, size: info.Size(), checksum: cs})
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].image < rows[j].image })

	format := strings.ToLower(lsOutput)
	if format == "plain" {
		for _, r := range rows {
			fmt.Println(r.image)
		}
		return nil
	}

	formatTable(rows)
	fmt.Printf("\n%d images\n", len(rows))
	return nil
}

func executeLsCheck(filePath string, path string, checkPath bool, cmd *cobra.Command) error {
	images, err := readImageList(filePath)
	if err != nil {
		return fmt.Errorf("read image list: %w", err)
	}

	var rows []lsCheckRow
	if checkPath {
		fi, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("path does not exist: %s", path)
			}
			return fmt.Errorf("path %s: %w", path, err)
		}
		if !fi.IsDir() {
			return fmt.Errorf("path is not a directory: %s", path)
		}
		fmt.Printf("Checking %d images from %s against path %s\n", len(images), filePath, path)
		for _, img := range images {
			tarName := containerruntime.TarFilenameFromImage(img)
			tarPath := filepath.Join(path, tarName)
			status := "missing"
			if _, err := os.Stat(tarPath); err == nil {
				status = "ok"
			}
			rows = append(rows, lsCheckRow{image: img, status: status})
		}
	} else {
		rt, err := selectLsRuntime()
		if err != nil {
			return fmt.Errorf("container runtime: %w", err)
		}
		fmt.Printf("Checking %d images from %s against %s\n", len(images), filePath, rt.Name())
		ctx := context.Background()
		localImages, err := rt.ListImages(ctx)
		if err != nil {
			return fmt.Errorf("list images: %w", err)
		}
		localSet := make(map[string]bool)
		for _, img := range localImages {
			localSet[img] = true
			if !strings.HasPrefix(img, "docker.io/") {
				localSet["docker.io/"+img] = true
			} else {
				localSet[strings.TrimPrefix(img, "docker.io/")] = true
			}
		}
		for _, img := range images {
			status := "missing"
			if localSet[img] {
				status = "ok"
			} else if strings.HasPrefix(img, "docker.io/") && localSet[strings.TrimPrefix(img, "docker.io/")] {
				status = "ok"
			} else if !strings.Contains(img, "/") && localSet["library/"+img] {
				status = "ok"
			}
			rows = append(rows, lsCheckRow{image: img, status: status})
		}
	}

	format := strings.ToLower(lsOutput)
	if format == "plain" {
		for _, r := range rows {
			fmt.Printf("%s\t%s\n", r.image, r.status)
		}
		return nil
	}

	formatCheckTable(rows)
	okCount := 0
	for _, r := range rows {
		if r.status == "ok" {
			okCount++
		}
	}
	fmt.Printf("\n%d ok, %d missing\n", okCount, len(rows)-okCount)
	return nil
}

func formatTable(rows []lsRow) {
	imageColWidth := 45
	for _, r := range rows {
		if len(r.image) > imageColWidth {
			imageColWidth = len(r.image)
		}
	}
	if imageColWidth > 80 {
		imageColWidth = 80
	}

	fmt.Printf("\n%-*s  %10s  %s\n", imageColWidth, "IMAGE", "SIZE", "CHECKSUM")
	fmt.Println(strings.Repeat("-", imageColWidth+10+10+2))

	for _, r := range rows {
		sizeStr := formatSize(r.size)
		fmt.Printf("%-*s  %10s  %s\n", imageColWidth, r.image, sizeStr, r.checksum)
	}
}

func formatCheckTable(rows []lsCheckRow) {
	imageColWidth := 50
	for _, r := range rows {
		if len(r.image) > imageColWidth {
			imageColWidth = len(r.image)
		}
	}
	if imageColWidth > 80 {
		imageColWidth = 80
	}

	fmt.Printf("\n%-*s  %s\n", imageColWidth, "IMAGE", "STATUS")
	fmt.Println(strings.Repeat("-", imageColWidth+10))

	for _, r := range rows {
		fmt.Printf("%-*s  %s\n", imageColWidth, r.image, r.status)
	}
}

func formatSize(n int64) string {
	const (
		KiB = 1024
		MiB = 1024 * KiB
		GiB = 1024 * MiB
	)
	switch {
	case n >= GiB:
		return fmt.Sprintf("%.1f GiB", float64(n)/GiB)
	case n >= MiB:
		return fmt.Sprintf("%.1f MiB", float64(n)/MiB)
	case n >= KiB:
		return fmt.Sprintf("%.1f KiB", float64(n)/KiB)
	default:
		return fmt.Sprintf("%d B", n)
	}
}
