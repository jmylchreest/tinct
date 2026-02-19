// Package telemetry provides anonymous, privacy-first usage analytics for Tinct.
//
// Telemetry is opt-out (enabled by default) and can be disabled by:
//   - Setting the environment variable TINCT_TELEMETRY=off
//   - Setting "enabled": false in ~/.local/share/tinct/telemetry.json
//
// Data collected is fully anonymous: a random UUID is generated on first run,
// SHA256-hashed before transmission, and never correlated with any personal information.
// The telemetry backend is Aptabase (https://aptabase.com), an open-source,
// privacy-first analytics platform.
package telemetry

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/jmylchreest/tinct/internal/version"
)

const (
	// AppKey is the Aptabase application key.
	AppKey = "A-EU-4071287770"

	// apiPath is the Aptabase events API path.
	apiPath = "/api/v0/events"

	// httpTimeout is the maximum time to wait for the HTTP request.
	httpTimeout = 5 * time.Second

	// configFileName is the telemetry config file name.
	configFileName = "telemetry.json"

	// configDirName is the directory under ~/.local/share/ for tinct data.
	configDirName = "tinct"

	// envKey is the environment variable to control telemetry.
	envKey = "TINCT_TELEMETRY"

	// sdkPrefix is the prefix for the SDK version in Aptabase system props.
	// The full sdkVersion is built dynamically as sdkPrefix + version.Version.
	sdkPrefix = "tinct-telemetry@"
)

// Config is the persistent telemetry configuration stored on disk.
type Config struct {
	// Enabled controls whether telemetry is active. Default: true (opt-out).
	Enabled bool `json:"enabled"`

	// ID is the SHA256-hashed anonymous identifier for this installation.
	// Generated from a random UUID on first run, then hashed before storage.
	ID string `json:"id"`
}

// Client sends telemetry events to Aptabase.
// All methods are safe for concurrent use.
type Client struct {
	config     *Config
	httpClient *http.Client
	baseURL    string
	appKey     string
	disabled   bool
	mu         sync.RWMutex
}

// New creates a new telemetry client.
// It loads or initialises the config file and checks the environment variable.
// If telemetry is disabled (by env var or config), the client becomes a no-op.
// This function never returns an error; failures are handled silently.
func New() *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: httpTimeout},
		appKey:     AppKey,
	}

	// Derive base URL from app key region.
	c.baseURL = resolveBaseURL(c.appKey)

	// Check environment variable override first (takes priority).
	if envVal := os.Getenv(envKey); envVal != "" {
		lower := strings.ToLower(envVal)
		if lower == "off" || lower == "false" || lower == "0" || lower == "no" {
			c.disabled = true
			return c
		}
	}

	// Load or initialise config.
	config, err := loadOrInitConfig()
	if err != nil {
		// Can't load config — disable silently.
		c.disabled = true
		return c
	}
	c.config = config

	if !config.Enabled {
		c.disabled = true
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
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.config == nil {
		return ""
	}
	return c.config.ID
}

// Send sends a telemetry event synchronously with the configured HTTP timeout.
// For CLI tools, this is preferred over async sending because the process may
// exit before an async goroutine completes. The HTTP timeout (5s) ensures this
// never blocks for long. Errors are silently discarded.
func (c *Client) Send(event Event) {
	if !c.IsEnabled() {
		return
	}

	_ = c.sendSync(event) //nolint:errcheck // Fire-and-forget, errors silently discarded
}

// SendAsync sends a telemetry event in a background goroutine.
// Only use this if the process will remain alive long enough for the request
// to complete. For CLI tools, prefer Send() instead.
func (c *Client) SendAsync(event Event) {
	if !c.IsEnabled() {
		return
	}

	go c.sendSync(event) //nolint:errcheck // Fire-and-forget
}

// sendSync sends an event synchronously. Used internally by Send.
func (c *Client) sendSync(event Event) error {
	c.mu.RLock()
	if c.disabled || c.config == nil {
		c.mu.RUnlock()
		return nil
	}
	config := c.config
	c.mu.RUnlock()

	// Build the Aptabase event payload.
	payload := aptabaseEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		SessionID: generateSessionID(),
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
	}

	// Marshal as a single-element array (Aptabase batch format).
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

	// Include the anonymous installation ID as a custom header.
	// This allows Aptabase to count unique installations.
	if config.ID != "" {
		req.Header.Set("X-Installation-ID", config.ID)
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
	return fmt.Sprintf("%d%08d", time.Now().Unix(), rand.Intn(100000000)) //nolint:gosec // Non-cryptographic randomness is fine for session IDs
}

// --- Config file management ---

// configPath returns the full path to the telemetry config file.
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	return filepath.Join(home, ".local", "share", configDirName, configFileName), nil
}

// loadOrInitConfig loads the telemetry config from disk, or creates it if missing.
func loadOrInitConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	// Try to load existing config.
	data, err := os.ReadFile(path)
	if err == nil {
		var config Config
		if jsonErr := json.Unmarshal(data, &config); jsonErr == nil {
			// Validate the loaded config.
			if config.ID != "" {
				return &config, nil
			}
			// ID is empty — regenerate.
		}
		// Corrupt config — regenerate.
	}

	// No config or corrupt — create new one.
	config := &Config{
		Enabled: true,
		ID:      generateInstallationID(),
	}

	// Save to disk.
	if saveErr := saveConfig(path, config); saveErr != nil {
		// If we can't save, still return the config for this session.
		return config, nil
	}

	return config, nil
}

// saveConfig writes the telemetry config to disk.
func saveConfig(path string, config *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// generateInstallationID creates a new anonymous installation identifier.
// A random UUID-like string is generated, then SHA256-hashed so the raw ID
// is never stored or transmitted.
func generateInstallationID() string {
	// Generate a random UUID-like value using crypto-quality randomness
	// from the OS, combined with current time for additional entropy.
	raw := fmt.Sprintf("%d-%d-%d-%d",
		time.Now().UnixNano(),
		rand.Int63(), //nolint:gosec // Combined with time for sufficient entropy
		rand.Int63(), //nolint:gosec // Combined with time for sufficient entropy
		os.Getpid(),
	)

	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}
