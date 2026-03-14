package main

import (
	"fmt"
	"strings"
)

// ImageParts holds parsed image components
type ImageParts struct {
	Registry string // Registry address (e.g. registry.com)
	Path     string // Path part (e.g. project or path/to/project, may be empty)
	Name     string // Image name (e.g. nginx)
	Tag      string // Tag (e.g. latest)
}

// parseImage parses an image reference into components, supporting multi-level paths.
// Examples:
//
//	old-registry.com/project/nginx:latest -> {Registry: "old-registry.com", Path: "project", Name: "nginx", Tag: "latest"}
//	old-registry.com/path/to/nginx:latest -> {Registry: "old-registry.com", Path: "path/to", Name: "nginx", Tag: "latest"}
//	nginx:latest -> {Registry: "", Path: "", Name: "nginx", Tag: "latest"}
//	registry.com/nginx:latest -> {Registry: "registry.com", Path: "", Name: "nginx", Tag: "latest"}
func parseImage(image string) ImageParts {
	parts := ImageParts{}

	// Split tag from image
	tagIndex := strings.LastIndex(image, ":")
	if tagIndex > 0 && tagIndex < len(image)-1 {
		// If nothing after : contains /, treat it as tag
		afterColon := image[tagIndex+1:]
		if !strings.Contains(afterColon, "/") {
			parts.Tag = afterColon
			image = image[:tagIndex]
		}
	}
	if parts.Tag == "" {
		parts.Tag = "latest" // default tag
	}

	// Split registry and path
	imageParts := strings.Split(image, "/")
	if len(imageParts) == 1 {
		// Image name only (e.g. nginx)
		parts.Name = imageParts[0]
		return parts
	}

	// Check if first part is registry (contains . or : or is localhost)
	firstPart := imageParts[0]
	if strings.Contains(firstPart, ".") || strings.Contains(firstPart, ":") || firstPart == "localhost" {
		// First part is registry
		parts.Registry = firstPart
		if len(imageParts) == 2 {
			// registry/image
			parts.Name = imageParts[1]
		} else {
			// registry/path/to/image
			parts.Path = strings.Join(imageParts[1:len(imageParts)-1], "/")
			parts.Name = imageParts[len(imageParts)-1]
		}
	} else {
		// First part is not registry, likely part of path
		if len(imageParts) == 2 {
			// path/image
			parts.Path = imageParts[0]
			parts.Name = imageParts[1]
		} else {
			// path/to/image
			parts.Path = strings.Join(imageParts[:len(imageParts)-1], "/")
			parts.Name = imageParts[len(imageParts)-1]
		}
	}

	return parts
}

// buildTargetImage builds the target image name.
// Three modes:
//  1. Unified project: if project is set, push all to registry/project/image:tag
//  2. Append path: if registry contains path, prepend to original path
//  3. Default preserve: replace registry only, keep original path structure
func buildTargetImage(sourceImage, registry, project string) string {
	// Parse source image
	sourceParts := parseImage(sourceImage)

	// Mode 1: --project specified, unified project
	if project != "" {
		// Extract base registry (if registry has path, take base only)
		baseRegistry := extractRegistryBase(registry)
		return fmt.Sprintf("%s/%s/%s:%s",
			baseRegistry, // new-registry.com
			project,     // newproject
			sourceParts.Name,
			sourceParts.Tag)
	}

	// Mode 2: --registry contains path (e.g. new-registry.com/path/xx)
	if hasPath(registry) {
		// Append path mode
		if sourceParts.Path != "" {
			return fmt.Sprintf("%s/%s/%s:%s",
				registry,         // new-registry.com/path/xx
				sourceParts.Path, // original path
				sourceParts.Name,
				sourceParts.Tag)
		}
		// Source has no path, append directly
		return fmt.Sprintf("%s/%s:%s",
			registry,
			sourceParts.Name,
			sourceParts.Tag)
	}

	// Mode 3: Default, replace registry only, preserve path
	// If source has no path: registry/name:tag
	if sourceParts.Path == "" {
		return fmt.Sprintf("%s/%s:%s",
			registry,
			sourceParts.Name,
			sourceParts.Tag)
	}

	// Source has path, preserve it
	return fmt.Sprintf("%s/%s/%s:%s",
		registry,         // new-registry.com
		sourceParts.Path, // original path
		sourceParts.Name,
		sourceParts.Tag)
}

// extractRegistryBase extracts the base registry from a registry string.
// E.g. new-registry.com/path/xx -> new-registry.com
func extractRegistryBase(registry string) string {
	parts := strings.Split(registry, "/")
	return parts[0]
}

// hasPath returns true if registry contains a path
func hasPath(registry string) bool {
	return strings.Contains(registry, "/")
}

// parseImageNameAndTag parses image name and tag (for backward compatibility)
func parseImageNameAndTag(image string) (string, string) {
	parts := parseImage(image)
	if parts.Path != "" {
		return fmt.Sprintf("%s/%s", parts.Path, parts.Name), parts.Tag
	}
	return parts.Name, parts.Tag
}

// extractProjectFromImage extracts project name from image (for backward compatibility)
func extractProjectFromImage(image string) string {
	parts := parseImage(image)
	if parts.Path != "" {
		// Return first segment of path as project
		pathParts := strings.Split(parts.Path, "/")
		return pathParts[0]
	}
	return "library" // default project
}
