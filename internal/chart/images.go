package chart

import (
	"bytes"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/harpoon/hpn/pkg/types"
	"gopkg.in/yaml.v3"
)

// imageLikePattern matches substrings that look like container image refs: at least two path segments
// (e.g. registry/repo:tag) to avoid matching apiVersion-like "apps/v1".
var imageLikePattern = regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9.-]+(/[a-zA-Z0-9][a-zA-Z0-9.-]+)+:[a-zA-Z0-9][a-zA-Z0-9._-]+`)

// ExtractImagesFromYAML parses a multi-document YAML stream (e.g. helm template output),
// recursively finds all string values that look like container image references (including
// those under keys other than "image", e.g. in args like --foo=registry/repo:tag),
// validates them with ParseImage, deduplicates and returns.
func ExtractImagesFromYAML(yamlBytes []byte) ([]string, error) {
	dec := yaml.NewDecoder(bytes.NewReader(yamlBytes))
	var collected []string
	seen := make(map[string]struct{})
	for {
		var doc interface{}
		if err := dec.Decode(&doc); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		walkYAML(doc, func(_ string, value string) {
			value = strings.TrimSpace(value)
			if value == "" {
				return
			}
			for _, candidate := range extractImageCandidates(value) {
				if !isPlausibleContainerImage(candidate) {
					continue
				}
				if _, err := types.ParseImage(candidate); err != nil {
					continue
				}
				if _, ok := seen[candidate]; ok {
					continue
				}
				seen[candidate] = struct{}{}
				collected = append(collected, candidate)
			}
		})
	}
	sort.Strings(collected)
	return collected, nil
}

// extractImageCandidates returns the string as a single candidate if it looks like a whole image ref,
// or finds all substrings that match an image-like pattern (e.g. in --flag=quay.io/repo/img:tag).
func extractImageCandidates(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	// Whole string is a valid-looking image ref (key "image" or standalone value).
	// Skip strings that look like flags (e.g. --foo=...) so we extract the image part via regex below.
	if looksLikeImageReference(s) && !strings.HasPrefix(s, "--") && !strings.Contains(s, "=") {
		if _, err := types.ParseImage(s); err == nil {
			return []string{s}
		}
	}
	// Find embedded image refs (e.g. in --prometheus-config-reloader=quay.io/...:tag)
	var out []string
	for _, sub := range imageLikePattern.FindAllString(s, -1) {
		sub = strings.TrimSpace(sub)
		if sub != "" {
			out = append(out, sub)
		}
	}
	return out
}

// looksLikeImageReference heuristically detects strings that are likely container image refs.
// Requires ":" (tag) to avoid treating apiVersion-like "apps/v1" as an image.
func looksLikeImageReference(s string) bool {
	if len(s) < 4 || strings.Contains(s, " ") {
		return false
	}
	return strings.Contains(s, ":")
}

// isPlausibleContainerImage filters out URLs and non-image strings (e.g. Prometheus metric names)
// that happen to match path:tag pattern. Requires at least one "/" and rejects http(s) URLs.
func isPlausibleContainerImage(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "unix://") || strings.HasPrefix(lower, "tcp://") || strings.HasPrefix(lower, "fd://") {
		return false
	}
	lastColon := strings.LastIndex(s, ":")
	if lastColon <= 0 || lastColon >= len(s)-1 {
		return false
	}
	tag := s[lastColon+1:]
	// Reject if tag looks like a port number only (e.g. "9090") or metric suffix (e.g. "sum", "rate5m")
	if len(tag) > 0 && tag[0] >= '0' && tag[0] <= '9' && strings.IndexFunc(tag, func(r rune) bool { return r < '0' || r > '9' }) < 0 {
		return false
	}
	// If no "/", allow only "name:tag" with exactly one colon and tag that looks like a version (e.g. latest, v1.0, 1.21-alpine)
	if !strings.Contains(s, "/") {
		if strings.Index(s, ":") != lastColon {
			return false
		}
		return looksLikeImageTag(tag)
	}
	return true
}

// looksLikeImageTag returns true if tag looks like a container image tag (version or "latest"), not a metric suffix.
func looksLikeImageTag(tag string) bool {
	if tag == "latest" {
		return true
	}
	if len(tag) == 0 {
		return false
	}
	// Version-like: starts with 'v' and digit, or contains a dot (e.g. 1.21-alpine, 12.3.1)
	if tag[0] == 'v' && len(tag) > 1 && (tag[1] >= '0' && tag[1] <= '9') {
		return true
	}
	if strings.Contains(tag, ".") {
		return true
	}
	// All digits (version number only)
	if strings.IndexFunc(tag, func(r rune) bool { return r < '0' || r > '9' }) < 0 {
		return len(tag) <= 5
	}
	return false
}

func walkYAML(node interface{}, fn func(key string, value string)) {
	switch n := node.(type) {
	case map[interface{}]interface{}:
		for k, v := range n {
			key, _ := k.(string)
			switch val := v.(type) {
			case string:
				fn(key, val)
			default:
				walkYAML(v, fn)
			}
		}
	case map[string]interface{}:
		for k, v := range n {
			switch val := v.(type) {
			case string:
				fn(k, val)
			default:
				walkYAML(v, fn)
			}
		}
	case []interface{}:
		for _, item := range n {
			walkYAML(item, fn)
		}
	case string:
		fn("", n)
	}
}
