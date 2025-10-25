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

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}
