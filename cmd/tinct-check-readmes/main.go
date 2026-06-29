// Command tinct-check-readmes is a warn-only doc drift detector.
//
// It parses the YAML frontmatter from each plugin's README.md and
// compares the declared `plugin:` block against the plugin's runtime
// behaviour (for built-in plugins: by importing the package; for
// external plugins: by exec'ing --plugin-info).
//
// All findings are written to stderr. The tool always exits 0 — it is
// intentionally non-blocking so a documentation drift never gates a
// merge. Wire it into CI as an informational job.
//
// Usage:
//
//	tinct-check-readmes builtin [--root internal/plugin/output]
//	tinct-check-readmes external --binary path/to/tinct-plugin-foo [--readme path/to/README.md]
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/pkg/plugin"
	"github.com/jmylchreest/tinct/pkg/plugin/hooks"

	// Built-in input plugins.
	"github.com/jmylchreest/tinct/internal/plugin/input/file"
	"github.com/jmylchreest/tinct/internal/plugin/input/googlegenai"
	"github.com/jmylchreest/tinct/internal/plugin/input/image"
	markdownin "github.com/jmylchreest/tinct/internal/plugin/input/markdown"
	"github.com/jmylchreest/tinct/internal/plugin/input/openrouter"
	"github.com/jmylchreest/tinct/internal/plugin/input/remotecss"
	"github.com/jmylchreest/tinct/internal/plugin/input/remotejson"

	// Built-in output plugin registry. Keep this list in sync with
	// internal/plugin/manager/manager.go — the check tool only verifies
	// plugins it can construct.
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
)

// Source values for the frontmatter `plugin.source` field. Also used as
// the subcommand names so a user can `tinct-check-readmes builtin` /
// `... external`.
const (
	sourceBuiltin  = "builtin"
	sourceExternal = "external"
)

// builtinRegistry returns one instance of every built-in output plugin
// the tool knows how to check. Keep aligned with the manager's registry.
func builtinRegistry() []output.Plugin {
	return []output.Plugin{
		alacritty.New(), awww.New(), btop.New(), dunst.New(),
		foot.New(), fuzzel.New(), ghostty.New(), gnomeshell.New(),
		gtk3.New(), gtk4.New(), helix.New(), histui.New(),
		hyprland.New(), hyprlock.New(),
		hyprpaper.New(), kdeplasma.New(), kitty.New(), konsole.New(),
		libadwaita.New(), markdownout.New(), mc.New(), neovim.New(),
		niri.New(), qt5.New(), qt6.New(), rofi.New(), rosec.New(),
		swayosd.New(), tmux.New(), walker.New(), warp.New(),
		waybar.New(), wbg.New(), wezterm.New(), wofi.New(),
		yazi.New(), zellij.New(),
	}
}

// builtinInputPlugin pairs an input plugin instance with its source
// subdirectory. Input plugin directory names (e.g. "remotejson") don't
// always match Plugin.Name() (e.g. "remote-json"), so we record the dir
// explicitly rather than deriving it.
type builtinInputPlugin struct {
	plugin input.Plugin
	dir    string
}

// inputBuiltinRegistry returns one instance of every built-in input
// plugin paired with its source directory.
func inputBuiltinRegistry() []builtinInputPlugin {
	return []builtinInputPlugin{
		{image.New(), "image"},
		{file.New(), "file"},
		{markdownin.New(), "markdown"},
		{remotejson.New(), "remotejson"},
		{remotecss.New(), "remotecss"},
		{googlegenai.New(), "googlegenai"},
		{openrouter.New(), "openrouter"},
	}
}

// readmeFrontmatter is the parsed YAML at the top of each plugin
// README. Only the plugin: block is checked; other fields (title,
// sidebar_position) are passed through untouched.
type readmeFrontmatter struct {
	Plugin readmePlugin `yaml:"plugin"`
}

type readmePlugin struct {
	Type             string        `yaml:"type"` // input | output
	Name             string        `yaml:"name"`
	Category         string        `yaml:"category"`
	Source           string        `yaml:"source"`
	App              string        `yaml:"app"`
	AppURL           string        `yaml:"app_url"`
	Requires         []string      `yaml:"requires"`
	Optional         []string      `yaml:"optional"`
	Pattern          string        `yaml:"pattern"`
	DefaultOutputDir string        `yaml:"default_output_dir"`
	GeneratedFiles   []string      `yaml:"generated_files"`
	Reload           *readmeReload `yaml:"reload"`

	// input-only.
	SourceType        string   `yaml:"source_type"`
	Description       string   `yaml:"description"`
	Service           string   `yaml:"service"`
	ServiceURL        string   `yaml:"service_url"`
	RequiresNetwork   bool     `yaml:"requires_network"`
	RequiresCreds     []string `yaml:"requires_credentials"`
	ProducesWallpaper bool     `yaml:"produces_wallpaper"`

	// external-only.
	Version         string `yaml:"version"`
	ProtocolVersion string `yaml:"protocol_version"`
	Repository      string `yaml:"repository"`
	Install         string `yaml:"install"`
}

type readmeReload struct {
	Method             string `yaml:"method"`
	Command            string `yaml:"command"`
	UserActionRequired bool   `yaml:"user_action_required"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case sourceBuiltin:
		runBuiltin(os.Args[2:])
	case sourceExternal:
		runExternal(os.Args[2:])
	case "-h", "--help":
		printUsage()
	default:
		// G705 false positive: os.Args is shell-quoted user input, but we're
		// writing to stderr in a CLI, not rendering HTML.
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n", os.Args[1]) //nolint:gosec
		printUsage()
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  tinct-check-readmes builtin [--root internal/plugin/output]")
	fmt.Fprintln(os.Stderr, "  tinct-check-readmes external --binary path/to/plugin [--readme path/to/README.md]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Warn-only. Always exits 0. Diagnostics go to stderr.")
}

func runBuiltin(args []string) {
	fs := flag.NewFlagSet(sourceBuiltin, flag.ExitOnError)
	root := fs.String("root", "internal/plugin/output", "root directory containing built-in plugin sub-dirs")
	categoriesFile := fs.String("categories", "internal/plugin/output/categories.json", "path to output category enum (single source of truth)")
	if err := fs.Parse(args); err != nil {
		return
	}

	knownCategories = loadKnownCategories(*categoriesFile)

	plugins := builtinRegistry()
	sort.Slice(plugins, func(i, j int) bool { return plugins[i].Name() < plugins[j].Name() })

	checked := 0
	for _, p := range plugins {
		name := p.Name()
		readmePath := filepath.Join(*root, name, "README.md")
		fm, ok := loadFrontmatter(readmePath, name)
		if !ok {
			continue
		}
		checked++
		compareBuiltin(p, readmePath, fm)
	}

	// Also walk input plugins. The --root flag is for output by
	// convention; derive the input root from it (sibling directory).
	inputRoot := filepath.Join(filepath.Dir(*root), "input")
	inputs := inputBuiltinRegistry()
	inputChecked := 0
	for _, ip := range inputs {
		readmePath := filepath.Join(inputRoot, ip.dir, "README.md")
		fm, ok := loadFrontmatter(readmePath, ip.plugin.Name())
		if !ok {
			continue
		}
		inputChecked++
		compareInputBuiltin(ip.plugin, readmePath, fm)
	}

	fmt.Fprintf(os.Stderr, "\ntinct-check-readmes: %d warning(s); %d of %d output, %d of %d input plugin(s) had parseable frontmatter\n",
		warningCount, checked, len(plugins), inputChecked, len(inputs))
}

func runExternal(args []string) {
	fs := flag.NewFlagSet(sourceExternal, flag.ExitOnError)
	binary := fs.String("binary", "", "path to the external plugin binary (required)")
	readme := fs.String("readme", "", "path to the README (default: README.md in same dir as --binary)")
	if err := fs.Parse(args); err != nil {
		return
	}
	if *binary == "" {
		fmt.Fprintln(os.Stderr, "external mode requires --binary")
		return
	}

	if *readme == "" {
		*readme = filepath.Join(filepath.Dir(*binary), "README.md")
	}

	info, err := runPluginInfo(*binary)
	if err != nil {
		warnf("%s: --plugin-info failed: %v", filepath.Base(*binary), err)
		return
	}

	fm, ok := loadFrontmatter(*readme, info.Name)
	if !ok {
		fmt.Fprintf(os.Stderr, "\ntinct-check-readmes: %d warning(s); 0 of 1 external plugin(s) had parseable frontmatter\n",
			warningCount)
		return
	}

	compareExternal(info, *readme, fm)
	fmt.Fprintf(os.Stderr, "\ntinct-check-readmes: %d warning(s); 1 of 1 external plugin(s) had parseable frontmatter\n",
		warningCount)
}

// loadFrontmatter reads a README and unmarshals its YAML frontmatter.
// Missing READMEs are warned and return ok=false; parse errors warn
// and return ok=false; success returns ok=true.
func loadFrontmatter(path, pluginName string) (readmeFrontmatter, bool) {
	var fm readmeFrontmatter
	data, err := os.ReadFile(path)
	if err != nil {
		warnf("%s: README missing or unreadable at %s (%v)", pluginName, path, err)
		return fm, false
	}
	raw, err := extractFrontmatter(data)
	if err != nil {
		warnf("%s: README at %s has no parseable frontmatter (%v)", pluginName, path, err)
		return fm, false
	}
	if err := yaml.Unmarshal(raw, &fm); err != nil {
		warnf("%s: frontmatter YAML parse error: %v", pluginName, err)
		return fm, false
	}
	return fm, true
}

// extractFrontmatter pulls the YAML block between the first two `---`
// fences at the start of the file.
func extractFrontmatter(data []byte) ([]byte, error) {
	const fence = "---"
	trimmed := bytes.TrimLeft(data, "\n\r ")
	if !bytes.HasPrefix(trimmed, []byte(fence)) {
		return nil, fmt.Errorf("file does not start with %q", fence)
	}
	rest := trimmed[len(fence):]
	// Skip the newline after the opening fence.
	if i := bytes.IndexByte(rest, '\n'); i >= 0 {
		rest = rest[i+1:]
	}
	body, _, ok := bytes.Cut(rest, []byte("\n"+fence))
	if !ok {
		return nil, fmt.Errorf("no closing %q fence found", fence)
	}
	return body, nil
}

func compareBuiltin(p output.Plugin, _ string, fm readmeFrontmatter) {
	name := p.Name()

	if fm.Plugin.Type != "output" {
		warnf("%s: frontmatter type=%q expected %q", name, fm.Plugin.Type, "output")
	}
	if fm.Plugin.Name != name {
		warnf("%s: frontmatter name=%q differs from Plugin.Name()=%q", name, fm.Plugin.Name, name)
	}
	if fm.Plugin.Source != sourceBuiltin {
		warnf("%s: frontmatter source=%q expected %q", name, fm.Plugin.Source, sourceBuiltin)
	}
	if len(knownCategories) > 0 && !knownCategories[fm.Plugin.Category] {
		warnf("%s: frontmatter category=%q not in categories.json; plugin will be invisible on the docs landing table and orphaned in the sidebar until the enum is extended",
			name, fm.Plugin.Category)
	}

	actualDir := foldHome(p.DefaultOutputDir())
	if fm.Plugin.DefaultOutputDir != actualDir {
		warnf("%s: frontmatter default_output_dir=%q differs from runtime %q",
			name, fm.Plugin.DefaultOutputDir, actualDir)
	}

	if hp, ok := p.(hooks.Provider); ok {
		spec := hp.Hooks()
		if !equalSets(fm.Plugin.Requires, spec.RequiredBinaries) {
			warnf("%s: frontmatter requires=%v differs from RequiredBinaries=%v",
				name, fm.Plugin.Requires, spec.RequiredBinaries)
		}
		if !equalSets(fm.Plugin.Optional, spec.OptionalBinaries) {
			warnf("%s: frontmatter optional=%v differs from OptionalBinaries=%v",
				name, fm.Plugin.Optional, spec.OptionalBinaries)
		}
		actualMethod := deriveReloadMethod(spec)
		if fm.Plugin.Reload == nil {
			if actualMethod != "none" {
				warnf("%s: frontmatter has no reload block but runtime reload method=%q",
					name, actualMethod)
			}
		} else if fm.Plugin.Reload.Method != actualMethod {
			warnf("%s: frontmatter reload.method=%q differs from runtime %q",
				name, fm.Plugin.Reload.Method, actualMethod)
		}
	}
}

//nolint:gocyclo // long but linear sequence of field-by-field comparisons; splitting would obscure the diff intent
func compareExternal(info plugin.PluginInfo, _ string, fm readmeFrontmatter) {
	name := info.Name

	if fm.Plugin.Name != name {
		warnf("%s: frontmatter name=%q differs from --plugin-info name=%q", name, fm.Plugin.Name, name)
	}
	if fm.Plugin.Source != sourceExternal {
		warnf("%s: frontmatter source=%q expected %q", name, fm.Plugin.Source, sourceExternal)
	}
	if fm.Plugin.Version != "" && fm.Plugin.Version != info.Version {
		warnf("%s: frontmatter version=%q differs from --plugin-info version=%q",
			name, fm.Plugin.Version, info.Version)
	}
	if fm.Plugin.ProtocolVersion != "" && fm.Plugin.ProtocolVersion != info.ProtocolVersion {
		warnf("%s: frontmatter protocol_version=%q differs from --plugin-info protocol_version=%q",
			name, fm.Plugin.ProtocolVersion, info.ProtocolVersion)
	}

	if info.Metadata == nil {
		warnf("%s: --plugin-info has no metadata block; cannot check requires/output_dir/reload",
			name)
		return
	}
	m := info.Metadata
	if !equalSets(fm.Plugin.Requires, m.RequiredBinaries) {
		warnf("%s: frontmatter requires=%v differs from metadata.required_binaries=%v",
			name, fm.Plugin.Requires, m.RequiredBinaries)
	}
	if !equalSets(fm.Plugin.Optional, m.OptionalBinaries) {
		warnf("%s: frontmatter optional=%v differs from metadata.optional_binaries=%v",
			name, fm.Plugin.Optional, m.OptionalBinaries)
	}
	if fm.Plugin.DefaultOutputDir != m.DefaultOutputDir {
		warnf("%s: frontmatter default_output_dir=%q differs from metadata.default_output_dir=%q",
			name, fm.Plugin.DefaultOutputDir, m.DefaultOutputDir)
	}
	if !equalSets(fm.Plugin.GeneratedFiles, m.GeneratedFiles) {
		warnf("%s: frontmatter generated_files=%v differs from metadata.generated_files=%v",
			name, fm.Plugin.GeneratedFiles, m.GeneratedFiles)
	}
	if fm.Plugin.Pattern != "" && fm.Plugin.Pattern != m.Pattern {
		warnf("%s: frontmatter pattern=%q differs from metadata.pattern=%q",
			name, fm.Plugin.Pattern, m.Pattern)
	}
	if m.Reload != nil && fm.Plugin.Reload != nil {
		if fm.Plugin.Reload.Method != m.Reload.Method {
			warnf("%s: frontmatter reload.method=%q differs from metadata.reload.method=%q",
				name, fm.Plugin.Reload.Method, m.Reload.Method)
		}
	}
}

// compareInputBuiltin validates an input plugin's README frontmatter
// against its runtime. Input plugins don't expose DefaultOutputDir,
// hooks.Spec, or generated files, so most fields are documentation-only.
// We verify what we can: type discriminator, plugin name, source,
// optionally produces_wallpaper (derived from WallpaperProvider).
func compareInputBuiltin(p input.Plugin, _ string, fm readmeFrontmatter) {
	name := p.Name()

	if fm.Plugin.Type != "input" {
		warnf("%s: frontmatter type=%q expected %q", name, fm.Plugin.Type, "input")
	}
	if fm.Plugin.Name != name {
		warnf("%s: frontmatter name=%q differs from Plugin.Name()=%q", name, fm.Plugin.Name, name)
	}
	if fm.Plugin.Source != sourceBuiltin {
		warnf("%s: frontmatter source=%q expected %q", name, fm.Plugin.Source, sourceBuiltin)
	}

	// If the plugin implements WallpaperProvider, produces_wallpaper
	// should be true. If not, true is suspicious (warn-only — there are
	// edge cases where wallpaper is produced via another mechanism).
	_, hasWallpaper := p.(input.WallpaperProvider)
	if hasWallpaper && !fm.Plugin.ProducesWallpaper {
		warnf("%s: plugin implements WallpaperProvider but frontmatter produces_wallpaper=false", name)
	}
	if !hasWallpaper && fm.Plugin.ProducesWallpaper {
		warnf("%s: frontmatter produces_wallpaper=true but plugin does not implement WallpaperProvider", name)
	}
}

// deriveReloadMethod maps a runtime hooks.Spec into the README's
// reload.method vocabulary (signal/ipc/watch/wallpaper-apply/none).
//
// Precedence: an explicit Reload/ReloadFn beats SupportsWallpaper, since
// a plugin that does both is primarily a config reloader that also
// happens to push a wallpaper. Wallpaper-only plugins (awww, wbg,
// hyprpaper) are the SupportsWallpaper-without-Reload case.
func deriveReloadMethod(spec hooks.Spec) string {
	if spec.Reload != nil {
		switch spec.Reload.Verb {
		case hooks.VerbSignal:
			return "signal"
		case hooks.VerbExec:
			return "ipc"
		}
	}
	if spec.ReloadFn != nil {
		// Custom ReloadFn — treat as ipc-equivalent (the plugin's Go code
		// does whatever it does, but it's an active action).
		return "ipc"
	}
	if spec.SupportsWallpaper {
		return "wallpaper-apply"
	}
	return "none"
}

// runPluginInfo execs <binary> --plugin-info and parses the JSON.
func runPluginInfo(binary string) (plugin.PluginInfo, error) {
	var info plugin.PluginInfo
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "--plugin-info")
	out, err := cmd.Output()
	if err != nil {
		return info, fmt.Errorf("exec %s --plugin-info: %w", binary, err)
	}
	if err := json.Unmarshal(out, &info); err != nil {
		return info, fmt.Errorf("invalid JSON: %w", err)
	}
	return info, nil
}

// foldHome rewrites $HOME-prefixed paths back to a leading ~ so they
// can be compared to the (unexpanded) frontmatter value.
func foldHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if path == home {
		return "~"
	}
	if strings.HasPrefix(path, home+string(filepath.Separator)) {
		return "~" + path[len(home):]
	}
	return path
}

// equalSets compares two slices as sets (order-independent, dedup'd).
// nil and empty slice are equivalent.
func equalSets(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	as := make(map[string]struct{}, len(a))
	bs := make(map[string]struct{}, len(b))
	for _, x := range a {
		as[x] = struct{}{}
	}
	for _, x := range b {
		bs[x] = struct{}{}
	}
	if len(as) != len(bs) {
		return false
	}
	for k := range as {
		if _, ok := bs[k]; !ok {
			return false
		}
	}
	return true
}

// knownCategories is the set of output plugin category IDs loaded
// from internal/plugin/output/categories.json. Empty until
// loadKnownCategories runs; an empty map disables the
// unknown-category warning (so the external mode, which doesn't
// require the file, still works).
var knownCategories map[string]bool

// loadKnownCategories reads the category enum from the given path.
// Failure is treated as "no enum available" — warned, not fatal —
// because the check tool's primary job (README ↔ runtime diff) still
// works without it.
func loadKnownCategories(path string) map[string]bool {
	data, err := os.ReadFile(path)
	if err != nil {
		warnf("could not load %s (%v) — category enum validation disabled", path, err)
		return nil
	}
	var cats []struct {
		ID    string `json:"id"`
		Label string `json:"label"`
	}
	if err := json.Unmarshal(data, &cats); err != nil {
		warnf("%s: failed to parse: %v — category enum validation disabled", path, err)
		return nil
	}
	known := make(map[string]bool, len(cats))
	for _, c := range cats {
		known[c.ID] = true
	}
	return known
}

// warningCount counts every warning emitted so the summary line at the
// end of a run reflects load-stage failures (missing READMEs, parse
// errors) as well as compare-stage drifts.
var warningCount int

func warnf(format string, args ...any) {
	warningCount++
	fmt.Fprintf(os.Stderr, "WARN "+format+"\n", args...)
}
