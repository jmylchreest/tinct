// Package wezterm provides an output plugin for WezTerm colour schemes.
package wezterm

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/shared/utils"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/jmylchreest/tinct/internal/pluginconfig"
	"github.com/jmylchreest/tinct/internal/version"
	"github.com/jmylchreest/tinct/pkg/plugin/hooks"
	"github.com/jmylchreest/tinct/pkg/plugin/paths"
)

//go:embed *.tmpl
var templates embed.FS

// Plugin implements the output.Plugin interface for WezTerm.
type Plugin struct {
	outputDir string
	themeName string
	verbose   bool
}

// New creates a new WezTerm output plugin.
func New() *Plugin { return &Plugin{themeName: "tinct"} }

// Name returns the plugin name.
func (p *Plugin) Name() string { return "wezterm" }

// Description returns the plugin description.
func (p *Plugin) Description() string { return "WezTerm colour scheme (TOML)" }

// Version returns the plugin version.
func (p *Plugin) Version() string { return version.Version }

// RegisterFlags registers plugin-specific flags.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "wezterm.output-dir", "", "Output directory (default: ~/.config/wezterm/colors)")
	cmd.Flags().StringVar(&p.themeName, "wezterm.theme-name", "tinct", "Theme name (filename: <name>.toml)")
}

// SetVerbose enables or disables verbose logging.
func (p *Plugin) SetVerbose(verbose bool) { p.verbose = verbose }

// GetEmbeddedFS returns the embedded template filesystem.
func (p *Plugin) GetEmbeddedFS() any { return templates }

// GetFlagHelp returns help info for plugin flags.
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
		{Name: "wezterm.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/wezterm/colors)"},
		{Name: "wezterm.theme-name", Type: "string", Default: "tinct", Description: "Theme name (filename: <name>.toml)"},
	}
}

// Validate checks plugin configuration.
func (p *Plugin) Validate() error {
	if p.themeName == "" {
		return fmt.Errorf("theme name cannot be empty")
	}
	return nil
}

// DefaultOutputDir returns the resolved output directory.
// Resolution: --wezterm.output-dir → TINCT_PLUGIN_WEZTERM_OUTPUT_DIR →
// [plugin.wezterm] output_dir → platform default.
func (p *Plugin) DefaultOutputDir() string {
	return pluginconfig.Resolve("wezterm", "output_dir", p.outputDir, p.platformDefault())
}

func (p *Plugin) platformDefault() string {
	// WezTerm respects $XDG_CONFIG_HOME everywhere; macOS users typically
	// use ~/.config/wezterm too rather than the bundle path.
	if os.Getenv("XDG_CONFIG_HOME") == "" && runtime.GOOS == "windows" {
		if appData := os.Getenv("AppData"); appData != "" {
			return filepath.Join(appData, "wezterm", "colors")
		}
	}
	return filepath.Join(paths.XDGConfigDir(), "wezterm", "colors")
}

// Hooks declares WezTerm's pre/post-execute behaviour. WezTerm watches
// its config and reloads automatically — only the user-facing
// instructions step is needed.
func (p *Plugin) Hooks() hooks.Spec {
	themeName := p.themeName
	return hooks.Spec{
		// WezTerm isn't always on PATH on macOS (it lives in the .app bundle),
		// so the binary is optional rather than required.
		OptionalBinaries: []string{"wezterm"},
		AutoCreateDir:    true,
		InstructionsFn: func(_ hooks.Context) string {
			return fmt.Sprintf("   Set 'config.color_scheme = %q' in your wezterm.lua. WezTerm reloads automatically.", themeName)
		},
	}
}

// Generate creates the theme file.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	loader := tmplloader.New("wezterm", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, _, err := loader.Load("tinct.toml.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New("wezterm").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return map[string][]byte{p.themeName + ".toml": buf.Bytes()}, nil
}
