package hooks

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jmylchreest/tinct/pkg/util/appdetect"
)

// RunPre evaluates the spec's pre-execute checks. It returns the same
// (skip, reason, err) shape the manager already uses for imperative
// pre-hooks, so it can be called either before or instead of one.
//
// hctx supplies the plugin name (used to prefix verbose-mode messages),
// the resolved output directory (consulted by AutoCreateDir), and the
// verbose flag. WrittenFiles / WallpaperPath are unused here but the
// shared Context shape keeps the runner API symmetric with RunPost.
func RunPre(spec Spec, hctx Context) (skip bool, reason string, err error) {
	for _, bin := range spec.RequiredBinaries {
		if !appdetect.IsPresentAny([]string{bin}, nil) {
			return true, fmt.Sprintf("%s executable not found on $PATH", bin), nil
		}
	}

	if hctx.Verbose {
		for _, bin := range spec.OptionalBinaries {
			if !appdetect.IsPresentAny([]string{bin}, nil) {
				fmt.Fprintf(os.Stderr, "%swarning: %s not found - some functionality may be limited\n", linePrefix(hctx.PluginName), bin)
			}
		}
	}

	for _, d := range spec.RequiredDirs {
		expanded := expandHome(d)
		if _, statErr := os.Stat(expanded); os.IsNotExist(statErr) {
			return true, fmt.Sprintf("required directory does not exist: %s", d), nil
		}
	}

	for _, group := range spec.RequiredAny {
		if group.satisfied() {
			continue
		}
		return true, group.skipReason(), nil
	}

	if spec.AutoCreateDir && hctx.OutputDir != "" {
		if _, statErr := os.Stat(hctx.OutputDir); os.IsNotExist(statErr) {
			if mkErr := os.MkdirAll(hctx.OutputDir, 0o750); mkErr != nil {
				return true, fmt.Sprintf("failed to create directory %s: %v", hctx.OutputDir, mkErr), nil
			}
			if hctx.Verbose {
				fmt.Fprintf(os.Stderr, "%screated directory: %s\n", linePrefix(hctx.PluginName), hctx.OutputDir)
			}
		}
	}

	return false, "", nil
}

// satisfied reports whether any single candidate in the group is
// present. Binaries go through appdetect (PATH, Flatpak, AppImage);
// paths are checked with os.Stat rather than appdetect's dirExists,
// because detection markers are frequently plain files (kdeglobals,
// config.toml) and appdetect only counts directories.
func (g AnyOf) satisfied() bool {
	for _, bin := range g.Binaries {
		if bin != "" && appdetect.IsPresentAny([]string{bin}, nil) {
			return true
		}
	}
	for _, d := range g.Dirs {
		if d == "" {
			continue
		}
		if _, err := os.Stat(expandHome(d)); err == nil {
			return true
		}
	}
	return false
}

// skipReason renders the message shown when an AnyOf group is not
// satisfied: the plugin's own Reason when it set one, otherwise a
// generated "none of ..." listing every candidate the group accepted.
func (g AnyOf) skipReason() string {
	if g.Reason != "" {
		return g.Reason
	}

	candidates := make([]string, 0, len(g.Binaries)+len(g.Dirs))
	for _, b := range g.Binaries {
		if b != "" {
			candidates = append(candidates, b+" (executable)")
		}
	}
	for _, d := range g.Dirs {
		if d != "" {
			candidates = append(candidates, d)
		}
	}

	switch len(candidates) {
	case 0:
		return "no detection candidates declared"
	case 1:
		return "not found: " + candidates[0]
	default:
		return "none of the following were found: " + strings.Join(candidates, ", ")
	}
}

// Indent formats a possibly multi-line message for display under a
// "⊘ Skipping <plugin>: " or "   <plugin>: " style header.
//
// The first line is returned as-is — it follows the header on the same
// line. Continuation lines are dedented by their common leading
// whitespace and then re-indented to `indent`, so a message written as
// a naturally-indented Go string literal lands aligned under the header
// while keeping its *relative* structure: a command nested under a
// numbered step stays nested. Blank lines stay blank rather than
// becoming trailing whitespace.
//
// Exported because the CLI formats skip reasons at its own indent level
// and should not reimplement this.
func Indent(msg, indent string) string {
	lines := strings.Split(msg, "\n")
	if len(lines) == 1 {
		return msg
	}

	common := commonIndent(lines[1:])

	out := make([]string, 0, len(lines))
	out = append(out, strings.TrimRight(lines[0], " \t"))
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			out = append(out, "")
			continue
		}
		out = append(out, indent+strings.TrimRight(strings.TrimPrefix(line, common), " \t"))
	}
	return strings.Join(out, "\n")
}

// commonIndent returns the longest leading run of spaces/tabs shared by
// every non-blank line, which Indent strips before re-indenting.
func commonIndent(lines []string) string {
	common := ""
	first := true
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lead := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		if first {
			common, first = lead, false
			continue
		}
		for !strings.HasPrefix(lead, common) {
			common = common[:len(common)-1]
		}
	}
	return common
}

// linePrefix returns the verbose-output prefix for a plugin. When name is
// non-empty we produce "   kitty: " so users can scan multi-plugin runs.
// When name is empty (older callers, tests) we keep the historical
// three-space indent so output stays aligned with the rest of tinct.
func linePrefix(name string) string {
	if name == "" {
		return "   "
	}
	return "   " + name + ": "
}

// expandHome expands a leading ~ (or ~/path) to the user's home dir so
// plugins can declare RequiredDirs in the natural shell-style form
// without having to call os.UserHomeDir themselves. Other ~user forms
// are intentionally left alone — runner.go isn't the place to ape every
// shell's tilde expansion.
func expandHome(p string) string {
	if !strings.HasPrefix(p, "~") {
		return p
	}
	if p != "~" && !strings.HasPrefix(p, "~/") {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	if p == "~" {
		return home
	}
	return filepath.Join(home, p[2:])
}

// RunPost evaluates the spec's post-execute actions: chmod marked files,
// reload the target app, set the wallpaper if applicable, and print
// instructions. Errors inside RunPost are non-fatal — they're logged in
// verbose mode and the function returns nil. This matches existing
// PostExecuteHook semantics.
func RunPost(ctx context.Context, spec Spec, hctx Context) error {
	if hctx.DryRun {
		return nil
	}

	if len(spec.MakeExecutable) > 0 {
		applyMakeExecutable(spec.MakeExecutable, hctx.WrittenFiles, hctx.PluginName, hctx.Verbose)
	}

	if spec.ReloadFn != nil {
		if err := spec.ReloadFn(ctx); err != nil && hctx.Verbose {
			fmt.Fprintf(os.Stderr, "%swarning: reload failed: %v\n", linePrefix(hctx.PluginName), err)
		}
	} else if spec.Reload != nil {
		runReloadVerb(ctx, *spec.Reload, hctx.PluginName, hctx.Verbose)
	}

	if spec.SupportsWallpaper && spec.Wallpaper != nil && hctx.WallpaperPath != "" {
		if err := spec.Wallpaper(ctx, hctx.WallpaperPath); err != nil && hctx.Verbose {
			fmt.Fprintf(os.Stderr, "%swarning: wallpaper apply failed: %v\n", linePrefix(hctx.PluginName), err)
		}
	}

	if hctx.Verbose && len(hctx.WrittenFiles) > 0 {
		printInstructions(spec, hctx)
	}

	return nil
}

func applyMakeExecutable(names, written []string, pluginName string, verbose bool) {
	for _, name := range names {
		for _, f := range written {
			if filepath.Base(f) != name {
				continue
			}
			if err := os.Chmod(f, 0o755); err != nil && verbose {
				fmt.Fprintf(os.Stderr, "%swarning: failed to chmod %s: %v\n", linePrefix(pluginName), f, err)
			}
		}
	}
}

// instructionIndent is the hanging indent for continuation lines of a
// multi-line Instructions block, sized to sit clear of the "   " that
// prefixes every tinct output line without tracking the plugin-name
// width (which varies per plugin).
const instructionIndent = "     "

func printInstructions(spec Spec, hctx Context) {
	var msg string
	switch {
	case spec.InstructionsFn != nil:
		msg = spec.InstructionsFn(hctx)
	case spec.Instructions != "":
		msg = renderTemplate(spec.Instructions, hctx)
	}
	if msg == "" {
		return
	}

	// Instructions have no header line of their own, so the plugin
	// prefix is added here and continuation lines get a hanging indent
	// rather than repeating "<plugin>: " on every row. TrimLeft keeps
	// older single-line Instructions — which embed their own leading
	// indent — from being double-indented.
	prefix := linePrefix(hctx.PluginName)
	fmt.Fprintln(os.Stderr, prefix+Indent(strings.TrimLeft(msg, " \t"), instructionIndent))
}

// renderTemplate runs the Instructions string through text/template with a
// minimal data shape. Failures fall back to the raw template string so a
// broken template doesn't silently swallow the message.
func renderTemplate(s string, hctx Context) string {
	tmpl, err := template.New("instructions").Parse(s)
	if err != nil {
		return s
	}
	var buf bytes.Buffer
	data := map[string]any{
		"OutputDir":     hctx.OutputDir,
		"WrittenFiles":  hctx.WrittenFiles,
		"WallpaperPath": hctx.WallpaperPath,
	}
	if execErr := tmpl.Execute(&buf, data); execErr != nil {
		return s
	}
	return buf.String()
}
