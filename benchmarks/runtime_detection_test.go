package benchmarks

import (
	"context"
	"testing"
	"time"

	"github.com/harpoon/hpn/internal/runtime"
)

// BenchmarkRuntimeDetection tests runtime detection performance
func BenchmarkRuntimeDetection(b *testing.B) {
	detector := runtime.NewDetector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detector.DetectAvailable()
	}
}

// BenchmarkRuntimeAvailabilityCheck tests individual runtime availability checks
func BenchmarkRuntimeAvailabilityCheck(b *testing.B) {
	runtimes := []runtime.ContainerRuntime{
		runtime.NewDockerRuntime(),
		runtime.NewPodmanRuntime(),
		runtime.NewNerdctlRuntime(),
	}

	for _, rt := range runtimes {
		b.Run(rt.Name(), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = rt.IsAvailable()
			}
		})
	}
}

// BenchmarkRuntimeVersionCheck tests runtime version retrieval performance
func BenchmarkRuntimeVersionCheck(b *testing.B) {
	detector := runtime.NewDetector()
	available := detector.DetectAvailable()
	
	if len(available) == 0 {
		b.Skip("No container runtime available for testing")
	}

	rt := available[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := rt.Version()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRuntimeSelection tests runtime selection performance
func BenchmarkRuntimeSelection(b *testing.B) {
	detector := runtime.NewDetector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detector.GetPreferred()
	}
}

// BenchmarkRuntimeByName tests runtime retrieval by name performance
func BenchmarkRuntimeByName(b *testing.B) {
	detector := runtime.NewDetector()
	runtimeNames := []string{"docker", "podman", "nerdctl"}

	for _, name := range runtimeNames {
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = detector.GetByName(name)
			}
		})
	}
}

// BenchmarkRuntimeOperationSetup tests the overhead of setting up runtime operations
func BenchmarkRuntimeOperationSetup(b *testing.B) {
	detector := runtime.NewDetector()
	available := detector.DetectAvailable()
	
	if len(available) == 0 {
		b.Skip("No container runtime available for testing")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		pullOptions := runtime.PullOptions{
			Timeout: 5 * time.Minute,
		}
		
		// Simulate operation setup without actual execution
		_ = ctx
		_ = pullOptions
		
		cancel()
	}
}

// BenchmarkRuntimeDetectionMemory tests memory allocation during runtime detection
func BenchmarkRuntimeDetectionMemory(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector := runtime.NewDetector()
		_ = detector.DetectAvailable()
	}
}