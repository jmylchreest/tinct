package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

// runPlugin runs in Tinct plugin mode (legacy JSON-stdio, deprecated)
func runPlugin() error {
	// Read palette from stdin (JSON)
	var palette map[string]interface{}
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&palette); err != nil {
		return fmt.Errorf("failed to decode palette: %w", err)
	}

	// Generate wob theme
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	themesDir := filepath.Join(homeDir, ".config", "wob", "themes")
	if err := os.MkdirAll(themesDir, 0755); err != nil { // #nosec G301 - Themes directory needs standard permissions
		return fmt.Errorf("failed to create themes directory: %w", err)
	}

	themeFile := filepath.Join(themesDir, "tinct.ini")

	// Generate theme content
	themeContent, err := generateWobThemeFromMap(palette)
	if err != nil {
		return fmt.Errorf("failed to generate theme: %w", err)
	}

	if err := os.WriteFile(themeFile, []byte(themeContent), 0644); err != nil { // #nosec G306 - Theme file needs standard read permissions
		return fmt.Errorf("failed to write theme file: %w", err)
	}

	// Install wrapper (copy self)
	scriptsDir := filepath.Join(homeDir, ".config", "wob", "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil { // #nosec G301 - Scripts directory needs standard permissions
		return fmt.Errorf("failed to create scripts directory: %w", err)
	}

	wrapperPath := filepath.Join(scriptsDir, "wob-tinct")
	selfPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Copy self to wrapper location
	if err := copyFile(selfPath, wrapperPath); err != nil {
		return fmt.Errorf("failed to install wrapper: %w", err)
	}

	if err := os.Chmod(wrapperPath, 0755); err != nil { // #nosec G302 - Wrapper executable needs execute permission
		return fmt.Errorf("failed to chmod wrapper: %w", err)
	}

	// Output success message with instructions
	fmt.Fprintf(os.Stderr, "\nGenerated wob theme: %s\n", themeFile)
	fmt.Fprintf(os.Stderr, "Installed wrapper: %s\n\n", wrapperPath)
	fmt.Fprintf(os.Stderr, "To use with Hyprland, add to your hyprland.conf:\n\n")
	fmt.Fprintf(os.Stderr, "  exec-once = %s start --base-config ~/.config/wob/base.ini \\\n", wrapperPath)
	fmt.Fprintf(os.Stderr, "                       --append-config ~/.config/wob/themes/tinct.ini\n\n")
	fmt.Fprintf(os.Stderr, "Then bind keys to send values:\n\n")
	fmt.Fprintf(os.Stderr, "  bind = , XF86AudioRaiseVolume, exec, wpctl set-volume @DEFAULT_SINK@ 5%%+ && \\\n")
	fmt.Fprintf(os.Stderr, "         %s send $(wpctl get-volume @DEFAULT_SINK@ | awk '{print $2 * 100}')\n\n", wrapperPath)

	return nil
}

// generateWobThemeFromMap creates wob INI content from palette map (JSON-stdio mode)
func generateWobThemeFromMap(palette map[string]interface{}) (string, error) {
	// Load template from embedded filesystem
	tmplContent, err := templatesFS.ReadFile("templates/tinct.ini.tmpl")
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	// Helper to get hex color from nested palette structure
	getColor := func(key string) string {
		// Access palette.colours[key].hex
		if colours, ok := palette["colours"].(map[string]interface{}); ok {
			if colorObj, ok := colours[key].(map[string]interface{}); ok {
				if hex, ok := colorObj["hex"].(string); ok {
					return strings.TrimPrefix(hex, "#")
				}
			}
		}
		return "000000"
	}

	// Create a simple color map for the template
	colorMap := map[string]string{
		"background": getColor("background"),
		"foreground": getColor("foreground"),
		"accent1":    getColor("accent1"),
		"success":    getColor("success"),
		"warning":    getColor("warning"),
		"danger":     getColor("danger"),
	}

	// Parse and execute template
	tmpl, err := template.New("wob").Funcs(template.FuncMap{
		"get": func(m map[string]string, key string) string {
			if color, ok := m[key]; ok {
				return color
			}
			return "000000"
		},
	}).Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, colorMap); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	source, err := os.Open(src) // #nosec G304 - User-specified source file, intended to be read
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst) // #nosec G304 - User-specified destination file
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
