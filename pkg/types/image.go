package types

import (
	"fmt"
	"strings"
)

// Image represents a container image with its components
type Image struct {
	Registry string `json:"registry"`
	Project  string `json:"project"`
	Name     string `json:"name"`
	Tag      string `json:"tag"`
	FullName string `json:"full_name"`
	Digest   string `json:"digest,omitempty"`
	Size     int64  `json:"size,omitempty"`
}

// String returns the full image name
func (i *Image) String() string {
	if i.Project != "" && i.Project != "library" {
		return fmt.Sprintf("%s/%s/%s:%s", i.Registry, i.Project, i.Name, i.Tag)
	}
	if i.Registry != "" && i.Registry != "docker.io" {
		return fmt.Sprintf("%s/%s:%s", i.Registry, i.Name, i.Tag)
	}
	return fmt.Sprintf("%s:%s", i.Name, i.Tag)
}

// ParseImage parses an image string into an Image struct
func ParseImage(imageStr string) (*Image, error) {
	if imageStr == "" {
		return nil, fmt.Errorf("image string cannot be empty")
	}

	image := &Image{
		FullName: imageStr,
	}

	// Split by tag first
	parts := strings.Split(imageStr, ":")
	if len(parts) == 1 {
		image.Tag = "latest"
	} else {
		image.Tag = parts[len(parts)-1]
		imageStr = strings.Join(parts[:len(parts)-1], ":")
	}

	// Split by registry and path
	pathParts := strings.Split(imageStr, "/")
	
	switch len(pathParts) {
	case 1:
		// Simple image name (e.g., "nginx")
		image.Registry = "docker.io"
		image.Project = "library"
		image.Name = pathParts[0]
	case 2:
		// Could be registry/image or project/image
		if strings.Contains(pathParts[0], ".") || strings.Contains(pathParts[0], ":") {
			// registry/image
			image.Registry = pathParts[0]
			image.Project = ""
			image.Name = pathParts[1]
		} else {
			// project/image (assume docker.io)
			image.Registry = "docker.io"
			image.Project = pathParts[0]
			image.Name = pathParts[1]
		}
	case 3:
		// registry/project/image
		image.Registry = pathParts[0]
		image.Project = pathParts[1]
		image.Name = pathParts[2]
	default:
		// More complex path (registry/nested/project/image)
		image.Registry = pathParts[0]
		image.Project = strings.Join(pathParts[1:len(pathParts)-1], "/")
		image.Name = pathParts[len(pathParts)-1]
	}

	return image, nil
}

// GenerateTarFilename generates a tar filename for the image
func (i *Image) GenerateTarFilename() string {
	registry := strings.ReplaceAll(i.Registry, ".", "_")
	registry = strings.ReplaceAll(registry, ":", "_")
	
	project := i.Project
	if project == "" || project == "library" {
		project = "library"
	}
	project = strings.ReplaceAll(project, "/", "_")
	
	name := strings.ReplaceAll(i.Name, "/", "_")
	tag := strings.ReplaceAll(i.Tag, ":", "_")
	
	return fmt.Sprintf("%s_%s_%s_%s.tar", registry, project, name, tag)
}