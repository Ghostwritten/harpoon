package runtime

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestTarFilenameFromImage(t *testing.T) {
	tests := []struct {
		image    string
		expected string
	}{
		{"nginx:latest", "nginx_latest.tar"},
		{"quay.io/foo/bar:v1", "quay.io_foo_bar_v1.tar"},
		{"registry.com/project/image:tag", "registry.com_project_image_tag.tar"},
	}
	for _, tt := range tests {
		got := TarFilenameFromImage(tt.image)
		if got != tt.expected {
			t.Errorf("TarFilenameFromImage(%q) = %q, want %q", tt.image, got, tt.expected)
		}
	}
}

func TestImageNameFromTarFilename(t *testing.T) {
	tests := []struct {
		tarFilename string
		expected    string
	}{
		{"nginx_latest.tar", "nginx:latest"},
		{"quay.io_foo_bar_v1.tar", "quay.io/foo/bar:v1"},
		{"registry.com_project_image_tag.tar", "registry.com/project/image:tag"},
		{"quay.io_prometheus_alertmanager_v0.30.1.tar", "quay.io/prometheus/alertmanager:v0.30.1"},
		{"docker.io_bats_bats_v1.4.1.tar", "docker.io/bats/bats:v1.4.1"},
		{"nginx.tar", "nginx:latest"},
		{"library_nginx_latest.tar", "library/nginx:latest"},
	}
	for _, tt := range tests {
		got := ImageNameFromTarFilename(tt.tarFilename)
		if got != tt.expected {
			t.Errorf("ImageNameFromTarFilename(%q) = %q, want %q", tt.tarFilename, got, tt.expected)
		}
	}
}

func TestImageNameFromTarFilename_RoundTrip(t *testing.T) {
	images := []string{"nginx:latest", "quay.io/foo/bar:v1", "docker.io/library/nginx:1.21"}
	for _, img := range images {
		tar := TarFilenameFromImage(img)
		got := ImageNameFromTarFilename(tar)
		if got != img {
			t.Errorf("round-trip: %q -> %q -> %q", img, tar, got)
		}
	}
}

func TestVerifyTarChecksum_NoChecksumFile_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	tarPath := filepath.Join(dir, "image.tar")
	_ = os.WriteFile(tarPath, []byte("data"), 0644)
	_, err := VerifyTarChecksum(tarPath)
	if err == nil {
		t.Fatal("expected error when checksum file missing")
	}
}

func TestVerifyTarChecksum_ValidChecksum_ReturnsTrue(t *testing.T) {
	dir := t.TempDir()
	tarPath := filepath.Join(dir, "image.tar")
	_ = os.WriteFile(tarPath, []byte("test data"), 0644)
	sum := sha256SumFile(tarPath)
	if sum == "" {
		t.Fatal("sha256SumFile failed")
	}
	_ = os.WriteFile(tarPath+".sha256", []byte(sum), 0644)
	ok, err := VerifyTarChecksum(tarPath)
	if err != nil {
		t.Fatalf("VerifyTarChecksum: %v", err)
	}
	if !ok {
		t.Error("expected checksum to match")
	}
}

func sha256SumFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
