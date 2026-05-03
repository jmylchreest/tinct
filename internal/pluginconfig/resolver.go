// Package pluginconfig provides a deterministic resolver for plugin-scoped
// configuration values. It merges three sources with explicit precedence:
//
//  1. cobra flag (caller passes the already-bound value)
//  2. environment variable (TINCT_PLUGIN_<NAME>_<KEY>)
//  3. tinct.toml [plugin.<name>] <key>
//  4. caller-supplied fallback
//
// The function form is intentional: plugins call Resolve directly from
// DefaultOutputDir or Validate. There is no struct-binding, no reflection,
// no implicit registration. Each call site reads one value with one
// statement; the precedence chain lives in one place
// (Resolve in this package).
//
// This is the smallest viable layer. If we later want auto-binding of every
// cobra flag to TOML/env without per-call wiring, the implementation can be
// swapped for koanf without changing call sites — the function signature
// remains stable.
package pluginconfig

import (
	"os"
	"strings"

	"github.com/jmylchreest/tinct/internal/config"
)

// Resolve returns the effective string value for plugin.<key>.
//
// flagValue is the value cobra has already bound onto the plugin struct.
// The convention across tinct plugins is that StringVar defaults to "" and
// any non-empty value implies the user supplied it via --<plugin>.<key>
// (with hyphens). Resolve treats empty string as "not set" for that
// reason — it doesn't otherwise distinguish "explicitly set to empty"
// from "unset", which is consistent with how the existing
// DefaultOutputDir methods test `if p.outputDir != ""`.
//
// Environment variable name is TINCT_PLUGIN_<NAME>_<KEY> with the plugin
// name uppercased and any hyphens or underscores in the key uppercased.
// Hyphens in the plugin name (none today) are converted to underscores.
//
// tinct.toml lookup reads [plugin.<name>] <key> from the cached Config.
// Errors loading config.toml are silently ignored (the resolver simply
// falls through to fallback) — the file is optional and any user with a
// broken config.toml has bigger problems than missing plugin overrides.
func Resolve(plugin, key, flagValue, fallback string) string {
	if flagValue != "" {
		return flagValue
	}
	if v := os.Getenv(envName(plugin, key)); v != "" {
		return v
	}
	if v := tomlValue(plugin, key); v != "" {
		return v
	}
	return fallback
}

// envName derives the environment variable name for a plugin/key. Public
// so callers can document the exact var name in flag help if they wish.
//
//	envName("kitty", "output_dir")     -> "TINCT_PLUGIN_KITTY_OUTPUT_DIR"
//	envName("hyprland", "reload")      -> "TINCT_PLUGIN_HYPRLAND_RELOAD"
//	envName("google-genai", "backend") -> "TINCT_PLUGIN_GOOGLE_GENAI_BACKEND"
func envName(plugin, key string) string {
	clean := func(s string) string {
		return strings.ToUpper(strings.ReplaceAll(s, "-", "_"))
	}
	return "TINCT_PLUGIN_" + clean(plugin) + "_" + clean(key)
}

// tomlValue reads plugin.<name>.<key> from tinct.toml.
func tomlValue(plugin, key string) string {
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		return ""
	}
	section, ok := cfg.Plugins[plugin]
	if !ok {
		return ""
	}
	v, ok := section[key]
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
