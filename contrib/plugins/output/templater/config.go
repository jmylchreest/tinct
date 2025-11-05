package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the plugin configuration
type Config struct {
	Templates []TemplateConfig `yaml:"templates" json:"templates"`
	Settings  Settings         `yaml:"settings" json:"settings"`
}

// TemplateConfig represents a single template configuration
type TemplateConfig struct {
	Name         string `yaml:"name" json:"name"`
	Description  string `yaml:"description" json:"description"`
	TemplatePath string `yaml:"template_path" json:"template_path"`
	OutputPath   string `yaml:"output_path" json:"output_path"`
	Enabled      bool   `yaml:"enabled" json:"enabled"`
}

// Settings represents global plugin settings
type Settings struct {
	CreateDirs   bool   `yaml:"create_dirs" json:"create_dirs"`
	Backup       bool   `yaml:"backup" json:"backup"`
	BackupSuffix string `yaml:"backup_suffix" json:"backup_suffix"`
	Verbose      bool   `yaml:"verbose" json:"verbose"`
	FileMode     uint32 `yaml:"file_mode" json:"file_mode"`
}

// LoadConfig loads configuration from a file (YAML or JSON)
func LoadConfig(path string) (*Config, error) {
	// Expand path
	path = expandPath(path)

	// Read file
	data, err := os.ReadFile(path) // #nosec G304 - User-specified config file, intended to be read
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Determine format by extension
	ext := strings.ToLower(filepath.Ext(path))

	var config Config
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, &config); err != nil {
			if jsonErr := json.Unmarshal(data, &config); jsonErr != nil {
				return nil, fmt.Errorf("failed to parse config as YAML or JSON: %w", err)
			}
		}
	}

	// Apply defaults
	applyDefaults(&config)

	// Validate config
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// applyDefaults applies default values to config
func applyDefaults(config *Config) {
	// Default settings
	if config.Settings.BackupSuffix == "" {
		config.Settings.BackupSuffix = ".backup"
	}
	if config.Settings.FileMode == 0 {
		config.Settings.FileMode = 0644
	}

	// Expand paths in templates
	for i := range config.Templates {
		config.Templates[i].TemplatePath = expandPath(config.Templates[i].TemplatePath)
		config.Templates[i].OutputPath = expandPath(config.Templates[i].OutputPath)
	}
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	if len(config.Templates) == 0 {
		return fmt.Errorf("no templates defined in config")
	}

	for i, tmpl := range config.Templates {
		if tmpl.Name == "" {
			return fmt.Errorf("template %d: name is required", i)
		}
		if tmpl.TemplatePath == "" {
			return fmt.Errorf("template %q: template_path is required", tmpl.Name)
		}
		if tmpl.OutputPath == "" {
			return fmt.Errorf("template %q: output_path is required", tmpl.Name)
		}
	}

	return nil
}
