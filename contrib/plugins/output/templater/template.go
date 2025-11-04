package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// TemplateProcessor handles template processing
type TemplateProcessor struct {
	config  *Config
	verbose bool
	dryRun  bool
}

// NewTemplateProcessor creates a new template processor
func NewTemplateProcessor(config *Config, verbose, dryRun bool) *TemplateProcessor {
	return &TemplateProcessor{
		config:  config,
		verbose: verbose,
		dryRun:  dryRun,
	}
}

// ProcessTemplates processes all enabled templates
func (tp *TemplateProcessor) ProcessTemplates(themeData *ThemeData) ([]ProcessingResult, error) {
	var results []ProcessingResult

	for _, tmplConfig := range tp.config.Templates {
		if !tmplConfig.Enabled {
			if tp.verbose {
				fmt.Fprintf(os.Stderr, "  Skipping disabled template: %s\n", tmplConfig.Name)
			}
			continue
		}

		result := tp.processTemplate(tmplConfig, themeData)
		results = append(results, result)
	}

	return results, nil
}

// processTemplate processes a single template
func (tp *TemplateProcessor) processTemplate(tmplConfig TemplateConfig, themeData *ThemeData) ProcessingResult {
	result := ProcessingResult{
		TemplateName: tmplConfig.Name,
		OutputPath:   tmplConfig.OutputPath,
		Success:      false,
	}

	if tp.verbose {
		fmt.Fprintf(os.Stderr, "\nProcessing template: %s\n", tmplConfig.Name)
		fmt.Fprintf(os.Stderr, "  Description: %s\n", tmplConfig.Description)
		fmt.Fprintf(os.Stderr, "  Template: %s\n", tmplConfig.TemplatePath)
		fmt.Fprintf(os.Stderr, "  Output: %s\n", tmplConfig.OutputPath)
	}

	// Read template file
	tmplContent, err := os.ReadFile(tmplConfig.TemplatePath)
	if err != nil {
		result.Error = fmt.Errorf("failed to read template: %w", err)
		return result
	}

	// Parse template with custom functions
	tmpl, err := template.New(tmplConfig.Name).
		Funcs(templateFuncs()).
		Parse(string(tmplContent))
	if err != nil {
		result.Error = fmt.Errorf("failed to parse template: %w", err)
		return result
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		result.Error = fmt.Errorf("failed to execute template: %w", err)
		return result
	}

	output := buf.Bytes()
	result.BytesWritten = len(output)

	// In dry-run mode, don't write files
	if tp.dryRun {
		result.Success = true
		return result
	}

	// Create output directory if needed
	if tp.config.Settings.CreateDirs {
		outputDir := filepath.Dir(tmplConfig.OutputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			result.Error = fmt.Errorf("failed to create output directory: %w", err)
			return result
		}
	}

	// Backup existing file if needed
	if tp.config.Settings.Backup {
		if _, err := os.Stat(tmplConfig.OutputPath); err == nil {
			backupPath := tmplConfig.OutputPath + tp.config.Settings.BackupSuffix
			if err := os.Rename(tmplConfig.OutputPath, backupPath); err != nil {
				if tp.verbose {
					fmt.Fprintf(os.Stderr, "  Warning: failed to create backup: %v\n", err)
				}
			} else if tp.verbose {
				fmt.Fprintf(os.Stderr, "  Created backup: %s\n", backupPath)
			}
		}
	}

	// Write output file
	fileMode := os.FileMode(tp.config.Settings.FileMode)
	if err := os.WriteFile(tmplConfig.OutputPath, output, fileMode); err != nil {
		result.Error = fmt.Errorf("failed to write output: %w", err)
		return result
	}

	result.Success = true
	return result
}

// templateFuncs returns custom template functions compatible with Tinct templates
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		// get - Get color by role name
		"get": func(data interface{}, role string) (*ColorValue, error) {
			td, ok := data.(*ThemeData)
			if !ok {
				return nil, fmt.Errorf("get: expected *ThemeData, got %T", data)
			}
			color, ok := td.Colors[role]
			if !ok {
				return nil, fmt.Errorf("color role %q not found", role)
			}
			return color, nil
		},

		// has - Check if color role exists
		"has": func(data interface{}, role string) bool {
			td, ok := data.(*ThemeData)
			if !ok {
				return false
			}
			_, ok = td.Colors[role]
			return ok
		},

		// themeType - Get theme type string
		"themeType": func(data interface{}) string {
			td, ok := data.(*ThemeData)
			if !ok {
				return "unknown"
			}
			return td.ThemeType()
		},

		// seq - Generate sequence of integers
		"seq": func(start, end int) []int {
			if start > end {
				return nil
			}
			seq := make([]int, 0, end-start+1)
			for i := start; i <= end; i++ {
				seq = append(seq, i)
			}
			return seq
		},

		// ansi - Get ANSI color code (for terminal themes)
		"ansi": func(data interface{}, index int) (*ColorValue, error) {
			td, ok := data.(*ThemeData)
			if !ok {
				return nil, fmt.Errorf("ansi: expected *ThemeData, got %T", data)
			}
			if index < 0 || index >= len(td.AllColors) {
				return nil, fmt.Errorf("ansi: index %d out of range (0-%d)", index, len(td.AllColors)-1)
			}
			return td.AllColors[index], nil
		},

		// rgbSpaces - Format RGB values with spaces (for some config formats)
		"rgbSpaces": func(color *ColorValue) string {
			return fmt.Sprintf("%d %d %d", color.R(), color.G(), color.B())
		},
	}
}
