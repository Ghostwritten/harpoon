package benchmarks

import (
	"testing"

	"github.com/harpoon/hpn/pkg/types"
)

// BenchmarkImageParsing tests the performance of image name parsing
func BenchmarkImageParsing(b *testing.B) {
	testCases := []struct {
		name  string
		image string
	}{
		{"Simple", "nginx:latest"},
		{"WithProject", "calico/node:v3.28.2"},
		{"WithRegistry", "registry.k8s.io/coredns:v1.11.1"},
		{"Complex", "harbor.example.com/production/microservice/api:v2.1.0"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := types.ParseImage(tc.image)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkImageParsingBatch tests batch image parsing performance
func BenchmarkImageParsingBatch(b *testing.B) {
	images := []string{
		"nginx:latest",
		"redis:7.0",
		"postgres:15",
		"calico/node:v3.28.2",
		"calico/cni:v3.28.2",
		"registry.k8s.io/coredns/coredns:v1.11.1",
		"registry.k8s.io/etcd:3.5.9-0",
		"registry.k8s.io/kube-apiserver:v1.28.2",
		"harbor.example.com/prod/app:v1.0.0",
		"quay.io/prometheus/prometheus:v2.45.0",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, image := range images {
			_, err := types.ParseImage(image)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkImageTarFilenameGeneration tests tar filename generation performance
func BenchmarkImageTarFilenameGeneration(b *testing.B) {
	image, err := types.ParseImage("registry.k8s.io/coredns/coredns:v1.11.1")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = image.GenerateTarFilename()
	}
}

// BenchmarkImageStringConversion tests image string conversion performance
func BenchmarkImageStringConversion(b *testing.B) {
	image := &types.Image{
		Registry: "registry.k8s.io",
		Project:  "coredns",
		Name:     "coredns",
		Tag:      "v1.11.1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = image.String()
	}
}

// BenchmarkImageParsingMemory tests memory allocation during image parsing
func BenchmarkImageParsingMemory(b *testing.B) {
	image := "registry.k8s.io/coredns/coredns:v1.11.1"
	
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := types.ParseImage(image)
		if err != nil {
			b.Fatal(err)
		}
	}
}