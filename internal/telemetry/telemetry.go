// Package telemetry provides anonymous, privacy-first usage analytics for Tinct.
//
// Telemetry is opt-out (enabled by default) and can be disabled by:
//   - Setting the environment variable TINCT_TELEMETRY=off
//   - Setting enabled = false in ~/.config/tinct/tinct.toml under [telemetry]
//
// Data collected is fully anonymous: a random UUID is generated on first run,
// SHA256-hashed before transmission, and never correlated with any personal information.
// The telemetry backend is Aptabase (https://aptabase.com), an open-source,
// privacy-first analytics platform.
package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/jmylchreest/tinct/internal/config"
	"github.com/jmylchreest/tinct/internal/version"
)

const (
	// AppKey is the Aptabase application key.
	AppKey = "A-EU-4071287770"

	// apiPath is the Aptabase events API path.
	apiPath = "/api/v0/events"

	// httpTimeout is the maximum time to wait for a single HTTP request.
	httpTimeout = 5 * time.Second

	// sdkPrefix is the prefix for the SDK version in Aptabase system props.
	// The full sdkVersion is built dynamically as sdkPrefix + version.Version.
	sdkPrefix = "tinct-telemetry@"

	// maxBatchSize is the maximum number of events per Aptabase API request.
	maxBatchSize = 25

	// queueSize is the capacity of the in-memory event queue. Sized generously
	// so that Send() never blocks under normal usage.
	queueSize = 128

	// debounceDelay is how long the worker waits after receiving the first event
	// before sending, to allow subsequent events to accumulate into one batch.
	debounceDelay = 5 * time.Millisecond
)

// telemetryConfig holds the resolved telemetry settings for the Client.
type telemetryConfig struct {
	Enabled bool
}

// Client sends telemetry events to Aptabase.
//
// Events are queued in memory and dispatched by a background worker goroutine.
// The worker uses a short debounce window so that a burst of Send() calls
// (e.g. one summary + N plugin events) is batched into as few HTTP requests
// as possible (Aptabase accepts up to 25 events per request).
//
// Call Flush(ctx) before the process exits to drain the queue.
// All methods are safe for concurrent use.
type Client struct {
	config     *telemetryConfig
	httpClient *http.Client
	baseURL    string
	appKey     string
	sessionID  string
	disabled   bool

	// queue receives events from Send(). Closed by Flush() to signal the worker to stop.
	queue chan Event
	// done is closed by the worker when it has finished draining the queue.
	done chan struct{}

	mu sync.RWMutex
}

// New creates a new telemetry client with a single session ID that is shared
// across all events sent by this client. For CLI tools, one Client per command
// invocation means one Aptabase session — all events (e.g. a "generate" summary
// plus individual "plugin_used" events) are grouped together.
//
// It loads the config from tinct.toml (with env var overrides already applied).
// If telemetry is disabled, the client becomes a no-op.
// This function never returns an error; failures are handled silently.
func New() *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: httpTimeout},
		appKey:     AppKey,
		sessionID:  generateSessionID(),
		queue:      make(chan Event, queueSize),
		done:       make(chan struct{}),
	}

	// Derive base URL from app key region.
	c.baseURL = resolveBaseURL(c.appKey)

	// Load config (env overrides are already applied by config.Load).
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		// Can't load config — disable silently.
		c.disabled = true
		close(c.done) // nothing to drain
		return c
	}

	c.config = &telemetryConfig{
		Enabled: cfg.Telemetry.Enabled,
	}

	if !c.config.Enabled {
		c.disabled = true
		close(c.done) // nothing to drain
	} else {
		// Eagerly initialise the installation ID so the telemetry.id file is
		// created on first run even if the caller never calls ID() or Send().
		_ = config.InstallationID()

		// Start the background worker that drains the queue.
		go c.worker()
	}

	return c
}

// IsEnabled returns whether telemetry is currently active.
func (c *Client) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.disabled
}

// ID returns the anonymous installation identifier (SHA256 hash).
// Returns empty string if telemetry is disabled.
func (c *Client) ID() string {
	if !c.IsEnabled() {
		return ""
	}
	return config.InstallationID()
}

// Send enqueues a telemetry event for asynchronous dispatch.
// It returns immediately; the event is sent by the background worker.
// If the client is disabled, the queue is full, or Flush has already been
// called, the event is silently dropped.
func (c *Client) Send(event Event) {
	if !c.IsEnabled() {
		return
	}

	// If the worker has already exited (Flush was called), done is closed.
	// Attempting to send on a closed queue channel would panic, so we check
	// done first. Both channels are only ever closed, never re-opened, so
	// this select is race-free.
	select {
	case <-c.done:
		// Worker already stopped — drop silently.
		return
	default:
	}

	select {
	case c.queue <- event:
	default:
		// Queue full — drop silently rather than block the caller.
	}
}

// Flush closes the event queue and waits for the background worker to finish
// sending all queued events. ctx controls the maximum wait time.
//
// After Flush returns, the client is no longer usable; further Send() calls
// will panic on a closed channel (callers must not Send after Flush).
func (c *Client) Flush(ctx context.Context) {
	if !c.IsEnabled() {
		return
	}

	// Signal the worker that no more events are coming.
	close(c.queue)

	// Wait for it to finish or for the context to expire.
	select {
	case <-c.done:
	case <-ctx.Done():
	}
}

// worker is the background goroutine that drains the event queue.
//
// It waits for the first event, then starts a short debounce timer.
// While the timer runs it collects additional events. When the timer fires
// (or the batch reaches maxBatchSize), it sends the batch and loops back
// to wait for the next event. When the queue is closed (by Flush), it
// sends any remaining events and exits.
func (c *Client) worker() {
	defer close(c.done)

	for {
		// Block until the first event arrives or the queue is closed.
		event, ok := <-c.queue
		if !ok {
			// Queue closed with nothing pending — we're done.
			return
		}

		// Seed the current batch with this first event.
		batch := make([]Event, 0, maxBatchSize)
		batch = append(batch, event)

		// Start the debounce timer to let more events accumulate.
		timer := time.NewTimer(debounceDelay)

	collect:
		for len(batch) < maxBatchSize {
			select {
			case e, ok := <-c.queue:
				if !ok {
					// Queue closed mid-debounce — send what we have and exit.
					timer.Stop()
					c.sendBatch(batch)
					return
				}
				batch = append(batch, e)
			case <-timer.C:
				// Debounce window expired — send what we have.
				break collect
			}
		}
		timer.Stop()

		c.sendBatch(batch)

		// If the batch hit maxBatchSize the queue may still have events.
		// The outer for loop will pick them up immediately.
	}
}

// sendBatch sends a slice of events in a single Aptabase API request.
// Errors are silently discarded (fire-and-forget).
func (c *Client) sendBatch(events []Event) {
	if len(events) == 0 {
		return
	}

	payloads := make([]aptabaseEvent, 0, len(events))
	for _, event := range events {
		payloads = append(payloads, aptabaseEvent{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			SessionID: c.sessionID,
			EventName: event.Name,
			SystemProps: aptabaseSystemProps{
				OSName:     runtime.GOOS,
				OSVersion:  "", // Not collecting OS version for privacy.
				Locale:     "",
				IsDebug:    version.Version == "0.0.0",
				AppVersion: version.Version,
				SDKVersion: sdkPrefix + version.Version,
			},
			Props: event.Props,
		})
	}

	body, err := json.Marshal(payloads)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+apiPath, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("App-Key", c.appKey)

	// Include the anonymous installation ID as a custom header.
	// This allows Aptabase to count unique installations.
	if id := config.InstallationID(); id != "" {
		req.Header.Set("X-Installation-ID", id)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

// sendSync sends a single event synchronously. Used by tests that need to
// verify the HTTP payload without going through the async worker.
func (c *Client) sendSync(event Event) error {
	c.mu.RLock()
	if c.disabled || c.config == nil {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	payload := aptabaseEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		SessionID: c.sessionID,
		EventName: event.Name,
		SystemProps: aptabaseSystemProps{
			OSName:     runtime.GOOS,
			OSVersion:  "",
			Locale:     "",
			IsDebug:    version.Version == "0.0.0",
			AppVersion: version.Version,
			SDKVersion: sdkPrefix + version.Version,
		},
		Props: event.Props,
	}

	body, err := json.Marshal([]aptabaseEvent{payload})
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+apiPath, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("App-Key", c.appKey)

	if id := config.InstallationID(); id != "" {
		req.Header.Set("X-Installation-ID", id)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send event: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// aptabaseEvent is the Aptabase event payload format.
type aptabaseEvent struct {
	Timestamp   string              `json:"timestamp"`
	SessionID   string              `json:"sessionId"`
	EventName   string              `json:"eventName"`
	SystemProps aptabaseSystemProps `json:"systemProps"`
	Props       map[string]any      `json:"props,omitempty"`
}

// aptabaseSystemProps contains automatic system metadata.
type aptabaseSystemProps struct {
	Locale     string `json:"locale"`
	OSName     string `json:"osName"`
	OSVersion  string `json:"osVersion"`
	IsDebug    bool   `json:"isDebug"`
	AppVersion string `json:"appVersion"`
	SDKVersion string `json:"sdkVersion"`
}

// resolveBaseURL determines the Aptabase API base URL from the app key.
// App keys are formatted as A-{REGION}-{ID}.
func resolveBaseURL(appKey string) string {
	parts := strings.SplitN(appKey, "-", 3)
	if len(parts) >= 2 {
		switch strings.ToUpper(parts[1]) {
		case "EU":
			return "https://eu.aptabase.com"
		case "US":
			return "https://us.aptabase.com"
		case "SH":
			// Self-hosted — needs custom base URL override.
			// For now, fall through to default.
		}
	}
	return "https://eu.aptabase.com"
}

// generateSessionID creates an Aptabase session ID.
// Format: epoch seconds + 8 random digits.
func generateSessionID() string {
	return fmt.Sprintf("%d%08d", time.Now().Unix(), rand.Intn(100000000)) //nolint:gosec // #nosec G404 -- non-cryptographic randomness is fine for ephemeral telemetry session IDs
}
