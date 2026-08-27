// Package cli provides the command-line interface for Tinct.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/manager"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/pkg/plugin/hooks"
)

// loadAndConfigurePlugins loads the plugin manifest and configures plugins.
func loadAndConfigurePlugins() error {
	lock, _, err := loadPluginManifest()
	if err != nil || lock == nil {
		return nil // No manifest is OK
	}

	// Register external plugins with absolute path resolution.
	registerExternalPluginsFromManifest(lock, true, generateVerbose)

	// Configure external plugins with dry-run and args.
	if lock.ExternalPlugins != nil {
		for _, meta := range lock.ExternalPlugins {
			pluginName := meta.Name
			if pluginName == "" {
				pluginName, _, _, _, _ = queryPluginMetadata(meta.Path)
			}
			if pluginName != "" {
				configureExternalPlugin(pluginName, meta.Type, generateDryRun, generatePluginArgs, generateVerbose)
			}
		}
	}

	return nil
}

// getAndValidateInputPlugin retrieves and validates the input plugin.
func getAndValidateInputPlugin() (input.Plugin, error) {
	plugin, ok := sharedPluginManager.GetInputPlugin(generateInputPlugin)
	if !ok {
		allInput := sharedPluginManager.AllInputPlugins()
		availablePlugins := make([]string, 0, len(allInput))
		for pluginName := range allInput {
			availablePlugins = append(availablePlugins, pluginName)
		}
		return nil, fmt.Errorf("unknown input plugin: %s (available: %s)", generateInputPlugin, strings.Join(availablePlugins, ", "))
	}

	// Set verbose before Validate: for external plugins Validate is the
	// first RPC, and it starts the subprocess whose logger is fixed for
	// the life of that process. Setting it afterwards would be too late
	// and the run would silently lose plugin debug output.
	if verbosePlugin, ok := plugin.(interface{ SetVerbose(bool) }); ok {
		verbosePlugin.SetVerbose(generateVerbose)
	}

	if err := plugin.Validate(); err != nil {
		return nil, fmt.Errorf("input plugin validation failed: %w", err)
	}

	return plugin, nil
}

// generateInputPalette generates a raw palette from the input plugin.
// Returns the palette, canonical wallpaper path, and raw wallpaper path.
func generateInputPalette(ctx context.Context, inputPlugin input.Plugin) (palette *colour.Palette, wallpaperPath, wallpaperRawPath string, err error) {
	if generateVerbose {
		fmt.Fprintf(os.Stderr, " Input plugin: %s\n", inputPlugin.Name())
		fmt.Fprintf(os.Stderr, "   %s\n", inputPlugin.Description())
	}

	// Prepare options for input plugin.
	inputOpts := buildInputOptions()

	// Generate raw palette from input plugin.
	palette, err = inputPlugin.Generate(ctx, inputOpts)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate palette: %w", err)
	}

	if generateVerbose {
		fmt.Fprintf(os.Stderr, "   Generated raw palette (%d colours)\n", len(palette.Colors))
	}

	// Extract wallpaper paths if available.
	wallpaperPath, wallpaperRawPath = extractWallpaperPaths(inputPlugin)
	if generateVerbose && wallpaperPath != "" {
		fmt.Fprintf(os.Stderr, "   Wallpaper source: %s\n", wallpaperPath)
	}

	return palette, wallpaperPath, wallpaperRawPath, nil
}

// buildInputOptions creates input plugin options.
func buildInputOptions() input.GenerateOptions {
	inputOpts := input.GenerateOptions{
		Verbose:         generateVerbose,
		DryRun:          generateDryRun,
		Backend:         generateBackend,
		ColourOverrides: []string{},
		PluginArgs:      make(map[string]any),
	}

	// Extract plugin-specific args if provided.
	if argsJSON, ok := generatePluginArgs[generateInputPlugin]; ok {
		var args map[string]any
		if err := json.Unmarshal([]byte(argsJSON), &args); err == nil {
			inputOpts.PluginArgs = args
			if generateVerbose {
				fmt.Fprintf(os.Stderr, "   Plugin args: %v\n", args)
			}
		} else if generateVerbose {
			fmt.Fprintf(os.Stderr, "   Failed to parse plugin args: %v\n", err)
		}
	}

	return inputOpts
}

// extractWallpaperPaths extracts both canonical and raw wallpaper paths from input plugin.
// Returns (canonicalPath, rawPath) - both may be empty if plugin doesn't provide wallpapers.
func extractWallpaperPaths(inputPlugin input.Plugin) (canonicalPath, rawPath string) {
	if provider, ok := inputPlugin.(input.WallpaperProvider); ok {
		return provider.WallpaperPath(), provider.WallpaperRawPath()
	}
	return "", ""
}

// categorizePalette categorizes a raw palette for both primary and alternate themes.
// Always returns both primary and alternate palettes to support dual-theme plugins.
func categorizePalette(rawPalette *colour.Palette, inputPlugin input.Plugin, desaturate bool) (primary, alternate *colour.CategorisedPalette) {
	themeType := determineThemeType(inputPlugin)

	// Determine primary theme
	primaryTheme := themeType
	if themeType == colour.ThemeAuto {
		config := colour.DefaultCategorisationConfig()
		config.ThemeType = colour.ThemeAuto
		tempPalette := colour.Categorise(rawPalette, config)
		primaryTheme = tempPalette.ThemeType
	}

	// Categorize for primary theme
	primaryConfig := colour.DefaultCategorisationConfig()
	primaryConfig.ThemeType = primaryTheme
	primaryPalette := colour.Categorise(rawPalette, primaryConfig)

	// Determine alternate theme (opposite of primary)
	var alternateTheme colour.ThemeType
	if primaryTheme == colour.ThemeDark {
		alternateTheme = colour.ThemeLight
	} else {
		alternateTheme = colour.ThemeDark
	}

	// For alternate theme, optionally desaturate the palette
	// to avoid highly saturated backgrounds in light themes
	alternatePaletteRaw := rawPalette
	if desaturate {
		alternatePaletteRaw = desaturatePalette(rawPalette, alternateTheme)
	}

	// Categorize for alternate theme
	alternateConfig := colour.DefaultCategorisationConfig()
	alternateConfig.ThemeType = alternateTheme
	alternatePalette := colour.Categorise(alternatePaletteRaw, alternateConfig)

	if generateVerbose {
		fmt.Fprintf(os.Stderr, "   Categorized palette (%d colours, %s theme)\n",
			len(primaryPalette.AllColours), primaryPalette.ThemeType.String())
		if desaturate {
			fmt.Fprintf(os.Stderr, "   Categorized alternate palette (%d colours, %s theme, desaturated)\n",
				len(alternatePalette.AllColours), alternatePalette.ThemeType.String())
		} else {
			fmt.Fprintf(os.Stderr, "   Categorized alternate palette (%d colours, %s theme)\n",
				len(alternatePalette.AllColours), alternatePalette.ThemeType.String())
		}
		fmt.Fprintf(os.Stderr, "   Plugin execution complete.\n")
	}

	return primaryPalette, alternatePalette
}

// desaturatePalette creates a new palette with inverted luminance and conditional desaturation for alternate themes.
// All colors are inverted. For light themes, very bright colors are desaturated to create neutral backgrounds.
func desaturatePalette(palette *colour.Palette, themeType colour.ThemeType) *colour.Palette {
	if palette == nil || len(palette.Colors) == 0 {
		return palette
	}

	const brightThreshold = 0.7 // Very bright colors get desaturated (after inversion)

	adjusted := make([]color.Color, len(palette.Colors))
	for i, c := range palette.Colors {
		r, g, b, a := c.RGBA()
		// Convert to 0-255 range (shift ensures values fit in uint8)
		r8 := uint8(r >> 8) //nolint:gosec // Safe: shifted uint32 always fits in uint8
		g8 := uint8(g >> 8) //nolint:gosec // Safe: shifted uint32 always fits in uint8
		b8 := uint8(b >> 8) //nolint:gosec // Safe: shifted uint32 always fits in uint8
		a8 := uint8(a >> 8) //nolint:gosec // Safe: shifted uint32 always fits in uint8

		// Convert to HSL
		h, s, l := rgbToHSL(r8, g8, b8)

		// Always invert luminance for alternate theme
		l = 1.0 - l

		// For light themes, desaturate very bright colors to avoid garish backgrounds
		if themeType == colour.ThemeLight && l > brightThreshold {
			s *= 0.4
		}

		// Convert back to RGB
		rNew, gNew, bNew := hslToRGB(h, s, l)
		adjusted[i] = color.RGBA{R: rNew, G: gNew, B: bNew, A: a8}
	}

	return &colour.Palette{
		Colors:    adjusted,
		Weights:   palette.Weights,
		RoleHints: palette.RoleHints,
	}
}

// rgbToHSL converts RGB (0-255) to HSL (0-1).
func rgbToHSL(r, g, b uint8) (h, s, l float64) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	maxVal := rf
	if gf > maxVal {
		maxVal = gf
	}
	if bf > maxVal {
		maxVal = bf
	}

	minVal := rf
	if gf < minVal {
		minVal = gf
	}
	if bf < minVal {
		minVal = bf
	}

	l = (maxVal + minVal) / 2.0

	if maxVal == minVal {
		h = 0
		s = 0
	} else {
		d := maxVal - minVal
		if l > 0.5 {
			s = d / (2.0 - maxVal - minVal)
		} else {
			s = d / (maxVal + minVal)
		}

		switch maxVal {
		case rf:
			h = (gf - bf) / d
			if gf < bf {
				h += 6.0
			}
		case gf:
			h = (bf-rf)/d + 2.0
		case bf:
			h = (rf-gf)/d + 4.0
		}
		h /= 6.0
	}

	return h, s, l
}

// hslToRGB converts HSL (0-1) to RGB (0-255).
func hslToRGB(h, s, l float64) (r, g, b uint8) {
	var rf, gf, bf float64

	if s == 0 {
		rf = l
		gf = l
		bf = l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1.0 + s)
		} else {
			q = l + s - l*s
		}
		p := 2.0*l - q

		rf = hueToRGB(p, q, h+1.0/3.0)
		gf = hueToRGB(p, q, h)
		bf = hueToRGB(p, q, h-1.0/3.0)
	}

	r = uint8(rf * 255.0)
	g = uint8(gf * 255.0)
	b = uint8(bf * 255.0)
	return r, g, b
}

// hueToRGB is a helper for HSL to RGB conversion.
func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1.0
	}
	if t > 1 {
		t -= 1.0
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6.0*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6.0
	}
	return p
}

// determineThemeType determines the theme type from global flag and plugin hints.
func determineThemeType(inputPlugin input.Plugin) colour.ThemeType {
	switch globalTheme {
	case "dark":
		return colour.ThemeDark
	case "light":
		return colour.ThemeLight
	case "auto":
		// Check if plugin provides a theme hint (optional).
		if hinter, ok := inputPlugin.(input.ThemeHinter); ok {
			hint := hinter.ThemeHint()
			if generateVerbose && hint != "" && hint != "auto" {
				fmt.Fprintf(os.Stderr, "   Plugin suggests theme: %s\n", hint)
			}
		}
		return colour.ThemeAuto
	default:
		return colour.ThemeAuto
	}
}

// handlePaletteOutput shows preview and saves palette if requested.
func handlePaletteOutput(palette *colour.CategorisedPalette) error {
	// Show preview if requested.
	if generatePreview {
		fmt.Println()
		fmt.Println(palette.StringWithPreview(true))
		fmt.Println()
	}

	// Save palette if requested.
	if generateSavePalette != "" {
		if err := savePalette(palette, generateSavePalette); err != nil {
			return fmt.Errorf("failed to save palette: %w", err)
		}
		if generateVerbose {
			fmt.Fprintf(os.Stderr, " Saved palette to: %s\n", generateSavePalette)
		}
	}

	return nil
}

// selectOutputPlugins determines which output plugins to run.
func selectOutputPlugins() ([]output.Plugin, error) {
	if len(generateOutputs) == 1 && generateOutputs[0] == pluginTypeAll {
		// Run all registered plugins.
		allPlugins := sharedPluginManager.AllOutputPlugins()
		plugins := make([]output.Plugin, 0, len(allPlugins))
		for _, plugin := range allPlugins {
			plugins = append(plugins, plugin)
		}
		if len(plugins) == 0 {
			return nil, fmt.Errorf("no output plugins available")
		}
		return plugins, nil
	}

	// Run specific plugins by name.
	plugins := make([]output.Plugin, 0, len(generateOutputs))
	for _, name := range generateOutputs {
		plugin, ok := sharedPluginManager.GetOutputPlugin(name)
		if !ok {
			availablePlugins := make([]string, 0)
			for pluginName := range sharedPluginManager.AllOutputPlugins() {
				availablePlugins = append(availablePlugins, pluginName)
			}
			return nil, fmt.Errorf("unknown output plugin: %s (available: %s)", name, strings.Join(availablePlugins, ", "))
		}
		plugins = append(plugins, plugin)
	}

	if len(plugins) == 0 {
		return nil, fmt.Errorf("no output plugins selected")
	}

	return plugins, nil
}

// pluginExecution tracks the execution state of an output plugin.
type pluginExecution struct {
	plugin       output.Plugin
	skip         bool
	skipReason   string
	writtenFiles []string
}

// preparePluginExecutions validates plugins and runs pre-execute hooks.
func preparePluginExecutions(ctx context.Context, plugins []output.Plugin) []pluginExecution {
	executions := make([]pluginExecution, 0, len(plugins))

	for _, plugin := range plugins {
		exec := pluginExecution{plugin: plugin}

		// Set verbose mode if plugin supports it.
		if verbosePlugin, ok := plugin.(output.VerbosePlugin); ok {
			verbosePlugin.SetVerbose(generateVerbose)
		}

		// Validate plugin.
		if err := plugin.Validate(); err != nil {
			if generateVerbose {
				fmt.Fprintf(os.Stderr, " Skipping %s: %v\n", plugin.Name(), err)
			}
			exec.skip = true
			exec.skipReason = fmt.Sprintf("validation failed: %v", err)
			executions = append(executions, exec)
			continue
		}

		// Run PreExecute hook if plugin implements it.
		if shouldSkipFromPreHook(ctx, plugin, &exec) {
			executions = append(executions, exec)
			continue
		}

		executions = append(executions, exec)
	}

	return executions
}

// shouldSkipFromPreHook runs the pre-execute checks and determines if a
// plugin should be skipped. It evaluates the declarative hooks.Spec first
// (if the plugin implements hooks.Provider), then the optional imperative
// PreExecuteHook. Either gating skip propagates to the caller.
func shouldSkipFromPreHook(ctx context.Context, plugin output.Plugin, exec *pluginExecution) bool {
	if provider, ok := plugin.(hooks.Provider); ok {
		spec := provider.Hooks()
		hctx := hooks.Context{
			PluginName: plugin.Name(),
			Verbose:    generateVerbose,
			OutputDir:  plugin.DefaultOutputDir(),
		}
		skip, reason, err := hooks.RunPre(spec, hctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, " %s pre-execution check failed: %v\n", plugin.Name(), err)
			exec.skip = true
			exec.skipReason = fmt.Sprintf("pre-hook error: %v", err)
			return true
		}
		if skip {
			if generateVerbose {
				fmt.Fprintf(os.Stderr, "⊘ Skipping %s: %s\n", plugin.Name(), reason)
			}
			exec.skip = true
			exec.skipReason = reason
			return true
		}
	}

	preHook, ok := plugin.(output.PreExecuteHook)
	if !ok {
		return false
	}

	hookCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	skip, reason, err := preHook.PreExecute(hookCtx)
	cancel()

	if err != nil {
		fmt.Fprintf(os.Stderr, " %s pre-execution check failed: %v\n", plugin.Name(), err)
		exec.skip = true
		exec.skipReason = fmt.Sprintf("pre-hook error: %v", err)
		return true
	}

	if skip {
		if generateVerbose {
			fmt.Fprintf(os.Stderr, "⊘ Skipping %s: %s\n", plugin.Name(), reason)
		}
		exec.skip = true
		exec.skipReason = reason
		return true
	}

	return false
}

// generateAndWriteFiles generates files from plugins and writes them to disk.
func generateAndWriteFiles(executions []pluginExecution, palette, alternatePalette *colour.CategorisedPalette, wallpaperPath, wallpaperRawPath string) int {
	successCount := 0
	firstOutputPlugin := true

	for i := range executions {
		exec := &executions[i]
		if exec.skip {
			continue
		}

		if generateVerbose && firstOutputPlugin {
			fmt.Fprintf(os.Stderr, "→ Running output plugins...\n")
			firstOutputPlugin = false
		}

		if processPluginGeneration(exec, palette, alternatePalette, wallpaperPath, wallpaperRawPath) {
			successCount++
		}
	}

	return successCount
}

// processPluginGeneration generates and writes files for a single plugin.
func processPluginGeneration(exec *pluginExecution, palette, alternatePalette *colour.CategorisedPalette, wallpaperPath, wallpaperRawPath string) bool {
	plugin := exec.plugin

	if generateVerbose {
		fmt.Fprintf(os.Stderr, " Output plugin: %s\n", plugin.Name())
		fmt.Fprintf(os.Stderr, "   %s\n", plugin.Description())
	}

	var files map[string][]byte
	var err error

	// Check if plugin supports dual-theme generation
	if dualThemePlugin, ok := plugin.(output.DualThemePlugin); ok && alternatePalette != nil {
		// Generate both themes
		primaryThemeData := colour.NewThemeData(palette, wallpaperPath, wallpaperRawPath, "")
		alternateThemeData := colour.NewThemeData(alternatePalette, wallpaperPath, wallpaperRawPath, "")

		files, err = dualThemePlugin.GenerateDualTheme(primaryThemeData, alternateThemeData)
		if err != nil {
			fmt.Fprintf(os.Stderr, " %s failed (dual-theme): %v\n", plugin.Name(), err)
			exec.skip = true
			exec.skipReason = fmt.Sprintf("dual-theme generation failed: %v", err)
			return false
		}
	} else {
		// Standard single-theme generation
		themeData := colour.NewThemeData(palette, wallpaperPath, wallpaperRawPath, "")

		// If this is an external plugin and we have an alternate palette, set it
		if alternatePalette != nil {
			if extPlugin, ok := plugin.(*manager.ExternalOutputPlugin); ok {
				extPlugin.SetAlternatePalette(alternatePalette)
			}
		}

		// Generate files.
		files, err = plugin.Generate(themeData)
		if err != nil {
			fmt.Fprintf(os.Stderr, " %s failed: %v\n", plugin.Name(), err)
			exec.skip = true
			exec.skipReason = fmt.Sprintf("generation failed: %v", err)
			return false
		}
	}

	// Write files.
	if !writePluginFiles(exec, plugin, files) {
		return false
	}

	if generateVerbose {
		fmt.Fprintf(os.Stderr, "   Plugin execution complete.\n")
	}

	return true
}

// writePluginFiles writes generated files to disk.
func writePluginFiles(exec *pluginExecution, plugin output.Plugin, files map[string][]byte) bool { //nolint:gocyclo,gocognit// file write with backup, conflict handling, and verbose output
	outputDir := plugin.DefaultOutputDir()
	exec.writtenFiles = make([]string, 0, len(files))

	for filename, content := range files {
		// Resolve the output path.
		// Internal plugins return relative paths (joined with outputDir).
		// External plugins (outputDir=="") return absolute paths they manage themselves.
		var fullPath string
		if outputDir == "" {
			fullPath = filepath.Clean(filename)
		} else {
			fullPath = filepath.Clean(filepath.Join(outputDir, filename))
		}

		// nil content means the plugin already wrote the file (protocol >= 0.2.0).
		// Track it in the manifest but don't re-write it.
		if content == nil { //nolint:nestif
			if generateDryRun {
				fmt.Printf("   Plugin wrote: %s\n", fullPath)
			} else {
				fmt.Printf("   %s (plugin-managed)\n", fullPath)
				exec.writtenFiles = append(exec.writtenFiles, fullPath)

				// Record file in manifest for tracking (read content from disk for hashing).
				if generateManifestManager != nil {
					diskContent, err := os.ReadFile(fullPath)
					if err != nil {
						if generateVerbose {
							fmt.Fprintf(os.Stderr, "   Warning: failed to read %s for manifest: %v\n", fullPath, err)
						}
					} else {
						if err := generateManifestManager.RecordFile(plugin.Name(), fullPath, diskContent); err != nil {
							if generateVerbose {
								fmt.Fprintf(os.Stderr, "   Warning: failed to record %s in manifest: %v\n", fullPath, err)
							}
						}
					}
				}
			}
			continue
		}

		if generateDryRun { //nolint:nestif
			fmt.Printf("   Would write: %s (%d bytes)\n", fullPath, len(content))
		} else {
			// Check for untracked file protection.
			if generateManifestManager != nil && !generateForce {
				// Check if file exists on disk but isn't tracked.
				if _, err := os.Stat(fullPath); err == nil {
					if !generateManifestManager.IsTracked(fullPath) {
						fmt.Fprintf(os.Stderr, "   Skipping %s: file exists but is not tracked by tinct\n", fullPath)
						fmt.Fprintf(os.Stderr, "   Use --force to overwrite, or 'tinct files scan' to adopt existing files\n")
						continue
					}
				}
			}

			if err := writeFile(fullPath, content, generateVerbose); err != nil {
				fmt.Fprintf(os.Stderr, " Failed to write %s: %v\n", fullPath, err)
				exec.skip = true
				exec.skipReason = fmt.Sprintf("write failed: %v", err)
				return false
			}
			fmt.Printf("   %s (%d bytes)\n", fullPath, len(content))
			exec.writtenFiles = append(exec.writtenFiles, fullPath)

			// Record file in manifest for tracking.
			if generateManifestManager != nil {
				if err := generateManifestManager.RecordFile(plugin.Name(), fullPath, content); err != nil {
					if generateVerbose {
						fmt.Fprintf(os.Stderr, "   Warning: failed to record %s in manifest: %v\n", fullPath, err)
					}
				}
			}
		}
	}

	return true
}

// runPostExecutionHooks runs post-execute behaviour for successful plugins.
// The plugin's optional imperative PostExecuteHook runs first (so it can
// emit custom state-aware warnings or messages), then the declarative
// hooks.Spec runs (chmod, reload, wallpaper, instructions).
func runPostExecutionHooks(ctx context.Context, executions []pluginExecution, wallpaperPath string) {
	for _, exec := range executions {
		if exec.skip || len(exec.writtenFiles) == 0 {
			continue
		}

		plugin := exec.plugin
		execContext := output.ExecutionContext{
			DryRun:        generateDryRun,
			Verbose:       generateVerbose,
			OutputDir:     plugin.DefaultOutputDir(),
			WallpaperPath: wallpaperPath,
		}

		if postHook, ok := plugin.(output.PostExecuteHook); ok {
			if generateVerbose {
				fmt.Fprintf(os.Stderr, "→ Running %s post-hook...\n", plugin.Name())
			}
			hookCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			if err := postHook.PostExecute(hookCtx, execContext, exec.writtenFiles); err != nil && generateVerbose {
				// Post-hooks are best-effort (e.g. notifying a daemon that may
				// not be running) — surface failures only in verbose mode.
				fmt.Fprintf(os.Stderr, "   %s post-hook failed: %v\n", plugin.Name(), err)
			}
			cancel()
		}

		if provider, ok := plugin.(hooks.Provider); ok {
			spec := provider.Hooks()
			hctx := hooks.Context{
				PluginName:    plugin.Name(),
				DryRun:        execContext.DryRun,
				Verbose:       execContext.Verbose,
				OutputDir:     execContext.OutputDir,
				WallpaperPath: execContext.WallpaperPath,
				WrittenFiles:  exec.writtenFiles,
			}
			specCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			if err := hooks.RunPost(specCtx, spec, hctx); err != nil && generateVerbose {
				fmt.Fprintf(os.Stderr, "   %s spec post-hook failed: %v\n", plugin.Name(), err)
			}
			cancel()
		}
	}
}

// printGenerationSummary prints the final summary of the generation process.
func printGenerationSummary(successCount int, executions []pluginExecution) error {
	if generateDryRun {
		return nil
	}

	fmt.Println()
	if successCount > 0 {
		fmt.Printf(" Done! Generated %d output plugin(s)\n", successCount)
		return nil
	}

	// Collect failure reasons to help user understand what went wrong
	var failureReasons []string
	for _, exec := range executions {
		if exec.skip && exec.skipReason != "" {
			failureReasons = append(failureReasons, fmt.Sprintf("  %s: %s", exec.plugin.Name(), exec.skipReason))
		}
	}

	if len(failureReasons) > 0 {
		fmt.Fprintf(os.Stderr, "\nPlugin failures:\n")
		for _, reason := range failureReasons {
			fmt.Fprintf(os.Stderr, "%s\n", reason)
		}
	}

	return fmt.Errorf("no output plugins succeeded")
}
