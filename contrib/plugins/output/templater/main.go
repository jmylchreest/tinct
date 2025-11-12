package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	goplug "github.com/hashicorp/go-plugin"
	"github.com/jmylchreest/tinct/pkg/plugin"
)

const (
	Version = "0.0.1"
	Name    = "templater"
)

// PluginInfo represents the plugin metadata for Tinct discovery
type PluginInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Author      string `json:"author"`
}

func main() {
	// Handle --plugin-info flag for Tinct discovery
	if len(os.Args) > 1 && os.Args[1] == "--plugin-info" {
		p := &TemplaterPlugin{}
		info := p.GetMetadata()
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(info); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding plugin info: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// If no arguments, run as go-plugin server (Tinct integration)
	if len(os.Args) == 1 {
		goplug.Serve(&goplug.ServeConfig{
			HandshakeConfig: plugin.Handshake,
			Plugins: map[string]goplug.Plugin{
				"output": &plugin.OutputPluginRPC{
					Impl: &TemplaterPlugin{},
				},
			},
		})
		return
	}

	// Legacy JSON-stdio mode (deprecated, kept for backward compatibility)
	runLegacyMode()
}

func runLegacyMode() {

	// Read palette data from stdin
	var input PaletteInput
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding palette input: %v\n", err)
		os.Exit(1)
	}

	// Get config path from plugin args or use default
	configPath := getConfigPath(input.PluginArgs)

	// Load configuration
	config, err := LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config from %s: %v\n", configPath, err)
		fmt.Fprintf(os.Stderr, "Hint: Create a config file at %s\n", configPath)
		fmt.Fprintf(os.Stderr, "See README.md for configuration examples\n")
		os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "Error processing templates: %v\n", err)
		os.Exit(1)
	}

	// Output results
	if len(results) == 0 {
		fmt.Println("No templates processed (all disabled or skipped)")
		os.Exit(0)
	}

	fmt.Printf("\n✓ Processed %d template(s)\n", len(results))
	for _, result := range results {
		if result.Success {
			if input.DryRun {
				fmt.Printf("  Would write: %s (%d bytes)\n", result.OutputPath, result.BytesWritten)
			} else {
				fmt.Printf("  Generated: %s (%d bytes)\n", result.OutputPath, result.BytesWritten)
			}
		} else {
			fmt.Printf("  ✗ Failed: %s - %v\n", result.TemplateName, result.Error)
		}
	}
}

// getConfigPath returns the config file path from plugin args or default location
func getConfigPath(args map[string]interface{}) string {
	// Check plugin args first
	if configPath, ok := args["config"].(string); ok && configPath != "" {
		return expandPathSimple(configPath)
	}

	// Try default locations in order
	defaults := []string{
		"~/.config/tinct/templater.yaml",
		"~/.config/tinct/templater.yml",
		"~/.config/tinct/templater.json",
	}

	for _, path := range defaults {
		expanded := expandPathSimple(path)
		if _, err := os.Stat(expanded); err == nil {
			return expanded
		}
	}

	// Return the first default even if it doesn't exist
	// (error will be handled by LoadConfig)
	return expandPathSimple(defaults[0])
}

// expandPathSimple expands ~ to home directory (JSON-stdio mode)
func expandPathSimple(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

// applyPluginArgs applies plugin arguments to override config settings
func applyPluginArgs(config *Config, args map[string]interface{}) {
	// Filter templates by name if specified
	if templates, ok := args["templates"].([]interface{}); ok {
		enabled := make(map[string]bool)
		for _, t := range templates {
			if name, ok := t.(string); ok {
				enabled[name] = true
			}
		}
		// Disable templates not in the list
		for i := range config.Templates {
			if !enabled[config.Templates[i].Name] {
				config.Templates[i].Enabled = false
			}
		}
	}

	// Skip specific templates if specified
	if skip, ok := args["skip"].([]interface{}); ok {
		skipMap := make(map[string]bool)
		for _, s := range skip {
			if name, ok := s.(string); ok {
				skipMap[name] = true
			}
		}
		// Disable skipped templates
		for i := range config.Templates {
			if skipMap[config.Templates[i].Name] {
				config.Templates[i].Enabled = false
			}
		}
	}
}
