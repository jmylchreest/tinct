// Package config provides centralised configuration for Tinct.
//
// Configuration is stored in two files under ~/.config/tinct/:
//
//   - tinct.toml   — human-editable settings (telemetry on/off, cache options)
//   - telemetry.id — plain-text anonymous installation ID (64-char SHA256 hex)
//
// Environment variables override file-based settings where documented.
package config

import (
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
)

const (
	// configFileName is the name of the TOML config file.
	configFileName = "tinct.toml"

	// idFileName is the plain-text file that holds the installation ID.
	idFileName = "telemetry.id"

	// configDirName is the directory name under the user config dir.
	configDirName = "tinct"

	// legacyTelemetryFile is the old JSON config at ~/.local/share/tinct/.
	// Migrated on first run.
	legacyTelemetryFile = "telemetry.json"

	// idLen is the expected length of a valid SHA256 hex installation ID.
	idLen = 64

	// Environment variable names.
	envTelemetry           = "TINCT_TELEMETRY"
	envImageCache          = "TINCT_IMAGE_CACHE"
	envImageCacheDir       = "TINCT_IMAGE_CACHE_DIR"
	envImageCacheFilename  = "TINCT_IMAGE_CACHE_FILENAME"
	envImageCacheOverwrite = "TINCT_IMAGE_CACHE_OVERWRITE"
)

// Config represents the full tinct.toml configuration.
type Config struct {
	Telemetry TelemetryConfig `toml:"telemetry"`
	Cache     CacheConfig     `toml:"cache"`
}

// TelemetryConfig holds telemetry settings stored in tinct.toml.
// The installation ID is kept separately in telemetry.id.
type TelemetryConfig struct {
	// Enabled controls whether telemetry is active. Default: true (opt-out).
	Enabled bool `toml:"enabled"`
}

// CacheConfig holds image cache settings.
type CacheConfig struct {
	// Images controls whether remote images are cached locally. Default: true.
	Images bool `toml:"images"`

	// Overwrite controls whether cached images are overwritten on re-download. Default: false.
	Overwrite bool `toml:"overwrite"`

	// Dir overrides the default image cache directory. Empty = default.
	Dir string `toml:"dir,omitempty"`

	// Filename overrides the default cached image filename. Empty = default.
	Filename string `toml:"filename,omitempty"`
}

// Singletons — lazy-initialised on first call to Load / InstallationID.
var (
	cachedConfig *Config
	cfgPath      string
	configOnce   sync.Once
	errConfig    error

	cachedID string
	idOnce   sync.Once
)

// Load loads the configuration, applying environment variable overrides.
// The first call reads from disk (or creates the file); subsequent calls
// return the cached value. Never returns a nil Config.
func Load() (*Config, error) {
	configOnce.Do(func() {
		cachedConfig, cfgPath, errConfig = loadFromDisk()
		if cachedConfig != nil {
			applyEnvOverrides(cachedConfig)
		}
	})

	if cachedConfig == nil {
		cfg := defaults()
		applyEnvOverrides(cfg)
		return cfg, errConfig
	}

	return cachedConfig, errConfig
}

// InstallationID returns the persistent anonymous installation identifier.
//
// The ID is stored as a plain 64-character lowercase SHA256 hex string in
// ~/.config/tinct/telemetry.id.  On first call the file is read and validated;
// if it is absent, empty, or does not pass validation (wrong length, non-hex
// characters) a new ID is generated, written to disk, and returned.
//
// Migrates from the legacy ~/.local/share/tinct/telemetry.json if present.
// Never returns an error; returns empty string only on catastrophic failure.
func InstallationID() string {
	idOnce.Do(func() {
		cachedID = loadOrCreateID()
	})
	return cachedID
}

// Path returns the path to the tinct.toml config file.
func Path() string {
	_, _ = Load() //nolint:errcheck // populate cfgPath, error not needed here
	return cfgPath
}

// defaults returns a Config with default values.
func defaults() *Config {
	return &Config{
		Telemetry: TelemetryConfig{Enabled: true},
		Cache: CacheConfig{
			Images:    true,
			Overwrite: false,
		},
	}
}

// configDir returns ~/.config/tinct (or the XDG equivalent).
func configDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("get config directory: %w", err)
	}
	return filepath.Join(dir, configDirName), nil
}

// configFilePath returns the full path to tinct.toml.
func configFilePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// idFilePath returns the full path to telemetry.id.
func idFilePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, idFileName), nil
}

// loadFromDisk reads tinct.toml, creating it with defaults if absent or corrupt.
func loadFromDisk() (*Config, string, error) {
	path, err := configFilePath()
	if err != nil {
		return defaults(), "", fmt.Errorf("determine config path: %w", err)
	}

	data, err := os.ReadFile(path) // #nosec G304 -- path is derived from os.UserConfigDir(), not user-controlled input
	if err == nil {
		cfg := defaults()
		if _, decErr := toml.Decode(string(data), cfg); decErr == nil {
			return cfg, path, nil
		}
		// Corrupt TOML — regenerate.
	}

	cfg := defaults()
	if saveErr := save(path, cfg); saveErr != nil {
		return cfg, path, fmt.Errorf("save new config: %w", saveErr)
	}

	return cfg, path, nil
}

// loadOrCreateID loads, validates, and (if necessary) generates the installation ID.
func loadOrCreateID() string {
	path, err := idFilePath()
	if err != nil {
		// Can't determine path — generate an ephemeral ID for this session.
		return generateInstallationID()
	}

	// Attempt to read and validate an existing ID.
	if data, readErr := os.ReadFile(path); readErr == nil { // #nosec G304
		if id := strings.TrimSpace(string(data)); isValidID(id) {
			return id
		}
		// File exists but content is invalid — regenerate below.
	}

	// No valid ID on disk — check for a legacy telemetry.json to migrate from.
	id := migrateLegacyID()
	if id == "" || !isValidID(id) {
		id = generateInstallationID()
	}

	// Persist the new (or migrated) ID.
	dir := filepath.Dir(path)
	if mkErr := os.MkdirAll(dir, 0o750); mkErr == nil {
		_ = os.WriteFile(path, []byte(id+"\n"), 0o600) //nolint:errcheck // best-effort
	}

	return id
}

// isValidID reports whether s is a valid installation ID:
// exactly idLen (64) lowercase hexadecimal characters.
func isValidID(s string) bool {
	if len(s) != idLen {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// migrateLegacyID reads the installation ID from the old JSON config at
// ~/.local/share/tinct/telemetry.json, removes that file, and returns the ID.
// Returns empty string if the file is absent, unreadable, or the ID is invalid.
func migrateLegacyID() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	legacyPath := filepath.Join(home, ".local", "share", "tinct", legacyTelemetryFile)
	data, err := os.ReadFile(legacyPath) // #nosec G304 - Legacy path
	if err != nil {
		return ""
	}

	var legacy struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return ""
	}

	if isValidID(legacy.ID) {
		_ = os.Remove(legacyPath) // best-effort cleanup
		return legacy.ID
	}

	return ""
}

// save writes the config to disk in TOML format.
func save(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) // #nosec G304 -- path is derived from os.UserConfigDir(), not user-controlled input
	if err != nil {
		return fmt.Errorf("open config file: %w", err)
	}
	defer f.Close()

	const header = "# Tinct configuration file.\n" +
		"# See: https://tinct.jmylchreest.dev/docs/configuration\n\n"
	if _, err := f.WriteString(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	return nil
}

// applyEnvOverrides applies environment variable overrides to the config.
func applyEnvOverrides(cfg *Config) {
	if val := os.Getenv(envTelemetry); val != "" {
		switch strings.ToLower(val) {
		case "off", "false", "0", "no":
			cfg.Telemetry.Enabled = false
		case "on", "true", "1", "yes":
			cfg.Telemetry.Enabled = true
		}
	}

	if val := os.Getenv(envImageCache); val != "" {
		if parsed, err := strconv.ParseBool(val); err == nil {
			cfg.Cache.Images = parsed
		}
	}

	if val := os.Getenv(envImageCacheDir); val != "" {
		cfg.Cache.Dir = val
	}

	if val := os.Getenv(envImageCacheFilename); val != "" {
		cfg.Cache.Filename = val
	}

	if val := os.Getenv(envImageCacheOverwrite); val != "" {
		if parsed, err := strconv.ParseBool(val); err == nil {
			cfg.Cache.Overwrite = parsed
		}
	}
}

// generateInstallationID creates a new anonymous installation identifier.
// 32 bytes of cryptographically-random data are SHA256-hashed, producing a
// 64-character lowercase hex string. Falls back to a time/pid-based seed if
// crypto/rand is unavailable (e.g. restricted sandbox).
func generateInstallationID() string {
	b := make([]byte, 32)
	if _, err := cryptorand.Read(b); err != nil {
		// Fallback: hash time + pid so we always produce a valid ID.
		raw := fmt.Sprintf("%d-%d", os.Getpid(), os.Getppid())
		hash := sha256.Sum256([]byte(raw))
		return hex.EncodeToString(hash[:])
	}
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:])
}

// ResetForTesting resets all cached state. Only call from test code.
func ResetForTesting() {
	cachedConfig = nil
	cfgPath = ""
	configOnce = sync.Once{}
	errConfig = nil
	cachedID = ""
	idOnce = sync.Once{}
}
