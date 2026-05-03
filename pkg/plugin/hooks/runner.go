package hooks

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
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
		if _, statErr := os.Stat(d); os.IsNotExist(statErr) {
			return true, fmt.Sprintf("required directory does not exist: %s", d), nil
		}
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

func printInstructions(spec Spec, hctx Context) {
	var msg string
	switch {
	case spec.InstructionsFn != nil:
		msg = spec.InstructionsFn(hctx)
	case spec.Instructions != "":
		msg = renderTemplate(spec.Instructions, hctx)
	}
	if msg != "" {
		fmt.Fprintln(os.Stderr, msg)
	}
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
