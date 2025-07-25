package version

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// Build information. These will be set by the build system.
var (
	// Version is the semantic version of the application
	Version = "v1.1.0-dev"
	
	// GitCommit is the git commit hash
	GitCommit = "unknown"
	
	// GitBranch is the git branch name
	GitBranch = "unknown"
	
	// BuildDate is the date when the binary was built
	BuildDate = "unknown"
	
	// BuildUser is the user who built the binary
	BuildUser = "unknown"
	
	// BuildHost is the host where the binary was built
	BuildHost = "unknown"
	
	// GoVersion is the Go version used to build the binary
	GoVersion = runtime.Version()
	
	// Platform is the target platform
	Platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)

// Info contains version information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	GitBranch string `json:"gitBranch"`
	BuildDate string `json:"buildDate"`
	BuildUser string `json:"buildUser"`
	BuildHost string `json:"buildHost"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

// Get returns the version information
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		GitBranch: GitBranch,
		BuildDate: BuildDate,
		BuildUser: BuildUser,
		BuildHost: BuildHost,
		GoVersion: GoVersion,
		Platform:  Platform,
	}
}

// GetVersion returns just the version string
func GetVersion() string {
	return Version
}

// GetShortCommit returns the short commit hash (first 7 characters)
func GetShortCommit() string {
	if len(GitCommit) >= 7 {
		return GitCommit[:7]
	}
	return GitCommit
}

// GetBuildInfo returns build information as a formatted string
func GetBuildInfo() string {
	return fmt.Sprintf("commit: %s, built: %s", GetShortCommit(), BuildDate)
}

// GetFullVersion returns the full version string with build info
func GetFullVersion() string {
	return fmt.Sprintf("%s (%s)", Version, GetBuildInfo())
}

// GetDetailedVersion returns detailed version information
func GetDetailedVersion() string {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("Version: %s", Version))
	
	if GitCommit != "unknown" {
		parts = append(parts, fmt.Sprintf("Git Commit: %s", GitCommit))
	}
	
	if GitBranch != "unknown" {
		parts = append(parts, fmt.Sprintf("Git Branch: %s", GitBranch))
	}
	
	if BuildDate != "unknown" {
		parts = append(parts, fmt.Sprintf("Build Date: %s", BuildDate))
	}
	
	if BuildUser != "unknown" {
		parts = append(parts, fmt.Sprintf("Build User: %s", BuildUser))
	}
	
	if BuildHost != "unknown" {
		parts = append(parts, fmt.Sprintf("Build Host: %s", BuildHost))
	}
	
	parts = append(parts, fmt.Sprintf("Go Version: %s", GoVersion))
	parts = append(parts, fmt.Sprintf("Platform: %s", Platform))
	
	return strings.Join(parts, "\n")
}

// GetUserAgent returns a user agent string for HTTP requests
func GetUserAgent() string {
	return fmt.Sprintf("harpoon/%s (%s; %s)", Version, runtime.GOOS, runtime.GOARCH)
}

// IsDevVersion returns true if this is a development version
func IsDevVersion() bool {
	return strings.Contains(Version, "dev") || 
		   strings.Contains(Version, "alpha") || 
		   strings.Contains(Version, "beta") || 
		   strings.Contains(Version, "rc")
}

// IsRelease returns true if this is a release version
func IsRelease() bool {
	return !IsDevVersion() && GitCommit != "unknown" && BuildDate != "unknown"
}

// GetBuildTime returns the build time as a time.Time object
func GetBuildTime() (time.Time, error) {
	if BuildDate == "unknown" {
		return time.Time{}, fmt.Errorf("build date is unknown")
	}
	
	// Try different time formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, BuildDate); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse build date: %s", BuildDate)
}

// GetAge returns the age of the build
func GetAge() (time.Duration, error) {
	buildTime, err := GetBuildTime()
	if err != nil {
		return 0, err
	}
	
	return time.Since(buildTime), nil
}

// PrintVersion prints version information to stdout
func PrintVersion() {
	fmt.Printf("Harpoon (hpn) %s\n", GetFullVersion())
}

// PrintDetailedVersion prints detailed version information to stdout
func PrintDetailedVersion() {
	fmt.Printf("Harpoon (hpn)\n%s\n", GetDetailedVersion())
}

// Validate checks if the version information is valid
func Validate() error {
	if Version == "" {
		return fmt.Errorf("version is empty")
	}
	
	// Check if version follows semantic versioning pattern
	if !strings.HasPrefix(Version, "v") {
		return fmt.Errorf("version should start with 'v': %s", Version)
	}
	
	return nil
}