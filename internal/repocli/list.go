package repocli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jmylchreest/tinct/internal/repomanager"
	"github.com/spf13/cobra"
)

// ListCmd returns the list command.
func ListCmd() *cobra.Command {
	var (
		manifestPath  string
		availableOnly bool
		format        string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List plugins in repository",
		Long: `List all plugins and versions in the repository.

Examples:
  # List all plugins
  tinct-repo-manager list

  # List only available versions
  tinct-repo-manager list --available-only

  # Output as JSON
  tinct-repo-manager list --format json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load manifest
			mgr, err := repomanager.LoadManifest(manifestPath)
			if err != nil {
				return fmt.Errorf("failed to load manifest: %w", err)
			}

			manifest := mgr.GetManifest()

			if format == "json" {
				// TODO: Implement JSON output
				return fmt.Errorf("JSON format not yet implemented")
			}

			// Table format
			fmt.Printf("Repository: %s\n", manifest.Name)
			if manifest.Description != "" {
				fmt.Printf("Description: %s\n", manifest.Description)
			}
			fmt.Printf("Plugins: %d\n", len(manifest.Plugins))
			fmt.Println()

			// Sort plugin names
			pluginNames := make([]string, 0, len(manifest.Plugins))
			for name := range manifest.Plugins {
				pluginNames = append(pluginNames, name)
			}
			sort.Strings(pluginNames)

			// Print table header
			fmt.Printf("%-20s %-8s %-10s %-20s %s\n", "PLUGIN", "TYPE", "VERSION", "PLATFORMS", "STATUS")
			fmt.Println(strings.Repeat("-", 80))

			// Print each plugin
			for _, pluginName := range pluginNames {
				plugin := manifest.Plugins[pluginName]

				// Sort versions (newest first, assuming they're already sorted)
				for _, version := range plugin.Versions {
					// Collect platforms
					platforms := make([]string, 0, len(version.Downloads))
					allAvailable := true
					someAvailable := false

					for platform, download := range version.Downloads {
						if availableOnly && !download.Available {
							continue
						}
						platforms = append(platforms, platform)
						if download.Available {
							someAvailable = true
						} else {
							allAvailable = false
						}
					}

					if availableOnly && len(platforms) == 0 {
						continue
					}

					sort.Strings(platforms)
					platformStr := strings.Join(platforms, ", ")
					if len(platformStr) > 20 {
						platformStr = platformStr[:17] + "..."
					}

					// Determine status
					status := "✓"
					if !allAvailable && someAvailable {
						status = "⚠"
					} else if !someAvailable {
						status = "✗"
					}

					fmt.Printf("%-20s %-8s %-10s %-20s %s\n",
						pluginName,
						plugin.Type,
						version.Version,
						platformStr,
						status)
				}
			}

			fmt.Println()
			fmt.Println("Status: ✓ = all available, ⚠ = partial, ✗ = unavailable")

			return nil
		},
	}

	cmd.Flags().StringVar(&manifestPath, "manifest", "repository.json", "Path to manifest")
	cmd.Flags().BoolVar(&availableOnly, "available-only", false, "Show only available versions")
	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json)")

	return cmd
}
