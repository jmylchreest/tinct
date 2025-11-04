// Package color provides color extraction and palette generation functionality.
package colour

import (
	"image/color"
	"testing"
)

func TestNewPalette(t *testing.T) {
	colors := []color.Color{
		color.RGBA{R: 255, G: 0, B: 0, A: 255},
		color.RGBA{R: 0, G: 255, B: 0, A: 255},
		color.RGBA{R: 0, G: 0, B: 255, A: 255},
	}

	palette := NewPalette(colors)

	if palette == nil {
		t.Fatal("NewPalette returned nil")
	}

	if palette.Len() != 3 {
		t.Errorf("Expected palette length 3, got %d", palette.Len())
	}
}

func TestPaletteLen(t *testing.T) {
	tests := []struct {
		name   string
		colors []color.Color
		want   int
	}{
		{
			name:   "empty palette",
			colors: []color.Color{},
			want:   0,
		},
		{
			name: "single color",
			colors: []color.Color{
				color.RGBA{R: 255, G: 0, B: 0, A: 255},
			},
			want: 1,
		},
		{
			name: "multiple colors",
			colors: []color.Color{
				color.RGBA{R: 255, G: 0, B: 0, A: 255},
				color.RGBA{R: 0, G: 255, B: 0, A: 255},
				color.RGBA{R: 0, G: 0, B: 255, A: 255},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			palette := NewPalette(tt.colors)
			if got := palette.Len(); got != tt.want {
				t.Errorf("Len() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestToRGB(t *testing.T) {
	tests := []struct {
		name  string
		color color.Color
		want  RGB
	}{
		{
			name:  "red",
			color: color.RGBA{R: 255, G: 0, B: 0, A: 255},
			want:  RGB{R: 255, G: 0, B: 0},
		},
		{
			name:  "green",
			color: color.RGBA{R: 0, G: 255, B: 0, A: 255},
			want:  RGB{R: 0, G: 255, B: 0},
		},
		{
			name:  "blue",
			color: color.RGBA{R: 0, G: 0, B: 255, A: 255},
			want:  RGB{R: 0, G: 0, B: 255},
		},
		{
			name:  "white",
			color: color.RGBA{R: 255, G: 255, B: 255, A: 255},
			want:  RGB{R: 255, G: 255, B: 255},
		},
		{
			name:  "black",
			color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
			want:  RGB{R: 0, G: 0, B: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToRGB(tt.color)
			if got != tt.want {
				t.Errorf("ToRGB() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestRGBHex(t *testing.T) {
	tests := []struct {
		name string
		rgb  RGB
		want string
	}{
		{
			name: "red",
			rgb:  RGB{R: 255, G: 0, B: 0},
			want: "#ff0000",
		},
		{
			name: "green",
			rgb:  RGB{R: 0, G: 255, B: 0},
			want: "#00ff00",
		},
		{
			name: "blue",
			rgb:  RGB{R: 0, G: 0, B: 255},
			want: "#0000ff",
		},
		{
			name: "white",
			rgb:  RGB{R: 255, G: 255, B: 255},
			want: "#ffffff",
		},
		{
			name: "black",
			rgb:  RGB{R: 0, G: 0, B: 0},
			want: "#000000",
		},
		{
			name: "grey",
			rgb:  RGB{R: 128, G: 128, B: 128},
			want: "#808080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rgb.Hex()
			if got != tt.want {
				t.Errorf("Hex() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRGBString(t *testing.T) {
	tests := []struct {
		name string
		rgb  RGB
		want string
	}{
		{
			name: "red",
			rgb:  RGB{R: 255, G: 0, B: 0},
			want: "rgb(255, 0, 0)",
		},
		{
			name: "green",
			rgb:  RGB{R: 0, G: 255, B: 0},
			want: "rgb(0, 255, 0)",
		},
		{
			name: "blue",
			rgb:  RGB{R: 0, G: 0, B: 255},
			want: "rgb(0, 0, 255)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rgb.String()
			if got != tt.want {
				t.Errorf("String() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestPaletteToHex(t *testing.T) {
	colors := []color.Color{
		color.RGBA{R: 255, G: 0, B: 0, A: 255},
		color.RGBA{R: 0, G: 255, B: 0, A: 255},
		color.RGBA{R: 0, G: 0, B: 255, A: 255},
	}

	palette := NewPalette(colors)
	hexColors := palette.ToHex()

	want := []string{"#ff0000", "#00ff00", "#0000ff"}

	if len(hexColors) != len(want) {
		t.Fatalf("ToHex() returned %d colors, want %d", len(hexColors), len(want))
	}

	for i, got := range hexColors {
		if got != want[i] {
			t.Errorf("ToHex()[%d] = %s, want %s", i, got, want[i])
		}
	}
}

func TestPaletteToRGBSlice(t *testing.T) {
	colors := []color.Color{
		color.RGBA{R: 255, G: 0, B: 0, A: 255},
		color.RGBA{R: 0, G: 255, B: 0, A: 255},
		color.RGBA{R: 0, G: 0, B: 255, A: 255},
	}

	palette := NewPalette(colors)
	rgbColors := palette.ToRGBSlice()

	want := []RGB{
		{R: 255, G: 0, B: 0},
		{R: 0, G: 255, B: 0},
		{R: 0, G: 0, B: 255},
	}

	if len(rgbColors) != len(want) {
		t.Fatalf("ToRGBSlice() returned %d colors, want %d", len(rgbColors), len(want))
	}

	for i, got := range rgbColors {
		if got != want[i] {
			t.Errorf("ToRGBSlice()[%d] = %+v, want %+v", i, got, want[i])
		}
	}
}

func TestPaletteToJSON(t *testing.T) {
	colors := []color.Color{
		color.RGBA{R: 255, G: 0, B: 0, A: 255},
		color.RGBA{R: 0, G: 255, B: 0, A: 255},
	}

	palette := NewPalette(colors)
	jsonBytes, err := palette.ToJSON()

	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	if len(jsonBytes) == 0 {
		t.Error("ToJSON() returned empty bytes")
	}

	// Check that it contains expected content.
	jsonStr := string(jsonBytes)
	expectedStrings := []string{
		`"count": 2`,
		`"hex": "#ff0000"`,
		`"hex": "#00ff00"`,
		`"r": 255`,
		`"g": 255`,
	}

	for _, expected := range expectedStrings {
		if !contains(jsonStr, expected) {
			t.Errorf("ToJSON() output missing expected string: %s", expected)
		}
	}
}

func TestPaletteGet(t *testing.T) {
	colors := []color.Color{
		color.RGBA{R: 255, G: 0, B: 0, A: 255},
		color.RGBA{R: 0, G: 255, B: 0, A: 255},
		color.RGBA{R: 0, G: 0, B: 255, A: 255},
	}

	palette := NewPalette(colors)

	tests := []struct {
		name    string
		index   int
		wantErr bool
	}{
		{
			name:    "valid index 0",
			index:   0,
			wantErr: false,
		},
		{
			name:    "valid index 1",
			index:   1,
			wantErr: false,
		},
		{
			name:    "valid index 2",
			index:   2,
			wantErr: false,
		},
		{
			name:    "negative index",
			index:   -1,
			wantErr: true,
		},
		{
			name:    "index out of bounds",
			index:   3,
			wantErr: true,
		},
		{
			name:    "index far out of bounds",
			index:   100,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := palette.Get(tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPaletteAll(t *testing.T) {
	colors := []color.Color{
		color.RGBA{R: 255, G: 0, B: 0, A: 255},
		color.RGBA{R: 0, G: 255, B: 0, A: 255},
		color.RGBA{R: 0, G: 0, B: 255, A: 255},
	}

	palette := NewPalette(colors)

	count := 0
	for i, c := range palette.All() {
		if i != count {
			t.Errorf("Expected index %d, got %d", count, i)
		}
		if c == nil {
			t.Errorf("Color at index %d is nil", i)
		}
		count++
	}

	if count != 3 {
		t.Errorf("Expected to iterate over 3 colors, got %d", count)
	}
}

func TestPaletteString(t *testing.T) {
	tests := []struct {
		name   string
		colors []color.Color
	}{
		{
			name:   "empty palette",
			colors: []color.Color{},
		},
		{
			name: "single color",
			colors: []color.Color{
				color.RGBA{R: 255, G: 0, B: 0, A: 255},
			},
		},
		{
			name: "multiple colors",
			colors: []color.Color{
				color.RGBA{R: 255, G: 0, B: 0, A: 255},
				color.RGBA{R: 0, G: 255, B: 0, A: 255},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			palette := NewPalette(tt.colors)
			str := palette.String()
			if str == "" {
				t.Error("String() returned empty string")
			}
		})
	}
}

// Helper function.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || substr == "" || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
