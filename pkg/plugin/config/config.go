// Package config provides a public, deterministic resolver for
// plugin-scoped configuration values. External Go plugins import this
// to honour the same env-var override mechanism that internal plugins
// receive via internal/pluginconfig.Resolve.
//
// Precedence (highest first):
//
//  1. flagValue  — caller passes the value cobra/pflag has already bound.
//  2. environment variable TINCT_PLUGIN_<NAME>_<KEY>.
//  3. fallback   — typically the plugin's platform default.
//
// The tinct.toml [plugin.<name>] <key> tier present in the internal
// resolver is intentionally absent here: external plugins live in their
// own modules and shouldn't depend on tinct's internal config loader.
// Tinct itself can propagate tinct.toml values to the plugin process by
// pre-setting the env var if/when that becomes desirable.
package config

import (
	"os"
	"strings"
)

// Resolve returns the effective string value for plugin.<name>.<key>.
//
// Plugin name and key are normalised for the env var: uppercase, with
// hyphens and underscores both treated as `_`. Examples:
//
//	Resolve("kitty", "output_dir", "", "/foo")     // → $TINCT_PLUGIN_KITTY_OUTPUT_DIR or "/foo"
//	Resolve("kde-plasma", "scheme", "", "default") // → $TINCT_PLUGIN_KDE_PLASMA_SCHEME or "default"
//
// Empty flagValue means "user did not pass --<plugin>.<key>"; the
// resolver falls through to env then fallback. This matches how the
// existing tinct plugins test `if p.outputDir != ""` to detect a
// caller-provided override.
func Resolve(pluginName, key, flagValue, fallback string) string {
	if flagValue != "" {
		return flagValue
	}
	if v := os.Getenv(EnvName(pluginName, key)); v != "" {
		return v
	}
	return fallback
}

// EnvName derives the environment variable name for a plugin/key. Public
// so callers can document the exact var name in flag help if they wish.
//
//	EnvName("kitty", "output_dir")     -> "TINCT_PLUGIN_KITTY_OUTPUT_DIR"
//	EnvName("google-genai", "backend") -> "TINCT_PLUGIN_GOOGLE_GENAI_BACKEND"
func EnvName(pluginName, key string) string {
	clean := func(s string) string {
		return strings.ToUpper(strings.ReplaceAll(s, "-", "_"))
	}
	return "TINCT_PLUGIN_" + clean(pluginName) + "_" + clean(key)
}
