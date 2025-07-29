package version

import (
	"fmt"
	"runtime"
)

// Version information - these will be set by build flags
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
	GoVersion = runtime.Version()
)

// GetVersion returns the current version
func GetVersion() string {
	return Version
}

// GetShortCommit returns the short commit hash
func GetShortCommit() string {
	if len(GitCommit) > 7 {
		return GitCommit[:7]
	}
	return GitCommit
}

// GetFullVersion returns the full version string
func GetFullVersion() string {
	if Version == "dev" {
		return fmt.Sprintf("%s-%s", Version, GetShortCommit())
	}
	return Version
}

// PrintVersion prints basic version information
func PrintVersion() {
	fmt.Printf("hpn version %s\n", GetFullVersion())
}

// PrintDetailedVersion prints detailed version information
func PrintDetailedVersion() {
	fmt.Printf("hpn version %s\n", GetFullVersion())
	fmt.Printf("Git commit: %s\n", GitCommit)
	fmt.Printf("Build date: %s\n", BuildDate)
	fmt.Printf("Go version: %s\n", GoVersion)
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}