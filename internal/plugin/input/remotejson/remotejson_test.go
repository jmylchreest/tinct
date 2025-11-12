// Package remotejson provides tests for the remote-json input plugin.
package remotejson

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
)

// TestValidateRequiresURL tests that Validate requires a URL.
func TestValidateRequiresURL(t *testing.T) {
	plugin := New()

	err := plugin.Validate()
	if err == nil {
		t.Error("Expected error when URL is not set")
	}
}

// TestValidateRequiresHTTPScheme tests that Validate requires http:// or https://.
func TestValidateRequiresHTTPScheme(t *testing.T) {
	plugin := New()
	plugin.url = "ftp://example.com/palette.json"

	err := plugin.Validate()
	if err == nil {
		t.Error("Expected error for non-HTTP scheme")
	}
}

// TestValidateAcceptsHTTPScheme tests that Validate accepts http://.
func TestValidateAcceptsHTTPScheme(t *testing.T) {
	plugin := New()
	plugin.url = "http://example.com/palette.json"

	err := plugin.Validate()
	if err != nil {
		t.Errorf("Unexpected error for http:// URL: %v", err)
	}
}

// TestValidateAcceptsHTTPSScheme tests that Validate accepts https://.
func TestValidateAcceptsHTTPSScheme(t *testing.T) {
	plugin := New()
	plugin.url = "https://example.com/palette.json"

	err := plugin.Validate()
	if err != nil {
		t.Errorf("Unexpected error for https:// URL: %v", err)
	}
}

// TestGenerateWithSimpleJSON tests generating palette from simple JSON.
func TestGenerateWithSimpleJSON(t *testing.T) {
	// Create test HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"background": "#1a1b26",
			"foreground": "#c0caf5",
			"accent": "#7aa2f7"
		}`))
	}))
	defer server.Close()

	plugin := New()
	plugin.url = server.URL

	ctx := context.Background()
	opts := input.GenerateOptions{}

	palette, err := plugin.Generate(ctx, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if palette == nil {
		t.Fatal("Generate() returned nil palette")
	}

	if len(palette.Colors) != 3 {
		t.Errorf("Expected 3 colors, got %d", len(palette.Colors))
	}
}

// TestGenerateWithNestedJSON tests generating palette from nested JSON.
func TestGenerateWithNestedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"colors": {
				"primary": {
					"background": "#1a1b26",
					"foreground": "#c0caf5"
				},
				"accents": {
					"blue": "#7aa2f7",
					"purple": "#bb9af7"
				}
			}
		}`))
	}))
	defer server.Close()

	plugin := New()
	plugin.url = server.URL

	ctx := context.Background()
	opts := input.GenerateOptions{}

	palette, err := plugin.Generate(ctx, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(palette.Colors) != 4 {
		t.Errorf("Expected 4 colors, got %d", len(palette.Colors))
	}
}

// TestGenerateWithJSONPathQuery tests applying JSONPath query.
func TestGenerateWithJSONPathQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"theme": {
				"colors": {
					"background": "#1a1b26",
					"foreground": "#c0caf5"
				}
			},
			"metadata": {
				"name": "test"
			}
		}`))
	}))
	defer server.Close()

	plugin := New()
	plugin.url = server.URL
	plugin.query = "$.theme.colors"

	ctx := context.Background()
	opts := input.GenerateOptions{}

	palette, err := plugin.Generate(ctx, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(palette.Colors) != 2 {
		t.Errorf("Expected 2 colors, got %d", len(palette.Colors))
	}
}

// TestGenerateWithColorMapping tests mapping colors to roles.
func TestGenerateWithColorMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"base": "#1a1b26",
			"text": "#c0caf5",
			"blue": "#7aa2f7"
		}`))
	}))
	defer server.Close()

	plugin := New()
	plugin.url = server.URL
	plugin.mapping = map[string]string{
		"base": "background",
		"text": "foreground",
		"blue": "accent1",
	}

	ctx := context.Background()
	opts := input.GenerateOptions{}

	palette, err := plugin.Generate(ctx, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(palette.Colors) != 3 {
		t.Errorf("Expected 3 colors, got %d", len(palette.Colors))
	}

	// Verify role hints are set.
	if len(palette.RoleHints) != 3 {
		t.Errorf("Expected 3 role hints, got %d", len(palette.RoleHints))
	}

	// Verify specific roles are present.
	expectedRoles := []colour.Role{
		colour.RoleBackground,
		colour.RoleForeground,
		colour.RoleAccent1,
	}

	for _, role := range expectedRoles {
		if _, ok := palette.RoleHints[role]; !ok {
			t.Errorf("Expected role %v to be present in role hints", role)
		}
	}
}

// TestGenerateWithCatppuccinFormat tests Catppuccin-style JSON format.
func TestGenerateWithCatppuccinFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"base": {
				"hex": "#1e1e2e",
				"rgb": {"r": 30, "g": 30, "b": 46}
			},
			"text": {
				"hex": "#cdd6f4",
				"rgb": {"r": 205, "g": 214, "b": 244}
			}
		}`))
	}))
	defer server.Close()

	plugin := New()
	plugin.url = server.URL

	ctx := context.Background()
	opts := input.GenerateOptions{}

	palette, err := plugin.Generate(ctx, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(palette.Colors) != 2 {
		t.Errorf("Expected 2 colors, got %d", len(palette.Colors))
	}
}

// TestGenerateWithInvalidJSON tests error handling for invalid JSON.
func TestGenerateWithInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	plugin := New()
	plugin.url = server.URL

	ctx := context.Background()
	opts := input.GenerateOptions{}

	_, err := plugin.Generate(ctx, opts)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// TestGenerateWithInvalidQuery tests error handling for invalid JSONPath query.
func TestGenerateWithInvalidQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"colors": {"bg": "#1a1b26"}}`))
	}))
	defer server.Close()

	plugin := New()
	plugin.url = server.URL
	plugin.query = "$.nonexistent.path"

	ctx := context.Background()
	opts := input.GenerateOptions{}

	_, err := plugin.Generate(ctx, opts)
	if err == nil {
		t.Error("Expected error for invalid query path")
	}
}

// TestGenerateWithNoColors tests error handling when no colors are extracted.
func TestGenerateWithNoColors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"metadata": {"name": "test"}}`))
	}))
	defer server.Close()

	plugin := New()
	plugin.url = server.URL

	ctx := context.Background()
	opts := input.GenerateOptions{}

	_, err := plugin.Generate(ctx, opts)
	if err == nil {
		t.Error("Expected error when no colors are extracted")
	}
}

// TestGenerateWithInvalidColorMapping tests error handling for invalid role mapping.
func TestGenerateWithInvalidColorMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"base": "#1a1b26"}`))
	}))
	defer server.Close()

	plugin := New()
	plugin.url = server.URL
	plugin.mapping = map[string]string{
		"base": "invalidrole",
	}

	ctx := context.Background()
	opts := input.GenerateOptions{}

	_, err := plugin.Generate(ctx, opts)
	if err == nil {
		t.Error("Expected error for invalid role in mapping")
	}
}

// TestGenerateWithHTTPError tests error handling for HTTP errors.
func TestGenerateWithHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	plugin := New()
	plugin.url = server.URL

	ctx := context.Background()
	opts := input.GenerateOptions{}

	_, err := plugin.Generate(ctx, opts)
	if err == nil {
		t.Error("Expected error for HTTP 404")
	}
}

// TestGenerateWithShorthandHex tests parsing shorthand hex colors.
func TestGenerateWithShorthandHex(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"white": "#fff",
			"black": "#000",
			"red": "#f00"
		}`))
	}))
	defer server.Close()

	plugin := New()
	plugin.url = server.URL

	ctx := context.Background()
	opts := input.GenerateOptions{}

	palette, err := plugin.Generate(ctx, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(palette.Colors) != 3 {
		t.Errorf("Expected 3 colors, got %d", len(palette.Colors))
	}
}
