package benchmarks

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// BenchmarkImageListReading tests image list file reading performance
func BenchmarkImageListReading(b *testing.B) {
	testCases := []struct {
		name  string
		count int
	}{
		{"Small", 10},
		{"Medium", 100},
		{"Large", 1000},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Create temporary image list file
			tmpFile, err := ioutil.TempFile("", "images-*.txt")
			if err != nil {
				b.Fatal(err)
			}
			defer os.Remove(tmpFile.Name())

			// Write test images
			for i := 0; i < tc.count; i++ {
				fmt.Fprintf(tmpFile, "nginx:v%d.0\n", i)
			}
			tmpFile.Close()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := readImageList(tmpFile.Name())
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkTarFileDiscovery tests tar file discovery performance
func BenchmarkTarFileDiscovery(b *testing.B) {
	// Create temporary directory with tar files
	tmpDir, err := ioutil.TempDir("", "tar-test-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test tar files
	for i := 0; i < 50; i++ {
		tarFile := filepath.Join(tmpDir, fmt.Sprintf("image%d.tar", i))
		if err := ioutil.WriteFile(tarFile, []byte("dummy"), 0644); err != nil {
			b.Fatal(err)
		}
	}

	b.Run("NonRecursive", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := findTarFiles(tmpDir, false)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Create subdirectories for recursive test
	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0755)
	for i := 0; i < 25; i++ {
		tarFile := filepath.Join(subDir, fmt.Sprintf("subimage%d.tar", i))
		if err := ioutil.WriteFile(tarFile, []byte("dummy"), 0644); err != nil {
			b.Fatal(err)
		}
	}

	b.Run("Recursive", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := findTarFiles(tmpDir, true)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkDirectoryCreation tests directory creation performance
func BenchmarkDirectoryCreation(b *testing.B) {
	tmpDir, err := ioutil.TempDir("", "dir-test-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dirPath := filepath.Join(tmpDir, fmt.Sprintf("test%d", i))
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFilePathOperations tests file path manipulation performance
func BenchmarkFilePathOperations(b *testing.B) {
	testPaths := []string{
		"./images/nginx_latest.tar",
		"./images/redis/redis_7.0.tar",
		"/tmp/harbor.example.com_production_app_v1.0.0.tar",
		"registry.k8s.io_coredns_coredns_v1.11.1.tar",
	}

	b.Run("Join", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, path := range testPaths {
				_ = filepath.Join("./images", path)
			}
		}
	})

	b.Run("Dir", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, path := range testPaths {
				_ = filepath.Dir(path)
			}
		}
	})

	b.Run("Base", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, path := range testPaths {
				_ = filepath.Base(path)
			}
		}
	})
}

// BenchmarkFileTarFilenameGeneration tests tar filename generation performance
func BenchmarkFileTarFilenameGeneration(b *testing.B) {
	images := []string{
		"nginx:latest",
		"redis:7.0",
		"calico/node:v3.28.2",
		"registry.k8s.io/coredns/coredns:v1.11.1",
		"harbor.example.com/production/microservice/api:v2.1.0",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, image := range images {
			_ = generateTarFilename(image)
		}
	}
}

// Helper functions (simplified versions of the actual implementation)

func readImageList(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var images []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			images = append(images, line)
		}
	}

	return images, scanner.Err()
}

func findTarFiles(dir string, recursive bool) ([]string, error) {
	var tarFiles []string

	if recursive {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".tar") {
				tarFiles = append(tarFiles, path)
			}
			return nil
		})
		return tarFiles, err
	} else {
		files, err := filepath.Glob(filepath.Join(dir, "*.tar"))
		return files, err
	}
}

func generateTarFilename(image string) string {
	filename := strings.ReplaceAll(image, "/", "_")
	filename = strings.ReplaceAll(filename, ":", "_")
	return filename + ".tar"
}