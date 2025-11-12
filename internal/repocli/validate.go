package repocli

import (
	"fmt"

	"github.com/jmylchreest/tinct/internal/repomanager"
	"github.com/spf13/cobra"
)

// ValidateCmd returns the validate command.
func ValidateCmd() *cobra.Command {
	var (
		manifestPath string
	)

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate repository manifest",
		Long: `Validate the repository manifest structure and contents.

Checks:
  - JSON syntax
  - Required fields
  - Plugin metadata completeness
  - Version format
  - Checksum format

Examples:
  tinct-repo-manager validate
  tinct-repo-manager validate --manifest /path/to/repository.json
`,
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Printf("Validating manifest: %s\n\n", manifestPath)

			// Load manifest (this validates JSON syntax)
			mgr, err := repomanager.LoadManifest(manifestPath)
			if err != nil {
				return fmt.Errorf("✗ Invalid manifest: %w", err)
			}

			manifest := mgr.GetManifest()

			var errors []string
			var warnings []string

			// Check manifest metadata
			if manifest.Name == "" {
				errors = append(errors, "manifest.name is required")
			}
			if manifest.Description == "" {
				warnings = append(warnings, "manifest.description is empty")
			}
			if manifest.URL == "" {
				warnings = append(warnings, "manifest.url is empty")
			}
			if manifest.Version == "" {
				errors = append(errors, "manifest.version is required")
			}

			// Check plugins
			if len(manifest.Plugins) == 0 {
				warnings = append(warnings, "no plugins in repository")
			}

			pluginCount := 0
			versionCount := 0

			for pluginName, plugin := range manifest.Plugins {
				pluginCount++

				// Check plugin metadata
				if plugin.Name == "" {
					errors = append(errors, fmt.Sprintf("plugin '%s': name is required", pluginName))
				}
				if plugin.Name != pluginName {
					errors = append(errors, fmt.Sprintf("plugin '%s': name mismatch (got '%s')", pluginName, plugin.Name))
				}
				if plugin.Type == "" {
					errors = append(errors, fmt.Sprintf("plugin '%s': type is required", pluginName))
				} else if plugin.Type != "input" && plugin.Type != "output" {
					errors = append(errors, fmt.Sprintf("plugin '%s': invalid type '%s' (must be 'input' or 'output')", pluginName, plugin.Type))
				}
				if plugin.Description == "" {
					warnings = append(warnings, fmt.Sprintf("plugin '%s': description is empty", pluginName))
				}

				// Check versions
				if len(plugin.Versions) == 0 {
					errors = append(errors, fmt.Sprintf("plugin '%s': no versions", pluginName))
				}

				for _, version := range plugin.Versions {
					versionCount++

					if version.Version == "" {
						errors = append(errors, fmt.Sprintf("plugin '%s': version string is required", pluginName))
					}

					// Check downloads
					if len(version.Downloads) == 0 {
						errors = append(errors, fmt.Sprintf("plugin '%s' version %s: no downloads", pluginName, version.Version))
					}

					for platform, download := range version.Downloads {
						if download.URL == "" {
							errors = append(errors, fmt.Sprintf("plugin '%s' version %s (%s): URL is required", pluginName, version.Version, platform))
						}
						if download.Checksum == "" {
							warnings = append(warnings, fmt.Sprintf("plugin '%s' version %s (%s): checksum is empty", pluginName, version.Version, platform))
						}
						if download.Size == 0 {
							warnings = append(warnings, fmt.Sprintf("plugin '%s' version %s (%s): size is 0", pluginName, version.Version, platform))
						}
					}
				}
			}

			// Print results
			fmt.Printf("Plugins: %d\n", pluginCount)
			fmt.Printf("Versions: %d\n", versionCount)
			fmt.Println()

			if len(warnings) > 0 {
				fmt.Printf("Warnings (%d):\n", len(warnings))
				for _, w := range warnings {
					fmt.Printf("  ⚠ %s\n", w)
				}
				fmt.Println()
			}

			if len(errors) > 0 {
				fmt.Printf("Errors (%d):\n", len(errors))
				for _, e := range errors {
					fmt.Printf("  ✗ %s\n", e)
				}
				fmt.Println()
				return fmt.Errorf("validation failed with %d error(s)", len(errors))
			}

			fmt.Println("✓ Manifest is valid")
			return nil
		},
	}

	cmd.Flags().StringVar(&manifestPath, "manifest", "repository.json", "Path to manifest")

	return cmd
}
