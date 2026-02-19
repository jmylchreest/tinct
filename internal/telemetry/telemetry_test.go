package telemetry

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jmylchreest/tinct/internal/version"
)

// TestGenerateInstallationID verifies that installation IDs are valid SHA256 hex strings.
func TestGenerateInstallationID(t *testing.T) {
	id := generateInstallationID()

	// SHA256 hex string should be 64 characters.
	if len(id) != 64 {
		t.Errorf("expected 64 char hex string, got %d chars: %s", len(id), id)
	}

	// Should be lowercase hex.
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("invalid hex character: %c", c)
		}
	}

	// Two consecutive calls should produce different IDs.
	id2 := generateInstallationID()
	if id == id2 {
		t.Error("two consecutive IDs should be different")
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

// TestConfigLoadAndSave tests the config file round-trip.
func TestConfigLoadAndSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "telemetry.json")

	config := &Config{
		Enabled: true,
		ID:      "abc123def456",
	}

	// Save.
	if err := saveConfig(path, config); err != nil {
		t.Fatalf("saveConfig() error = %v", err)
	}

	// Read back and verify.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if loaded.Enabled != config.Enabled {
		t.Errorf("Enabled = %v, want %v", loaded.Enabled, config.Enabled)
	}
	if loaded.ID != config.ID {
		t.Errorf("ID = %q, want %q", loaded.ID, config.ID)
	}
}

// TestConfigFilePermissions verifies the config file has restrictive permissions.
func TestConfigFilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "telemetry.json")

	config := &Config{Enabled: true, ID: "test"}
	if err := saveConfig(path, config); err != nil {
		t.Fatalf("saveConfig() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	// Should be 0600 (owner read/write only).
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("file permissions = %o, want 600", perm)
	}
}

// TestClientDisabledByEnvVar verifies TINCT_TELEMETRY env var disables the client.
func TestClientDisabledByEnvVar(t *testing.T) {
	disableValues := []string{"off", "false", "0", "no", "OFF", "False", "No"}

	for _, val := range disableValues {
		t.Run(val, func(t *testing.T) {
			t.Setenv(envKey, val)
			client := New()
			if client.IsEnabled() {
				t.Errorf("client should be disabled with %s=%s", envKey, val)
			}
		})
	}
}

// TestClientEnabledByDefault verifies that the client is enabled when
// no env var is set and config allows it.
func TestClientEnabledByDefault(t *testing.T) {
	// Clear the env var to ensure default behaviour.
	t.Setenv(envKey, "")

	// Override HOME to use a temp directory for the config file.
	dir := t.TempDir()
	t.Setenv("HOME", dir)

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
	t.Setenv(envKey, "")

	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// Create config directory and file with enabled=false.
	configDir := filepath.Join(dir, ".local", "share", "tinct")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	config := Config{Enabled: false, ID: "test-id"}
	data, _ := json.Marshal(config)
	if err := os.WriteFile(filepath.Join(configDir, "telemetry.json"), data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	client := New()
	if client.IsEnabled() {
		t.Error("client should be disabled when config has enabled=false")
	}
}

// TestClientFirstRunCreatesConfig verifies that a config file is created on first run.
func TestClientFirstRunCreatesConfig(t *testing.T) {
	t.Setenv(envKey, "")

	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_ = New()

	// Config file should now exist.
	path := filepath.Join(dir, ".local", "share", "tinct", "telemetry.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("config file should have been created on first run")
	}

	// Read and validate.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if !config.Enabled {
		t.Error("default config should have enabled=true")
	}
	if len(config.ID) != 64 {
		t.Errorf("config ID should be 64 char SHA256 hex, got %d chars", len(config.ID))
	}
}

// TestClientPersistentID verifies that the installation ID persists across client instances.
func TestClientPersistentID(t *testing.T) {
	t.Setenv(envKey, "")

	dir := t.TempDir()
	t.Setenv("HOME", dir)

	client1 := New()
	id1 := client1.ID()

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

	t.Setenv(envKey, "")
	dir := t.TempDir()
	t.Setenv("HOME", dir)

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

	t.Setenv(envKey, "off")
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

	t.Setenv(envKey, "")
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	client := New()
	client.baseURL = server.URL

	// sendSync returns an error, but Send (async) swallows it.
	err := client.sendSync(NewEvent("test"))
	if err != nil {
		t.Error("sendSync should not return error for HTTP 500 (fire-and-forget)")
	}
}

// TestCorruptConfigFileRegenerated verifies that a corrupt config is replaced.
func TestCorruptConfigFileRegenerated(t *testing.T) {
	t.Setenv(envKey, "")

	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// Write corrupt config.
	configDir := filepath.Join(dir, ".local", "share", "tinct")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "telemetry.json"), []byte("not json"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	client := New()
	if !client.IsEnabled() {
		t.Error("client should be enabled after regenerating corrupt config")
	}
	if len(client.ID()) != 64 {
		t.Errorf("ID should be regenerated, got %d chars", len(client.ID()))
	}
}

// TestConfigWithEmptyIDRegenerated verifies that a config with empty ID is fixed.
func TestConfigWithEmptyIDRegenerated(t *testing.T) {
	t.Setenv(envKey, "")

	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// Write config with empty ID.
	configDir := filepath.Join(dir, ".local", "share", "tinct")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	config := Config{Enabled: true, ID: ""}
	data, _ := json.Marshal(config)
	if err := os.WriteFile(filepath.Join(configDir, "telemetry.json"), data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	client := New()
	if len(client.ID()) != 64 {
		t.Errorf("empty ID should be regenerated, got %d chars", len(client.ID()))
	}
}

// TestEnvVarOverridesConfig verifies that the env var takes priority over config file.
func TestEnvVarOverridesConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// Write config with enabled=true.
	configDir := filepath.Join(dir, ".local", "share", "tinct")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	config := Config{Enabled: true, ID: "test-id-123"}
	data, _ := json.Marshal(config)
	if err := os.WriteFile(filepath.Join(configDir, "telemetry.json"), data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// But env var says off.
	t.Setenv(envKey, "off")

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

	t.Setenv(envKey, "")
	dir := t.TempDir()
	t.Setenv("HOME", dir)

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

	t.Setenv(envKey, "")
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	client := New()
	client.baseURL = server.URL

	_ = client.sendSync(NewEvent("test"))

	var events []aptabaseEvent
	if err := json.Unmarshal(receivedBody, &events); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify timestamp is parseable as RFC3339.
	_, err := time.Parse(time.RFC3339Nano, events[0].Timestamp)
	if err != nil {
		t.Errorf("timestamp %q is not valid RFC3339: %v", events[0].Timestamp, err)
	}

	// Verify timestamp ends with Z (UTC).
	if !strings.HasSuffix(events[0].Timestamp, "Z") {
		t.Errorf("timestamp should be UTC (end with Z), got %q", events[0].Timestamp)
	}
}
