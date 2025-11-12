package repocli

import (
	"fmt"

	"github.com/jmylchreest/tinct/internal/repomanager"
	"github.com/spf13/cobra"
)

// RemoveCmd returns the remove command.
func RemoveCmd() *cobra.Command {
	var (
		pluginName   string
		version      string
		allVersions  bool
		manifestPath string
		dryRun       bool
	)

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove plugin or version from repository",
		Long: `Remove a plugin or specific version from the repository.

Examples:
  # Remove specific version
  tinct-repo-manager remove --plugin random --version 1.0.0

  # Remove all versions of a plugin
  tinct-repo-manager remove --plugin old-plugin --all-versions
`,
		RunE: func(_ *cobra.Command, _ []string) error {
			if pluginName == "" {
				return fmt.Errorf("--plugin is required")
			}
			if !allVersions && version == "" {
				return fmt.Errorf("either --version or --all-versions must be specified")
			}

			// Load manifest
			mgr, err := repomanager.LoadManifest(manifestPath)
			if err != nil {
				return fmt.Errorf("failed to load manifest: %w", err)
			}

			if allVersions {
				// Remove entire plugin
				fmt.Printf("Removing plugin '%s' (all versions)...\n", pluginName)
				if !dryRun {
					if err := mgr.RemovePlugin(pluginName); err != nil {
						return fmt.Errorf("failed to remove plugin: %w", err)
					}
				}
			} else {
				// Remove specific version
				fmt.Printf("Removing plugin '%s' version %s...\n", pluginName, version)
				if !dryRun {
					if err := mgr.RemovePluginVersion(pluginName, version); err != nil {
						return fmt.Errorf("failed to remove version: %w", err)
					}
				}
			}

			// Save manifest
			if !dryRun {
				if err := mgr.Save(); err != nil {
					return fmt.Errorf("failed to save manifest: %w", err)
				}
				fmt.Printf("✓ Manifest saved: %s\n", manifestPath)
			} else {
				fmt.Println("(Dry run - no changes saved)")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&pluginName, "plugin", "", "Plugin name")
	cmd.Flags().StringVar(&version, "version", "", "Version to remove")
	cmd.Flags().BoolVar(&allVersions, "all-versions", false, "Remove all versions")
	cmd.Flags().StringVar(&manifestPath, "manifest", "repository.json", "Path to manifest")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without saving")

	return cmd
}
