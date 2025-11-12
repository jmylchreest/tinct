package repocli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmylchreest/tinct/internal/plugin/repository"
	"github.com/jmylchreest/tinct/internal/repomanager"
	"github.com/spf13/cobra"
)

// AddCmd returns the add command.
func AddCmd() *cobra.Command {
	var (
		pluginName    string
		pluginType    string
		pluginVersion string
		filePath      string
		url           string
		platform      string
		manifestPath  string
		skipQuery     bool
		dryRun        bool
		runtime       string
		dependencies  []string
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add plugin manually to repository",
		Long: `Add a plugin version manually to the repository.

Can add from:
  - Local file (script or binary)
  - URL (for remote plugins)

Examples:
  # Add script plugin from file
  tinct-repo-manager add --plugin notify-send --type output \
    --file contrib/plugins/output/tinct-plugin-notify-send.py \
    --platform any --runtime python3 --version 0.0.1

  # Add compiled plugin from URL
  tinct-repo-manager add --plugin random --type input \
    --url "https://github.com/user/repo/releases/download/v1.0.0/plugin.tar.gz" \
    --platform linux_amd64 --version 1.0.0
`,
		RunE: func(_ *cobra.Command, _ []string) error {
			// Validate inputs
			if filePath == "" && url == "" {
				return fmt.Errorf("either --file or --url must be specified")
			}
			if filePath != "" && url != "" {
				return fmt.Errorf("cannot specify both --file and --url")
			}
			if pluginName == "" {
				return fmt.Errorf("--plugin is required")
			}
			if pluginType == "" {
				return fmt.Errorf("--type is required")
			}
			if platform == "" {
				return fmt.Errorf("--platform is required")
			}
			if pluginVersion == "" {
				return fmt.Errorf("--version is required")
			}

			// Load manifest
			mgr, err := repomanager.LoadManifest(manifestPath)
			if err != nil {
				return fmt.Errorf("failed to load manifest: %w", err)
			}

			var (
				downloadURL string
				checksum    string
				size        int64
				metadata    *repomanager.PluginMetadata
			)

			if filePath != "" {
				// Handle local file
				absPath, err := filepath.Abs(filePath)
				if err != nil {
					return fmt.Errorf("invalid file path: %w", err)
				}

				// Check file exists
				info, err := os.Stat(absPath)
				if err != nil {
					return fmt.Errorf("file not found: %w", err)
				}
				size = info.Size()

				// For local files, we'll use file:// URL
				downloadURL = "file://" + absPath

				// TODO: Calculate checksum for local files
				fmt.Println("Warning: Checksum calculation for local files not yet implemented")

				// Query plugin if not skipped
				if !skipQuery {
					fmt.Println("Querying plugin metadata...")
					metadata, err = repomanager.QueryScriptPlugin(absPath)
					if err != nil {
						fmt.Printf("Warning: query failed: %v\n", err)
					} else {
						fmt.Printf("  Name: %s\n", metadata.Name)
						fmt.Printf("  Type: %s\n", metadata.Type)
						fmt.Printf("  Description: %s\n", metadata.Description)
					}
				}
			} else {
				// Handle URL
				downloadURL = url

				fmt.Println("Calculating checksum...")
				checksum, size, err = repomanager.CalculateChecksum(url)
				if err != nil {
					return fmt.Errorf("failed to calculate checksum: %w", err)
				}
				fmt.Printf("  Checksum: sha256:%s\n", checksum)
				fmt.Printf("  Size: %d bytes\n", size)

				// Query plugin if not skipped
				if !skipQuery {
					fmt.Println("Querying plugin metadata...")
					metadata, err = repomanager.QueryPlugin(url)
					if err != nil {
						fmt.Printf("Warning: query failed: %v\n", err)
					} else {
						fmt.Printf("  Name: %s\n", metadata.Name)
						fmt.Printf("  Type: %s\n", metadata.Type)
						fmt.Printf("  Description: %s\n", metadata.Description)
					}
				}
			}

			// Create download entry
			download := &repository.Download{
				URL:       downloadURL,
				Checksum:  fmt.Sprintf("sha256:%s", checksum),
				Size:      size,
				Available: true,
			}

			if runtime != "" {
				download.Runtime = runtime
			}
			if len(dependencies) > 0 {
				download.Dependencies = dependencies
			}

			// Create version entry
			version := &repository.Version{
				Version: pluginVersion,
				Downloads: map[string]*repository.Download{
					platform: download,
				},
			}

			// Add compatibility if we have metadata
			if metadata != nil {
				version.Compatibility = repomanager.CalculateCompatibility(metadata.ProtocolVersion)
			}

			// Add to manifest
			if !dryRun {
				if err := mgr.AddOrUpdatePluginVersion(pluginName, version); err != nil {
					return fmt.Errorf("failed to add plugin: %w", err)
				}

				// Set plugin metadata
				if metadata != nil {
					mgr.SetPluginMetadata(pluginName, metadata)
				} else {
					// Use provided values
					mgr.SetPluginMetadata(pluginName, &repomanager.PluginMetadata{
						Type: pluginType,
					})
				}

				// Save manifest
				if err := mgr.Save(); err != nil {
					return fmt.Errorf("failed to save manifest: %w", err)
				}

				fmt.Printf("\n✓ Added %s %s (%s) to manifest\n", pluginName, pluginVersion, platform)
				fmt.Printf("✓ Manifest saved: %s\n", manifestPath)
			} else {
				fmt.Println("\n(Dry run - no changes saved)")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&pluginName, "plugin", "", "Plugin name")
	cmd.Flags().StringVar(&pluginType, "type", "", "Plugin type (input/output)")
	cmd.Flags().StringVar(&pluginVersion, "version", "", "Plugin version")
	cmd.Flags().StringVar(&filePath, "file", "", "Local file path")
	cmd.Flags().StringVar(&url, "url", "", "Download URL")
	cmd.Flags().StringVar(&platform, "platform", "", "Platform (e.g., linux_amd64, any)")
	cmd.Flags().StringVar(&manifestPath, "manifest", "repository.json", "Path to manifest")
	cmd.Flags().BoolVar(&skipQuery, "skip-query", false, "Skip querying plugin metadata")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without saving")
	cmd.Flags().StringVar(&runtime, "runtime", "", "Runtime requirement (e.g., python3, bash)")
	cmd.Flags().StringSliceVar(&dependencies, "dependencies", []string{}, "Runtime dependencies")

	return cmd
}
