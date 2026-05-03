package pluginconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmylchreest/tinct/internal/config"
)

// writeTinctTOML drops a tinct.toml in a temp $XDG_CONFIG_HOME and points
// config.Load at it. Returns nothing — Load is cached so callers must
// arrange test isolation themselves (typically by running each test as a
// fresh subtest with a fresh tempdir, AND using config.Reset between
// them — see resetConfig below).
func writeTinctTOML(t *testing.T, contents string) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	dir := filepath.Join(tmp, "tinct")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tinct.toml"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	// Force a re-read on the next config.Load() call.
	config.ResetForTesting()
}

func TestResolve_FallbackOnly(t *testing.T) {
	writeTinctTOML(t, "") // empty config
	got := Resolve("kitty", "output_dir", "", "/fallback")
	if got != "/fallback" {
		t.Errorf("Resolve fallback = %q, want /fallback", got)
	}
}

func TestResolve_FlagWins(t *testing.T) {
	t.Setenv("TINCT_PLUGIN_KITTY_OUTPUT_DIR", "/from-env")
	writeTinctTOML(t, `[plugin.kitty]
output_dir = "/from-toml"
`)
	got := Resolve("kitty", "output_dir", "/from-flag", "/fallback")
	if got != "/from-flag" {
		t.Errorf("Resolve flag-wins = %q, want /from-flag", got)
	}
}

func TestResolve_EnvOverridesToml(t *testing.T) {
	t.Setenv("TINCT_PLUGIN_KITTY_OUTPUT_DIR", "/from-env")
	writeTinctTOML(t, `[plugin.kitty]
output_dir = "/from-toml"
`)
	got := Resolve("kitty", "output_dir", "", "/fallback")
	if got != "/from-env" {
		t.Errorf("Resolve env-over-toml = %q, want /from-env", got)
	}
}

func TestResolve_TomlOverridesFallback(t *testing.T) {
	t.Setenv("TINCT_PLUGIN_KITTY_OUTPUT_DIR", "")
	writeTinctTOML(t, `[plugin.kitty]
output_dir = "/from-toml"
`)
	got := Resolve("kitty", "output_dir", "", "/fallback")
	if got != "/from-toml" {
		t.Errorf("Resolve toml-over-fallback = %q, want /from-toml", got)
	}
}

func TestResolve_EmptyTomlSection(t *testing.T) {
	writeTinctTOML(t, `[plugin.neovim]
output_dir = "/from-toml"
`)
	// Querying a different plugin should fall through to fallback.
	got := Resolve("kitty", "output_dir", "", "/fallback")
	if got != "/fallback" {
		t.Errorf("Resolve cross-plugin = %q, want /fallback", got)
	}
}

func TestResolve_MultipleKeys(t *testing.T) {
	writeTinctTOML(t, `[plugin.neovim]
output_dir = "/themes"
theme_name = "tinct-dark"
`)
	if got := Resolve("neovim", "output_dir", "", ""); got != "/themes" {
		t.Errorf("output_dir = %q, want /themes", got)
	}
	if got := Resolve("neovim", "theme_name", "", "tinct"); got != "tinct-dark" {
		t.Errorf("theme_name = %q, want tinct-dark", got)
	}
}

func TestEnvName(t *testing.T) {
	tests := []struct {
		plugin, key, want string
	}{
		{"kitty", "output_dir", "TINCT_PLUGIN_KITTY_OUTPUT_DIR"},
		{"hyprland", "reload", "TINCT_PLUGIN_HYPRLAND_RELOAD"},
		{"google-genai", "backend", "TINCT_PLUGIN_GOOGLE_GENAI_BACKEND"},
		{"awww", "transition-type", "TINCT_PLUGIN_AWWW_TRANSITION_TYPE"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := envName(tt.plugin, tt.key); got != tt.want {
				t.Errorf("envName(%q, %q) = %q, want %q", tt.plugin, tt.key, got, tt.want)
			}
		})
	}
}

func TestResolve_WrongTypeIgnored(t *testing.T) {
	// If a user puts a non-string value in tinct.toml, the resolver falls
	// through to fallback rather than panicking or returning an empty.
	writeTinctTOML(t, `[plugin.kitty]
output_dir = 42
`)
	got := Resolve("kitty", "output_dir", "", "/fallback")
	if got != "/fallback" {
		t.Errorf("Resolve wrong-type = %q, want /fallback (non-string TOML value should fall through)", got)
	}
}
