// Package manager provides plugin management with configuration support.
package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/executor"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/input/file"
	"github.com/jmylchreest/tinct/internal/plugin/input/googlegenai"
	"github.com/jmylchreest/tinct/internal/plugin/input/image"
	"github.com/jmylchreest/tinct/internal/plugin/input/markdown"
	"github.com/jmylchreest/tinct/internal/plugin/input/minimax"
	"github.com/jmylchreest/tinct/internal/plugin/input/openrouter"
	"github.com/jmylchreest/tinct/internal/plugin/input/remotecss"
	"github.com/jmylchreest/tinct/internal/plugin/input/remotejson"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/palettebuilder"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/alacritty"
	"github.com/jmylchreest/tinct/internal/plugin/output/awww"
	"github.com/jmylchreest/tinct/internal/plugin/output/btop"
	"github.com/jmylchreest/tinct/internal/plugin/output/dunst"
	"github.com/jmylchreest/tinct/internal/plugin/output/foot"
	"github.com/jmylchreest/tinct/internal/plugin/output/fuzzel"
	"github.com/jmylchreest/tinct/internal/plugin/output/ghostty"
	gnomeshell "github.com/jmylchreest/tinct/internal/plugin/output/gnome-shell"
	"github.com/jmylchreest/tinct/internal/plugin/output/gtk3"
	"github.com/jmylchreest/tinct/internal/plugin/output/gtk4"
	"github.com/jmylchreest/tinct/internal/plugin/output/helix"
	"github.com/jmylchreest/tinct/internal/plugin/output/histui"
	"github.com/jmylchreest/tinct/internal/plugin/output/hyprland"
	"github.com/jmylchreest/tinct/internal/plugin/output/hyprlock"
	"github.com/jmylchreest/tinct/internal/plugin/output/hyprpaper"
	kdeplasma "github.com/jmylchreest/tinct/internal/plugin/output/kde-plasma"
	"github.com/jmylchreest/tinct/internal/plugin/output/kitty"
	"github.com/jmylchreest/tinct/internal/plugin/output/konsole"
	"github.com/jmylchreest/tinct/internal/plugin/output/libadwaita"
	markdownout "github.com/jmylchreest/tinct/internal/plugin/output/markdown"
	"github.com/jmylchreest/tinct/internal/plugin/output/mc"
	"github.com/jmylchreest/tinct/internal/plugin/output/neovim"
	"github.com/jmylchreest/tinct/internal/plugin/output/niri"
	"github.com/jmylchreest/tinct/internal/plugin/output/qt5"
	"github.com/jmylchreest/tinct/internal/plugin/output/qt6"
	"github.com/jmylchreest/tinct/internal/plugin/output/rofi"
	"github.com/jmylchreest/tinct/internal/plugin/output/rosec"
	"github.com/jmylchreest/tinct/internal/plugin/output/swayosd"
	"github.com/jmylchreest/tinct/internal/plugin/output/tmux"
	"github.com/jmylchreest/tinct/internal/plugin/output/walker"
	"github.com/jmylchreest/tinct/internal/plugin/output/warp"
	"github.com/jmylchreest/tinct/internal/plugin/output/waybar"
	"github.com/jmylchreest/tinct/internal/plugin/output/wbg"
	"github.com/jmylchreest/tinct/internal/plugin/output/wezterm"
	"github.com/jmylchreest/tinct/internal/plugin/output/wofi"
	"github.com/jmylchreest/tinct/internal/plugin/output/yazi"
	"github.com/jmylchreest/tinct/internal/plugin/output/zellij"
	"github.com/jmylchreest/tinct/internal/plugin/protocol"
	"github.com/jmylchreest/tinct/pkg/plugin"
	"github.com/jmylchreest/tinct/pkg/plugin/hooks"
)

const (
	versionUnknown = "unknown"
)

// Builder provides a fluent interface for constructing a Manager.
type Builder struct {
	inputRegistry  *input.Registry
	outputRegistry *output.Registry
}

// NewBuilder creates a new Manager builder with default settings.
func NewBuilder() *Builder {
	return &Builder{
		inputRegistry:  input.NewRegistry(),
		outputRegistry: output.NewRegistry(),
	}
}

// WithCustomRegistries allows providing custom plugin registries (useful for testing).
func (b *Builder) WithCustomRegistries(inputReg *input.Registry, outputReg *output.Registry) *Builder {
	b.inputRegistry = inputReg
	b.outputRegistry = outputReg
	return b
}

// Build constructs the Manager with the configured settings.
func (b *Builder) Build() *Manager {
	m := &Manager{
		inputRegistry:  b.inputRegistry,
		outputRegistry: b.outputRegistry,
	}

	// Register built-in plugins.
	m.registerBuiltinPlugins()

	return m
}

// Manager owns plugin registries and provides plugin lookup.
type Manager struct {
	inputRegistry  *input.Registry
	outputRegistry *output.Registry
}

// registerBuiltinPlugins registers all built-in plugins.
func (m *Manager) registerBuiltinPlugins() {
	// Register input plugins.
	m.inputRegistry.Register(image.New())
	m.inputRegistry.Register(file.New())
	m.inputRegistry.Register(markdown.New())
	m.inputRegistry.Register(remotejson.New())
	m.inputRegistry.Register(remotecss.New())
	m.inputRegistry.Register(googlegenai.New())
	m.inputRegistry.Register(minimax.New())
	m.inputRegistry.Register(openrouter.New())

	// Register output plugins.
	m.outputRegistry.Register(alacritty.New())
	m.outputRegistry.Register(awww.New())
	m.outputRegistry.Register(btop.New())
	m.outputRegistry.Register(dunst.New())
	m.outputRegistry.Register(foot.New())
	m.outputRegistry.Register(fuzzel.New())
	m.outputRegistry.Register(ghostty.New())
	m.outputRegistry.Register(gnomeshell.New())
	m.outputRegistry.Register(gtk3.New())
	m.outputRegistry.Register(gtk4.New())
	m.outputRegistry.Register(helix.New())
	m.outputRegistry.Register(histui.New())
	m.outputRegistry.Register(hyprland.New())
	m.outputRegistry.Register(hyprlock.New())
	m.outputRegistry.Register(hyprpaper.New())
	m.outputRegistry.Register(kdeplasma.New())
	m.outputRegistry.Register(kitty.New())
	m.outputRegistry.Register(konsole.New())
	m.outputRegistry.Register(libadwaita.New())
	m.outputRegistry.Register(mc.New())
	m.outputRegistry.Register(markdownout.New())
	m.outputRegistry.Register(neovim.New())
	m.outputRegistry.Register(niri.New())
	m.outputRegistry.Register(rofi.New())
	m.outputRegistry.Register(rosec.New())
	m.outputRegistry.Register(qt5.New())
	m.outputRegistry.Register(qt6.New())
	m.outputRegistry.Register(swayosd.New())
	m.outputRegistry.Register(tmux.New())
	m.outputRegistry.Register(walker.New())
	m.outputRegistry.Register(warp.New())
	m.outputRegistry.Register(waybar.New())
	m.outputRegistry.Register(wbg.New())
	m.outputRegistry.Register(wezterm.New())
	m.outputRegistry.Register(wofi.New())
	m.outputRegistry.Register(yazi.New())
	m.outputRegistry.Register(zellij.New())
}

// InputRegistry returns the input plugin registry.
func (m *Manager) InputRegistry() *input.Registry {
	return m.inputRegistry
}

// OutputRegistry returns the output plugin registry.
func (m *Manager) OutputRegistry() *output.Registry {
	return m.outputRegistry
}

// GetInputPlugin retrieves an input plugin by name.
func (m *Manager) GetInputPlugin(name string) (input.Plugin, bool) {
	p, ok := m.inputRegistry.Get(name)
	return p, ok
}

// GetOutputPlugin retrieves an output plugin by name.
func (m *Manager) GetOutputPlugin(name string) (output.Plugin, bool) {
	p, ok := m.outputRegistry.Get(name)
	return p, ok
}

// AllInputPlugins returns all registered input plugins.
func (m *Manager) AllInputPlugins() map[string]input.Plugin {
	return m.inputRegistry.All()
}

// AllOutputPlugins returns all registered output plugins.
func (m *Manager) AllOutputPlugins() map[string]output.Plugin {
	return m.outputRegistry.All()
}

// CloseAllExternalPlugins releases any plugin subprocess still held by
// cached executors. ExternalInputPlugin keeps lastExecutor live so
// post-Generate getters (WallpaperPath, ThemeHint) can read from it;
// the CLI calls this from cobra.OnFinalize at exit.
func (m *Manager) CloseAllExternalPlugins() {
	for _, p := range m.inputRegistry.All() {
		if c, ok := p.(interface{ Close() }); ok {
			c.Close()
		}
	}
}

// RegisterExternalPlugin registers an external plugin with the manager.
func (m *Manager) RegisterExternalPlugin(name, pluginType, path, description string) error {
	// Validate plugin path - must be absolute and should exist.
	if !filepath.IsAbs(path) {
		return fmt.Errorf("plugin path must be absolute: %s", path)
	}

	// Check if the plugin file exists and is executable.
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("plugin not found or not accessible: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("plugin path is a directory, not a file: %s", path)
	}

	// Query plugin info to check protocol version.
	pluginInfo, err := queryPluginInfo(path)
	if err != nil {
		return fmt.Errorf("failed to query plugin info: %w", err)
	}

	// Check protocol version compatibility.
	if pluginInfo.ProtocolVersion != "" {
		compatible, err := protocol.IsCompatibleForType(pluginInfo.ProtocolVersion, pluginType)
		if err != nil || !compatible {
			errMsg := "unknown error"
			if err != nil {
				errMsg = err.Error()
			}
			return &IncompatiblePluginError{
				Name:          name,
				PluginVersion: pluginInfo.ProtocolVersion,
				HostVersion:   protocol.ProtocolVersion,
				MinVersion:    protocol.MinCompatibleVersionForType(pluginType),
				Reason:        errMsg,
			}
		}
	}
	// Note: If protocol_version is missing, we allow the plugin (backward compatibility)
	// but this should be warned about in verbose mode

	switch pluginType {
	case "output":
		p := NewExternalOutputPlugin(name, description, path)
		m.outputRegistry.Register(p)
		return nil
	case "input":
		p := NewExternalInputPlugin(name, description, path)
		m.inputRegistry.Register(p)
		return nil
	default:
		return fmt.Errorf("unknown plugin type: %s", pluginType)
	}
}

// PluginInfo holds metadata returned by a plugin's --plugin-info command.
type PluginInfo struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	Version         string `json:"version"`
	ProtocolVersion string `json:"protocol_version"`
	Description     string `json:"description"`
}

// queryPluginInfo queries a plugin for its metadata.
func queryPluginInfo(pluginPath string) (PluginInfo, error) {
	cmd := exec.Command(pluginPath, "--plugin-info")
	cmdOutput, err := cmd.Output()
	if err != nil {
		return PluginInfo{}, fmt.Errorf("failed to execute plugin: %w", err)
	}

	var info PluginInfo
	if err := json.Unmarshal(cmdOutput, &info); err != nil {
		return PluginInfo{}, fmt.Errorf("failed to parse plugin info: %w", err)
	}

	return info, nil
}

// IncompatiblePluginError reports a plugin rejected for speaking a
// protocol version this host no longer supports.
//
// It is a distinct type because the condition is user-fixable and must
// not be swallowed: the plugin vanishes from the registry, so without a
// message the only symptom is "unknown plugin", which points the user at
// the wrong problem entirely.
type IncompatiblePluginError struct {
	Name          string
	PluginVersion string
	HostVersion   string
	MinVersion    string
	Reason        string
}

func (e *IncompatiblePluginError) Error() string {
	return fmt.Sprintf(
		"plugin %q speaks protocol %s, but tinct %s requires at least %s: %s",
		e.Name, e.PluginVersion, e.HostVersion, e.MinVersion, e.Reason,
	)
}

// externalPluginBase holds fields and methods shared by all external plugin wrappers.
//
// session caches the plugin subprocess for the lifetime of one run. Every
// RPC used to spawn its own executor and Kill() it immediately, which
// meant a single output plugin was launched five times per generate
// (Validate, GetHooks, PreExecute, Generate, PostExecute) — five
// handshakes, five yamux sessions, and five chances to race the
// teardown (the stray "yamux: Failed to write header" lines). Worse,
// each spawn got a fresh plugin object, so anything a plugin learned in
// one call (its args, for instance) was gone by the next.
//
// Reusing one process fixes both: the plugin keeps its state across the
// call sequence, and Configure can push args in once, before the
// pre-execute checks that need them.
type externalPluginBase struct {
	name        string
	description string
	path        string
	args        map[string]any
	dryRun      bool
	// verbose must be set before the first session() call: it selects
	// the go-plugin logger, which is fixed for the life of the
	// subprocess. Both input and output wrappers expose SetVerbose so
	// the CLI can set it before validation starts.
	verbose bool

	mu          sync.Mutex
	execSession *executor.PluginExecutor
	// configured guards the one-shot Configure RPC so re-entering
	// session() mid-run does not re-push args.
	configured bool
}

// session returns the cached executor, starting the plugin subprocess on
// first use. kind is "input" or "output" and selects which Configure RPC
// is attempted.
//
// Configure is optional: plugins built against older SDKs (or that do
// not implement Configurable) report "can't find method" and the host
// carries on, exactly as with the other optional 0.3.0+ methods. Its
// failure is never fatal — a plugin that cannot be configured still
// generates, it just does not see its args early.
func (b *externalPluginBase) session(kind string) (*executor.PluginExecutor, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.execSession != nil {
		return b.execSession, nil
	}

	pluginExec, err := executor.NewWithVerbose(b.path, b.verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin executor: %w", err)
	}
	b.execSession = pluginExec

	if !b.configured {
		b.configured = true
		pluginExec.Configure(context.Background(), kind, plugin.ConfigureRequest{
			Args:    b.args,
			DryRun:  b.dryRun,
			Verbose: b.verbose,
		})
	}

	return b.execSession, nil
}

// Close shuts down the cached plugin subprocess. Safe to call more than
// once and on a plugin that never started one. The host calls this when
// a run finishes; the next session() starts a fresh process.
func (b *externalPluginBase) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.execSession != nil {
		b.execSession.Close()
		b.execSession = nil
	}
	b.configured = false
}

// Name returns the plugin's name.
func (b *externalPluginBase) Name() string { return b.name }

// Description returns the plugin's description.
func (b *externalPluginBase) Description() string { return b.description }

// Version returns the plugin's version by querying the plugin executable.
func (b *externalPluginBase) Version() string {
	info, err := queryPluginInfo(b.path)
	if err != nil {
		return versionUnknown
	}
	if info.Version == "" {
		return versionUnknown
	}
	return info.Version
}

// SetArgs sets custom arguments for this plugin.
func (b *externalPluginBase) SetArgs(args map[string]any) { b.args = args }

// GetArgs returns custom arguments for this plugin.
func (b *externalPluginBase) GetArgs() map[string]any { return b.args }

// SetVerbose sets the verbose flag. It only affects sessions started
// afterwards — a running plugin subprocess keeps the logger it was
// created with — so the host sets it before the first RPC.
func (b *externalPluginBase) SetVerbose(verbose bool) { b.verbose = verbose }

// GetVerbose returns the verbose flag.
func (b *externalPluginBase) GetVerbose() bool { return b.verbose }

// SetDryRun sets the dry-run mode for this plugin.
func (b *externalPluginBase) SetDryRun(dryRun bool) { b.dryRun = dryRun }

// GetDryRun returns the dry-run mode for this plugin.
func (b *externalPluginBase) GetDryRun() bool { return b.dryRun }

// getFlagHelp queries the plugin executable for flag help via RPC.
func (b *externalPluginBase) getFlagHelp() []input.FlagHelp {
	pluginExec, err := executor.NewWithVerbose(b.path, false)
	if err != nil {
		return []input.FlagHelp{}
	}
	defer pluginExec.Close()

	flagHelp, err := pluginExec.GetFlagHelp(context.Background())
	if err != nil {
		return []input.FlagHelp{}
	}

	return flagHelp
}

// ExternalInputPlugin wraps an external executable as an input plugin.
type ExternalInputPlugin struct {
	externalPluginBase
	lastExecutor *executor.PluginExecutor // Store last executor to query wallpaper path
}

// NewExternalInputPlugin creates a new external input plugin wrapper.
func NewExternalInputPlugin(name, description, path string) *ExternalInputPlugin {
	return &ExternalInputPlugin{
		externalPluginBase: externalPluginBase{
			name:        name,
			description: description,
			path:        path,
		},
	}
}

// Generate executes the external plugin and returns a palette.
// Uses the hybrid executor which automatically detects and uses the appropriate
// protocol (go-plugin RPC or JSON-stdio).
func (p *ExternalInputPlugin) Generate(ctx context.Context, opts input.GenerateOptions) (*colour.Palette, error) {
	pluginExec, err := p.session("input")
	if err != nil {
		return nil, err
	}

	// Kept for the wallpaper-path accessors below; the session owns the
	// process lifetime, so this must not be closed here.
	p.lastExecutor = pluginExec

	// Merge plugin args from opts with plugin's own args.
	mergedArgs := make(map[string]any)
	maps.Copy(mergedArgs, p.args)
	maps.Copy(mergedArgs, opts.PluginArgs)

	// Convert to protocol format.
	protocolOpts := plugin.InputOptions{
		Verbose:         opts.Verbose,
		DryRun:          opts.DryRun || p.dryRun,
		Backend:         opts.Backend,
		ColourOverrides: opts.ColourOverrides,
		PluginArgs:      mergedArgs,
	}

	// Debug: show what's being sent to plugin.
	if opts.Verbose {
		if optsJSON, err := json.Marshal(protocolOpts); err == nil {
			fmt.Fprintf(os.Stderr, "   Sending to plugin: %s\n", string(optsJSON))
		}
	}

	// Execute input plugin.
	colors, err := pluginExec.ExecuteInput(ctx, protocolOpts)
	if err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w", err)
	}

	// Map any role-name → index hints supplied by the plugin onto colour.Role
	// values, dropping unknown role names with a verbose warning.
	rawHints := pluginExec.GetLastRoleHints()
	if len(rawHints) == 0 {
		return colour.NewPalette(colors), nil
	}

	roleHints := make(map[colour.Role]int, len(rawHints))
	for name, idx := range rawHints {
		role, perr := palettebuilder.ParseColourRole(name)
		if perr != nil {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "   Skipping unknown role hint %q from plugin: %v\n", name, perr)
			}
			continue
		}
		if idx < 0 || idx >= len(colors) {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "   Skipping role hint %q: index %d out of range (palette has %d colours)\n", name, idx, len(colors))
			}
			continue
		}
		roleHints[role] = idx
	}

	if len(roleHints) == 0 {
		return colour.NewPalette(colors), nil
	}

	colourColors := make([]color.Color, len(colors))
	copy(colourColors, colors)
	return colour.NewPaletteWithRoleHints(colourColors, roleHints), nil
}

// ThemeHint returns the theme hint reported by the most recent Generate call
// (protocol >= 0.3.0). Implements the input.ThemeHinter interface for
// external plugins; returns "" when no hint was provided.
func (p *ExternalInputPlugin) ThemeHint() string {
	if p.lastExecutor == nil {
		return ""
	}
	return p.lastExecutor.GetLastThemeHint()
}

// RegisterFlags is a no-op for external plugins (they don't have flags).
func (p *ExternalInputPlugin) RegisterFlags(_ *cobra.Command) {
	// External plugins don't register flags in Tinct.
	// They handle their own arguments if needed.
}

// Validate fail-fast checks the plugin's own configuration via the
// optional Validator RPC. The args set by SetArgs (from --plugin-args
// and the dynamic --<plugin>.<arg> flags) are forwarded so the plugin
// can validate keys it cares about; the map is empty when nothing was
// configured for this plugin. Plugins that don't
// implement Validator (or run on JSON stdio) report success and
// surface their errors at Generate time as before.
func (p *ExternalInputPlugin) Validate() error {
	pluginExec, err := p.session("input")
	if err != nil {
		return err
	}
	return pluginExec.Validate(context.Background(), "input", p.args)
}

// GetFlagHelp returns help information for plugin flags.
// For external plugins, this queries the plugin executable via RPC.
func (p *ExternalInputPlugin) GetFlagHelp() []input.FlagHelp {
	return p.getFlagHelp()
}

// WallpaperPath returns the canonical wallpaper path from the last plugin execution.
// Implements the input.WallpaperProvider interface for external plugins.
func (p *ExternalInputPlugin) WallpaperPath() string {
	if p.lastExecutor == nil {
		return ""
	}
	return p.lastExecutor.GetWallpaperPath()
}

// WallpaperRawPath returns the raw wallpaper path from the last plugin execution.
// This is the literal path as provided by the user before any canonicalization.
// Implements the input.WallpaperProvider interface for external plugins.
func (p *ExternalInputPlugin) WallpaperRawPath() string {
	if p.lastExecutor == nil {
		return ""
	}
	return p.lastExecutor.GetWallpaperRawPath()
}

// Close releases the cached executor (and its plugin subprocess) held
// for post-Generate getters. Idempotent.
func (p *ExternalInputPlugin) Close() {
	if p.lastExecutor != nil {
		p.lastExecutor.Close()
		p.lastExecutor = nil
	}
}

// ExternalOutputPlugin wraps an external executable as an output plugin.
type ExternalOutputPlugin struct {
	externalPluginBase
	alternatePalette *colour.CategorisedPalette
}

// NewExternalOutputPlugin creates a new external output plugin wrapper.
func NewExternalOutputPlugin(name, description, path string) *ExternalOutputPlugin {
	return &ExternalOutputPlugin{
		externalPluginBase: externalPluginBase{
			name:        name,
			description: description,
			path:        path,
		},
	}
}

// SetAlternatePalette sets the alternate palette for dual-theme generation.
func (p *ExternalOutputPlugin) SetAlternatePalette(palette *colour.CategorisedPalette) {
	p.alternatePalette = palette
}

// Generate executes the external plugin and returns its output.
func (p *ExternalOutputPlugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	pluginExec, err := p.session("output")
	if err != nil {
		return nil, err
	}

	// Extract palette from themeData.
	palette := themeData.Palette()

	// Convert to protocol format, including alternate palette if set.
	paletteData := convertCategorisedPaletteToProtocolWithAlternate(palette, p.alternatePalette, p.args, p.dryRun, p.verbose)

	// Execute output plugin.
	files, err := pluginExec.ExecuteOutput(context.Background(), paletteData)
	if err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w", err)
	}

	// If no files generated, return empty map.
	if len(files) == 0 {
		return make(map[string][]byte), nil
	}

	return files, nil
}

// RegisterFlags is a no-op for external plugins (they don't have flags).
func (p *ExternalOutputPlugin) RegisterFlags(_ *cobra.Command) {
	// External plugins don't register flags in Tinct.
	// They handle their own arguments if needed.
}

// Validate fail-fast checks the plugin's own configuration via the
// optional Validator RPC. The args set by SetArgs (from --plugin-args
// and the dynamic --<plugin>.<arg> flags) are forwarded so the plugin
// can validate keys it cares about; the map is empty when nothing was
// configured for this plugin. Plugins that don't
// implement Validator (or run on JSON stdio) report success and
// surface their errors at Generate time as before.
func (p *ExternalOutputPlugin) Validate() error {
	pluginExec, err := p.session("output")
	if err != nil {
		return err
	}
	return pluginExec.Validate(context.Background(), "output", p.args)
}

// Hooks fetches the plugin's static hooks.Spec via the optional
// HooksProvider RPC. Plugins that don't implement HooksProvider return
// an empty Spec, which the runner treats as a no-op. Implements the
// hooks.Provider interface so the existing CLI hook runner picks up
// external plugin specs without any extra wiring.
func (p *ExternalOutputPlugin) Hooks() hooks.Spec {
	pluginExec, err := p.session("output")
	if err != nil {
		return hooks.Spec{}
	}
	spec, _ := pluginExec.GetHooks(context.Background())
	return spec
}

// Templates fetches the plugin's bundled templates via the optional
// TemplateLister RPC. Plugins that don't implement it return nil so
// `tinct plugins templates list/dump` can silently skip them.
//
// Each call spawns a fresh executor — this is invoked from the
// templates CLI which only ever runs once per session, so the cost
// of process startup is acceptable in exchange for not having to
// thread template state through the long-lived plugin manager.
func (p *ExternalOutputPlugin) Templates() map[string][]byte {
	pluginExec, err := executor.NewWithVerbose(p.path, p.verbose)
	if err != nil {
		return nil
	}
	defer pluginExec.Close()
	tmpls, _ := pluginExec.GetTemplates(context.Background())
	return tmpls
}

// DefaultOutputDir returns the default output directory (not used for external plugins).
func (p *ExternalOutputPlugin) DefaultOutputDir() string {
	return "" // External plugins handle their own output
}

// PreExecute calls the external plugin's pre-execute hook.
// Implements the output.PreExecuteHook interface.
func (p *ExternalOutputPlugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
	pluginExec, err := p.session("output")
	if err != nil {
		return false, "", err
	}

	// Execute pre-execute hook.
	return pluginExec.PreExecute(ctx)
}

// GetFlagHelp returns help information for plugin flags.
// For external plugins, this queries the plugin executable via RPC.
func (p *ExternalOutputPlugin) GetFlagHelp() []input.FlagHelp {
	return p.getFlagHelp()
}

// PostExecute calls the external plugin's post-execute hook.
// Implements the output.PostExecuteHook interface.
//
// The wire-level PostExecute RPC takes only (ctx, writtenFiles) — the
// execution context isn't currently propagated to external plugins
// because the protocol predates ExecutionContext. The execCtx
// parameter is accepted here so this method satisfies the 3-arg
// output.PostExecuteHook interface (the hook check in
// internal/cli/generate_helpers.go performs a type assertion against
// it). When the protocol grows ExecutionContext fields, fold them
// into pluginExec.PostExecute and the parameter becomes load-bearing
// without changing this signature.
func (p *ExternalOutputPlugin) PostExecute(ctx context.Context, _ output.ExecutionContext, writtenFiles []string) error {
	pluginExec, err := p.session("output")
	if err != nil {
		return err
	}

	// Execute post-execute hook.
	return pluginExec.PostExecute(ctx, writtenFiles)
}

// toCategorisedColour converts a single colour.CategorisedColour to the protocol representation.
func toCategorisedColour(c colour.CategorisedColour) plugin.CategorisedColour {
	return plugin.CategorisedColour{
		RGB: plugin.RGBColour{
			R: c.RGB.R,
			G: c.RGB.G,
			B: c.RGB.B,
		},
		Hex:        c.Hex,
		Role:       string(c.Role),
		Luminance:  c.Luminance,
		IsLight:    c.IsLight,
		Hue:        c.Hue,
		Saturation: c.Saturation,
		Index:      c.Index,
	}
}

// convertCategorisedPaletteToMaps converts the colour maps and slices of a CategorisedPalette
// to the protocol format, returning (coloursMap, allColoursSlice, themeType).
func convertCategorisedPaletteToMaps(p *colour.CategorisedPalette) (colours map[string]plugin.CategorisedColour, allColours []plugin.CategorisedColour, themeType string) {
	colours = make(map[string]plugin.CategorisedColour, len(p.Colours))
	for role, c := range p.Colours {
		colours[string(role)] = toCategorisedColour(c)
	}

	allColours = make([]plugin.CategorisedColour, len(p.AllColours))
	for i, c := range p.AllColours {
		allColours[i] = toCategorisedColour(c)
	}

	themeType = "dark"
	if p.ThemeType == colour.ThemeLight {
		themeType = "light"
	}

	return colours, allColours, themeType
}

// convertCategorisedPaletteToProtocol converts a CategorisedPalette to plugin.PaletteData.
func convertCategorisedPaletteToProtocol(palette *colour.CategorisedPalette, pluginArgs map[string]any, dryRun, verbose bool) plugin.PaletteData {
	return convertCategorisedPaletteToProtocolWithAlternate(palette, nil, pluginArgs, dryRun, verbose)
}

// convertCategorisedPaletteToProtocolWithAlternate converts a CategorisedPalette to plugin.PaletteData with optional alternate theme.
func convertCategorisedPaletteToProtocolWithAlternate(palette, alternatePalette *colour.CategorisedPalette, pluginArgs map[string]any, dryRun, verbose bool) plugin.PaletteData {
	colours, allColours, themeType := convertCategorisedPaletteToMaps(palette)

	paletteData := plugin.PaletteData{
		Colours:    colours,
		AllColours: allColours,
		ThemeType:  themeType,
		PluginArgs: pluginArgs,
		DryRun:     dryRun,
		Verbose:    verbose,
	}

	if alternatePalette != nil {
		altColours, altAllColours, altThemeType := convertCategorisedPaletteToMaps(alternatePalette)
		paletteData.AlternateTheme = &plugin.AlternateThemeData{
			Colours:    altColours,
			AllColours: altAllColours,
			ThemeType:  altThemeType,
		}
	}

	return paletteData
}
