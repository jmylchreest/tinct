// Package telemetry provides anonymous, privacy-first usage analytics for Tinct.
//
// Telemetry is opt-out (enabled by default) and can be disabled by:
//   - Setting the environment variable TINCT_TELEMETRY=off
//   - Setting enabled = false in ~/.config/tinct/tinct.toml under [telemetry]
//
// Data collected is fully anonymous: a random UUID is generated on first run,
// SHA256-hashed before transmission, and never correlated with any personal information.
// The telemetry backend is statsfactory (https://github.com/jmylchreest/statsfactory),
// a privacy-first analytics platform.
package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	statsfactory "github.com/jmylchreest/statsfactory/packages/sdk-go"
	"github.com/jmylchreest/tinct/internal/config"
	"github.com/jmylchreest/tinct/internal/version"
)

const (
	// httpTimeout is the maximum time to wait for a single HTTP request.
	httpTimeout = 5 * time.Second

	// flushInterval is how often the SDK background worker flushes queued events.
	// Set short since tinct is a CLI tool that exits quickly.
	flushInterval = 1 * time.Second
)

// Client sends telemetry events to statsfactory.
//
// Events are queued in memory and dispatched by the statsfactory SDK's background
// worker. The SDK batches events (up to 25 per request) and flushes periodically.
//
// Call Flush(ctx) or Close() before the process exits to drain the queue.
// All methods are safe for concurrent use.
type Client struct {
	sf          *statsfactory.Client
	disabled    bool
	verbose     bool
	headerShown bool
	mu          sync.RWMutex
}

// Option configures optional Client behaviour.
type Option func(*Client)

// WithVerbose enables debug logging of telemetry events to stderr.
func WithVerbose(v bool) Option {
	return func(c *Client) { c.verbose = v }
}

// logf prints a debug message to stderr when verbose mode is enabled.
// On the first call it prints the "→ Sending telemetry..." section header.
func (c *Client) logf(format string, args ...any) {
	if c.verbose {
		if !c.headerShown {
			fmt.Fprintf(os.Stderr, "→ Sending telemetry...\n")
			c.headerShown = true
		}
		fmt.Fprintf(os.Stderr, "   "+format+"\n", args...)
	}
}

// New creates a new telemetry client backed by the statsfactory SDK.
//
// It loads the config from tinct.toml (with env var overrides already applied).
// If telemetry is disabled, or if the server URL / app key are not configured,
// the client becomes a no-op.
// This function never returns an error; failures are handled silently.
func New(opts ...Option) *Client {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}

	// Check if telemetry credentials are available.
	serverURL := version.GetTelemetryServerURL()
	appKey := version.GetTelemetryAppKey()
	if serverURL == "" || appKey == "" {
		c.disabled = true
		c.logf("disabled: missing credentials (app_key=%q, server_url=%q)", appKey, serverURL)
		return c
	}

	// Load config (env overrides are already applied by config.Load).
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		c.disabled = true
		c.logf("disabled: config load failed (err=%v)", err)
		return c
	}

	if !cfg.Telemetry.Enabled {
		c.disabled = true
		c.logf("disabled: telemetry.enabled=false in config")
		return c
	}

	// Eagerly initialise the installation ID so the telemetry.id file is
	// created on first run even if the caller never calls Send().
	_ = config.InstallationID()

	sfConfig := statsfactory.Config{
		ServerURL:     serverURL,
		AppKey:        appKey,
		ClientName:    "tinct",
		ClientVersion: version.Version,
		FlushInterval: flushInterval,
		// SessionID left empty — the SDK generates a unique session per client
		// instance. Since tinct creates one client per `generate` invocation,
		// this naturally correlates generate + plugin_used events from the same run.
		HTTPClient: &http.Client{
			Timeout: httpTimeout,
		},
	}

	if c.verbose {
		sfConfig.OnError = func(err error) {
			fmt.Fprintf(os.Stderr, "   SDK error: %v\n", err)
		}
	}

	c.sf = statsfactory.New(sfConfig)

	return c
}

// NewWithClient creates a Client wrapping a pre-configured statsfactory.Client.
// Used by tests that need to inject a mock server.
func NewWithClient(sf *statsfactory.Client) *Client {
	return &Client{sf: sf}
}

// IsEnabled returns whether telemetry is currently active.
func (c *Client) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.disabled
}

// Send enqueues a telemetry event for asynchronous dispatch.
// It returns immediately; the event is sent by the SDK's background worker.
// If the client is disabled the event is silently dropped.
func (c *Client) Send(event Event) {
	if !c.IsEnabled() {
		return
	}

	dims := make(statsfactory.Dims, len(event.Props))
	for k, v := range event.Props {
		dims[k] = v
	}

	distinctID := config.InstallationID()

	if c.verbose {
		if b, err := json.Marshal(dims); err == nil {
			c.logf("send %q: distinct_id=%s dims=%s", event.Name, distinctID, string(b))
		}
	}

	c.sf.TrackWithOptions(event.Name, dims, statsfactory.TrackOptions{
		DistinctID: distinctID,
	})
}

// Flush sends all queued events and waits until complete or ctx expires.
// After Flush returns the client can still accept new events (unlike Close).
func (c *Client) Flush(ctx context.Context) {
	if !c.IsEnabled() {
		return
	}

	c.logf("flushing queued events...")

	// The statsfactory SDK's Flush blocks until all queued events are sent
	// or the context is cancelled. We ignore the error since telemetry is
	// fire-and-forget.
	err := c.sf.Flush(ctx)
	if err != nil {
		c.logf("flush error: %v", err)
	} else {
		c.logf("flush complete")
	}
}

// Close flushes remaining events and shuts down the background worker.
// After Close, Send calls are silently dropped.
func (c *Client) Close() {
	if !c.IsEnabled() {
		return
	}

	c.logf("closing client")
	_ = c.sf.Close()
}
