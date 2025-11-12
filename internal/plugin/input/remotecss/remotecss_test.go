// Package remotecss provides tests for the remote-css input plugin.
package remotecss

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
	plugin.url = "ftp://example.com/palette.css"

	err := plugin.Validate()
	if err == nil {
		t.Error("Expected error for non-HTTP scheme")
	}
}

// TestValidateAcceptsHTTPScheme tests that Validate accepts http://.
func TestValidateAcceptsHTTPScheme(t *testing.T) {
	plugin := New()
	plugin.url = "http://example.com/palette.css"

	err := plugin.Validate()
	if err != nil {
		t.Errorf("Unexpected error for http:// URL: %v", err)
	}
}

// TestValidateAcceptsHTTPSScheme tests that Validate accepts https://.
func TestValidateAcceptsHTTPSScheme(t *testing.T) {
	plugin := New()
	plugin.url = "https://example.com/palette.css"

	err := plugin.Validate()
	if err != nil {
		t.Errorf("Unexpected error for https:// URL: %v", err)
	}
}

// TestGenerateWithCSSVariables tests generating palette from CSS custom properties.
func TestGenerateWithCSSVariables(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`:root {
	--color-background: #1a1b26;
	--color-foreground: #c0caf5;
	--color-accent: #7aa2f7;
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

// TestGenerateWithColorProperties tests extracting colors from CSS color properties.
func TestGenerateWithColorProperties(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`.class {
	color: #1a1b26;
	background-color: #c0caf5;
	border-color: #7aa2f7;
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

// TestGenerateWithRGBColors tests parsing RGB color format.
func TestGenerateWithRGBColors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`:root {
	--color-1: rgb(26, 27, 38);
	--color-2: rgb(192, 202, 245);
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

// TestGenerateWithHSLColors tests parsing HSL color format.
func TestGenerateWithHSLColors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`:root {
	--color-1: hsl(224, 31%, 15%);
	--color-2: hsl(226, 64%, 88%);
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

// TestGenerateWithOKLCHColors tests parsing OKLCH color format.
func TestGenerateWithOKLCHColors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`:root {
	--color-1: oklch(0.25 0.05 264);
	--color-2: oklch(0.85 0.05 264);
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

// TestGenerateWithOKLABColors tests parsing OKLAB color format.
func TestGenerateWithOKLABColors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`:root {
	--color-1: oklab(0.25 0.02 -0.05);
	--color-2: oklab(0.85 0.02 -0.05);
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

// TestGenerateWithColorMapping tests mapping colors to roles.
func TestGenerateWithColorMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`:root {
	--color-base: #1a1b26;
	--color-text: #c0caf5;
	--color-blue: #7aa2f7;
}`))
	}))
	defer server.Close()

	plugin := New()
	plugin.url = server.URL
	plugin.mapping = map[string]string{
		"color-base": "background",
		"color-text": "foreground",
		"color-blue": "accent1",
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

// TestGenerateWithMixedFormats tests parsing multiple color formats in one CSS.
func TestGenerateWithMixedFormats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`:root {
	--color-1: #1a1b26;
	--color-2: rgb(192, 202, 245);
	--color-3: hsl(224, 64%, 70%);
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

// TestGenerateWithShorthandHex tests parsing shorthand hex colors.
func TestGenerateWithShorthandHex(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`:root {
	--color-white: #fff;
	--color-black: #000;
	--color-red: #f00;
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

// TestGenerateWithNoColors tests error handling when no colors are found.
func TestGenerateWithNoColors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`.class {
	font-size: 16px;
	padding: 10px;
}`))
	}))
	defer server.Close()

	plugin := New()
	plugin.url = server.URL

	ctx := context.Background()
	opts := input.GenerateOptions{}

	_, err := plugin.Generate(ctx, opts)
	if err == nil {
		t.Error("Expected error when no colors are found")
	}
}

// TestGenerateWithInvalidColorMapping tests error handling for invalid role mapping.
func TestGenerateWithInvalidColorMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`:root {
	--color-base: #1a1b26;
}`))
	}))
	defer server.Close()

	plugin := New()
	plugin.url = server.URL
	plugin.mapping = map[string]string{
		"color-base": "invalidrole",
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

// TestGenerateWithRGBAColors tests parsing RGBA color format (alpha is ignored).
func TestGenerateWithRGBAColors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`:root {
	--color-1: rgba(26, 27, 38, 0.9);
	--color-2: rgba(192, 202, 245, 1.0);
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
