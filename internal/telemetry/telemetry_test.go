package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jmylchreest/tinct/internal/config"
	"github.com/jmylchreest/tinct/internal/version"
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

// writeTestConfig writes a tinct.toml file into the test home's config dir.
func writeTestConfig(t *testing.T, homeDir string, content string) {
	t.Helper()
	configDir := filepath.Join(homeDir, ".config", "tinct")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "tinct.toml"), []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

// TestGenerateSessionID verifies session ID format.
func TestGenerateSessionID(t *testing.T) {
	sid := generateSessionID()

	// Should be a numeric string (epoch + 8 digits).
	if len(sid) < 10 {
		t.Errorf("session ID too short: %s", sid)
	}

	for _, c := range sid {
		if c < '0' || c > '9' {
			t.Errorf("session ID should be numeric, got char: %c", c)
		}
	}
}

// TestResolveBaseURL verifies URL resolution from app keys.
func TestResolveBaseURL(t *testing.T) {
	tests := []struct {
		appKey   string
		expected string
	}{
		{"A-EU-1234567890", "https://eu.aptabase.com"},
		{"A-US-1234567890", "https://us.aptabase.com"},
		{"A-SH-1234567890", "https://eu.aptabase.com"}, // Self-hosted falls through to default.
		{"invalid", "https://eu.aptabase.com"},         // Invalid key falls through.
		{"", "https://eu.aptabase.com"},                // Empty key falls through.
		{"A-eu-1234567890", "https://eu.aptabase.com"}, // Case-insensitive.
		{"A-us-1234567890", "https://us.aptabase.com"}, // Case-insensitive.
	}

	for _, tt := range tests {
		t.Run(tt.appKey, func(t *testing.T) {
			got := resolveBaseURL(tt.appKey)
			if got != tt.expected {
				t.Errorf("resolveBaseURL(%q) = %q, want %q", tt.appKey, got, tt.expected)
			}
		})
	}
}

// TestClientDisabledByEnvVar verifies TINCT_TELEMETRY env var disables the client.
func TestClientDisabledByEnvVar(t *testing.T) {
	disableValues := []string{"off", "false", "0", "no", "OFF", "False", "No"}

	for _, val := range disableValues {
		t.Run(val, func(t *testing.T) {
			_ = setupTestHome(t)
			t.Setenv("TINCT_TELEMETRY", val)
			client := New()
			if client.IsEnabled() {
				t.Errorf("client should be disabled with TINCT_TELEMETRY=%s", val)
			}
		})
	}
}

// TestClientEnabledByDefault verifies that the client is enabled when
// no env var is set and config allows it.
func TestClientEnabledByDefault(t *testing.T) {
	_ = setupTestHome(t)
	t.Setenv("TINCT_TELEMETRY", "")

	client := New()
	if !client.IsEnabled() {
		t.Error("client should be enabled by default")
	}

	// Should have generated an ID.
	if client.ID() == "" {
		t.Error("client should have a non-empty ID")
	}
}

// TestClientDisabledByConfig verifies that setting enabled=false in config disables telemetry.
func TestClientDisabledByConfig(t *testing.T) {
	dir := setupTestHome(t)
	t.Setenv("TINCT_TELEMETRY", "")

	// Write config with telemetry disabled.
	writeTestConfig(t, dir, `
[telemetry]
enabled = false
id = "test-id-123"
`)

	client := New()
	if client.IsEnabled() {
		t.Error("client should be disabled when config has enabled=false")
	}
}

// TestClientFirstRunCreatesConfig verifies that a config file is created on first run.
func TestClientFirstRunCreatesConfig(t *testing.T) {
	dir := setupTestHome(t)
	t.Setenv("TINCT_TELEMETRY", "")

	_ = New()

	// Config file should now exist at the new location.
	path := filepath.Join(dir, ".config", "tinct", "tinct.toml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("config file should have been created on first run")
	}

	// Read and verify it contains a telemetry ID.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "[telemetry]") {
		t.Error("config should contain [telemetry] section")
	}
	if !strings.Contains(content, "enabled = true") {
		t.Error("config should have enabled = true by default")
	}
	// Installation ID lives in telemetry.id, not tinct.toml.
	if strings.Contains(content, "id = ") {
		t.Error("tinct.toml should NOT contain an id field")
	}

	// telemetry.id file should exist alongside tinct.toml.
	idPath := filepath.Join(dir, ".config", "tinct", "telemetry.id")
	if _, err := os.Stat(idPath); os.IsNotExist(err) {
		t.Error("telemetry.id should have been created on first run")
	}
}

// TestClientPersistentID verifies that the installation ID persists across client instances.
func TestClientPersistentID(t *testing.T) {
	_ = setupTestHome(t)
	t.Setenv("TINCT_TELEMETRY", "")

	client1 := New()
	id1 := client1.ID()

	// The config package uses sync.Once, so creating a new Client will
	// reuse the same config (same process). This is the correct behaviour.
	client2 := New()
	id2 := client2.ID()

	if id1 != id2 {
		t.Errorf("installation ID should persist: got %q then %q", id1, id2)
	}
}

// TestSendSync verifies the HTTP request format sent to Aptabase.
func TestSendSync(t *testing.T) {
	var receivedBody []byte
	var receivedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dir := setupTestHome(t)
	t.Setenv("TINCT_TELEMETRY", "")

	// Pre-create config with a known ID.
	writeTestConfig(t, dir, `
[telemetry]
enabled = true
id = "test-id-for-send-sync"
`)

	client := New()
	client.baseURL = server.URL

	event := NewEvent("test_event").
		Set("key1", "value1").
		Set("key2", 42)

	err := client.sendSync(event)
	if err != nil {
		t.Fatalf("sendSync() error = %v", err)
	}

	// Verify headers.
	if receivedHeaders.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", receivedHeaders.Get("Content-Type"))
	}
	if receivedHeaders.Get("App-Key") != AppKey {
		t.Errorf("App-Key = %q, want %q", receivedHeaders.Get("App-Key"), AppKey)
	}
	if receivedHeaders.Get("X-Installation-ID") == "" {
		t.Error("X-Installation-ID header should be set")
	}

	// Verify body is a JSON array with one event.
	var events []aptabaseEvent
	if err := json.Unmarshal(receivedBody, &events); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	evt := events[0]
	if evt.EventName != "test_event" {
		t.Errorf("EventName = %q, want %q", evt.EventName, "test_event")
	}
	if evt.SystemProps.OSName == "" {
		t.Error("SystemProps.OSName should not be empty")
	}
	expectedSDK := sdkPrefix + version.Version
	if evt.SystemProps.SDKVersion != expectedSDK {
		t.Errorf("SystemProps.SDKVersion = %q, want %q", evt.SystemProps.SDKVersion, expectedSDK)
	}
	if evt.SessionID == "" {
		t.Error("SessionID should not be empty")
	}

	// Verify custom props.
	if evt.Props["key1"] != "value1" {
		t.Errorf("Props[key1] = %v, want value1", evt.Props["key1"])
	}
	// JSON numbers deserialise as float64.
	if evt.Props["key2"] != float64(42) {
		t.Errorf("Props[key2] = %v, want 42", evt.Props["key2"])
	}
}

// TestSendDisabledClient verifies that Send is a no-op when disabled.
func TestSendDisabledClient(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_ = setupTestHome(t)
	t.Setenv("TINCT_TELEMETRY", "off")
	client := New()
	client.baseURL = server.URL

	event := NewEvent("should_not_send")
	client.Send(event)

	if requestCount != 0 {
		t.Errorf("disabled client should not send requests, got %d", requestCount)
	}
}

// TestSendHTTPFailureSilent verifies that HTTP errors don't propagate.
func TestSendHTTPFailureSilent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_ = setupTestHome(t)
	t.Setenv("TINCT_TELEMETRY", "")

	client := New()
	client.baseURL = server.URL

	// sendSync returns nil for HTTP errors (fire-and-forget).
	err := client.sendSync(NewEvent("test"))
	if err != nil {
		t.Error("sendSync should not return error for HTTP 500 (fire-and-forget)")
	}
}

// TestEnvVarOverridesConfig verifies that the env var takes priority over config file.
func TestEnvVarOverridesConfig(t *testing.T) {
	dir := setupTestHome(t)

	// Write config with telemetry enabled.
	writeTestConfig(t, dir, `
[telemetry]
enabled = true
id = "test-id-123"
`)

	// But env var says off.
	t.Setenv("TINCT_TELEMETRY", "off")

	client := New()
	if client.IsEnabled() {
		t.Error("env var should override config file")
	}
}

// TestSendRequestToCorrectPath verifies the API path is correct.
func TestSendRequestToCorrectPath(t *testing.T) {
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_ = setupTestHome(t)
	t.Setenv("TINCT_TELEMETRY", "")

	client := New()
	client.baseURL = server.URL

	_ = client.sendSync(NewEvent("test"))

	if receivedPath != "/api/v0/events" {
		t.Errorf("request path = %q, want /api/v0/events", receivedPath)
	}
}

// TestEventTimestampFormat verifies the timestamp is RFC3339 format.
func TestEventTimestampFormat(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_ = setupTestHome(t)
	t.Setenv("TINCT_TELEMETRY", "")

	client := New()
	client.baseURL = server.URL

	_ = client.sendSync(NewEvent("test"))

	var events []aptabaseEvent
	if err := json.Unmarshal(receivedBody, &events); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify timestamp is parseable as RFC3339 (without nanoseconds).
	_, err := time.Parse(time.RFC3339, events[0].Timestamp)
	if err != nil {
		t.Errorf("timestamp %q is not valid RFC3339: %v", events[0].Timestamp, err)
	}

	// Verify timestamp ends with Z (UTC).
	if !strings.HasSuffix(events[0].Timestamp, "Z") {
		t.Errorf("timestamp should be UTC (end with Z), got %q", events[0].Timestamp)
	}

	// Verify no fractional seconds (prevents regression to RFC3339Nano).
	if strings.Contains(events[0].Timestamp, ".") {
		t.Errorf("timestamp should not contain fractional seconds, got %q", events[0].Timestamp)
	}
}

// TestSessionIDConsistentAcrossEvents verifies that all events sent by the
// same Client share a single session ID, so Aptabase groups them together.
func TestSessionIDConsistentAcrossEvents(t *testing.T) {
	var allBodies [][]byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		allBodies = append(allBodies, body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_ = setupTestHome(t)
	t.Setenv("TINCT_TELEMETRY", "")

	client := New()
	client.baseURL = server.URL

	// Send three events through the same client.
	_ = client.sendSync(NewEvent("event_one"))
	_ = client.sendSync(NewEvent("event_two"))
	_ = client.sendSync(NewEvent("event_three"))

	if len(allBodies) != 3 {
		t.Fatalf("expected 3 requests, got %d", len(allBodies))
	}

	// Extract session IDs from each request.
	sessionIDs := make([]string, 0, 3)
	for i, body := range allBodies {
		var events []aptabaseEvent
		if err := json.Unmarshal(body, &events); err != nil {
			t.Fatalf("request %d: failed to unmarshal: %v", i, err)
		}
		if len(events) != 1 {
			t.Fatalf("request %d: expected 1 event, got %d", i, len(events))
		}
		sessionIDs = append(sessionIDs, events[0].SessionID)
	}

	// All session IDs must be identical.
	for i := 1; i < len(sessionIDs); i++ {
		if sessionIDs[i] != sessionIDs[0] {
			t.Errorf("session ID mismatch: event 0 = %q, event %d = %q", sessionIDs[0], i, sessionIDs[i])
		}
	}

	if sessionIDs[0] == "" {
		t.Error("session ID should not be empty")
	}
}

// TestDifferentClientsGetDifferentSessions verifies that separate Client
// instances produce different session IDs (each CLI invocation = new session).
func TestDifferentClientsGetDifferentSessions(t *testing.T) {
	var allBodies [][]byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		allBodies = append(allBodies, body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_ = setupTestHome(t)
	t.Setenv("TINCT_TELEMETRY", "")

	client1 := New()
	client1.baseURL = server.URL

	client2 := New()
	client2.baseURL = server.URL

	_ = client1.sendSync(NewEvent("from_client1"))
	_ = client2.sendSync(NewEvent("from_client2"))

	if len(allBodies) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(allBodies))
	}

	var sid1, sid2 string
	for i, body := range allBodies {
		var events []aptabaseEvent
		if err := json.Unmarshal(body, &events); err != nil {
			t.Fatalf("request %d: failed to unmarshal: %v", i, err)
		}
		if i == 0 {
			sid1 = events[0].SessionID
		} else {
			sid2 = events[0].SessionID
		}
	}

	if sid1 == sid2 {
		t.Errorf("different clients should have different session IDs, both got %q", sid1)
	}
}

// TestSendAndFlushBatchesEvents verifies that multiple Send() calls are
// coalesced into a single HTTP request by the debounce worker.
func TestSendAndFlushBatchesEvents(t *testing.T) {
	var mu sync.Mutex
	var receivedBodies [][]byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		receivedBodies = append(receivedBodies, body)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_ = setupTestHome(t)
	t.Setenv("TINCT_TELEMETRY", "")

	client := New()
	client.baseURL = server.URL

	// Send several events in quick succession — they should all fit in one batch.
	const numEvents = 5
	for i := range numEvents {
		client.Send(NewEvent(fmt.Sprintf("event_%d", i)))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client.Flush(ctx)

	mu.Lock()
	defer mu.Unlock()

	// All events should have been coalesced into a single request.
	if len(receivedBodies) != 1 {
		t.Fatalf("expected 1 batched request, got %d", len(receivedBodies))
	}

	var events []aptabaseEvent
	if err := json.Unmarshal(receivedBodies[0], &events); err != nil {
		t.Fatalf("unmarshal batch: %v", err)
	}
	if len(events) != numEvents {
		t.Errorf("expected %d events in batch, got %d", numEvents, len(events))
	}
}

// TestSendAndFlushSplitsOversizeBatch verifies that more than maxBatchSize events
// are split into multiple requests of at most maxBatchSize each.
func TestSendAndFlushSplitsOversizeBatch(t *testing.T) {
	var mu sync.Mutex
	var receivedBodies [][]byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		receivedBodies = append(receivedBodies, body)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_ = setupTestHome(t)
	t.Setenv("TINCT_TELEMETRY", "")

	client := New()
	client.baseURL = server.URL

	// Send more events than maxBatchSize (25).
	const numEvents = 30
	for i := range numEvents {
		client.Send(NewEvent(fmt.Sprintf("event_%d", i)))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client.Flush(ctx)

	mu.Lock()
	defer mu.Unlock()

	// Count total events received across all requests.
	total := 0
	for _, body := range receivedBodies {
		var events []aptabaseEvent
		if err := json.Unmarshal(body, &events); err != nil {
			t.Fatalf("unmarshal batch: %v", err)
		}
		if len(events) > maxBatchSize {
			t.Errorf("batch size %d exceeds maxBatchSize %d", len(events), maxBatchSize)
		}
		total += len(events)
	}

	if total != numEvents {
		t.Errorf("expected %d total events across all batches, got %d", numEvents, total)
	}
}

// TestFlushWithCancelledContext verifies that Flush returns promptly when the
// context is already cancelled, even if the worker hasn't finished.
func TestFlushWithCancelledContext(t *testing.T) {
	// Use a server that hangs so the worker's HTTP request never completes.
	hang := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		<-hang
	}))

	_ = setupTestHome(t)
	t.Setenv("TINCT_TELEMETRY", "")

	client := New()
	client.baseURL = server.URL

	client.Send(NewEvent("will_not_complete"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	start := time.Now()
	client.Flush(ctx)
	elapsed := time.Since(start)

	// Flush should return nearly immediately because the context is already done.
	if elapsed > 500*time.Millisecond {
		t.Errorf("Flush took too long with cancelled context: %v", elapsed)
	}

	// Unblock the hanging server and wait for the worker goroutine to fully exit
	// before returning. This prevents the worker from racing with config.ResetForTesting()
	// in subsequent tests.
	close(hang)
	server.Close()
	<-client.done
}

// TestSendAfterFlushDropsSilently verifies that Send() after Flush() does not panic.
// (The queue is closed; the select default branch drops the event.)
func TestSendAfterFlushDropsSilently(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_ = setupTestHome(t)
	t.Setenv("TINCT_TELEMETRY", "")

	client := New()
	client.baseURL = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client.Flush(ctx)

	// This must not panic.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Send after Flush panicked: %v", r)
		}
	}()
	client.Send(NewEvent("post_flush"))
}

// TestLegacyMigration verifies that telemetry ID is migrated from the old
// ~/.local/share/tinct/telemetry.json to the new telemetry.id file.
func TestLegacyMigration(t *testing.T) {
	dir := setupTestHome(t)
	t.Setenv("TINCT_TELEMETRY", "")

	// A valid 64-char lowercase hex ID (as produced by generateInstallationID).
	legacyID := "dead1234beef5678cafe9012dead1234beef5678cafe9012dead1234beef5678"

	// Write legacy telemetry.json.
	legacyDir := filepath.Join(dir, ".local", "share", "tinct")
	if err := os.MkdirAll(legacyDir, 0o750); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	legacyData, _ := json.Marshal(struct {
		Enabled bool   `json:"enabled"`
		ID      string `json:"id"`
	}{
		Enabled: true,
		ID:      legacyID,
	})
	if err := os.WriteFile(filepath.Join(legacyDir, "telemetry.json"), legacyData, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	client := New()
	if client.ID() != legacyID {
		t.Errorf("expected migrated ID %q, got %q", legacyID, client.ID())
	}

	// Legacy file should be removed.
	if _, err := os.Stat(filepath.Join(legacyDir, "telemetry.json")); !os.IsNotExist(err) {
		t.Error("legacy telemetry.json should be removed after migration")
	}

	// New config file should exist.
	newPath := filepath.Join(dir, ".config", "tinct", "tinct.toml")
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Error("new tinct.toml should be created after migration")
	}
}
