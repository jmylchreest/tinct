package repocli

import (
	"fmt"
	"time"

	"github.com/jmylchreest/tinct/internal/repomanager"
	"github.com/spf13/cobra"
)

// PruneCmd returns the prune command.
func PruneCmd() *cobra.Command {
	var (
		manifestPath    string
		verifyAll       bool
		markUnavailable bool
		removeAfter     string
		dryRun          bool
		verbose         bool
	)

	cmd := &cobra.Command{
		Use:        "prune",
		Short:      "Verify URLs and prune unavailable entries",
		Deprecated: "Use 'sync --prune' instead. Prune is now integrated into the sync command.",
		Long: `Verify download URLs and prune unavailable entries.

DEPRECATED: This command is deprecated. Use 'sync --prune' instead.
Pruning functionality is now integrated into the sync command for convenience.

Modes:
  - Mark unavailable: Set available=false for dead URLs
  - Remove after period: Remove entries unavailable for X time

Examples:
  # Use sync with pruning instead (recommended)
  tinct-repo-manager sync --github owner/repo --version latest \
    --plugin-filter "plugin-*" --prune --prune-remove-after 720h

  # Verify all URLs and mark unavailable (deprecated)
  tinct-repo-manager prune --verify-all --mark-unavailable

  # Remove entries unavailable for 30+ days (deprecated)
  tinct-repo-manager prune --remove-after 720h

  # Dry run to preview changes (deprecated)
  tinct-repo-manager prune --verify-all --dry-run
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load manifest
			mgr, err := repomanager.LoadManifest(manifestPath)
			if err != nil {
				return fmt.Errorf("failed to load manifest: %w", err)
			}

			manifest := mgr.GetManifest()
			verifier := repomanager.NewVerifier()

			var (
				checked     int
				unavailable int
				removed     int
			)

			// Parse remove-after duration
			var removeAfterDuration time.Duration
			if removeAfter != "" {
				removeAfterDuration, err = time.ParseDuration(removeAfter)
				if err != nil {
					return fmt.Errorf("invalid duration: %w", err)
				}
			}

			fmt.Println("Pruning repository...")
			if dryRun {
				fmt.Println("(Dry run mode)")
			}
			fmt.Println()

			// Iterate through all plugins
			for pluginName, plugin := range manifest.Plugins {
				if verbose {
					fmt.Printf("Checking plugin: %s\n", pluginName)
				}

				for vi, version := range plugin.Versions {
					for platform, download := range version.Downloads {
						checked++

						if verbose {
							fmt.Printf("  %s %s (%s): ", pluginName, version.Version, platform)
						}

						// Check if should remove (already unavailable and past threshold)
						if removeAfter != "" && download.UnavailableSince != nil {
							unavailableDuration := time.Since(*download.UnavailableSince)
							if unavailableDuration > removeAfterDuration {
								fmt.Printf("  Removing: %s %s (%s) - unavailable for %v\n",
									pluginName, version.Version, platform, unavailableDuration.Round(time.Hour))

								if !dryRun {
									delete(version.Downloads, platform)
									removed++
								}
								continue
							}
						}

						// Verify URL if requested
						if verifyAll || download.Available {
							available, reason := verifier.VerifyURL(download.URL)

							if !available {
								unavailable++
								fmt.Printf("  ✗ Unavailable: %s %s (%s) - %s\n",
									pluginName, version.Version, platform, reason)

								if markUnavailable && !dryRun {
									download.Available = false
									now := time.Now()
									download.UnavailableSince = &now
									download.UnavailableReason = reason
								}
							} else {
								if verbose {
									fmt.Println("✓")
								}
								if !dryRun {
									download.Available = true
									now := time.Now()
									download.LastVerified = &now
									download.UnavailableSince = nil
									download.UnavailableReason = ""
								}
							}
						}
					}

					// Clean up versions with no downloads
					if !dryRun && len(version.Downloads) == 0 {
						plugin.Versions = append(plugin.Versions[:vi], plugin.Versions[vi+1:]...)
						fmt.Printf("  Removed version %s (no downloads left)\n", version.Version)
					}
				}

				// Clean up plugins with no versions
				if !dryRun && len(plugin.Versions) == 0 {
					delete(manifest.Plugins, pluginName)
					fmt.Printf("  Removed plugin %s (no versions left)\n", pluginName)
				}
			}

			fmt.Printf("\n=== Summary ===\n")
			fmt.Printf("Checked: %d\n", checked)
			fmt.Printf("Unavailable: %d\n", unavailable)
			if removed > 0 {
				fmt.Printf("Removed: %d\n", removed)
			}

			// Save manifest
			if !dryRun && (unavailable > 0 || removed > 0) {
				now := time.Now()
				manifest.LastPruned = &now
				if err := mgr.Save(); err != nil {
					return fmt.Errorf("failed to save manifest: %w", err)
				}
				fmt.Printf("\n✓ Manifest saved: %s\n", manifestPath)
			} else if dryRun {
				fmt.Println("\n(Dry run - no changes saved)")
			} else {
				fmt.Println("\n(No changes to save)")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&manifestPath, "manifest", "repository.json", "Path to manifest")
	cmd.Flags().BoolVar(&verifyAll, "verify-all", false, "Verify all URLs, not just available ones")
	cmd.Flags().BoolVar(&markUnavailable, "mark-unavailable", false, "Mark dead URLs as unavailable")
	cmd.Flags().StringVar(&removeAfter, "remove-after", "", "Remove entries unavailable for duration (e.g., 720h)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without saving")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	return cmd
}
