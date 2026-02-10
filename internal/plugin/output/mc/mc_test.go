package mc

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/shared/testing"
	pkgcolour "github.com/jmylchreest/tinct/pkg/colour"
)

// TestMCPlugin runs all standard plugin tests using shared utilities.
func TestMCPlugin(t *testing.T) {
	plugin := New()

	config := plugintesting.TestConfig{
		ExpectedName:         "mc",
		ExpectedFiles:        []string{"tinct.ini"},
		ExpectedBinaryName:   "mc",
		ExpectedDirSubstring: "mc/skins",
	}

	plugintesting.RunAllTests(t, plugin, config)
}

// TestMCPlugin_ContentValidation tests mc-specific content requirements.
func TestMCPlugin_ContentValidation(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["tinct.ini"])

	// Check for required INI sections.
	requiredSections := []string{
		"[skin]",
		"[lines]",
		"[core]",
		"[dialog]",
		"[error]",
		"[filehighlight]",
		"[menu]",
		"[popupmenu]",
		"[buttonbar]",
		"[statusbar]",
		"[help]",
		"[editor]",
		"[viewer]",
		"[diffviewer]",
		"[widget-panel]",
		"[widget-scrollbar]",
		"[widget-editor]",
	}

	for _, section := range requiredSections {
		if !strings.Contains(content, section) {
			t.Errorf("Generated content missing required section: %s", section)
		}
	}

	// Check 256colors flag is set.
	if !strings.Contains(content, "256colors = true") {
		t.Error("Generated content missing 256colors = true")
	}

	// Check that the skin description is present.
	if !strings.Contains(content, "description = Tinct") {
		t.Error("Generated content missing skin description")
	}
}

// TestMCPlugin_ColorFormat tests that all colors use the colorN format.
func TestMCPlugin_ColorFormat(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["tinct.ini"])

	// Find all colorN references and verify they're in valid range (16-255).
	colorPattern := regexp.MustCompile(`color(\d+)`)
	matches := colorPattern.FindAllStringSubmatch(content, -1)

	if len(matches) == 0 {
		t.Fatal("No colorN references found in generated content")
	}

	for _, match := range matches {
		var idx int
		if _, err := fmt.Sscanf(match[1], "%d", &idx); err != nil {
			t.Errorf("Failed to parse color index from %s: %v", match[0], err)
			continue
		}
		if idx < 16 || idx > 255 {
			t.Errorf("Color index %d out of valid range (16-255): %s", idx, match[0])
		}
	}
}

// TestMCPlugin_DarkTheme tests dark theme generation.
func TestMCPlugin_DarkTheme(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["tinct.ini"])

	if !strings.Contains(content, "Detected theme: dark") {
		t.Error("Generated content missing dark theme type")
	}

	if !strings.Contains(content, "description = Tinct - dark theme") {
		t.Error("Generated content missing dark theme description")
	}
}

// TestMCPlugin_LightTheme tests light theme generation.
func TestMCPlugin_LightTheme(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeLight)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["tinct.ini"])

	if !strings.Contains(content, "Detected theme: light") {
		t.Error("Generated content missing light theme type")
	}

	if !strings.Contains(content, "description = Tinct - light theme") {
		t.Error("Generated content missing light theme description")
	}
}

// TestMCPlugin_FileHighlightColors tests that file highlight colors follow
// Unix conventions (directories=accent1, executables=success, etc.)
func TestMCPlugin_FileHighlightColors(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["tinct.ini"])

	// Check that filehighlight section has key entries.
	requiredEntries := []string{
		"directory",
		"executable",
		"symlink",
		"stalelink",
		"archive",
		"source",
		"doc",
	}

	// Extract filehighlight section.
	fhStart := strings.Index(content, "[filehighlight]")
	if fhStart == -1 {
		t.Fatal("Missing [filehighlight] section")
	}
	// Find next section.
	nextSection := strings.Index(content[fhStart+1:], "[")
	var fhContent string
	if nextSection == -1 {
		fhContent = content[fhStart:]
	} else {
		fhContent = content[fhStart : fhStart+1+nextSection]
	}

	for _, entry := range requiredEntries {
		if !strings.Contains(fhContent, entry) {
			t.Errorf("File highlight section missing entry: %s", entry)
		}
	}
}

// TestMCPlugin_DiffViewerColors tests diff viewer color assignments.
func TestMCPlugin_DiffViewerColors(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["tinct.ini"])

	// Check diffviewer section exists with key entries.
	dvStart := strings.Index(content, "[diffviewer]")
	if dvStart == -1 {
		t.Fatal("Missing [diffviewer] section")
	}
	nextSection := strings.Index(content[dvStart+1:], "[")
	var dvContent string
	if nextSection == -1 {
		dvContent = content[dvStart:]
	} else {
		dvContent = content[dvStart : dvStart+1+nextSection]
	}

	requiredEntries := []string{
		"added",
		"removed",
		"changed",
	}

	for _, entry := range requiredEntries {
		if !strings.Contains(dvContent, entry) {
			t.Errorf("Diff viewer section missing entry: %s", entry)
		}
	}
}

// TestMCPlugin_CustomOutputDir tests custom output directory handling.
func TestMCPlugin_CustomOutputDir(t *testing.T) {
	plugin := New()
	plugin.outputDir = "/custom/path"

	dir := plugin.DefaultOutputDir()
	if dir != "/custom/path" {
		t.Errorf("DefaultOutputDir() = %s, want /custom/path", dir)
	}
}

// TestMCPlugin_XDGDataHome tests XDG_DATA_HOME respect.
func TestMCPlugin_XDGDataHome(t *testing.T) {
	plugin := New()
	t.Setenv("XDG_DATA_HOME", "/tmp/test-xdg-data")

	dir := plugin.DefaultOutputDir()
	expected := "/tmp/test-xdg-data/mc/skins"
	if dir != expected {
		t.Errorf("DefaultOutputDir() = %s, want %s", dir, expected)
	}
}

// TestMCPlugin_GetEmbeddedTemplates tests embedded template access.
func TestMCPlugin_GetEmbeddedTemplates(t *testing.T) {
	fs := GetEmbeddedTemplates()

	entries, err := fs.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	found := false
	for _, entry := range entries {
		if entry.Name() == "tinct.ini.tmpl" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Template file tinct.ini.tmpl not found in embedded filesystem")
	}
}

// TestMCPlugin_LineDrawingChars tests that line drawing characters are present.
func TestMCPlugin_LineDrawingChars(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["tinct.ini"])

	// Check line drawing section has proper Unicode characters.
	lineChars := []string{
		"horiz = ─",
		"vert = │",
		"lefttop = ┌",
		"righttop = ┐",
		"leftbottom = └",
		"rightbottom = ┘",
	}

	for _, lc := range lineChars {
		if !strings.Contains(content, lc) {
			t.Errorf("Missing line drawing character: %s", lc)
		}
	}
}

// TestXterm256_UsedByTemplate tests that the xterm256 template function produces
// valid colorN strings for representative palette colors.
func TestXterm256_UsedByTemplate(t *testing.T) {
	cv := pkgcolour.NewColorValue(
		pkgcolour.RGBA{R: 122, G: 162, B: 247, A: 255},
		"accent1",
		0,
	)

	result := pkgcolour.NearestXterm256String(cv.R(), cv.G(), cv.B())

	if !strings.HasPrefix(result, "color") {
		t.Errorf("NearestXterm256String() = %s, expected format colorN", result)
	}

	// Parse the index.
	var idx int
	if _, err := fmt.Sscanf(result, "color%d", &idx); err != nil {
		t.Fatalf("Failed to parse index from %s: %v", result, err)
	}
	if idx < 16 || idx > 255 {
		t.Errorf("NearestXterm256String() index %d out of range [16,255]", idx)
	}
}
