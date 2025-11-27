// Package konsole provides an output plugin for Konsole terminal color themes.
package konsole

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
	"github.com/jmylchreest/tinct/internal/plugin/output/shared/utils"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/jmylchreest/tinct/pkg/dbus"
	"github.com/jmylchreest/tinct/pkg/util/appdetect"
)

//go:embed *.tmpl
var templates embed.FS

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for Konsole terminal.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new Konsole output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "konsole"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Konsole terminal color theme"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return "0.0.1"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "konsole.output-dir", "", "Output directory (default: ~/.local/share/konsole)")
}

// SetVerbose enables or disables verbose logging for the plugin.
// Implements the output.VerbosePlugin interface.
func (p *Plugin) SetVerbose(verbose bool) {
	p.verbose = verbose
}

// GetEmbeddedFS returns the embedded template filesystem.
// Implements the output.TemplateProvider interface.
func (p *Plugin) GetEmbeddedFS() any {
	return templates
}

// GetFlagHelp returns help information for all plugin flags.
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
		{Name: "konsole.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.local/share/konsole)", Required: false},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	return nil
}

// D-Bus helper functions for Konsole session management

const (
	// Konsole D-Bus service name pattern
	konsoleServicePrefix = "org.kde.konsole"

	// Konsole D-Bus interfaces
	konsoleSessionInterface = "org.kde.konsole.Session"
)

// session represents a Konsole session accessible via D-Bus.
type session struct {
	ServiceName string
	SessionPath string
}

// applyColorSchemeToAllSessions applies a color scheme to all active Konsole sessions via D-Bus.
func applyColorSchemeToAllSessions(ctx context.Context, colorScheme string) (int, error) {
	if !dbus.IsAvailable() {
		return 0, nil
	}

	sessions, err := listSessions(ctx)
	if err != nil {
		return 0, err
	}

	if len(sessions) == 0 {
		return 0, nil
	}

	applied := 0
	var lastErr error

	for _, sess := range sessions {
		if err := sess.setColorScheme(ctx, colorScheme); err != nil {
			lastErr = err
			continue
		}
		applied++
	}

	if applied == 0 && lastErr != nil {
		return 0, fmt.Errorf("failed to apply color scheme to any session: %w", lastErr)
	}

	return applied, nil
}

// listSessions finds all active Konsole sessions on the session bus.
func listSessions(ctx context.Context) ([]session, error) {
	conn, err := dbus.SessionBus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to session bus: %w", err)
	}
	defer conn.Close()

	// List all available names on the bus
	names, err := conn.ListNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list bus names: %w", err)
	}

	var sessions []session

	// Find all Konsole services (e.g., org.kde.konsole-12345)
	for _, name := range names {
		if !strings.HasPrefix(name, konsoleServicePrefix) {
			continue
		}

		// Get session paths for this Konsole instance
		sessionPaths, err := getSessionPaths(ctx, conn, name)
		if err != nil {
			// Skip this instance if we can't enumerate sessions
			continue
		}

		for _, path := range sessionPaths {
			sessions = append(sessions, session{
				ServiceName: name,
				SessionPath: path,
			})
		}
	}

	return sessions, nil
}

// getSessionPaths retrieves all session object paths for a Konsole service.
func getSessionPaths(ctx context.Context, conn *dbus.Connection, serviceName string) ([]string, error) {
	var paths []string

	// Konsole typically has sessions at /Sessions/1, /Sessions/2, etc.
	// We'll try to enumerate them by attempting to access known patterns
	for i := 1; i <= 100; i++ { // Reasonable limit
		sessionPath := fmt.Sprintf("/Sessions/%d", i)
		obj := conn.Object(serviceName, sessionPath)

		// Try to get a property to verify the session exists
		_, err := obj.GetProperty(ctx, konsoleSessionInterface+".ProcessId")
		if err != nil {
			// Session doesn't exist or we can't access it
			continue
		}

		paths = append(paths, sessionPath)
	}

	return paths, nil
}

// setColorScheme sets the color scheme for a Konsole session.
func (s *session) setColorScheme(ctx context.Context, colorScheme string) error {
	conn, err := dbus.SessionBus(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to session bus: %w", err)
	}
	defer conn.Close()

	obj := conn.Object(s.ServiceName, s.SessionPath)

	// Call setProfile method with color scheme
	err = obj.Call(ctx, konsoleSessionInterface+".setProfile", colorScheme)
	if err != nil {
		return fmt.Errorf("failed to set color scheme: %w", err)
	}

	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
func (p *Plugin) DefaultOutputDir() string {
	if p.outputDir != "" {
		return p.outputDir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ".local/share/konsole"
	}
	return filepath.Join(home, ".local", "share", "konsole")
}

// Generate creates the theme file.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	// Populate template metadata fields.
	themeData.OutputDir = p.DefaultOutputDir()
	themeData.ColorFileName = "Tinct.colorscheme"

	files := make(map[string][]byte)

	// Generate theme file.
	themeContent, err := p.generateTheme(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate theme: %w", err)
	}

	// Use theme-specific filename (TinctDark or TinctLight).
	var fileName string
	if themeData.ThemeType() == colour.ThemeDark {
		fileName = "TinctDark.colorscheme"
	} else {
		fileName = "TinctLight.colorscheme"
	}

	files[fileName] = themeContent

	return files, nil
}

// generateTheme creates the theme configuration file.
func (p *Plugin) generateTheme(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New("konsole", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct.colorscheme.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.colorscheme.tmpl\n")
	}

	tmpl, err := template.New("theme").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse theme template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute theme template: %w", err)
	}

	return buf.Bytes(), nil
}

// PreExecute checks if konsole is available before generating the theme.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	// Check if konsole executable exists (native, Flatpak, or AppImage).
	if !appdetect.IsPresentAny([]string{"konsole"}, nil) {
		return true, "konsole executable not found on $PATH", nil
	}

	// Check if config directory exists, create if it doesn't.
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// Try to create the config directory.
		if err := os.MkdirAll(configDir, 0o755); err != nil { // #nosec G301 - Config directory needs standard permissions
			return true, fmt.Sprintf("failed to create konsole config directory: %s", configDir), nil
		}
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Created konsole config directory: %s\n", configDir)
		}
	}

	return false, "", nil
}

// PostExecute provides usage instructions for applying the theme and attempts to reload.
// Implements the output.PostExecuteHook interface.
func (p *Plugin) PostExecute(ctx context.Context, execCtx output.ExecutionContext, generatedFiles []string) error {
	if len(generatedFiles) == 0 {
		return nil
	}

	// Extract scheme name from first generated file (e.g., "TinctDark.colorscheme" -> "TinctDark")
	fileName := filepath.Base(generatedFiles[0])
	schemeName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	// Strategy 1: Try D-Bus to update all Konsole sessions (Linux only)
	if dbus.IsAvailable() {
		applied, err := applyColorSchemeToAllSessions(ctx, schemeName)
		if err == nil && applied > 0 {
			if p.verbose {
				fmt.Fprintf(os.Stderr, "\n")
				fmt.Fprintf(os.Stderr, "   Konsole color scheme applied to %d session(s) via D-Bus: %s\n", applied, schemeName)
				fmt.Fprintf(os.Stderr, "\n")
			}
			return nil
		}
		// D-Bus failed, fall through to konsoleprofile
		if p.verbose && err != nil {
			fmt.Fprintf(os.Stderr, "   D-Bus update failed: %v\n", err)
		}
	}

	// Strategy 2: Try konsoleprofile (only works if running inside Konsole)
	cmd := exec.CommandContext(ctx, "konsoleprofile", fmt.Sprintf("colors=%s", schemeName))
	if err := cmd.Run(); err != nil {
		// konsoleprofile only works when running inside Konsole
		if p.verbose {
			fmt.Fprintf(os.Stderr, "\n")
			fmt.Fprintf(os.Stderr, "   Konsole color scheme generated: %s\n", schemeName)
			fmt.Fprintf(os.Stderr, "\n")
			fmt.Fprintf(os.Stderr, "   The theme will apply automatically to new Konsole windows.\n")
			fmt.Fprintf(os.Stderr, "   To apply to the current session, run: konsoleprofile colors=%s\n", schemeName)
			fmt.Fprintf(os.Stderr, "\n")
		}
	} else if p.verbose {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "   Konsole color scheme applied to current session: %s\n", schemeName)
		fmt.Fprintf(os.Stderr, "\n")
	}

	return nil
}
