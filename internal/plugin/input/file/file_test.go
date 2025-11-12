// Package file provides tests for the file input plugin.
package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
)

func TestParseColourRole(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    colour.Role
		wantErr bool
	}{
		// British spelling (primary).
		{
			name:  "background - British",
			input: "background",
			want:  colour.RoleBackground,
		},
		{
			name:  "backgroundMuted - British camelCase",
			input: "backgroundMuted",
			want:  colour.RoleBackgroundMuted,
		},
		{
			name:  "foreground - British",
			input: "foreground",
			want:  colour.RoleForeground,
		},
		{
			name:  "foregroundMuted - British camelCase",
			input: "foregroundMuted",
			want:  colour.RoleForegroundMuted,
		},

		// American spelling (aliases - note: these currently work because.
		// the role names don't contain "colour/color", but we test the
		// normalization logic)
		{
			name:  "background - with underscores",
			input: "back_ground",
			want:  colour.RoleBackground,
		},
		{
			name:  "background - with hyphens",
			input: "back-ground",
			want:  colour.RoleBackground,
		},
		{
			name:  "background_muted - mixed style",
			input: "background_muted",
			want:  colour.RoleBackgroundMuted,
		},

		// Case variations.
		{
			name:  "BACKGROUND - uppercase",
			input: "BACKGROUND",
			want:  colour.RoleBackground,
		},
		{
			name:  "ForegroundMuted - mixed case",
			input: "ForegroundMuted",
			want:  colour.RoleForegroundMuted,
		},

		// Accent roles.
		{
			name:  "accent1",
			input: "accent1",
			want:  colour.RoleAccent1,
		},
		{
			name:  "accent2",
			input: "accent2",
			want:  colour.RoleAccent2,
		},
		{
			name:  "accent3",
			input: "accent3",
			want:  colour.RoleAccent3,
		},
		{
			name:  "accent4",
			input: "accent4",
			want:  colour.RoleAccent4,
		},

		// Semantic roles.
		{
			name:  "danger",
			input: "danger",
			want:  colour.RoleDanger,
		},
		{
			name:  "warning",
			input: "warning",
			want:  colour.RoleWarning,
		},
		{
			name:  "success",
			input: "success",
			want:  colour.RoleSuccess,
		},
		{
			name:  "info",
			input: "info",
			want:  colour.RoleInfo,
		},
		{
			name:  "notification",
			input: "notification",
			want:  colour.RoleNotification,
		},

		// Error cases.
		{
			name:    "invalid role",
			input:   "invalidrole",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "unknown colour name",
			input:   "purple",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseColourRole(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseColourRole() expected error for input %q, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("parseColourRole() unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("parseColourRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseColourRoleNormalisation(t *testing.T) {
	// Test that normalisation works correctly by testing equivalent inputs.
	equivalentInputs := []struct {
		name   string
		inputs []string
		want   colour.Role
	}{
		{
			name: "background variants",
			inputs: []string{
				"background",
				"BACKGROUND",
				"Background",
				"back-ground",
				"back_ground",
				"BACK_GROUND",
			},
			want: colour.RoleBackground,
		},
		{
			name: "backgroundMuted variants",
			inputs: []string{
				"backgroundMuted",
				"backgroundmuted",
				"BACKGROUNDMUTED",
				"background-muted",
				"background_muted",
				"BACKGROUND_MUTED",
			},
			want: colour.RoleBackgroundMuted,
		},
		{
			name: "foreground variants",
			inputs: []string{
				"foreground",
				"FOREGROUND",
				"Foreground",
				"fore-ground",
				"fore_ground",
			},
			want: colour.RoleForeground,
		},
	}

	for _, tt := range equivalentInputs {
		t.Run(tt.name, func(t *testing.T) {
			for _, input := range tt.inputs {
				got, err := parseColourRole(input)
				if err != nil {
					t.Errorf("parseColourRole(%q) unexpected error: %v", input, err)
					continue
				}
				if got != tt.want {
					t.Errorf("parseColourRole(%q) = %v, want %v", input, got, tt.want)
				}
			}
		})
	}
}

// TestGenerateWithoutInputs tests that Generate requires either a file or colour overrides.
func TestGenerateWithoutInputs(t *testing.T) {
	plugin := New()

	ctx := context.Background()
	opts := input.GenerateOptions{}

	_, err := plugin.Generate(ctx, opts)
	if err == nil {
		t.Error("Expected error when no file or colour overrides provided")
	}
}

// TestGenerateWithColourOverrides tests generating palette from colour overrides only.
func TestGenerateWithColourOverrides(t *testing.T) {
	plugin := New()
	plugin.colourOverrides = []string{
		"background=#1a1b26",
		"foreground=#c0caf5",
		"accent1=#7aa2f7",
	}

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

	// Verify role hints are set.
	if len(palette.RoleHints) != 3 {
		t.Errorf("Expected 3 role hints, got %d", len(palette.RoleHints))
	}
}

// TestGenerateWithTextFile tests generating palette from a simple text file.
func TestGenerateWithTextFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tinct-file-tests-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file with hex colors (without # prefix to avoid comment parsing).
	textFile := filepath.Join(tempDir, "palette.txt")
	content := `1a1b26
c0caf5
7aa2f7
bb9af7
`
	if err := os.WriteFile(textFile, []byte(content), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	plugin := New()
	plugin.path = textFile

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

// TestGenerateWithRoleBasedTextFile tests generating palette from text file with role assignments.
func TestGenerateWithRoleBasedTextFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tinct-file-tests-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file with role=hex format.
	textFile := filepath.Join(tempDir, "palette-roles.txt")
	content := `background=#1a1b26
foreground=#c0caf5
accent1=#7aa2f7
danger=#f7768e
`
	if err := os.WriteFile(textFile, []byte(content), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	plugin := New()
	plugin.path = textFile

	ctx := context.Background()
	opts := input.GenerateOptions{}

	palette, err := plugin.Generate(ctx, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(palette.Colors) != 4 {
		t.Errorf("Expected 4 colors, got %d", len(palette.Colors))
	}

	// Verify role hints are set.
	if len(palette.RoleHints) != 4 {
		t.Errorf("Expected 4 role hints, got %d", len(palette.RoleHints))
	}

	// Verify specific roles are present.
	expectedRoles := []colour.Role{
		colour.RoleBackground,
		colour.RoleForeground,
		colour.RoleAccent1,
		colour.RoleDanger,
	}

	for _, role := range expectedRoles {
		if _, ok := palette.RoleHints[role]; !ok {
			t.Errorf("Expected role %v to be present in role hints", role)
		}
	}
}

// TestGenerateWithFileAndOverrides tests merging file palette with colour overrides.
func TestGenerateWithFileAndOverrides(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tinct-file-tests-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create base palette file.
	textFile := filepath.Join(tempDir, "base.txt")
	content := `background=#1a1b26
foreground=#c0caf5
`
	if err := os.WriteFile(textFile, []byte(content), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	plugin := New()
	plugin.path = textFile
	plugin.colourOverrides = []string{
		"foreground=#ffffff", // Override existing
		"accent1=#7aa2f7",    // Add new
	}

	ctx := context.Background()
	opts := input.GenerateOptions{}

	palette, err := plugin.Generate(ctx, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Should have 3 colors: background (from file), foreground (overridden), accent1 (new).
	if len(palette.Colors) != 3 {
		t.Errorf("Expected 3 colors, got %d", len(palette.Colors))
	}

	if len(palette.RoleHints) != 3 {
		t.Errorf("Expected 3 role hints, got %d", len(palette.RoleHints))
	}
}

// TestGenerateWithInvalidFile tests error handling for invalid files.
func TestGenerateWithInvalidFile(t *testing.T) {
	plugin := New()
	plugin.path = "/nonexistent/file.txt"

	ctx := context.Background()
	opts := input.GenerateOptions{}

	_, err := plugin.Generate(ctx, opts)
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

// TestGenerateWithInvalidHexInFile tests error handling for invalid hex colors in file.
func TestGenerateWithInvalidHexInFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tinct-file-tests-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	textFile := filepath.Join(tempDir, "invalid.txt")
	content := `1a1b26
GGGGGG
`
	if err := os.WriteFile(textFile, []byte(content), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	plugin := New()
	plugin.path = textFile

	ctx := context.Background()
	opts := input.GenerateOptions{}

	_, err = plugin.Generate(ctx, opts)
	if err == nil {
		t.Error("Expected error for invalid hex color")
	}
}

// TestGenerateWithInvalidRole tests error handling for invalid role names.
func TestGenerateWithInvalidRole(t *testing.T) {
	plugin := New()
	plugin.colourOverrides = []string{
		"invalidrole=#1a1b26",
	}

	ctx := context.Background()
	opts := input.GenerateOptions{}

	_, err := plugin.Generate(ctx, opts)
	if err == nil {
		t.Error("Expected error for invalid role name")
	}
}

// TestGenerateWithShorthandHex tests parsing shorthand hex format (#RGB).
func TestGenerateWithShorthandHex(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tinct-file-tests-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	textFile := filepath.Join(tempDir, "shorthand.txt")
	content := `fff
000
f00
`
	if err := os.WriteFile(textFile, []byte(content), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	plugin := New()
	plugin.path = textFile

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

// TestParseTextFormat tests the parseTextFormat function directly.
func TestParseTextFormat(t *testing.T) {
	plugin := New()

	content := `1a1b26
c0caf5
7aa2f7
`

	colors, hints, err := plugin.parseTextFormat(content)
	if err != nil {
		t.Fatalf("parseTextFormat() error = %v", err)
	}

	if len(colors) != 3 {
		t.Errorf("Expected 3 colors, got %d", len(colors))
	}

	if len(hints) != 0 {
		t.Errorf("Expected 0 hints (no roles), got %d", len(hints))
	}
}
