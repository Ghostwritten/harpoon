package main

import (
	"fmt"
	"strings"
)

// ImageParts 表示解析后的镜像组件
type ImageParts struct {
	Registry string // 仓库地址（如 registry.com）
	Path     string // 路径部分（如 project 或 path/to/project，可能为空）
	Name     string // 镜像名（如 nginx）
	Tag      string // 标签（如 latest）
}

// parseImage 解析镜像名为组件，支持多级路径
// 示例：
//   - old-registry.com/project/nginx:latest -> {Registry: "old-registry.com", Path: "project", Name: "nginx", Tag: "latest"}
//   - old-registry.com/path/to/nginx:latest -> {Registry: "old-registry.com", Path: "path/to", Name: "nginx", Tag: "latest"}
//   - nginx:latest -> {Registry: "", Path: "", Name: "nginx", Tag: "latest"}
//   - registry.com/nginx:latest -> {Registry: "registry.com", Path: "", Name: "nginx", Tag: "latest"}
func parseImage(image string) ImageParts {
	parts := ImageParts{}

	// 分离 tag
	tagIndex := strings.LastIndex(image, ":")
	if tagIndex > 0 && tagIndex < len(image)-1 {
		// 检查 : 后面是否有 /，如果没有则认为是 tag
		afterColon := image[tagIndex+1:]
		if !strings.Contains(afterColon, "/") {
			parts.Tag = afterColon
			image = image[:tagIndex]
		}
	}
	if parts.Tag == "" {
		parts.Tag = "latest" // 默认 tag
	}

	// 分离 registry 和路径
	imageParts := strings.Split(image, "/")
	if len(imageParts) == 1 {
		// 只有镜像名，如 nginx
		parts.Name = imageParts[0]
		return parts
	}

	// 判断第一部分是否是 registry（包含 . 或 : 或 localhost）
	firstPart := imageParts[0]
	if strings.Contains(firstPart, ".") || strings.Contains(firstPart, ":") || firstPart == "localhost" {
		// 第一部分是 registry
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
		// 第一部分不是 registry，可能是路径的一部分
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

// buildTargetImage 构建目标镜像名
// 实现三种命名模式：
//   1. 统一项目模式：如果指定了 project，所有镜像统一推送到 registry/project/image:tag
//   2. 追加路径模式：如果 registry 包含路径，在原有路径前追加
//   3. 默认保持路径模式：只替换仓库，保持原有路径结构
func buildTargetImage(sourceImage, registry, project string) string {
	// 解析源镜像
	sourceParts := parseImage(sourceImage)

	// 场景1: 指定了 --project，统一替换项目
	if project != "" {
		// 提取 registry 的基础部分（如果 registry 包含路径，只取基础部分）
		baseRegistry := extractRegistryBase(registry)
		return fmt.Sprintf("%s/%s/%s:%s",
			baseRegistry, // new-registry.com
			project,      // newproject
			sourceParts.Name,
			sourceParts.Tag)
	}

	// 场景2: --registry 包含路径（如 new-registry.com/path/xx）
	if hasPath(registry) {
		// 追加路径模式
		if sourceParts.Path != "" {
			return fmt.Sprintf("%s/%s/%s:%s",
				registry,         // new-registry.com/path/xx
				sourceParts.Path, // project (原有路径)
				sourceParts.Name,
				sourceParts.Tag)
		}
		// 如果源镜像没有路径，直接追加
		return fmt.Sprintf("%s/%s:%s",
			registry,
			sourceParts.Name,
			sourceParts.Tag)
	}

	// 场景3: 默认模式，只替换仓库，保持路径
	// 如果源镜像没有路径，则直接 registry/name:tag
	if sourceParts.Path == "" {
		return fmt.Sprintf("%s/%s:%s",
			registry,
			sourceParts.Name,
			sourceParts.Tag)
	}

	// 如果源镜像有路径，保持路径
	return fmt.Sprintf("%s/%s/%s:%s",
		registry,         // new-registry.com
		sourceParts.Path, // project (保持原有路径)
		sourceParts.Name,
		sourceParts.Tag)
}

// extractRegistryBase 从 registry 中提取基础仓库名
// 例如：new-registry.com/path/xx -> new-registry.com
func extractRegistryBase(registry string) string {
	parts := strings.Split(registry, "/")
	return parts[0]
}

// hasPath 检查 registry 是否包含路径
func hasPath(registry string) bool {
	return strings.Contains(registry, "/")
}

// parseImageNameAndTag 解析镜像名和标签（保持向后兼容）
func parseImageNameAndTag(image string) (string, string) {
	parts := parseImage(image)
	if parts.Path != "" {
		return fmt.Sprintf("%s/%s", parts.Path, parts.Name), parts.Tag
	}
	return parts.Name, parts.Tag
}

// extractProjectFromImage 从镜像中提取项目名（保持向后兼容）
func extractProjectFromImage(image string) string {
	parts := parseImage(image)
	if parts.Path != "" {
		// 返回路径的第一部分作为项目名
		pathParts := strings.Split(parts.Path, "/")
		return pathParts[0]
	}
	return "library" // 默认项目名
}
