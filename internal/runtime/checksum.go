package runtime

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
)

// VerifyTarChecksum verifies the checksum of a tar file against its .sha256 file.
// Returns (true, nil) if the file exists and matches, (false, error) if the checksum
// file is missing or the checksum does not match.
func VerifyTarChecksum(tarPath string) (bool, error) {
	checksumPath := tarPath + ".sha256"
	expectedChecksumBytes, err := os.ReadFile(checksumPath)
	if err != nil {
		return false, fmt.Errorf("failed to read checksum file: %w", err)
	}
	expectedChecksum := strings.TrimSpace(string(expectedChecksumBytes))

	file, err := os.Open(tarPath)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false, fmt.Errorf("failed to calculate checksum: %w", err)
	}
	actualChecksum := fmt.Sprintf("%x", hash.Sum(nil))
	return actualChecksum == expectedChecksum, nil
}

// TarFilenameFromImage generates a tar filename from an image name (e.g. nginx:latest -> nginx_latest.tar).
func TarFilenameFromImage(image string) string {
	filename := strings.ReplaceAll(image, "/", "_")
	filename = strings.ReplaceAll(filename, ":", "_")
	return filename + ".tar"
}

// ImageNameFromTarFilename reverses TarFilenameFromImage (best-effort).
// e.g. nginx_latest.tar -> nginx:latest, quay.io_prometheus_alertmanager_v0.30.1.tar -> quay.io/prometheus/alertmanager:v0.30.1
// Heuristic: last segment before .tar is the tag, rest join with /
func ImageNameFromTarFilename(tarFilename string) string {
	stem := strings.TrimSuffix(tarFilename, ".tar")
	segments := strings.Split(stem, "_")
	if len(segments) == 1 {
		if segments[0] == "" {
			return "latest"
		}
		return segments[0] + ":latest"
	}
	tag := segments[len(segments)-1]
	name := strings.Join(segments[:len(segments)-1], "/")
	return name + ":" + tag
}

// generateChecksum generates a SHA256 checksum file for the given file
// Returns the checksum file path and error
func generateChecksum(filePath string) (string, error) {
	checksumPath := filePath + ".sha256"
	
	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for checksum: %w", err)
	}
	defer file.Close()
	
	// Calculate SHA256 hash
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}
	
	// Write checksum to file
	checksum := fmt.Sprintf("%x", hash.Sum(nil))
	if err := os.WriteFile(checksumPath, []byte(checksum), 0644); err != nil {
		return "", fmt.Errorf("failed to write checksum file: %w", err)
	}
	
	return checksumPath, nil
}

// verifyChecksum verifies the checksum of a file against its .sha256 file
// Returns true if checksum matches, false otherwise
func verifyChecksum(filePath string) (bool, error) {
	checksumPath := filePath + ".sha256"
	
	// Check if checksum file exists
	if _, err := os.Stat(checksumPath); os.IsNotExist(err) {
		return false, nil
	}
	
	// Read expected checksum
	expectedChecksum, err := os.ReadFile(checksumPath)
	if err != nil {
		return false, fmt.Errorf("failed to read checksum file: %w", err)
	}
	
	// Calculate actual checksum
	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to open file for verification: %w", err)
	}
	defer file.Close()
	
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false, fmt.Errorf("failed to calculate checksum: %w", err)
	}
	
	actualChecksum := fmt.Sprintf("%x", hash.Sum(nil))
	expected := string(expectedChecksum)
	// Remove any whitespace from expected checksum
	expected = trimWhitespace(expected)
	
	return actualChecksum == expected, nil
}

// trimWhitespace removes leading and trailing whitespace
func trimWhitespace(s string) string {
	start := 0
	end := len(s)
	
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	
	return s[start:end]
}

