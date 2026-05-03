package config

import "testing"

func TestResolve_FlagWins(t *testing.T) {
	t.Setenv("TINCT_PLUGIN_FOO_OUTPUT_DIR", "/from-env")
	got := Resolve("foo", "output_dir", "/from-flag", "/fallback")
	if got != "/from-flag" {
		t.Errorf("Resolve flag precedence: got %q, want %q", got, "/from-flag")
	}
}

func TestResolve_EnvWinsOverFallback(t *testing.T) {
	t.Setenv("TINCT_PLUGIN_FOO_OUTPUT_DIR", "/from-env")
	got := Resolve("foo", "output_dir", "", "/fallback")
	if got != "/from-env" {
		t.Errorf("Resolve env precedence: got %q, want %q", got, "/from-env")
	}
}

func TestResolve_FallbackWhenUnset(t *testing.T) {
	got := Resolve("never-set-anywhere-xyz", "output_dir", "", "/fallback")
	if got != "/fallback" {
		t.Errorf("Resolve fallback: got %q, want %q", got, "/fallback")
	}
}

func TestEnvName(t *testing.T) {
	tests := []struct {
		plugin, key, want string
	}{
		{"kitty", "output_dir", "TINCT_PLUGIN_KITTY_OUTPUT_DIR"},
		{"kde-plasma", "scheme", "TINCT_PLUGIN_KDE_PLASMA_SCHEME"},
		{"google-genai", "backend", "TINCT_PLUGIN_GOOGLE_GENAI_BACKEND"},
		{"foo", "with-dash", "TINCT_PLUGIN_FOO_WITH_DASH"},
	}
	for _, tt := range tests {
		if got := EnvName(tt.plugin, tt.key); got != tt.want {
			t.Errorf("EnvName(%q, %q) = %q, want %q", tt.plugin, tt.key, got, tt.want)
		}
	}
}
