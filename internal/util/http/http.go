// Package http provides HTTP utilities for fetching remote resources.
package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jmylchreest/tinct/internal/version"
)

const (
	// UserAgentName is the application name used in the User-Agent header.
	UserAgentName = "tinct"

	// DefaultTimeout is the default HTTP request timeout.
	DefaultTimeout = 10 * time.Second

	// DefaultMaxResponseBytes is the maximum response body size (500 MB).
	// Prevents memory exhaustion from malicious or unexpectedly large responses.
	DefaultMaxResponseBytes int64 = 500 * 1024 * 1024
)

// FetchOptions configures HTTP fetch behavior.
type FetchOptions struct {
	// Timeout specifies the HTTP request timeout.
	// If zero, DefaultTimeout is used.
	Timeout time.Duration

	// Headers specifies additional HTTP headers to send with the request.
	Headers map[string]string

	// ProgressCallback is called periodically with download progress.
	// Arguments are: bytesDownloaded, totalBytes (or -1 if unknown).
	ProgressCallback func(current, total int64)

	// MaxResponseBytes limits the maximum response body size.
	// If zero, DefaultMaxResponseBytes is used.
	// Set to -1 to disable the limit (not recommended).
	MaxResponseBytes int64
}

// Fetch retrieves content from a URL with context and timeout support.
// It automatically sets the User-Agent header and handles common HTTP errors.
func Fetch(ctx context.Context, url string, opts FetchOptions) ([]byte, error) {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent with dynamic version
	userAgent := fmt.Sprintf("%s/%s", UserAgentName, version.Version)
	req.Header.Set("User-Agent", userAgent)

	// Set additional headers
	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req) // #nosec G704 -- URL is validated by the caller
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Determine max response size.
	maxBytes := opts.MaxResponseBytes
	if maxBytes == 0 {
		maxBytes = DefaultMaxResponseBytes
	}

	// Wrap body in a size-limited reader to prevent memory exhaustion.
	// A value of -1 disables the limit (not recommended).
	body := resp.Body
	if maxBytes > 0 {
		body = io.NopCloser(io.LimitReader(resp.Body, maxBytes+1))
	}

	// If progress callback is provided, use it to track download progress
	if opts.ProgressCallback != nil {
		return fetchWithProgress(resp, body, maxBytes, opts.ProgressCallback)
	}

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if maxBytes > 0 && int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("response body exceeds maximum size limit (%d bytes)", maxBytes)
	}

	return data, nil
}

// fetchWithProgress reads the response body while calling the progress callback.
// body should already be wrapped in a size-limited reader if maxBytes > 0.
func fetchWithProgress(resp *http.Response, body io.Reader, maxBytes int64, callback func(current, total int64)) ([]byte, error) {
	totalBytes := resp.ContentLength // -1 if unknown

	// Use a buffer to collect data while tracking progress
	buf := make([]byte, 32*1024) // 32KB chunks
	var data []byte
	var downloaded int64

	for {
		n, err := body.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
			downloaded += int64(n)
			callback(downloaded, totalBytes)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
	}

	if maxBytes > 0 && int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("response body exceeds maximum size limit (%d bytes)", maxBytes)
	}

	return data, nil
}
