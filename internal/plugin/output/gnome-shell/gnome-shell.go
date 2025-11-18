// Package gnomeshell provides an output plugin for GNOME Shell theming.
package gnomeshell

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

	ps "github.com/mitchellh/go-ps"
	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/common"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/jmylchreest/tinct/internal/version"
)

//go:embed *.tmpl
var templates embed.FS

const userThemeExtensionID = "user-theme@gnome-shell-extensions.gcampax.github.com"

// Plugin implements the output.Plugin interface for GNOME Shell themes.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new GNOME Shell plugin instance.
func New() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "gnome-shell"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "GNOME Shell theme (panel, overview, system menus)"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers command-line flags for this plugin.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "gnome-shell.output-dir", "",
		"Output directory for GNOME Shell theme (default: ~/.local/share/themes/tinct/gnome-shell)")
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
func (p *Plugin) GetFlagHelp() []input.FlagHelp {
	return []input.FlagHelp{
		{
			Name:        "gnome-shell.output-dir",
			Type:        "string",
			Default:     "~/.local/share/themes/tinct/gnome-shell",
			Description: "Output directory for GNOME Shell theme",
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
		return ".local/share/themes/tinct/gnome-shell"
	}
	return filepath.Join(home, ".local", "share", "themes", "tinct", "gnome-shell")
}

// PreExecute checks if GNOME Shell and User Themes extension are installed.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	// Check if gnome-shell is installed (not running, just installed)
	_, err = exec.LookPath("gnome-shell")
	if err != nil {
		return true, "gnome-shell is not installed", nil
	}

	// Check if gsettings exists (required for applying theme/wallpaper)
	_, err = exec.LookPath("gsettings")
	if err != nil {
		return true, "gsettings command not found. Install gsettings (part of glib package).", nil
	}

	// Check if gnome-extensions command exists
	_, err = exec.LookPath("gnome-extensions")
	if err != nil {
		return true, "gnome-extensions command not found. Install gnome-shell-extensions package.", nil
	}

	// Check if User Themes extension exists (not necessarily enabled)
	home, err := os.UserHomeDir()
	if err != nil {
		return true, "Could not determine home directory", nil
	}
	userThemesPath := filepath.Join(home, ".local", "share", "gnome-shell", "extensions", userThemeExtensionID)
	systemThemesPath := filepath.Join("/usr/share/gnome-shell/extensions", userThemeExtensionID)

	userExists := false
	systemExists := false

	if _, err := os.Stat(userThemesPath); err == nil {
		userExists = true
	}
	if _, err := os.Stat(systemThemesPath); err == nil {
		systemExists = true
	}

	if !userExists && !systemExists {
		return true, fmt.Sprintf("User Themes extension is not installed. Install with: gnome-extensions install %s", userThemeExtensionID), nil
	}

	return false, "", nil
}

// Generate creates the GNOME Shell theme files.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate the GNOME Shell CSS file
	cssContent, err := p.generateCSS(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate GNOME Shell CSS: %w", err)
	}
	files["gnome-shell.css"] = cssContent

	return files, nil
}

// generateCSS generates the GNOME Shell CSS content.
func (p *Plugin) generateCSS(themeData *colour.ThemeData) ([]byte, error) {
	loader := tmplloader.New("gnome-shell", templates)
	if p.verbose {
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}

	tmplContent, _, err := loader.Load("gnome-shell.css.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	tmpl, err := template.New("gnome-shell").
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

// PostExecute applies theme settings and sets wallpaper automatically.
func (p *Plugin) PostExecute(ctx context.Context, execCtx output.ExecutionContext, _ []string) error {
	if execCtx.DryRun {
		return nil
	}

	// Optional: Check if gnome-shell is actually running
	gnomeShellRunning := false
	processes, err := ps.Processes()
	if err == nil {
		for _, proc := range processes {
			if proc.Executable() == "gnome-shell" {
				gnomeShellRunning = true
				break
			}
		}
	}

	if !gnomeShellRunning && p.verbose {
		fmt.Fprintf(os.Stderr, "   Warning: gnome-shell is not currently running\n")
	}

	// Get current theme setting
	cmd := exec.CommandContext(ctx, "gsettings", "get", "org.gnome.shell.extensions.user-theme", "name")
	output, err := cmd.Output()
	currentTheme := ""
	if err == nil {
		currentTheme = strings.Trim(strings.TrimSpace(string(output)), "'\"")
	}

	// Only apply the theme if it's not already set to 'tinct'
	if currentTheme != "tinct" {
		if currentTheme != "" && p.verbose {
			fmt.Fprintf(os.Stderr, "   Previous GNOME Shell theme: %s\n", currentTheme)
		}

		cmd = exec.CommandContext(ctx, "gsettings", "set",
			"org.gnome.shell.extensions.user-theme", "name", "tinct")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set GNOME Shell theme: %w", err)
		}

		// Always show when we actually change the theme
		fmt.Fprintf(os.Stderr, "   gsettings: GNOME Shell theme applied: tinct\n")
	} else if p.verbose {
		fmt.Fprintf(os.Stderr, "   GNOME Shell theme already set to: tinct\n")
	}

	// Set wallpaper if available
	if execCtx.WallpaperPath != "" {
		wallpaperURI := "file://" + execCtx.WallpaperPath

		// Get current wallpaper (check both light and dark)
		cmd = exec.CommandContext(ctx, "gsettings", "get", "org.gnome.desktop.background", "picture-uri")
		currentLight, err := cmd.Output()
		currentLightURI := ""
		if err == nil && len(currentLight) > 0 {
			currentLightURI = strings.Trim(strings.TrimSpace(string(currentLight)), "'\"")
		}

		cmd = exec.CommandContext(ctx, "gsettings", "get", "org.gnome.desktop.background", "picture-uri-dark")
		currentDark, err := cmd.Output()
		currentDarkURI := ""
		if err == nil && len(currentDark) > 0 {
			currentDarkURI = strings.Trim(strings.TrimSpace(string(currentDark)), "'\"")
		}

		// Only set wallpaper if it's different from current
		if currentLightURI != wallpaperURI || currentDarkURI != wallpaperURI {
			if currentLightURI != "" && currentLightURI != wallpaperURI && p.verbose {
				fmt.Fprintf(os.Stderr, "   Previous wallpaper: %s\n", currentLightURI)
			}

			// Set for light mode
			cmd = exec.CommandContext(ctx, "gsettings", "set",
				"org.gnome.desktop.background", "picture-uri", wallpaperURI)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to set wallpaper: %w", err)
			}

			// Set for dark mode
			cmd = exec.CommandContext(ctx, "gsettings", "set",
				"org.gnome.desktop.background", "picture-uri-dark", wallpaperURI)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to set dark wallpaper: %w", err)
			}

			// Set picture options to zoom
			cmd = exec.CommandContext(ctx, "gsettings", "set",
				"org.gnome.desktop.background", "picture-options", "zoom")
			_ = cmd.Run() //nolint:errcheck // Best effort, non-critical setting

			// Always show when we actually change the wallpaper
			fmt.Fprintf(os.Stderr, "   gsettings: wallpaper set: %s\n", execCtx.WallpaperPath)
		} else if p.verbose {
			fmt.Fprintf(os.Stderr, "   Wallpaper already set to: %s\n", execCtx.WallpaperPath)
		}
	}

	return nil
}
