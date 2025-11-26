// Package kdeplasma provides an output plugin for KDE Plasma desktop environment theming.
package kdeplasma

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/common"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/jmylchreest/tinct/internal/version"
)

//go:embed *.tmpl
var templates embed.FS

// Plugin implements the output.Plugin interface for KDE Plasma themes.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new KDE Plasma plugin instance.
func New() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "kde-plasma"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "KDE Plasma desktop environment theme (color schemes)"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers command-line flags for this plugin.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "kde-plasma.output-dir", "",
		"Output directory for KDE color scheme files (default: ~/.local/share/color-schemes)")
}

// SetVerbose enables verbose output.
func (p *Plugin) SetVerbose(verbose bool) {
	p.verbose = verbose
}

// GetEmbeddedFS returns the embedded filesystem containing templates.
func (p *Plugin) GetEmbeddedFS() any {
	return templates
}

// GetFlagHelp returns help text for plugin flags.
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
		{
			Name:        "kde-plasma.output-dir",
			Type:        "string",
			Default:     "~/.local/share/color-schemes",
			Description: "Output directory for KDE color scheme files",
			Required:    false,
		},
	}
}

// Validate validates the plugin configuration.
func (p *Plugin) Validate() error {
	return nil
}

// DefaultOutputDir returns the default output directory.
func (p *Plugin) DefaultOutputDir() string {
	if p.outputDir != "" {
		return p.outputDir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ".local/share/color-schemes"
	}
	return filepath.Join(home, ".local", "share", "color-schemes")
}

// PreExecute checks if KDE Plasma is installed.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	// Check if KDE config directory exists
	home, err := os.UserHomeDir()
	if err != nil {
		return true, "Could not determine home directory", nil
	}

	kdeConfigDir := filepath.Join(home, ".config", "kdeglobals")
	plasmarcPath := filepath.Join(home, ".config", "plasmarc")

	// Check for either kdeglobals or plasmarc to determine if KDE is installed
	_, kdeErr := os.Stat(kdeConfigDir)
	_, plasmaErr := os.Stat(plasmarcPath)

	if os.IsNotExist(kdeErr) && os.IsNotExist(plasmaErr) {
		return true, fmt.Sprintf(
			"KDE Plasma does not appear to be installed.\n"+
				"  Neither %s nor %s exist.\n"+
				"  This plugin is for KDE Plasma desktop environment.\n"+
				"  For Qt apps on other desktops, use the qt5 or qt6 plugins instead.",
			kdeConfigDir, plasmarcPath,
		), nil
	}

	return false, "", nil
}

// Generate creates the KDE Plasma color scheme file.
// This is called when only a single theme variant is requested.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate the KDE color scheme file
	colorsContent, err := p.generateColorScheme(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate KDE color scheme: %w", err)
	}

	var themeName string
	if themeData.ThemeType() == colour.ThemeDark {
		themeName = "TinctDark"
	} else {
		themeName = "TinctLight"
	}

	files[themeName+".colors"] = colorsContent

	return files, nil
}

// GenerateDualTheme creates both light and dark KDE Plasma color scheme files.
// This allows KDE to automatically switch themes based on system preference.
func (p *Plugin) GenerateDualTheme(primaryTheme, alternateTheme *colour.ThemeData) (map[string][]byte, error) {
	if primaryTheme == nil {
		return nil, fmt.Errorf("primary theme data cannot be nil")
	}
	if alternateTheme == nil {
		return nil, fmt.Errorf("alternate theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate primary theme
	primaryContent, err := p.generateColorScheme(primaryTheme)
	if err != nil {
		return nil, fmt.Errorf("failed to generate primary color scheme: %w", err)
	}

	var primaryName string
	if primaryTheme.ThemeType() == colour.ThemeDark {
		primaryName = "TinctDark"
	} else {
		primaryName = "TinctLight"
	}
	files[primaryName+".colors"] = primaryContent

	// Generate alternate theme
	alternateContent, err := p.generateColorScheme(alternateTheme)
	if err != nil {
		return nil, fmt.Errorf("failed to generate alternate color scheme: %w", err)
	}

	var alternateName string
	if alternateTheme.ThemeType() == colour.ThemeDark {
		alternateName = "TinctDark"
	} else {
		alternateName = "TinctLight"
	}
	files[alternateName+".colors"] = alternateContent

	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Generated dual-theme: %s and %s\n", primaryName, alternateName)
	}

	return files, nil
}

// generateColorScheme generates the KDE color scheme content.
func (p *Plugin) generateColorScheme(themeData *colour.ThemeData) ([]byte, error) {
	loader := tmplloader.New("kde-plasma", templates)
	if p.verbose {
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}

	tmplContent, _, err := loader.Load("tinct.colors.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	tmpl, err := template.New("kde-plasma").
		Funcs(common.TemplateFuncs()).
		Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// PostExecute automatically applies the KDE Plasma color scheme.
// When dual-theme is generated, it applies the theme matching the current system preference.
func (p *Plugin) PostExecute(ctx context.Context, execCtx output.ExecutionContext, generatedFiles []string) error {
	if execCtx.DryRun {
		return nil
	}

	// Determine which theme files were generated
	hasDark := false
	hasLight := false
	for _, file := range generatedFiles {
		if strings.HasSuffix(file, "TinctDark.colors") {
			hasDark = true
		} else if strings.HasSuffix(file, "TinctLight.colors") {
			hasLight = true
		}
	}

	if !hasDark && !hasLight {
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Warning: Could not determine theme name from generated files\n")
		}
		return nil
	}

	// Check if plasma-apply-colorscheme is available
	if _, err := exec.LookPath("plasma-apply-colorscheme"); err != nil {
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Warning: plasma-apply-colorscheme not found, theme not auto-applied\n")
			if hasDark && hasLight {
				fmt.Fprintf(os.Stderr, "   To apply manually: plasma-apply-colorscheme TinctDark (or TinctLight)\n")
			} else if hasDark {
				fmt.Fprintf(os.Stderr, "   To apply manually: plasma-apply-colorscheme TinctDark\n")
			} else {
				fmt.Fprintf(os.Stderr, "   To apply manually: plasma-apply-colorscheme TinctLight\n")
			}
		}
		return nil
	}

	// Determine which theme to apply
	themeName := ""
	if hasDark && hasLight {
		// Dual-theme: detect system preference
		themeName = p.detectSystemColorScheme(ctx)
		if themeName == "" {
			// Couldn't detect, default to dark
			themeName = "TinctDark"
			if p.verbose {
				fmt.Fprintf(os.Stderr, "   Could not detect system color preference, defaulting to TinctDark\n")
			}
		}
	} else if hasDark {
		themeName = "TinctDark"
	} else {
		themeName = "TinctLight"
	}

	// Apply the color scheme
	cmd := exec.CommandContext(ctx, "plasma-apply-colorscheme", themeName)
	if err := cmd.Run(); err != nil {
		// Don't fail, just warn
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Warning: Failed to apply color scheme: %v\n", err)
			fmt.Fprintf(os.Stderr, "   To apply manually: plasma-apply-colorscheme %s\n", themeName)
		}
		return nil
	}

	fmt.Fprintf(os.Stderr, "   plasma-apply-colorscheme: Applied %s\n", themeName)

	// Apply wallpaper if available
	if execCtx.WallpaperPath != "" {
		if err := p.applyWallpaper(ctx, execCtx.WallpaperPath); err != nil {
			if p.verbose {
				fmt.Fprintf(os.Stderr, "   Warning: Failed to apply wallpaper: %v\n", err)
			}
		}
	}

	return nil
}

// applyWallpaper applies the wallpaper using plasma-apply-wallpaperimage.
func (p *Plugin) applyWallpaper(ctx context.Context, wallpaperPath string) error {
	// Check if plasma-apply-wallpaperimage is available
	if _, err := exec.LookPath("plasma-apply-wallpaperimage"); err != nil {
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Warning: plasma-apply-wallpaperimage not found, wallpaper not auto-applied\n")
			fmt.Fprintf(os.Stderr, "   To apply manually: plasma-apply-wallpaperimage %s\n", wallpaperPath)
		}
		return err
	}

	// Apply the wallpaper
	cmd := exec.CommandContext(ctx, "plasma-apply-wallpaperimage", wallpaperPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute plasma-apply-wallpaperimage: %w", err)
	}

	fmt.Fprintf(os.Stderr, "   plasma-apply-wallpaperimage: Applied %s\n", wallpaperPath)
	return nil
}

// detectSystemColorScheme detects the current KDE system color scheme preference.
// Returns "TinctDark" or "TinctLight", or empty string if detection fails.
func (p *Plugin) detectSystemColorScheme(ctx context.Context) string {
	// KDE stores color scheme preference in kdeglobals
	// We can check if the current color scheme name contains "dark" or "light"
	cmd := exec.CommandContext(ctx, "kreadconfig5", "--file", "kdeglobals", "--group", "General", "--key", "ColorScheme")
	output, err := cmd.Output()
	if err == nil {
		currentScheme := strings.ToLower(strings.TrimSpace(string(output)))
		if strings.Contains(currentScheme, "dark") {
			return "TinctDark"
		} else if strings.Contains(currentScheme, "light") {
			return "TinctLight"
		}
	}

	// Fallback: check if plasma is in dark mode via lookAndFeelPackage
	cmd = exec.CommandContext(ctx, "kreadconfig5", "--file", "kdeglobals", "--group", "KDE", "--key", "LookAndFeelPackage")
	output, err = cmd.Output()
	if err == nil {
		lnf := strings.ToLower(strings.TrimSpace(string(output)))
		if strings.Contains(lnf, "dark") {
			return "TinctDark"
		} else if strings.Contains(lnf, "light") {
			return "TinctLight"
		}
	}

	return ""
}
