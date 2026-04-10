package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := defaults()

	if !cfg.Telemetry.Enabled {
		t.Error("expected telemetry enabled by default")
	}
	if !cfg.Cache.Images {
		t.Error("expected image cache enabled by default")
	}
	if cfg.Cache.Overwrite {
		t.Error("expected cache overwrite disabled by default")
	}
	if cfg.Cache.Dir != "" {
		t.Error("expected empty cache dir by default")
	}
	if cfg.Cache.Filename != "" {
		t.Error("expected empty cache filename by default")
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	tests := []struct {
		name           string
		envs           map[string]string
		checkTelemetry *bool
		checkCache     *bool
		checkDir       string
		checkFilename  string
		checkOverwrite *bool
	}{
		{
			name:           "TINCT_TELEMETRY=off disables telemetry",
			envs:           map[string]string{"TINCT_TELEMETRY": "off"},
			checkTelemetry: new(false),
		},
		{
			name:           "TINCT_TELEMETRY=on enables telemetry",
			envs:           map[string]string{"TINCT_TELEMETRY": "on"},
			checkTelemetry: new(true),
		},
		{
			name:           "TINCT_TELEMETRY=false disables telemetry",
			envs:           map[string]string{"TINCT_TELEMETRY": "false"},
			checkTelemetry: new(false),
		},
		{
			name:       "TINCT_IMAGE_CACHE=false disables cache",
			envs:       map[string]string{"TINCT_IMAGE_CACHE": "false"},
			checkCache: new(false),
		},
		{
			name:     "TINCT_IMAGE_CACHE_DIR overrides dir",
			envs:     map[string]string{"TINCT_IMAGE_CACHE_DIR": "/tmp/test-cache"},
			checkDir: "/tmp/test-cache",
		},
		{
			name:          "TINCT_IMAGE_CACHE_FILENAME overrides filename",
			envs:          map[string]string{"TINCT_IMAGE_CACHE_FILENAME": "custom.jpg"},
			checkFilename: "custom.jpg",
		},
		{
			name:           "TINCT_IMAGE_CACHE_OVERWRITE=true enables overwrite",
			envs:           map[string]string{"TINCT_IMAGE_CACHE_OVERWRITE": "true"},
			checkOverwrite: new(true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars.
			for k, v := range tt.envs {
				t.Setenv(k, v)
			}

			cfg := defaults()
			applyEnvOverrides(cfg)

			if tt.checkTelemetry != nil && cfg.Telemetry.Enabled != *tt.checkTelemetry {
				t.Errorf("telemetry enabled = %v, want %v", cfg.Telemetry.Enabled, *tt.checkTelemetry)
			}
			if tt.checkCache != nil && cfg.Cache.Images != *tt.checkCache {
				t.Errorf("cache images = %v, want %v", cfg.Cache.Images, *tt.checkCache)
			}
			if tt.checkDir != "" && cfg.Cache.Dir != tt.checkDir {
				t.Errorf("cache dir = %q, want %q", cfg.Cache.Dir, tt.checkDir)
			}
			if tt.checkFilename != "" && cfg.Cache.Filename != tt.checkFilename {
				t.Errorf("cache filename = %q, want %q", cfg.Cache.Filename, tt.checkFilename)
			}
			if tt.checkOverwrite != nil && cfg.Cache.Overwrite != *tt.checkOverwrite {
				t.Errorf("cache overwrite = %v, want %v", cfg.Cache.Overwrite, *tt.checkOverwrite)
			}
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "tinct.toml")

	cfg := defaults()
	cfg.Cache.Dir = "/custom/cache"

	// Save.
	if err := save(path, cfg); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Verify file exists.
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	// Read back.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	content := string(data)

	// Check header comment.
	if !contains(content, "# Tinct configuration file.") {
		t.Error("missing header comment")
	}

	// ID must NOT appear in tinct.toml (it lives in telemetry.id).
	if contains(content, "id =") {
		t.Error("installation ID should not be written to tinct.toml")
	}

	// Cache dir should be present.
	if !contains(content, `dir = "/custom/cache"`) {
		t.Error("missing cache dir in output")
	}
}

func TestMigrateLegacyID(t *testing.T) {
	// A valid 64-char lowercase hex ID (as produced by generateInstallationID).
	validID := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))

	legacyDir := filepath.Join(tmpDir, ".local", "share", "tinct")
	if err := os.MkdirAll(legacyDir, 0o750); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	type legacyJSON struct {
		Enabled bool   `json:"enabled"`
		ID      string `json:"id"`
	}
	data, _ := json.Marshal(legacyJSON{Enabled: true, ID: validID})
	legacyPath := filepath.Join(legacyDir, "telemetry.json")
	if err := os.WriteFile(legacyPath, data, 0o600); err != nil {
		t.Fatalf("write legacy file failed: %v", err)
	}

	got := migrateLegacyID()
	if got != validID {
		t.Errorf("migrateLegacyID() = %q, want %q", got, validID)
	}

	// Legacy file must be removed after migration.
	if _, err := os.Stat(legacyPath); !os.IsNotExist(err) {
		t.Error("legacy telemetry.json should be removed after migration")
	}
}

func TestMigrateLegacyIDInvalidHex(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	legacyDir := filepath.Join(tmpDir, ".local", "share", "tinct")
	if err := os.MkdirAll(legacyDir, 0o750); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	// Write a legacy file with an invalid ID (not 64 hex chars).
	type legacyJSON struct {
		Enabled bool   `json:"enabled"`
		ID      string `json:"id"`
	}
	data, _ := json.Marshal(legacyJSON{Enabled: true, ID: "not-a-valid-hex-id"})
	legacyPath := filepath.Join(legacyDir, "telemetry.json")
	if err := os.WriteFile(legacyPath, data, 0o600); err != nil {
		t.Fatalf("write legacy file failed: %v", err)
	}

	got := migrateLegacyID()
	if got != "" {
		t.Errorf("migrateLegacyID() with invalid ID should return empty string, got %q", got)
	}
}

func TestIsValidID(t *testing.T) {
	validID := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid 64-char hex", validID, true},
		{"too short", validID[:63], false},
		{"too long", validID + "a", false},
		{"uppercase hex", "A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2", false},
		{"non-hex chars", "z1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidID(tt.input); got != tt.want {
				t.Errorf("isValidID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateInstallationID(t *testing.T) {
	id1 := generateInstallationID()
	id2 := generateInstallationID()

	if id1 == "" {
		t.Error("expected non-empty ID")
	}
	if len(id1) != 64 { // SHA256 hex = 64 chars
		t.Errorf("expected 64-char hex string, got %d chars", len(id1))
	}
	if id1 == id2 {
		t.Error("expected unique IDs")
	}
}
func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
