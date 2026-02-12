package main

import (
	"testing"
)

func TestParseImage(t *testing.T) {
	tests := []struct {
		image    string
		registry string
		path     string
		name     string
		tag      string
	}{
		{"nginx:latest", "", "", "nginx", "latest"},
		{"nginx", "", "", "nginx", "latest"},
		{"registry.com/nginx:latest", "registry.com", "", "nginx", "latest"},
		{"registry.com/project/nginx:1.20", "registry.com", "project", "nginx", "1.20"},
		{"old-registry.com/path/to/nginx:latest", "old-registry.com", "path/to", "nginx", "latest"},
		{"quay.io/foo/bar:v1", "quay.io", "foo", "bar", "v1"},
		{"localhost:5000/myimg:tag", "localhost:5000", "", "myimg", "tag"},
		{"docker.io/library/alpine:3.18", "docker.io", "library", "alpine", "3.18"},
	}
	for _, tt := range tests {
		got := parseImage(tt.image)
		if got.Registry != tt.registry || got.Path != tt.path || got.Name != tt.name || got.Tag != tt.tag {
			t.Errorf("parseImage(%q) = Registry=%q Path=%q Name=%q Tag=%q, want Registry=%q Path=%q Name=%q Tag=%q",
				tt.image, got.Registry, got.Path, got.Name, got.Tag,
				tt.registry, tt.path, tt.name, tt.tag)
		}
	}
}

func TestBuildTargetImage_DefaultMode_OnlyRegistry(t *testing.T) {
	// Only replace registry, preserve path
	got := buildTargetImage("old.com/proj/nginx:latest", "new.com", "")
	want := "new.com/proj/nginx:latest"
	if got != want {
		t.Errorf("buildTargetImage(only registry) = %q, want %q", got, want)
	}
}

func TestBuildTargetImage_DefaultMode_NoPath(t *testing.T) {
	got := buildTargetImage("nginx:latest", "new.com", "")
	want := "new.com/nginx:latest"
	if got != want {
		t.Errorf("buildTargetImage(no path) = %q, want %q", got, want)
	}
}

func TestBuildTargetImage_UnifiedProject(t *testing.T) {
	// --project set: all images to registry/project/name:tag
	got := buildTargetImage("old.com/foo/bar:v1", "new.com", "production")
	want := "new.com/production/bar:v1"
	if got != want {
		t.Errorf("buildTargetImage(unified project) = %q, want %q", got, want)
	}
}

func TestBuildTargetImage_RegistryWithPath_AppendMode(t *testing.T) {
	// registry has path: new.com/path/xx, append source path
	got := buildTargetImage("old.com/proj/nginx:latest", "new.com/path/xx", "")
	want := "new.com/path/xx/proj/nginx:latest"
	if got != want {
		t.Errorf("buildTargetImage(append path) = %q, want %q", got, want)
	}
}

func TestBuildTargetImage_RegistryWithPath_NoSourcePath(t *testing.T) {
	got := buildTargetImage("nginx:latest", "new.com/path/xx", "")
	want := "new.com/path/xx/nginx:latest"
	if got != want {
		t.Errorf("buildTargetImage(registry path, no source path) = %q, want %q", got, want)
	}
}

func TestExtractRegistryBase(t *testing.T) {
	if got := extractRegistryBase("new.com/path/xx"); got != "new.com" {
		t.Errorf("extractRegistryBase = %q, want new.com", got)
	}
	if got := extractRegistryBase("registry.com"); got != "registry.com" {
		t.Errorf("extractRegistryBase = %q, want registry.com", got)
	}
}

func TestHasPath(t *testing.T) {
	if !hasPath("reg.com/path") {
		t.Error("hasPath(reg.com/path) should be true")
	}
	if hasPath("reg.com") {
		t.Error("hasPath(reg.com) should be false")
	}
}
