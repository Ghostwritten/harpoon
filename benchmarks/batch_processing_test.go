package benchmarks

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/harpoon/hpn/pkg/types"
)

// BenchmarkSerialImageProcessing tests serial image processing performance
func BenchmarkSerialImageProcessing(b *testing.B) {
	testCases := []struct {
		name  string
		count int
	}{
		{"Small", 10},
		{"Medium", 50},
		{"Large", 100},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			images := generateTestImages(tc.count)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := processImagesSerial(images)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkConcurrentImageProcessing tests concurrent image processing performance
func BenchmarkConcurrentImageProcessing(b *testing.B) {
	testCases := []struct {
		name    string
		count   int
		workers int
	}{
		{"Small_2Workers", 10, 2},
		{"Small_4Workers", 10, 4},
		{"Medium_2Workers", 50, 2},
		{"Medium_4Workers", 50, 4},
		{"Medium_8Workers", 50, 8},
		{"Large_4Workers", 100, 4},
		{"Large_8Workers", 100, 8},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			images := generateTestImages(tc.count)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := processImagesConcurrent(images, tc.workers)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkBatchImageListProcessing tests batch processing of image lists
func BenchmarkBatchImageListProcessing(b *testing.B) {
	// Create temporary image list files
	testFiles := make([]string, 0)
	defer func() {
		for _, file := range testFiles {
			os.Remove(file)
		}
	}()

	// Create files with different sizes
	fileSizes := []int{10, 50, 100, 500}
	for _, size := range fileSizes {
		tmpFile, err := ioutil.TempFile("", fmt.Sprintf("images-%d-*.txt", size))
		if err != nil {
			b.Fatal(err)
		}
		
		for i := 0; i < size; i++ {
			fmt.Fprintf(tmpFile, "nginx:v%d.0\n", i)
		}
		tmpFile.Close()
		testFiles = append(testFiles, tmpFile.Name())
	}

	b.Run("Serial", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, file := range testFiles {
				images, err := readImageList(file)
				if err != nil {
					b.Fatal(err)
				}
				err = processImagesSerial(images)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("Concurrent", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, file := range testFiles {
				images, err := readImageList(file)
				if err != nil {
					b.Fatal(err)
				}
				err = processImagesConcurrent(images, 4)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}

// BenchmarkMemoryUsageScaling tests memory usage scaling with image count
func BenchmarkMemoryUsageScaling(b *testing.B) {
	testCases := []int{100, 500, 1000, 2000}

	for _, count := range testCases {
		b.Run(fmt.Sprintf("Images_%d", count), func(b *testing.B) {
			images := generateTestImages(count)
			
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := processImagesSerial(images)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkWorkerPoolOverhead tests worker pool creation overhead
func BenchmarkWorkerPoolOverhead(b *testing.B) {
	images := generateTestImages(50)
	workerCounts := []int{1, 2, 4, 8, 16}

	for _, workers := range workerCounts {
		b.Run(fmt.Sprintf("Workers_%d", workers), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := processImagesConcurrent(images, workers)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkContextCancellation tests context cancellation performance
func BenchmarkContextCancellation(b *testing.B) {
	images := generateTestImages(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		
		// Start processing
		go func() {
			processImagesWithContext(ctx, images)
		}()
		
		// Cancel immediately to test cancellation overhead
		cancel()
	}
}

// Helper functions

func generateTestImages(count int) []string {
	images := make([]string, count)
	registries := []string{"docker.io", "registry.k8s.io", "harbor.example.com"}
	projects := []string{"library", "coredns", "production"}
	names := []string{"nginx", "redis", "postgres", "app", "api"}
	
	for i := 0; i < count; i++ {
		registry := registries[i%len(registries)]
		project := projects[i%len(projects)]
		name := names[i%len(names)]
		tag := fmt.Sprintf("v%d.0", i%10)
		
		if i%3 == 0 {
			images[i] = fmt.Sprintf("%s:%s", name, tag)
		} else if i%3 == 1 {
			images[i] = fmt.Sprintf("%s/%s:%s", project, name, tag)
		} else {
			images[i] = fmt.Sprintf("%s/%s/%s:%s", registry, project, name, tag)
		}
	}
	
	return images
}

func processImagesSerial(images []string) error {
	for _, image := range images {
		if err := processImage(image); err != nil {
			return err
		}
	}
	return nil
}

func processImagesConcurrent(images []string, workers int) error {
	jobs := make(chan string, len(images))
	results := make(chan error, len(images))
	
	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for image := range jobs {
				results <- processImage(image)
			}
		}()
	}
	
	// Send jobs
	for _, image := range images {
		jobs <- image
	}
	close(jobs)
	
	// Wait for completion
	wg.Wait()
	close(results)
	
	// Check for errors
	for err := range results {
		if err != nil {
			return err
		}
	}
	
	return nil
}

func processImagesWithContext(ctx context.Context, images []string) error {
	for _, image := range images {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := processImage(image); err != nil {
				return err
			}
		}
	}
	return nil
}

func processImage(image string) error {
	// Simulate image processing work
	parsed, err := types.ParseImage(image)
	if err != nil {
		return err
	}
	
	// Simulate some processing time
	time.Sleep(100 * time.Microsecond)
	
	// Generate tar filename (simulating save operation)
	_ = parsed.GenerateTarFilename()
	
	return nil
}