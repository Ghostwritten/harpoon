package chart

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFetchChart_LocalDir(t *testing.T) {
	dir := t.TempDir()
	// Minimal chart: Chart.yaml + templates/deployment.yaml with one image
	if err := os.MkdirAll(filepath.Join(dir, "templates"), 0755); err != nil {
		t.Fatal(err)
	}
	chartYaml := `apiVersion: v2
name: testchart
version: 0.1.0
`
	if err := os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte(chartYaml), 0644); err != nil {
		t.Fatal(err)
	}
	tpl := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test
spec:
  template:
    spec:
      containers:
      - name: app
        image: nginx:1.21
`
	if err := os.WriteFile(filepath.Join(dir, "templates", "deployment.yaml"), []byte(tpl), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	result, err := FetchChart(ctx, Options{Chart: dir})
	if err != nil {
		t.Fatalf("FetchChart: %v", err)
	}
	defer result.Cleanup()
	if result.Dir != dir {
		t.Errorf("expected Dir %q, got %q", dir, result.Dir)
	}
}

func TestGetImages_LocalChart_RequiresHelm(t *testing.T) {
	if _, err := exec.LookPath("helm"); err != nil {
		t.Skip("helm not in PATH, skipping integration test")
	}
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "templates"), 0755); err != nil {
		t.Fatal(err)
	}
	chartYaml := `apiVersion: v2
name: testchart
version: 0.1.0
`
	if err := os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte(chartYaml), 0644); err != nil {
		t.Fatal(err)
	}
	tpl := `apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  containers:
  - name: nginx
    image: nginx:latest
`
	if err := os.WriteFile(filepath.Join(dir, "templates", "pod.yaml"), []byte(tpl), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	images, err := GetImages(ctx, GetImagesOptions{
		Chart:       dir,
		ReleaseName: "test",
	})
	if err != nil {
		t.Fatalf("GetImages: %v", err)
	}
	if len(images) != 1 || images[0] != "nginx:latest" {
		t.Errorf("expected [nginx:latest], got %v", images)
	}
}
