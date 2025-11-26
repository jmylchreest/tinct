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
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/common"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/jmylchreest/tinct/internal/version"
	"github.com/jmylchreest/tinct/pkg/util/appdetect"
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
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
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
	return filepath.Join(home, ".local", "share", "themes")
}

// PreExecute checks if GNOME Shell and User Themes extension are installed.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	// Check if all required binaries are installed
	if !appdetect.IsPresentAll([]string{"gnome-shell", "gsettings", "gnome-extensions"}, nil) {
		if !appdetect.IsPresentAny([]string{"gnome-shell"}, nil) {
			return true, "gnome-shell is not installed", nil
		}
		if !appdetect.IsPresentAny([]string{"gsettings"}, nil) {
			return true, "gsettings command not found. Install gsettings (part of glib package).", nil
		}
		if !appdetect.IsPresentAny([]string{"gnome-extensions"}, nil) {
			return true, "gnome-extensions command not found. Install gnome-shell-extensions package.", nil
		}
	}

	// Check if User Themes extension exists (not necessarily enabled)
	home, err := os.UserHomeDir()
	if err != nil {
		return true, "cannot determine home directory", nil
	}

	userThemesPath := filepath.Join(home, ".local", "share", "gnome-shell", "extensions", userThemeExtensionID)
	systemThemesPath := filepath.Join("/usr/share/gnome-shell/extensions", userThemeExtensionID)

	// Check both user and system installation locations
	if !appdetect.IsPresentAny(nil, []string{userThemesPath, systemThemesPath}) {
		return true, "User Themes extension is not installed. Install with:\n  Fedora/RHEL: sudo dnf install gnome-shell-extension-user-theme\n  Ubuntu/Debian: sudo apt install gnome-shell-extension-user-theme\n  Arch: sudo pacman -S gnome-shell-extension-user-theme\n  Or via browser: https://extensions.gnome.org/extension/19/user-themes/", nil
	}

	// Check if User Themes extension is enabled
	cmd := exec.Command("gnome-extensions", "info", userThemeExtensionID)
	output, err := cmd.Output()
	if err == nil {
		// Parse the output to check if State: ENABLED
		outputStr := string(output)
		if !strings.Contains(outputStr, "State: ENABLED") && !strings.Contains(outputStr, "State: ACTIVE") {
			return true, fmt.Sprintf("User Themes extension is installed but not enabled. Enable with:\n  gnome-extensions enable %s\n  Then log out and back in", userThemeExtensionID), nil
		}
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

	// Generate both tinct-a and tinct-b themes to allow flicker-free toggling
	files["tinct-a/gnome-shell/gnome-shell.css"] = cssContent
	files["tinct-b/gnome-shell/gnome-shell.css"] = cssContent

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

	// Determine which theme to switch to (alternate between tinct-a and tinct-b)
	var targetTheme string
	themeChanged := false

	if currentTheme != "tinct-a" && currentTheme != "tinct-b" {
		// Not currently using tinct, start with tinct-a
		targetTheme = "tinct-a"
		themeChanged = true
		if currentTheme != "" && p.verbose {
			fmt.Fprintf(os.Stderr, "   Previous GNOME Shell theme: %s\n", currentTheme)
		}
	} else {
		// Already using tinct, toggle between a and b to reload CSS
		if currentTheme == "tinct-a" {
			targetTheme = "tinct-b"
		} else {
			targetTheme = "tinct-a"
		}
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Toggling theme from %s to %s to reload CSS...\n", currentTheme, targetTheme)
		}
	}

	// Set the theme
	cmd = exec.CommandContext(ctx, "gsettings", "set",
		"org.gnome.shell.extensions.user-theme", "name", targetTheme)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set GNOME Shell theme: %w", err)
	}

	if themeChanged {
		fmt.Fprintf(os.Stderr, "   gsettings: GNOME Shell theme applied: %s\n", targetTheme)
	} else {
		fmt.Fprintf(os.Stderr, "   gsettings: GNOME Shell theme reloaded: %s\n", targetTheme)
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
