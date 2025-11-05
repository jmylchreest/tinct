// Package file provides tests for the file input plugin.
package file

import (
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
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
