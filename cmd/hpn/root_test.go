package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/harpoon/hpn/internal/runtime"
)

func TestReadImageList_ValidFile_ReturnsLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "images.txt")
	content := "nginx:latest\nalpine:3.18\n# comment\n\nquay.io/foo/bar:v1\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	images, err := readImageList(path)
	if err != nil {
		t.Fatalf("readImageList: %v", err)
	}
	if len(images) != 3 {
		t.Fatalf("expected 3 images (skip comment and empty), got %d: %v", len(images), images)
	}
	if images[0] != "nginx:latest" || images[1] != "alpine:3.18" || images[2] != "quay.io/foo/bar:v1" {
		t.Errorf("unexpected list: %v", images)
	}
}

func TestReadImageList_EmptyAndCommentsOnly_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	content := "\n# only comment\n  \n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := readImageList(path)
	if err == nil {
		t.Fatal("expected error when no images in file")
	}
}

func TestReadImageList_FileNotFound_ReturnsError(t *testing.T) {
	_, err := readImageList("/nonexistent/file.txt")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestReadImageList_SingleLine_ReturnsOne(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "one.txt")
	if err := os.WriteFile(path, []byte("busybox:1.36\n"), 0644); err != nil {
		t.Fatal(err)
	}
	images, err := readImageList(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 1 || images[0] != "busybox:1.36" {
		t.Errorf("expected [busybox:1.36], got %v", images)
	}
}

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
		got := runtime.TarFilenameFromImage(tt.image)
		if got != tt.expected {
			t.Errorf("TarFilenameFromImage(%q) = %q, want %q", tt.image, got, tt.expected)
		}
	}
}

func TestVerifyTarChecksum_NoChecksumFile_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	tarPath := filepath.Join(dir, "image.tar")
	if err := os.WriteFile(tarPath, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := runtime.VerifyTarChecksum(tarPath)
	if err == nil {
		t.Fatal("expected error when checksum file missing")
	}
}

func TestVerifyTarChecksum_ValidChecksum_ReturnsTrue(t *testing.T) {
	dir := t.TempDir()
	tarPath := filepath.Join(dir, "image.tar")
	data := []byte("test data")
	if err := os.WriteFile(tarPath, data, 0644); err != nil {
		t.Fatal(err)
	}
	sum := sha256Sum(tarPath)
	if sum == "" {
		t.Fatal("sha256Sum failed")
	}
	checksumPath := tarPath + ".sha256"
	if err := os.WriteFile(checksumPath, []byte(sum), 0644); err != nil {
		t.Fatal(err)
	}
	ok, err := runtime.VerifyTarChecksum(tarPath)
	if err != nil {
		t.Fatalf("VerifyTarChecksum: %v", err)
	}
	if !ok {
		t.Error("expected checksum to match")
	}
}

// sha256Sum returns hex-encoded SHA256 of file for use in tests.
func sha256Sum(path string) string {
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
