package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultResponseLimit int64 = 64 << 20

// APIError reports an HTTP failure without including response bodies or credentials.
type APIError struct{ StatusCode int }

func (e *APIError) Error() string { return fmt.Sprintf("AI endpoint returned HTTP %d", e.StatusCode) }

func (o Options) deadline(ctx context.Context) (context.Context, context.CancelFunc) {
	timeout := o.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	return context.WithTimeout(ctx, timeout)
}
func (o Options) endpoint(fallback, path string) string {
	base := o.BaseURL
	if base == "" {
		base = fallback
	}
	return strings.TrimRight(base, "/") + path
}
func (o Options) request(ctx context.Context, method, endpoint string, payload any, headers map[string]string) ([]byte, error) {
	ctx, cancel := o.deadline(ctx)
	defer cancel()
	var body io.Reader = http.NoBody
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := o.Client
	if client == nil {
		client = &http.Client{}
	}
	// Copy the client to avoid mutating a caller-owned client. Do not redirect
	// credentials or replay a paid generation request at another endpoint.
	local := *client
	local.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := local.Do(req) // #nosec G704 -- endpoint is explicitly supplied by the caller, including local servers.
	if err != nil {
		// net/http includes the full URL in transport errors. Download URLs may
		// contain signed credentials, so retain the cause without the URL.
		if urlErr, ok := errors.AsType[*url.Error](err); ok {
			err = urlErr.Err
		}
		return nil, fmt.Errorf("AI request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{StatusCode: resp.StatusCode}
	}
	limit := o.MaxResponseBytes
	if limit <= 0 {
		limit = defaultResponseLimit
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("AI response exceeds %d bytes", limit)
	}
	return data, nil
}
func (o Options) post(ctx context.Context, endpoint string, payload, result any, headers map[string]string) error {
	data, err := o.request(ctx, http.MethodPost, endpoint, payload, headers)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
func (o Options) bearer() map[string]string {
	h := map[string]string{}
	if o.APIKey != "" {
		h["Authorization"] = "Bearer " + o.APIKey
	}
	return h
}
func decodeImage(encoded, mime string) (Image, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return Image{}, fmt.Errorf("decode image: %w", err)
	}
	if len(data) == 0 {
		return Image{}, fmt.Errorf("empty image in response")
	}
	if mime == "" {
		mime = http.DetectContentType(data)
	}
	return Image{Bytes: data, MIMEType: mime}, nil
}
func (o Options) download(ctx context.Context, imageURL string) (Image, error) {
	data, err := o.request(ctx, http.MethodGet, imageURL, nil, nil)
	if err != nil {
		return Image{}, err
	}
	if len(data) == 0 {
		return Image{}, fmt.Errorf("empty image in response")
	}
	return Image{Bytes: data, MIMEType: http.DetectContentType(data)}, nil
}
