// Package template provides template utilities and functions for tinct plugins.
package template

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmylchreest/tinct/pkg/util/semver"
)

// Loader handles loading templates with support for custom overrides and versioned templates.
// It checks for custom templates in ~/.config/tinct/templates/{pluginName}/
// and falls back to embedded templates if custom ones don't exist.
//
// Versioned templates allow plugins to support multiple versions of target applications.
// Templates can be organized in version subdirectories (e.g., templates/0.53/).
// The loader selects the highest version that doesn't exceed the target version.
type Loader struct {
	pluginName    string
	embedFS       embed.FS
	customBase    string // Base directory for custom templates
	verbose       bool   // Enable verbose logging
	logger        Logger // Logger for verbose output
	targetVersion string // Target application version for versioned templates
}

// Logger is a simple interface for logging messages.
type Logger interface {
	Printf(format string, v ...any)
}

// New creates a new template loader for the specified plugin.
// embedFS should be the embedded filesystem containing the plugin's default templates.
// pluginName is used to locate custom templates in ~/.config/tinct/templates/{pluginName}/.
func New(pluginName string, embedFS embed.FS) *Loader {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "" // Fallback to empty if home dir unavailable
	}
	customBase := filepath.Join(home, ".config", "tinct", "templates")

	return &Loader{
		pluginName: pluginName,
		embedFS:    embedFS,
		customBase: customBase,
		verbose:    false,
		logger:     nil,
	}
}

// WithCustomBase sets a custom base directory for template storage.
// This is useful for dumping templates to a non-default location.
func (l *Loader) WithCustomBase(customBase string) *Loader {
	l.customBase = customBase
	return l
}

// WithVerbose enables verbose logging for template operations.
func (l *Loader) WithVerbose(verbose bool, logger Logger) *Loader {
	l.verbose = verbose
	l.logger = logger
	return l
}

// WithTargetVersion sets the target application version for versioned template selection.
// When set, the loader will search for version-specific templates in subdirectories
// like templates/0.53/ and select the highest version that doesn't exceed targetVersion.
func (l *Loader) WithTargetVersion(version string) *Loader {
	l.targetVersion = version
	return l
}

// Load reads a template file, checking for custom overrides and versioned templates.
// filename should be the template filename (e.g., "theme.conf.tmpl").
// Returns the template content and whether it was loaded from a custom override.
//
// Template resolution order:
//  1. Custom template in ~/.config/tinct/templates/{pluginName}/{filename}
//  2. Versioned template in templates/{version}/{filename} (if targetVersion is set)
//  3. Default embedded template at {filename}
func (l *Loader) Load(filename string) (content []byte, fromCustom bool, err error) {
	// Try custom template first (highest priority).
	customPath := filepath.Join(l.customBase, l.pluginName, filename)

	if content, err := os.ReadFile(customPath); err == nil { // #nosec G304 - User-specified template directory, validated by caller
		if l.verbose && l.logger != nil {
			l.logger.Printf("   Using custom template: %s", customPath)
		}
		return content, true, nil
	}

	// Try versioned template if targetVersion is set.
	if l.targetVersion != "" {
		versionedContent, versionUsed, vErr := l.loadVersionedTemplate(filename)
		if vErr == nil && versionedContent != nil {
			if l.verbose && l.logger != nil {
				l.logger.Printf("   Using versioned template: templates/%s/%s (target: %s)",
					versionUsed, filename, l.targetVersion)
			}
			return versionedContent, false, nil
		}
		if l.verbose && l.logger != nil && vErr != nil {
			l.logger.Printf("   Versioned template not available for %s: %v", filename, vErr)
		}
	}

	// Fall back to embedded template.
	if l.verbose && l.logger != nil {
		l.logger.Printf("   Using default embedded template: %s", filename)
	}

	content, err = l.embedFS.ReadFile(filename)
	if err != nil {
		return nil, false, fmt.Errorf("failed to load template %q: %w", filename, err)
	}

	return content, false, nil
}

// loadVersionedTemplate attempts to load a template from a version subdirectory.
// It finds all version directories (e.g., templates/0.53/), selects the highest
// version that doesn't exceed targetVersion, and loads the template from there.
// Returns the content, the version used, and any error.
func (l *Loader) loadVersionedTemplate(filename string) ([]byte, string, error) {
	versions, err := l.findVersionDirectories()
	if err != nil || len(versions) == 0 {
		return nil, "", err
	}

	bestVersion := semver.FindBestMatch(l.targetVersion, versions)
	if bestVersion == "" {
		return nil, "", nil // No suitable version found
	}

	versionedPath := filepath.Join("templates", bestVersion, filename)
	content, err := l.embedFS.ReadFile(versionedPath)
	if err != nil {
		return nil, "", nil // Template doesn't exist in this version directory
	}

	return content, bestVersion, nil
}

// findVersionDirectories scans the embedded filesystem for version directories
// under templates/. Returns a list of version strings (e.g., ["0.52", "0.53"]).
func (l *Loader) findVersionDirectories() ([]string, error) {
	var versions []string

	entries, err := fs.ReadDir(l.embedFS, "templates")
	if err != nil {
		return nil, nil // templates directory doesn't exist — that's fine
	}

	for _, entry := range entries {
		if entry.IsDir() {
			v := semver.Parse(entry.Name())
			if v != nil {
				versions = append(versions, entry.Name())
			}
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		return semver.CompareStrings(versions[i], versions[j]) < 0
	})

	return versions, nil
}

// CustomPath returns the path where a custom template would be located.
func (l *Loader) CustomPath(filename string) string {
	return filepath.Join(l.customBase, l.pluginName, filename)
}

// CustomDir returns the directory where custom templates for this plugin would be located.
func (l *Loader) CustomDir() string {
	return filepath.Join(l.customBase, l.pluginName)
}

// HasCustomTemplate checks if a custom template exists for the given filename.
func (l *Loader) HasCustomTemplate(filename string) bool {
	customPath := l.CustomPath(filename)
	_, err := os.Stat(customPath)
	return err == nil
}

// ListEmbeddedTemplates returns a list of all embedded template files.
func (l *Loader) ListEmbeddedTemplates() ([]string, error) {
	var templates []string

	err := fs.WalkDir(l.embedFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".tmpl" {
			templates = append(templates, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list embedded templates: %w", err)
	}

	return templates, nil
}

// DumpTemplate writes an embedded template to the custom templates directory.
// If force is false, it will not overwrite existing custom templates.
func (l *Loader) DumpTemplate(filename string, force bool) error {
	// Read embedded template.
	content, err := l.embedFS.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read embedded template %q: %w", filename, err)
	}

	// Determine output path.
	outputPath := l.CustomPath(filename)

	// Check if file already exists.
	if !force {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("custom template already exists: %s (use --force to overwrite)", outputPath)
		}
	}

	// Create directory if it doesn't exist.
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", outputDir, err)
	}

	// Write file.
	if err := os.WriteFile(outputPath, content, 0o600); err != nil {
		return fmt.Errorf("failed to write template to %q: %w", outputPath, err)
	}

	return nil
}

// DumpAllTemplates writes all embedded templates to the custom templates directory.
// Returns the list of successfully dumped templates and any errors encountered.
// If force is false and some templates already exist, it will skip those but continue
// processing remaining templates.
func (l *Loader) DumpAllTemplates(force bool) ([]string, error) {
	templates, err := l.ListEmbeddedTemplates()
	if err != nil {
		return nil, err
	}

	dumped := make([]string, 0, len(templates))
	var errors []string

	for _, tmpl := range templates {
		if err := l.DumpTemplate(tmpl, force); err != nil {
			// If it's an "already exists" error and force is false, collect the error but continue.
			if !force && strings.Contains(err.Error(), "already exists") {
				errors = append(errors, err.Error())
				continue
			}
			// For other errors, return immediately.
			return dumped, err
		}
		dumped = append(dumped, l.CustomPath(tmpl))
	}

	// If we have collected errors (skipped files), return them as a combined error.
	if len(errors) > 0 {
		return dumped, fmt.Errorf("%s", strings.Join(errors, "; "))
	}

	return dumped, nil
}

// Info contains information about a template.
type Info struct {
	Filename       string
	EmbeddedExists bool
	CustomExists   bool
	CustomPath     string
	UsingCustom    bool
}

// GetInfo returns information about a specific template.
func (l *Loader) GetInfo(filename string) Info {
	_, embeddedErr := l.embedFS.ReadFile(filename)
	customExists := l.HasCustomTemplate(filename)

	return Info{
		Filename:       filename,
		EmbeddedExists: embeddedErr == nil,
		CustomExists:   customExists,
		CustomPath:     l.CustomPath(filename),
		UsingCustom:    customExists,
	}
}
