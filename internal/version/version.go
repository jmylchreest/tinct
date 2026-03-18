// Package version provides build-time version information for Tinct.
// Version information is injected at build time using ldflags.
package version

import (
	"fmt"
	"os"
	"runtime"
)

var (
	// Version is the semantic version of the application.
	// Injected at build time via: -ldflags "-X github.com/jmylchreest/tinct/internal/version.Version=x.y.z".
	Version = "0.0.0"

	// Commit is the git commit hash of the build.
	// Injected at build time via: -ldflags "-X github.com/jmylchreest/tinct/internal/version.Commit=$(git rev-parse HEAD)".
	Commit = "unknown"

	// Date is the build date in RFC3339 format.
	// Injected at build time via: -ldflags "-X github.com/jmylchreest/tinct/internal/version.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)".
	Date = "unknown"

	// TelemetryAppKey is the statsfactory API key for telemetry.
	// Injected at build time via: -ldflags "-X github.com/jmylchreest/tinct/internal/version.TelemetryAppKey=sf_live_...".
	// Can be overridden at runtime by the TINCT_TELEMETRY_APP_KEY environment variable.
	TelemetryAppKey = ""

	// TelemetryServerURL is the statsfactory server URL for telemetry.
	// Injected at build time via: -ldflags "-X github.com/jmylchreest/tinct/internal/version.TelemetryServerURL=https://...".
	// Can be overridden at runtime by the TINCT_TELEMETRY_SERVER_URL environment variable.
	TelemetryServerURL = ""

	// GoVersion is the Go version used to build the binary.
	GoVersion = runtime.Version()
)

// GetTelemetryAppKey returns the telemetry app key, preferring the environment
// variable TINCT_TELEMETRY_APP_KEY over the build-time injected value.
func GetTelemetryAppKey() string {
	if v := os.Getenv("TINCT_TELEMETRY_APP_KEY"); v != "" {
		return v
	}
	return TelemetryAppKey
}

// GetTelemetryServerURL returns the telemetry server URL, preferring the environment
// variable TINCT_TELEMETRY_SERVER_URL over the build-time injected value.
func GetTelemetryServerURL() string {
	if v := os.Getenv("TINCT_TELEMETRY_SERVER_URL"); v != "" {
		return v
	}
	return TelemetryServerURL
}

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
