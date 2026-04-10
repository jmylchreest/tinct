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
	kdedbus "github.com/jmylchreest/tinct/internal/plugin/output/shared/dbus_kde"
	"github.com/jmylchreest/tinct/internal/plugin/output/shared/utils"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/jmylchreest/tinct/internal/version"
	"github.com/jmylchreest/tinct/pkg/dbus"
)

//go:embed *.tmpl
var templates embed.FS

const (
	themeDark  = "TinctDark"
	themeLight = "TinctLight"
)

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

// Generate creates the KDE Plasma color scheme files.
// This is called when only a single theme variant is requested.
// Generates both variant 1 and 2 to enable alternating for plasma-apply-colorscheme reload workaround.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Determine base theme name
	var baseThemeName string
	if themeData.ThemeType() == colour.ThemeDark {
		baseThemeName = themeDark
	} else {
		baseThemeName = themeLight
	}

	// Generate both variants (workaround for plasma-apply-colorscheme not reloading same scheme name)
	for variant := 1; variant <= 2; variant++ {
		colorsContent, err := p.generateColorSchemeVariant(themeData, variant)
		if err != nil {
			return nil, fmt.Errorf("failed to generate KDE color scheme variant %d: %w", variant, err)
		}

		fileName := fmt.Sprintf("%s%d.colors", baseThemeName, variant)
		files[fileName] = colorsContent
	}

	return files, nil
}

// GenerateDualTheme creates both light and dark KDE Plasma color scheme files.
// This allows KDE to automatically switch themes based on system preference.
// Generates both variant 1 and 2 for each theme to enable alternating for plasma-apply-colorscheme reload workaround.
func (p *Plugin) GenerateDualTheme(primaryTheme, alternateTheme *colour.ThemeData) (map[string][]byte, error) {
	if primaryTheme == nil {
		return nil, fmt.Errorf("primary theme data cannot be nil")
	}
	if alternateTheme == nil {
		return nil, fmt.Errorf("alternate theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Determine base theme names
	var primaryBaseName string
	if primaryTheme.ThemeType() == colour.ThemeDark {
		primaryBaseName = themeDark
	} else {
		primaryBaseName = themeLight
	}

	var alternateBaseName string
	if alternateTheme.ThemeType() == colour.ThemeDark {
		alternateBaseName = themeDark
	} else {
		alternateBaseName = themeLight
	}

	// Generate both variants for primary theme
	for variant := 1; variant <= 2; variant++ {
		primaryContent, err := p.generateColorSchemeVariant(primaryTheme, variant)
		if err != nil {
			return nil, fmt.Errorf("failed to generate primary color scheme variant %d: %w", variant, err)
		}
		fileName := fmt.Sprintf("%s%d.colors", primaryBaseName, variant)
		files[fileName] = primaryContent
	}

	// Generate both variants for alternate theme
	for variant := 1; variant <= 2; variant++ {
		alternateContent, err := p.generateColorSchemeVariant(alternateTheme, variant)
		if err != nil {
			return nil, fmt.Errorf("failed to generate alternate color scheme variant %d: %w", variant, err)
		}
		fileName := fmt.Sprintf("%s%d.colors", alternateBaseName, variant)
		files[fileName] = alternateContent
	}

	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Generated dual-theme with variants: %s1/2 and %s1/2\n", primaryBaseName, alternateBaseName)
	}

	return files, nil
}

// generateColorSchemeVariant generates the KDE color scheme content with a specific variant number.
// The variant number (1 or 2) is used in the ColorScheme and Name fields to enable
// the workaround for plasma-apply-colorscheme not reloading when the scheme name is unchanged.
func (p *Plugin) generateColorSchemeVariant(themeData *colour.ThemeData, variant int) ([]byte, error) {
	loader := tmplloader.New("kde-plasma", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}

	tmplContent, _, err := loader.Load("tinct.colors.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	tmpl, err := template.New("kde-plasma").
		Funcs(utils.TemplateFuncs()).
		Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Post-process to add variant number to ColorScheme and Name
	content := buf.String()

	// Determine theme type
	var themeType string
	if themeData.ThemeType() == colour.ThemeDark {
		themeType = "Dark"
	} else {
		themeType = "Light"
	}

	// Replace ColorScheme and Name to include variant
	variantStr := fmt.Sprintf("%d", variant)
	content = strings.Replace(content,
		fmt.Sprintf("ColorScheme=Tinct%s\n", themeType),
		fmt.Sprintf("ColorScheme=Tinct%s%s\n", themeType, variantStr), 1)
	content = strings.Replace(content,
		fmt.Sprintf("Name=Tinct %s\n", themeType),
		fmt.Sprintf("Name=Tinct %s %s\n", themeType, variantStr), 1)

	return []byte(content), nil
}

// PostExecute automatically applies the KDE Plasma color scheme.
// When dual-theme is generated, it applies the theme matching the current system preference.
// Both plasma-apply-colorscheme (with variant toggling) and D-Bus reload are used together:
// - plasma-apply-colorscheme applies the .colors file system-wide
// - D-Bus reload ensures KWin and Plasma Shell reload their configurations immediately.
func (p *Plugin) PostExecute(ctx context.Context, execCtx output.ExecutionContext, generatedFiles []string) error { //nolint:gocyclo,gocognit // KDE theme application with variant toggling and D-Bus reload
	if execCtx.DryRun {
		return nil
	}

	// Determine which theme files were generated (check for variants)

	// Determine which theme files were generated (check for variants)
	hasDark := false
	hasLight := false
	for _, file := range generatedFiles {
		if strings.Contains(file, themeDark) && strings.HasSuffix(file, ".colors") {
			hasDark = true
		} else if strings.Contains(file, themeLight) && strings.HasSuffix(file, ".colors") {
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
			switch {
			case hasDark && hasLight:
				fmt.Fprintf(os.Stderr, "   To apply manually: plasma-apply-colorscheme TinctDark1 (or TinctLight1)\n")
			case hasDark:
				fmt.Fprintf(os.Stderr, "   To apply manually: plasma-apply-colorscheme TinctDark1\n")
			default:
				fmt.Fprintf(os.Stderr, "   To apply manually: plasma-apply-colorscheme TinctLight1\n")
			}
		}
		return nil
	}

	// Determine base theme name (without variant) based on system preference
	var baseThemeName string
	switch {
	case hasDark && hasLight:
		// Dual-theme: detect system preference
		detected := p.detectSystemColorScheme(ctx)
		switch {
		case strings.Contains(detected, "Dark"):
			baseThemeName = themeDark
		case strings.Contains(detected, "Light"):
			baseThemeName = themeLight
		default:
			baseThemeName = themeDark
			if p.verbose {
				fmt.Fprintf(os.Stderr, "   Could not detect system color preference, defaulting to TinctDark\n")
			}
		}
	case hasDark:
		baseThemeName = themeDark
	default:
		baseThemeName = themeLight
	}

	// Determine which variant to apply by checking current scheme
	variantToApply := p.determineVariantToApply(ctx, baseThemeName)
	themeNameToApply := fmt.Sprintf("%s%d", baseThemeName, variantToApply)

	// Apply the color scheme using plasma-apply-colorscheme with variant toggling
	cmd := exec.CommandContext(ctx, "plasma-apply-colorscheme", themeNameToApply)
	if err := cmd.Run(); err != nil {
		// Don't fail, just warn
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Warning: Failed to apply color scheme: %v\n", err)
			fmt.Fprintf(os.Stderr, "   To apply manually: plasma-apply-colorscheme %s\n", themeNameToApply)
		}
		return nil
	}

	fmt.Fprintf(os.Stderr, "   plasma-apply-colorscheme: Applied %s\n", themeNameToApply)

	// Trigger D-Bus reload to ensure KWin and Plasma Shell pick up the changes immediately
	if dbus.IsAvailable() {
		reloaded, err := kdedbus.ReloadTheme(ctx)
		if reloaded && p.verbose {
			fmt.Fprintf(os.Stderr, "   KDE Plasma theme reloaded via D-Bus (KWin + Plasma Shell)\n")
		} else if err != nil && p.verbose {
			fmt.Fprintf(os.Stderr, "   Note: D-Bus reload failed (theme still applied via plasma-apply-colorscheme): %v\n", err)
		}
	}

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
	cmd := exec.CommandContext(ctx, "plasma-apply-wallpaperimage", wallpaperPath) // #nosec G204 -- binary name is hardcoded; exec.CommandContext does not invoke a shell so wallpaperPath cannot cause injection
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute plasma-apply-wallpaperimage: %w", err)
	}

	fmt.Fprintf(os.Stderr, "   plasma-apply-wallpaperimage: Applied %s\n", wallpaperPath)
	return nil
}

// detectSystemColorScheme detects the current KDE system color scheme preference.
// Returns themeDark or themeLight, or empty string if detection fails.
// determineVariantToApply checks the currently applied color scheme and returns
// the variant to apply (1 or 2). This implements the alternating workaround for
// plasma-apply-colorscheme not reloading when the scheme name is unchanged.
//
// Logic:
// - If current scheme is TinctDark1 or TinctLight1, return 2
// - If current scheme is TinctDark2 or TinctLight2, return 1
// - Otherwise (first run or different scheme), return 1.
func (p *Plugin) determineVariantToApply(ctx context.Context, baseThemeName string) int {
	cmd := exec.CommandContext(ctx, "kreadconfig5", "--file", "kdeglobals", "--group", "General", "--key", "ColorScheme")
	cmdOutput, err := cmd.Output()
	if err != nil {
		return 1
	}

	currentScheme := strings.TrimSpace(string(cmdOutput))

	// Check if current scheme is one of our variants
	if currentScheme == fmt.Sprintf("%s1", baseThemeName) {
		return 2 // Switch to variant 2
	} else if currentScheme == fmt.Sprintf("%s2", baseThemeName) {
		return 1 // Switch to variant 1
	}

	// Not a Tinct variant, or first run - use variant 1
	return 1
}

func (p *Plugin) detectSystemColorScheme(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, "kreadconfig5", "--file", "kdeglobals", "--group", "General", "--key", "ColorScheme")
	cmdOutput, err := cmd.Output()
	if err == nil {
		currentScheme := strings.ToLower(strings.TrimSpace(string(cmdOutput)))
		if strings.Contains(currentScheme, "dark") {
			return themeDark
		} else if strings.Contains(currentScheme, "light") {
			return themeLight
		}
	}

	// Fallback: check if plasma is in dark mode via lookAndFeelPackage
	cmd = exec.CommandContext(ctx, "kreadconfig5", "--file", "kdeglobals", "--group", "KDE", "--key", "LookAndFeelPackage")
	cmdOutput, err = cmd.Output()
	if err == nil {
		lnf := strings.ToLower(strings.TrimSpace(string(cmdOutput)))
		if strings.Contains(lnf, "dark") {
			return themeDark
		} else if strings.Contains(lnf, "light") {
			return themeLight
		}
	}

	return ""
}
