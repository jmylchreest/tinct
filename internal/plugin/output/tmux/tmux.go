// Package tmux provides an output plugin for tmux colour themes.
package tmux

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
	"github.com/jmylchreest/tinct/internal/pluginconfig"
	"github.com/jmylchreest/tinct/internal/version"
	"github.com/jmylchreest/tinct/pkg/plugin/hooks"
	"github.com/jmylchreest/tinct/pkg/plugin/paths"
)

//go:embed *.tmpl
var templates embed.FS

// Plugin implements the output.Plugin interface for tmux.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new tmux output plugin.
func New() *Plugin { return &Plugin{} }

// Name returns the plugin name.
func (p *Plugin) Name() string { return "tmux" }

// Description returns the plugin description.
func (p *Plugin) Description() string { return "tmux colour theme (status bar, panes, modes)" }

// Version returns the plugin version.
func (p *Plugin) Version() string { return version.Version }

// RegisterFlags registers plugin-specific flags.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "tmux.output-dir", "", "Output directory (default: ~/.config/tmux)")
}

// SetVerbose enables or disables verbose logging.
func (p *Plugin) SetVerbose(verbose bool) { p.verbose = verbose }

// GetEmbeddedFS returns the embedded template filesystem.
func (p *Plugin) GetEmbeddedFS() any { return templates }

// GetFlagHelp returns help info for plugin flags.
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
		{Name: "tmux.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/tmux)"},
	}
}

// Validate checks plugin configuration.
func (p *Plugin) Validate() error { return nil }

// DefaultOutputDir returns the resolved output directory.
// Resolution: --tmux.output-dir → TINCT_PLUGIN_TMUX_OUTPUT_DIR →
// [plugin.tmux] output_dir → platform default.
func (p *Plugin) DefaultOutputDir() string {
	return pluginconfig.Resolve("tmux", "output_dir", p.outputDir,
		filepath.Join(paths.XDGConfigDir(), "tmux"))
}

// Hooks declares tmux's pre/post-execute behaviour. tmux doesn't watch
// config files, so we attempt an automatic reload via ReloadFn: if any
// tmux server is running (`tmux ls` succeeds with at least one session),
// we source the freshly generated tinct.conf into it. The verb-based
// ReloadSpec can't express "only run if a server is up", so we use the
// ReloadFn closure to keep the conditional silent in the no-server case
// (which is the common case on a clean shell). The runner already
// verbose-gates any error chatter; we don't add any normal-output
// diagnostic ourselves. InstructionsFn (which is itself verbose-only in
// the runner) still describes the manual user action.
func (p *Plugin) Hooks() hooks.Spec {
	return hooks.Spec{
		RequiredBinaries: []string{"tmux"},
		AutoCreateDir:    true,
		ReloadFn: func(ctx context.Context) error {
			outputDir := p.DefaultOutputDir()
			path := filepath.Join(outputDir, "tinct.conf")
			return reloadRunningSessions(ctx, path)
		},
		InstructionsFn: func(hctx hooks.Context) string {
			path := filepath.Join(hctx.OutputDir, "tinct.conf")
			return fmt.Sprintf("   Add 'source-file %s' to your tmux.conf, then reload with 'tmux source-file ~/.config/tmux/tmux.conf' or prefix-r.", path)
		},
	}
}

// reloadRunningSessions sources the generated config into any running
// tmux server. Returns nil (no error) when no server is running so the
// runner stays silent; surfaces real errors back to the runner, which
// only logs them in verbose mode.
func reloadRunningSessions(ctx context.Context, path string) error {
	out, err := exec.CommandContext(ctx, "tmux", "ls").Output()
	if err != nil {
		// `tmux ls` exits non-zero (and writes "no server running") when
		// there's no server — treat that as "nothing to reload", which
		// is the normal case on a fresh shell.
		return nil
	}
	if strings.TrimSpace(string(out)) == "" {
		return nil
	}
	return exec.CommandContext(ctx, "tmux", "source-file", path).Run()
}

// Generate creates the theme files. Two files are produced: a palette
// (tinct-palette.conf) that exposes every tinct role as a tmux user
// option (@tinct-<role>), and a main file (tinct.conf) that sources the
// palette and applies opinionated styling on top.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	palette, err := p.renderTemplate("tinct-palette.conf.tmpl", themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to render palette: %w", err)
	}

	main, err := p.renderTemplate("tinct.conf.tmpl", themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to render main config: %w", err)
	}

	return map[string][]byte{
		"tinct-palette.conf": palette,
		"tinct.conf":         main,
	}, nil
}

// renderTemplate loads and executes a single template (with custom
// override support) against the given theme data.
func (p *Plugin) renderTemplate(name string, themeData *colour.ThemeData) ([]byte, error) {
	loader := tmplloader.New("tmux", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, _, err := loader.Load(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read template %s: %w", name, err)
	}

	tmpl, err := template.New(name).Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", name, err)
	}
	return buf.Bytes(), nil
}
