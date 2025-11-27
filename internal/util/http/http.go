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

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// If progress callback is provided, use it to track download progress
	if opts.ProgressCallback != nil {
		return fetchWithProgress(resp, opts.ProgressCallback)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

// fetchWithProgress reads the response body while calling the progress callback.
func fetchWithProgress(resp *http.Response, callback func(current, total int64)) ([]byte, error) {
	totalBytes := resp.ContentLength // -1 if unknown

	// Use a buffer to collect data while tracking progress
	buf := make([]byte, 32*1024) // 32KB chunks
	var data []byte
	var downloaded int64

	for {
		n, err := resp.Body.Read(buf)
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

	return data, nil
}
