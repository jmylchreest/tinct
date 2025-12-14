package markdown

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
)

func TestPlugin_Name(t *testing.T) {
	p := New()
	if got := p.Name(); got != "markdown" {
		t.Errorf("Name() = %v, want markdown", got)
	}
}

func TestPlugin_Description(t *testing.T) {
	p := New()
	if got := p.Description(); got == "" {
		t.Error("Description() returned empty string")
	}
}

func TestPlugin_Version(t *testing.T) {
	p := New()
	if got := p.Version(); got == "" {
		t.Error("Version() returned empty string")
	}
}

func TestPlugin_Validate(t *testing.T) {
	tests := []struct {
		name      string
		themeName string
		wantErr   bool
	}{
		{
			name:      "missing name",
			themeName: "",
			wantErr:   true,
		},
		{
			name:      "name provided",
			themeName: "My Theme",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			p.name = tt.themeName
			err := p.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlugin_DefaultOutputDir(t *testing.T) {
	p := New()

	// Test with custom output dir
	p.outputDir = "/custom/path"
	if got := p.DefaultOutputDir(); got != "/custom/path" {
		t.Errorf("DefaultOutputDir() = %v, want /custom/path", got)
	}

	// Test default (should use home dir)
	p.outputDir = ""
	dir := p.DefaultOutputDir()
	if !strings.Contains(dir, ".config/tinct/themes") {
		t.Errorf("DefaultOutputDir() = %v, should contain .config/tinct/themes", dir)
	}
}

func TestPlugin_Generate(t *testing.T) {
	// Create a mock palette
	palette := colour.NewCategorisedPalette(colour.ThemeDark)
	palette.Colours[colour.RoleBackground] = colour.CategorisedColour{
		Hex:    "#1a1b26",
		Weight: 0.25,
		Role:   colour.RoleBackground,
	}
	palette.Colours[colour.RoleForeground] = colour.CategorisedColour{
		Hex:    "#c0caf5",
		Weight: 0.15,
		Role:   colour.RoleForeground,
	}
	palette.Colours[colour.RoleAccent1] = colour.CategorisedColour{
		Hex:    "#7aa2f7",
		Weight: 0.10,
		Role:   colour.RoleAccent1,
	}

	themeData := colour.NewThemeData(palette, "", "Test Theme")

	p := New()
	p.name = "Test Theme"

	files, err := p.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Generate() returned %d files, want 1", len(files))
	}

	// Check filename
	var filename string
	var content []byte
	for k, v := range files {
		filename = k
		content = v
	}

	if filename != "test-theme.md" {
		t.Errorf("Generate() filename = %v, want test-theme.md", filename)
	}

	// Check content has required sections
	contentStr := string(content)
	if !strings.Contains(contentStr, "---") {
		t.Error("Generate() missing YAML front matter delimiters")
	}
	if !strings.Contains(contentStr, "name: Test Theme") {
		t.Error("Generate() missing theme name in front matter")
	}
	if !strings.Contains(contentStr, "# Test Theme") {
		t.Error("Generate() missing markdown title")
	}
	if !strings.Contains(contentStr, "## Color Palette") {
		t.Error("Generate() missing color palette section")
	}
	if !strings.Contains(contentStr, "#1a1b26") {
		t.Error("Generate() missing background color")
	}
}

func TestPlugin_Generate_WithWallpaper(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal PNG file
	pngData := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde,
		0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41, 0x54,
		0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00, 0x00,
		0x00, 0x03, 0x00, 0x01, 0x00, 0x05, 0xfe, 0xd4,
		0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44,
		0xae, 0x42, 0x60, 0x82,
	}

	wallpaperPath := filepath.Join(tmpDir, "wallpaper.png")
	if err := os.WriteFile(wallpaperPath, pngData, 0644); err != nil {
		t.Fatalf("Failed to write wallpaper: %v", err)
	}

	palette := colour.NewCategorisedPalette(colour.ThemeDark)
	palette.Colours[colour.RoleBackground] = colour.CategorisedColour{
		Hex:    "#000000",
		Weight: 1.0,
		Role:   colour.RoleBackground,
	}

	themeData := colour.NewThemeData(palette, "", "Test Theme")
	themeData.WallpaperPath = wallpaperPath

	p := New()
	p.name = "Wallpaper Theme"
	p.noEmbed = false

	files, err := p.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	var content []byte
	for _, v := range files {
		content = v
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "embedded: true") {
		t.Error("Generate() wallpaper not embedded")
	}
	if !strings.Contains(contentStr, "format: png") {
		t.Error("Generate() missing wallpaper format")
	}
}

func TestPlugin_Generate_ExternalWallpaper(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal PNG file
	pngData := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde,
		0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41, 0x54,
		0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00, 0x00,
		0x00, 0x03, 0x00, 0x01, 0x00, 0x05, 0xfe, 0xd4,
		0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44,
		0xae, 0x42, 0x60, 0x82,
	}

	wallpaperPath := filepath.Join(tmpDir, "wallpaper.png")
	if err := os.WriteFile(wallpaperPath, pngData, 0644); err != nil {
		t.Fatalf("Failed to write wallpaper: %v", err)
	}

	palette := colour.NewCategorisedPalette(colour.ThemeDark)
	palette.Colours[colour.RoleBackground] = colour.CategorisedColour{
		Hex:    "#000000",
		Weight: 1.0,
		Role:   colour.RoleBackground,
	}

	themeData := colour.NewThemeData(palette, "", "Test Theme")
	themeData.WallpaperPath = wallpaperPath

	p := New()
	p.name = "External Wallpaper Theme"
	p.noEmbed = true // Don't embed, reference externally

	files, err := p.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	var content []byte
	for _, v := range files {
		content = v
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "embedded: false") {
		t.Error("Generate() wallpaper should not be embedded")
	}
	if !strings.Contains(contentStr, wallpaperPath) {
		t.Error("Generate() missing external wallpaper path")
	}
}

func TestPlugin_Generate_WithAmbientColors(t *testing.T) {
	palette := colour.NewCategorisedPalette(colour.ThemeDark)
	palette.Colours[colour.RoleBackground] = colour.CategorisedColour{
		Hex:    "#000000",
		Weight: 1.0,
		Role:   colour.RoleBackground,
	}
	// Add position/ambient colors
	palette.Colours[colour.RolePositionTopLeft] = colour.CategorisedColour{
		Hex:    "#ff0000",
		Weight: 0.5,
		Role:   colour.RolePositionTopLeft,
	}
	palette.Colours[colour.RolePositionBottomRight] = colour.CategorisedColour{
		Hex:    "#0000ff",
		Weight: 0.5,
		Role:   colour.RolePositionBottomRight,
	}

	themeData := colour.NewThemeData(palette, "", "Test Theme")

	p := New()
	p.name = "Ambient Theme"

	files, err := p.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	var content []byte
	for _, v := range files {
		content = v
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "ambience:") {
		t.Error("Generate() missing ambience section")
	}
	if !strings.Contains(contentStr, "#ff0000") {
		t.Error("Generate() missing top-left ambient color")
	}
	if !strings.Contains(contentStr, "#0000ff") {
		t.Error("Generate() missing bottom-right ambient color")
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Simple Name", "simple-name"},
		{"UPPERCASE", "uppercase"},
		{"Mixed Case Name", "mixed-case-name"},
		{"Name-With-Dashes", "name-with-dashes"},
		{"Name_With_Underscores", "namewithunderscores"},
		{"Name  With  Multiple   Spaces", "name-with-multiple-spaces"},
		{"Special@#$%^Characters!", "specialcharacters"},
		{"---Leading and Trailing---", "leading-and-trailing"},
		{"Tokyo Night Storm", "tokyo-night-storm"},
		{"Gruvbox Dark Hard", "gruvbox-dark-hard"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := slugify(tt.input)
			if got != tt.want {
				t.Errorf("slugify(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestPlugin_GetFlagHelp(t *testing.T) {
	p := New()
	help := p.GetFlagHelp()

	if len(help) == 0 {
		t.Error("GetFlagHelp() returned empty slice")
	}

	// Check for required flags
	hasName := false
	for _, h := range help {
		if h.Name == "markdown.name" {
			hasName = true
			if !h.Required {
				t.Error("markdown.name should be required")
			}
		}
	}

	if !hasName {
		t.Error("GetFlagHelp() missing markdown.name flag")
	}
}
