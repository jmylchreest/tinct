package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmylchreest/tinct/pkg/plugin"
)

// TemplaterPlugin implements the plugin.OutputPlugin interface
type TemplaterPlugin struct{}

// Generate processes templates and returns generated files
func (p *TemplaterPlugin) Generate(ctx context.Context, palette plugin.PaletteData) (map[string][]byte, error) {
	// Convert plugin.PaletteData to internal PaletteInput format
	input := convertProtocolPalette(palette)

	// Get config path from plugin args or use default
	configPath := getConfigPath(input.PluginArgs)

	// Load configuration
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("error loading config from %s: %w", configPath, err)
	}

	// Apply plugin args overrides
	applyPluginArgs(config, input.PluginArgs)

	// Set verbose mode from config or plugin args
	verbose := config.Settings.Verbose
	if v, ok := input.PluginArgs["verbose"].(bool); ok {
		verbose = v
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Templater v%s\n", Version)
		fmt.Fprintf(os.Stderr, "Config: %s\n", configPath)
		fmt.Fprintf(os.Stderr, "Templates: %d\n", len(config.Templates))
		fmt.Fprintf(os.Stderr, "Dry-run: %v\n\n", input.DryRun)
	}

	// Convert palette to ThemeData structure
	themeData := convertToThemeData(&input)

	// Process each enabled template
	processor := NewTemplateProcessor(config, verbose, input.DryRun)
	results, err := processor.ProcessTemplates(themeData)
	if err != nil {
		return nil, fmt.Errorf("error processing templates: %w", err)
	}

	// Convert results to file map
	files := make(map[string][]byte)
	for _, result := range results {
		if result.Success && !input.DryRun {
			// Expand output path
			outputPath, err := expandPath(result.OutputPath)
			if err != nil {
				return nil, fmt.Errorf("error expanding path %s: %w", result.OutputPath, err)
			}
			files[outputPath] = []byte(result.Content)
		}
	}

	// Output summary to stderr
	if len(results) == 0 {
		fmt.Fprintln(os.Stderr, "No templates processed (all disabled or skipped)")
	} else {
		fmt.Fprintf(os.Stderr, "\n✓ Processed %d template(s)\n", len(results))
		for _, result := range results {
			if result.Success {
				if input.DryRun {
					fmt.Fprintf(os.Stderr, "  Would write: %s (%d bytes)\n", result.OutputPath, result.BytesWritten)
				} else {
					fmt.Fprintf(os.Stderr, "  ✓ %s (%d bytes)\n", result.OutputPath, result.BytesWritten)
				}
			} else {
				fmt.Fprintf(os.Stderr, "  ✗ %s: %s\n", result.TemplateName, result.Error)
			}
		}
	}

	return files, nil
}

// PreExecute checks if config file exists
func (p *TemplaterPlugin) PreExecute(ctx context.Context) (bool, string, error) {
	// Check if default config exists
	configPath := getConfigPath(nil)
	expandedPath, err := expandPath(configPath)
	if err != nil {
		return false, "", fmt.Errorf("error expanding config path: %w", err)
	}

	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		msg := fmt.Sprintf("Config file not found: %s\nHint: Create a config file. See README.md for examples", expandedPath)
		return true, msg, nil
	}

	return false, "", nil
}

// PostExecute runs post-template commands if configured
func (p *TemplaterPlugin) PostExecute(ctx context.Context, writtenFiles []string) error {
	// Could implement post-commands here if needed
	// For now, templater handles this internally
	return nil
}

// GetMetadata returns plugin metadata
func (p *TemplaterPlugin) GetMetadata() plugin.PluginInfo {
	return plugin.PluginInfo{
		Name:            Name,
		Type:            "output",
		Version:         Version,
		ProtocolVersion: plugin.ProtocolVersion,
		Description:     "Templater for custom configuration files",
		PluginProtocol:  "go-plugin",
	}
}

// GetFlagHelp returns help information for plugin flags
func (p *TemplaterPlugin) GetFlagHelp() []plugin.FlagHelp {
	return []plugin.FlagHelp{
		{
			Name:        "config",
			Type:        "string",
			Default:     "~/.config/tinct/templater.yaml",
			Description: "Path to templater configuration file",
			Required:    false,
		},
		{
			Name:        "templates",
			Type:        "[]string",
			Default:     "",
			Description: "Filter to only process specified templates by name",
			Required:    false,
		},
		{
			Name:        "skip",
			Type:        "[]string",
			Default:     "",
			Description: "Skip specified templates by name",
			Required:    false,
		},
		{
			Name:        "verbose",
			Type:        "bool",
			Default:     "false",
			Description: "Enable verbose output",
			Required:    false,
		},
	}
}

// convertProtocolPalette converts plugin.PaletteData to internal PaletteInput
func convertProtocolPalette(palette plugin.PaletteData) PaletteInput {
	colours := make(map[string]CategorisedColour)
	for role, color := range palette.Colours {
		colours[role] = CategorisedColour{
			Hex:   color.Hex,
			Role:  role,
			Index: color.Index,
			RGB: RGB{
				R: color.RGB.R,
				G: color.RGB.G,
				B: color.RGB.B,
			},
		}
	}

	allColours := make([]CategorisedColour, len(palette.AllColours))
	for i, color := range palette.AllColours {
		allColours[i] = CategorisedColour{
			Hex:   color.Hex,
			Role:  color.Role,
			Index: i,
			RGB: RGB{
				R: color.RGB.R,
				G: color.RGB.G,
				B: color.RGB.B,
			},
		}
	}

	themeType := 0 // dark
	if palette.ThemeType == "light" {
		themeType = 1
	}

	return PaletteInput{
		Colours:    colours,
		AllColours: allColours,
		ThemeType:  themeType,
		PluginArgs: palette.PluginArgs,
		DryRun:     palette.DryRun,
	}
}

// expandPath expands ~ and environment variables in paths
func expandPath(path string) (string, error) {
	if len(path) == 0 {
		return path, nil
	}

	// Expand ~
	if path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(homeDir, path[1:])
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	return path, nil
}
