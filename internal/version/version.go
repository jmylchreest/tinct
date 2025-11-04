// Package version provides build-time version information for Tinct.
// Version information is injected at build time using ldflags.
package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the semantic version of the application.
	// Injected at build time via: -ldflags "-X github.com/jmylchreest/tinct/internal/version.Version=x.y.z".
	Version = "dev"

	// Commit is the git commit hash of the build.
	// Injected at build time via: -ldflags "-X github.com/jmylchreest/tinct/internal/version.Commit=$(git rev-parse HEAD)".
	Commit = "unknown"

	// Date is the build date in RFC3339 format.
	// Injected at build time via: -ldflags "-X github.com/jmylchreest/tinct/internal/version.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)".
	Date = "unknown"

	// GoVersion is the Go version used to build the binary.
	GoVersion = runtime.Version()
)

// Info holds all version information for the application.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// GetInfo returns all version information as a structured type.
func GetInfo() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		GoVersion: GoVersion,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a human-readable version string.
func String() string {
	info := GetInfo()
	if Commit != "unknown" && Date != "unknown" {
		return fmt.Sprintf("tinct version %s (commit: %s, built: %s, %s, %s)",
			info.Version, info.Commit[:8], info.Date, info.GoVersion, info.Platform)
	}
	return fmt.Sprintf("tinct version %s (%s, %s)", info.Version, info.GoVersion, info.Platform)
}

// Short returns a short version string suitable for CLI output.
func Short() string {
	return Version
}
