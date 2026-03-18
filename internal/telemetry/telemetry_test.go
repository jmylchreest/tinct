package telemetry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	statsfactory "github.com/jmylchreest/statsfactory/packages/sdk-go"
	"github.com/jmylchreest/tinct/internal/config"
)

// setupTestHome creates a temp dir and sets HOME + XDG_CONFIG_HOME so that
// config.Load() writes to a predictable location. Returns the temp dir path.
// Also resets the config singleton so each test gets a fresh load.
func setupTestHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, ".config"))
	config.ResetForTesting()
	return dir
}

// setTelemetryCredentials sets env vars so New() has valid credentials.
func setTelemetryCredentials(t *testing.T, serverURL string) {
	t.Helper()
	t.Setenv("TINCT_TELEMETRY_APP_KEY", "sf_test_key")
	t.Setenv("TINCT_TELEMETRY_SERVER_URL", serverURL)
}

// --- statsfactory wire format types for test assertions ---

// sfIngestRequest is the JSON body sent to POST /v1/events.
type sfIngestRequest struct {
	Events []sfEvent `json:"events"`
}

// sfEvent is a single event in the statsfactory wire format.
type sfEvent struct {
	Event      string         `json:"event"`
	EventKey   string         `json:"event_key,omitempty"`
	Timestamp  string         `json:"timestamp,omitempty"`
	SessionID  string         `json:"session_id,omitempty"`
	DistinctID string         `json:"distinct_id,omitempty"`
	Dimensions map[string]any `json:"dimensions,omitempty"`
}

// capturedRequest holds the data from a single request to the mock server.
type capturedRequest struct {
	Method    string
	Path      string
	UserAgent string
	AuthToken string
	Body      sfIngestRequest
}

// mockServer creates an httptest server that captures requests and returns 200.
func mockServer(t *testing.T) (*httptest.Server, *[]capturedRequest) {
	t.Helper()
	var mu sync.Mutex
	var captured []capturedRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body sfIngestRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("mock server: decode error: %v", err)
			w.WriteHeader(400)
			return
		}

		mu.Lock()
		captured = append(captured, capturedRequest{
			Method:    r.Method,
			Path:      r.URL.Path,
			UserAgent: r.Header.Get("User-Agent"),
			AuthToken: r.Header.Get("Authorization"),
			Body:      body,
		})
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]any{
			"accepted": len(body.Events),
		})
	}))

	t.Cleanup(srv.Close)
	return srv, &captured
}

func getCaptured(captured *[]capturedRequest) []capturedRequest {
	return *captured
}

// --- Tests for New() enable/disable logic ---

func TestClientDisabledByEnvVar(t *testing.T) {
	setupTestHome(t)
	srv, _ := mockServer(t)
	setTelemetryCredentials(t, srv.URL)
	t.Setenv("TINCT_TELEMETRY", "off")

	client := New()
	if client.IsEnabled() {
		t.Error("expected client to be disabled when TINCT_TELEMETRY=off")
	}
}

func TestClientEnabledByDefault(t *testing.T) {
	setupTestHome(t)
	srv, _ := mockServer(t)
	setTelemetryCredentials(t, srv.URL)
	t.Setenv("TINCT_TELEMETRY", "") // clear any override

	client := New()
	defer client.Close()

	if !client.IsEnabled() {
		t.Error("expected client to be enabled by default")
	}
}

func TestClientDisabledByConfig(t *testing.T) {
	dir := setupTestHome(t)
	srv, _ := mockServer(t)
	setTelemetryCredentials(t, srv.URL)

	// Write config with telemetry disabled.
	configDir := filepath.Join(dir, ".config", "tinct")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "tinct.toml"), []byte("[telemetry]\nenabled = false\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	client := New()
	if client.IsEnabled() {
		t.Error("expected client to be disabled when config says enabled=false")
	}
}

func TestClientDisabledWithoutCredentials(t *testing.T) {
	setupTestHome(t)
	// Don't set credentials — should be disabled.
	t.Setenv("TINCT_TELEMETRY_APP_KEY", "")
	t.Setenv("TINCT_TELEMETRY_SERVER_URL", "")

	client := New()
	if client.IsEnabled() {
		t.Error("expected client to be disabled when credentials are missing")
	}
}

func TestClientDisabledWithPartialCredentials(t *testing.T) {
	setupTestHome(t)
	t.Setenv("TINCT_TELEMETRY_APP_KEY", "sf_test_key")
	t.Setenv("TINCT_TELEMETRY_SERVER_URL", "") // missing server URL

	client := New()
	if client.IsEnabled() {
		t.Error("expected client to be disabled when server URL is missing")
	}
}

func TestEnvVarOverridesConfig(t *testing.T) {
	dir := setupTestHome(t)
	srv, _ := mockServer(t)
	setTelemetryCredentials(t, srv.URL)

	// Config says enabled, env says off.
	configDir := filepath.Join(dir, ".config", "tinct")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "tinct.toml"), []byte("[telemetry]\nenabled = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TINCT_TELEMETRY", "off")

	client := New()
	if client.IsEnabled() {
		t.Error("expected TINCT_TELEMETRY=off to override config")
	}
}

// --- Tests for Send/Flush via mock server ---

func TestSendAndFlushDispatchesEvents(t *testing.T) {
	srv, captured := mockServer(t)

	sfClient := statsfactory.New(statsfactory.Config{
		ServerURL:     srv.URL,
		AppKey:        "sf_test_key",
		ClientName:    "tinct",
		ClientVersion: "0.0.0-test",
		FlushInterval: time.Hour, // don't auto-flush
	})

	client := NewWithClient(sfClient)

	client.Send(NewEvent("test_event").Set("key", "value"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client.Flush(ctx)
	client.Close()

	reqs := getCaptured(captured)
	if len(reqs) == 0 {
		t.Fatal("expected at least one request to the mock server")
	}

	// Find the event across all requests.
	var found bool
	for _, req := range reqs {
		if req.Method != "POST" {
			t.Errorf("Method = %q, want POST", req.Method)
		}
		if req.Path != "/v1/events" {
			t.Errorf("Path = %q, want /v1/events", req.Path)
		}
		if req.AuthToken != "Bearer sf_test_key" {
			t.Errorf("Auth = %q, want Bearer sf_test_key", req.AuthToken)
		}
		for _, ev := range req.Body.Events {
			if ev.Event == "test_event" {
				found = true
				if ev.Dimensions["key"] != "value" {
					t.Errorf("dimension key = %v, want value", ev.Dimensions["key"])
				}
			}
		}
	}
	if !found {
		t.Error("test_event not found in captured requests")
	}
}

func TestSendDisabledClientIsNoop(t *testing.T) {
	srv, captured := mockServer(t)

	sfClient := statsfactory.New(statsfactory.Config{
		ServerURL:     srv.URL,
		AppKey:        "key",
		FlushInterval: time.Hour,
	})

	client := &Client{sf: sfClient, disabled: true}
	client.Send(NewEvent("should_not_send"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	client.Flush(ctx)

	// Close the underlying sf client to avoid goroutine leaks.
	sfClient.Close()

	reqs := getCaptured(captured)
	if len(reqs) != 0 {
		t.Errorf("expected no requests for disabled client, got %d", len(reqs))
	}
}

func TestSendMultipleEventsBatched(t *testing.T) {
	srv, captured := mockServer(t)

	sfClient := statsfactory.New(statsfactory.Config{
		ServerURL:     srv.URL,
		AppKey:        "key",
		FlushInterval: time.Hour,
	})

	client := NewWithClient(sfClient)

	// Send a generate + 3 plugin_used events (typical usage).
	client.Send(NewGenerateEvent(GenerateEventParams{
		InputPlugin:   "image",
		OutputPlugins: []string{"kitty", "waybar", "gtk4"},
		ThemeType:     "dark",
	}))
	client.Send(NewPluginUsedEvent("kitty", "0.1.27", false, "ok"))
	client.Send(NewPluginUsedEvent("waybar", "0.1.27", false, "ok"))
	client.Send(NewPluginUsedEvent("gtk4", "0.1.27", false, "failed"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client.Flush(ctx)
	client.Close()

	reqs := getCaptured(captured)
	totalEvents := 0
	for _, req := range reqs {
		totalEvents += len(req.Body.Events)
	}
	if totalEvents != 4 {
		t.Errorf("total events = %d, want 4", totalEvents)
	}
}

func TestWireFormatDotNotation(t *testing.T) {
	srv, captured := mockServer(t)

	sfClient := statsfactory.New(statsfactory.Config{
		ServerURL:     srv.URL,
		AppKey:        "key",
		FlushInterval: time.Hour,
	})

	client := NewWithClient(sfClient)

	client.Send(NewPluginUsedEvent("kitty", "0.1.27", false, "ok"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client.Flush(ctx)
	client.Close()

	reqs := getCaptured(captured)
	if len(reqs) == 0 {
		t.Fatal("expected at least one request")
	}

	ev := reqs[0].Body.Events[0]
	if ev.Event != "plugin_used" {
		t.Errorf("event = %q, want plugin_used", ev.Event)
	}

	// Verify dot-notation dimension keys.
	assertDim := func(key string, want any) {
		t.Helper()
		got, ok := ev.Dimensions[key]
		if !ok {
			t.Errorf("missing dimension %q", key)
			return
		}
		// JSON numbers decode as float64, bools stay bool.
		switch w := want.(type) {
		case bool:
			if g, ok := got.(bool); !ok || g != w {
				t.Errorf("dim %q = %v (%T), want %v", key, got, got, want)
			}
		case string:
			if g, ok := got.(string); !ok || g != w {
				t.Errorf("dim %q = %v (%T), want %v", key, got, got, want)
			}
		}
	}

	assertDim("plugin.name", "kitty")
	assertDim("plugin.version", "0.1.27")
	assertDim("plugin.external", false)
	assertDim("plugin.status", "ok")
}

func TestWireFormatArrayDimension(t *testing.T) {
	srv, captured := mockServer(t)

	sfClient := statsfactory.New(statsfactory.Config{
		ServerURL:     srv.URL,
		AppKey:        "key",
		FlushInterval: time.Hour,
	})

	client := NewWithClient(sfClient)

	client.Send(NewGenerateEvent(GenerateEventParams{
		InputPlugin:   "image",
		OutputPlugins: []string{"kitty", "waybar", "gtk4"},
		ThemeType:     "dark",
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client.Flush(ctx)
	client.Close()

	reqs := getCaptured(captured)
	if len(reqs) == 0 {
		t.Fatal("expected at least one request")
	}

	ev := reqs[0].Body.Events[0]
	if ev.Event != "generate" {
		t.Errorf("event = %q, want generate", ev.Event)
	}

	// output.plugins should arrive as a JSON array ([]interface{} after decode).
	raw, ok := ev.Dimensions["output.plugins"]
	if !ok {
		t.Fatal("missing dimension output.plugins")
	}
	arr, ok := raw.([]interface{})
	if !ok {
		t.Fatalf("output.plugins is %T, want []interface{} (JSON array)", raw)
	}
	want := []string{"kitty", "waybar", "gtk4"}
	if len(arr) != len(want) {
		t.Fatalf("output.plugins has %d elements, want %d", len(arr), len(want))
	}
	for i, w := range want {
		if s, ok := arr[i].(string); !ok || s != w {
			t.Errorf("output.plugins[%d] = %v (%T), want %q", i, arr[i], arr[i], w)
		}
	}
}

func TestWireFormatSessionIDIsPerClient(t *testing.T) {
	srv, captured := mockServer(t)

	// Create two separate clients — each should get a distinct session_id.
	var sessionIDs [2]string
	for i := 0; i < 2; i++ {
		sfClient := statsfactory.New(statsfactory.Config{
			ServerURL:     srv.URL,
			AppKey:        "key",
			FlushInterval: time.Hour,
		})
		client := NewWithClient(sfClient)
		client.Send(NewEvent("test"))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		client.Flush(ctx)
		cancel()
		client.Close()
	}

	reqs := getCaptured(captured)
	if len(reqs) < 2 {
		t.Fatalf("expected at least 2 requests, got %d", len(reqs))
	}
	sessionIDs[0] = reqs[0].Body.Events[0].SessionID
	sessionIDs[1] = reqs[1].Body.Events[0].SessionID

	if sessionIDs[0] == "" {
		t.Error("session_id should not be empty")
	}
	if sessionIDs[0] == sessionIDs[1] {
		t.Errorf("expected different session_ids for different clients, both got %q", sessionIDs[0])
	}
}

func TestSendSetsDistinctID(t *testing.T) {
	dir := setupTestHome(t)
	srv, captured := mockServer(t)

	// Write an installation ID file.
	configDir := filepath.Join(dir, ".config", "tinct")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	installID := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	if err := os.WriteFile(filepath.Join(configDir, "telemetry.id"), []byte(installID), 0o600); err != nil {
		t.Fatal(err)
	}
	// Reset config singleton so it reads our ID file.
	config.ResetForTesting()

	sfClient := statsfactory.New(statsfactory.Config{
		ServerURL:     srv.URL,
		AppKey:        "key",
		FlushInterval: time.Hour,
	})

	client := NewWithClient(sfClient)
	client.Send(NewEvent("test"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client.Flush(ctx)
	client.Close()

	reqs := getCaptured(captured)
	if len(reqs) == 0 {
		t.Fatal("expected at least one request")
	}

	ev := reqs[0].Body.Events[0]
	if ev.DistinctID != installID {
		t.Errorf("distinct_id = %q, want %q", ev.DistinctID, installID)
	}
}

func TestFlushWithCancelledContext(t *testing.T) {
	// Server that blocks forever.
	unblock := make(chan struct{})
	slowSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-unblock
		w.WriteHeader(200)
	}))
	defer slowSrv.Close()
	defer close(unblock)

	sfClient := statsfactory.New(statsfactory.Config{
		ServerURL:     slowSrv.URL,
		AppKey:        "key",
		FlushInterval: time.Hour,
	})

	client := NewWithClient(sfClient)
	client.Send(NewEvent("test"))

	// Flush with an already-expired context.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	client.Flush(ctx)

	// Should not panic or hang. Close the underlying client to avoid leaks.
	sfClient.Close()
}

func TestClientFirstRunCreatesConfig(t *testing.T) {
	dir := setupTestHome(t)
	srv, _ := mockServer(t)
	setTelemetryCredentials(t, srv.URL)

	client := New()
	defer client.Close()

	if !client.IsEnabled() {
		t.Fatal("expected client to be enabled")
	}

	// Config file should have been created.
	configPath := filepath.Join(dir, ".config", "tinct", "tinct.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("expected tinct.toml to be created on first run")
	}

	// Installation ID file should have been created.
	idPath := filepath.Join(dir, ".config", "tinct", "telemetry.id")
	if _, err := os.Stat(idPath); os.IsNotExist(err) {
		t.Error("expected telemetry.id to be created on first run")
	}
}

func TestLegacyMigration(t *testing.T) {
	dir := setupTestHome(t)
	srv, _ := mockServer(t)
	setTelemetryCredentials(t, srv.URL)

	// Write a legacy telemetry.json file.
	legacyDir := filepath.Join(dir, ".local", "share", "tinct")
	if err := os.MkdirAll(legacyDir, 0o700); err != nil {
		t.Fatal(err)
	}
	legacyID := "deadbeef" + "deadbeef" + "deadbeef" + "deadbeef" + "deadbeef" + "deadbeef" + "deadbeef" + "deadbeef"
	legacyJSON := `{"id":"` + legacyID + `"}`
	if err := os.WriteFile(filepath.Join(legacyDir, "telemetry.json"), []byte(legacyJSON), 0o600); err != nil {
		t.Fatal(err)
	}

	client := New()
	defer client.Close()

	if !client.IsEnabled() {
		t.Fatal("expected client to be enabled")
	}

	// The installation ID should be the migrated legacy ID.
	got := config.InstallationID()
	if got != legacyID {
		t.Errorf("InstallationID = %q, want %q (legacy migration)", got, legacyID)
	}

	// Legacy file should be removed after migration.
	legacyPath := filepath.Join(legacyDir, "telemetry.json")
	if _, err := os.Stat(legacyPath); !os.IsNotExist(err) {
		t.Error("expected legacy telemetry.json to be removed after migration")
	}
}
