package runtime

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AuthConfig represents Docker/Podman auth configuration
type AuthConfig struct {
	Auths map[string]AuthEntry `json:"auths"`
}

// AuthEntry represents a single registry auth entry
type AuthEntry struct {
	Auth string `json:"auth"` // base64(username:password)
}

// GetAuthFilePath returns the auth file path for the given runtime
func GetAuthFilePath(runtimeName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtimeName {
	case "docker", "nerdctl":
		return filepath.Join(homeDir, ".docker", "config.json"), nil
	case "podman", "skopeo":
		return filepath.Join(homeDir, ".config", "containers", "auth.json"), nil
	default:
		return "", fmt.Errorf("unknown runtime: %s", runtimeName)
	}
}

// ReadAuthConfig reads auth configuration from file
func ReadAuthConfig(authFile string) (*AuthConfig, error) {
	config := &AuthConfig{
		Auths: make(map[string]AuthEntry),
	}

	data, err := os.ReadFile(authFile)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil // Return empty config if file doesn't exist
		}
		return nil, err
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse auth config: %v", err)
	}

	return config, nil
}

// WriteAuthConfig writes auth configuration to file
func WriteAuthConfig(authFile string, config *AuthConfig) error {
	// Ensure directory exists
	dir := filepath.Dir(authFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create auth directory: %v", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal auth config: %v", err)
	}

	if err := os.WriteFile(authFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write auth file: %v", err)
	}

	return nil
}

// EncodeAuth encodes username:password to base64
func EncodeAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
