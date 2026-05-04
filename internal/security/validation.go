// Package security provides security validation utilities for Tinct.
package security

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// ValidateGitURL validates a git repository URL for safe cloning.
// Only allows https:// and git:// protocols from trusted domains.
func ValidateGitURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("empty git URL")
	}

	// Parse the URL
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid git URL: %w", err)
	}

	// Only allow specific protocols
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "https" && scheme != "git" {
		return fmt.Errorf("invalid git URL protocol (only https:// and git:// allowed): %s", scheme)
	}

	// Require a host
	if parsed.Host == "" {
		return fmt.Errorf("git URL must have a hostname")
	}

	// Block localhost and private IPs to prevent SSRF
	host := strings.ToLower(parsed.Hostname())
	if isLocalOrPrivateHost(host) {
		return fmt.Errorf("git URL cannot point to local or private hosts: %s", host)
	}

	return nil
}

// ValidateHTTPURL validates an HTTP(S) URL for safe downloads.
// Only allows HTTPS from non-local hosts.
func ValidateHTTPURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("empty URL")
	}

	// Parse the URL
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Only allow HTTPS (not HTTP)
	if !strings.EqualFold(parsed.Scheme, "https") {
		return fmt.Errorf("only HTTPS URLs are allowed (got %s)", parsed.Scheme)
	}

	// Require a host
	if parsed.Host == "" {
		return fmt.Errorf("URL must have a hostname")
	}

	// Block localhost and private IPs to prevent SSRF
	host := strings.ToLower(parsed.Hostname())
	if isLocalOrPrivateHost(host) {
		return fmt.Errorf("URL cannot point to local or private hosts: %s", host)
	}

	return nil
}

// ValidateRuntimeHTTPURL validates a URL used for runtime data fetching
// (e.g. remote-css, remote-json, wallpaper URLs in theme files).
//
// By default it enforces the same rules as ValidateHTTPURL: HTTPS-only,
// no localhost/private IPs. If the TINCT_ALLOW_INSECURE_HTTP environment
// variable is set to "1" or "true", plain HTTP is also allowed but private
// IP / localhost blocking still applies.
func ValidateRuntimeHTTPURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("empty URL")
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	allowInsecure := isInsecureHTTPAllowed()

	switch {
	case scheme == "https":
		// OK
	case scheme == "http" && allowInsecure:
		// OK — env var permits plain HTTP
	case scheme == "http":
		return fmt.Errorf("only HTTPS URLs are allowed (got http); set TINCT_ALLOW_INSECURE_HTTP=1 to permit plain HTTP")
	default:
		return fmt.Errorf("unsupported URL scheme: %s", scheme)
	}

	if parsed.Host == "" {
		return fmt.Errorf("URL must have a hostname")
	}

	host := strings.ToLower(parsed.Hostname())
	if isLocalOrPrivateHost(host) {
		return fmt.Errorf("URL cannot point to local or private hosts: %s", host)
	}

	return nil
}

// isInsecureHTTPAllowed checks whether the TINCT_ALLOW_INSECURE_HTTP
// environment variable permits plain HTTP URLs.
func isInsecureHTTPAllowed() bool {
	v := strings.ToLower(os.Getenv("TINCT_ALLOW_INSECURE_HTTP"))
	return v == "1" || v == "true"
}

// ValidateFilePath validates a file path within an archive to prevent directory traversal.
func ValidateFilePath(filePath, baseDir string) error {
	if filePath == "" {
		return fmt.Errorf("empty file path")
	}

	// Check for dangerous patterns
	if strings.Contains(filePath, "..") {
		return fmt.Errorf("file path contains directory traversal (..) - not allowed")
	}

	if filepath.IsAbs(filePath) {
		return fmt.Errorf("absolute paths in archives are not allowed")
	}

	// Ensure the final path would be within baseDir
	finalPath := filepath.Join(baseDir, filePath)
	cleanFinal := filepath.Clean(finalPath)
	cleanBase := filepath.Clean(baseDir)

	if !strings.HasPrefix(cleanFinal, cleanBase+string(filepath.Separator)) &&
		cleanFinal != cleanBase {

		return fmt.Errorf("file path would escape base directory")
	}

	return nil
}

// LimitedReader wraps an io.Reader and limits the total bytes that can be read.
// This prevents decompression bomb attacks when extracting archives.
type LimitedReader struct {
	R         io.Reader
	Remaining int64
}

// Read implements io.Reader with size limits.
func (l *LimitedReader) Read(p []byte) (int, error) {
	if l.Remaining <= 0 {
		return 0, fmt.Errorf("decompression size limit exceeded")
	}
	if int64(len(p)) > l.Remaining {
		p = p[:l.Remaining]
	}
	n, err := l.R.Read(p)
	l.Remaining -= int64(n)
	return n, err
}

// NewLimitedReader creates a new LimitedReader with the specified size limit.
func NewLimitedReader(r io.Reader, maxBytes int64) *LimitedReader {
	return &LimitedReader{
		R:         r,
		Remaining: maxBytes,
	}
}

// isLocalOrPrivateHost checks if a hostname is localhost or a private IP.
func isLocalOrPrivateHost(host string) bool { //nolint:gocyclo
	// Check for localhost variations
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}

	// Check for private IP ranges
	if strings.HasPrefix(host, "192.168.") ||
		strings.HasPrefix(host, "10.") ||
		strings.HasPrefix(host, "172.16.") ||
		strings.HasPrefix(host, "172.17.") ||
		strings.HasPrefix(host, "172.18.") ||
		strings.HasPrefix(host, "172.19.") ||
		strings.HasPrefix(host, "172.20.") ||
		strings.HasPrefix(host, "172.21.") ||
		strings.HasPrefix(host, "172.22.") ||
		strings.HasPrefix(host, "172.23.") ||
		strings.HasPrefix(host, "172.24.") ||
		strings.HasPrefix(host, "172.25.") ||
		strings.HasPrefix(host, "172.26.") ||
		strings.HasPrefix(host, "172.27.") ||
		strings.HasPrefix(host, "172.28.") ||
		strings.HasPrefix(host, "172.29.") ||
		strings.HasPrefix(host, "172.30.") ||
		strings.HasPrefix(host, "172.31.") ||
		strings.HasPrefix(host, "169.254.") {

		return true
	}

	// Check for link-local IPv6
	if strings.HasPrefix(host, "fe80:") || strings.HasPrefix(host, "fc00:") || strings.HasPrefix(host, "fd00:") {
		return true
	}

	return false
}
