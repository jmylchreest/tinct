// Package repocli provides CLI commands for the repository manager tool.
package repocli

import (
	"fmt"
	"time"

	"github.com/jmylchreest/tinct/internal/repomanager"
	"github.com/spf13/cobra"
)

// SyncCmd returns the sync command.
func SyncCmd() *cobra.Command {
	var (
		configPath         string
		minProtocolVersion string
		githubRepo         string
		version            string
		pluginFilter       []string
		exclude            []string
		manifestPath       string
		skipQuery          bool
		dryRun             bool
		verbose            bool
		prune              bool
		pruneRemoveAfter   string
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync plugins from GitHub release or config file",
		Long: `Sync plugins from GitHub release(s) or configuration file with optional pruning.

Mode 1: GitHub sync (requires --github, --version, --plugin-filter)
  Version specifiers:
    - latest: Latest non-prerelease
    - all: All non-prerelease versions
    - v1.2.3: Specific version tag

Mode 2: Config file sync (requires --config)
  Uses JSONL configuration file to define all sync sources.
  See sync-config.jsonl for format.

Protocol filtering:
  Use --min-protocol-version to skip plugins with older protocol versions.
  Automatically skips all older versions of a plugin once any version fails the check.

Pruning:
  Use --prune to verify and clean up unavailable entries after sync.
  Use --prune-remove-after to remove entries unavailable for a duration (e.g., 720h = 30 days).

Examples:
  # Sync from config file (recommended)
  tinct-repo-manager sync --config sync-config.jsonl --min-protocol-version 0.0.1 --prune

  # Sync from GitHub release
  tinct-repo-manager sync --github jmylchreest/tinct --version latest \
    --plugin-filter "tinct-plugin-*" --min-protocol-version 0.0.1 --prune

  # Sync all releases
  tinct-repo-manager sync --github jmylchreest/tinct --version all \
    --plugin-filter "tinct-plugin-*" --min-protocol-version 0.0.1
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate mode selection
			if configPath == "" && githubRepo == "" {
				return fmt.Errorf("must specify either --config or --github")
			}

			// If config is specified, delegate to config-based sync
			if configPath != "" {
				return syncFromConfig(configPath, manifestPath, minProtocolVersion, skipQuery, dryRun, verbose, prune, pruneRemoveAfter)
			}

			// GitHub mode - validate required flags
			if version == "" {
				return fmt.Errorf("--version is required when using --github")
			}
			if len(pluginFilter) == 0 {
				return fmt.Errorf("--plugin-filter is required when using --github")
			}

			// Create GitHub client
			client := repomanager.NewGitHubClient()

			// Load manifest
			mgr, err := repomanager.LoadManifest(manifestPath)
			if err != nil {
				return fmt.Errorf("failed to load manifest: %w", err)
			}

			// Create protocol version tracker for cascade filtering
			tracker := NewProtocolVersionTracker()

			// Create metadata hydration cache for retroactive metadata application
			// This allows us to reuse metadata from one architecture for all others,
			// even if they failed before we successfully queried one
			hydrationCache := NewMetadataHydrationCache()

			// Create a synthetic source for the shared processing function
			source := &repomanager.SyncSource{
				Type:    repomanager.SyncSourceGitHub,
				Repo:    githubRepo,
				Version: version,
				Filter:  pluginFilter,
				Exclude: exclude,
			}

			// Process using the shared function
			totalAdded, totalSkipped, totalErrors := ProcessGitHubSourceWithProtocol(
				source, client, mgr, minProtocolVersion, tracker, hydrationCache,
				skipQuery, dryRun, verbose,
			)

			// Summary
			fmt.Printf("\n=== Sync Summary ===\n")
			fmt.Printf("Added: %d\n", totalAdded)
			fmt.Printf("Skipped: %d\n", totalSkipped)
			if totalErrors > 0 {
				fmt.Printf("Errors: %d\n", totalErrors)
			}

			// Prune if requested
			var pruneStats *PruneStats

			if prune {
				fmt.Printf("\n=== Pruning ===\n")

				var removeAfterDuration time.Duration
				if pruneRemoveAfter != "" {
					removeAfterDuration, err = time.ParseDuration(pruneRemoveAfter)
					if err != nil {
						return fmt.Errorf("invalid prune-remove-after duration: %w", err)
					}
				}

				pruneStats = PruneManifest(mgr, removeAfterDuration, dryRun, verbose)

				fmt.Printf("\n=== Prune Summary ===\n")
				fmt.Printf("Checked: %d\n", pruneStats.Checked)
				fmt.Printf("Unavailable: %d\n", pruneStats.Unavailable)
				if pruneStats.FilterFailed > 0 {
					fmt.Printf("Filter failed: %d\n", pruneStats.FilterFailed)
				}
				if pruneStats.Removed > 0 {
					fmt.Printf("Removed: %d\n", pruneStats.Removed)
				}
			}

			// Save manifest
			saveNeeded := totalAdded > 0 || (prune && pruneStats != nil && (pruneStats.Unavailable > 0 || pruneStats.Removed > 0))

			if !dryRun && saveNeeded {
				// Only update LastPruned if we actually removed entries
				if prune && pruneStats != nil && pruneStats.Removed > 0 {
					now := time.Now()
					mgr.GetManifest().LastPruned = &now
				}

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

	// Mode selection flags
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to sync configuration file (JSONL)")
	cmd.Flags().StringVar(&githubRepo, "github", "", "GitHub repository (owner/repo)")

	// GitHub mode flags
	cmd.Flags().StringVar(&version, "version", "", "Release version/tag (or 'latest'/'all')")
	cmd.Flags().StringSliceVar(&pluginFilter, "plugin-filter", []string{}, "Plugin patterns to include")
	cmd.Flags().StringSliceVar(&exclude, "exclude", []string{}, "Patterns to exclude")

	// Common flags
	cmd.Flags().StringVar(&minProtocolVersion, "min-protocol-version", "", "Minimum plugin protocol version (e.g., 0.0.1)")
	cmd.Flags().StringVar(&manifestPath, "manifest", "repository.json", "Path to manifest")
	cmd.Flags().BoolVar(&skipQuery, "skip-query", false, "Skip querying plugin metadata")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without saving")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	cmd.Flags().BoolVar(&prune, "prune", false, "Verify and prune unavailable entries after sync")
	cmd.Flags().StringVar(&pruneRemoveAfter, "prune-remove-after", "720h", "Remove entries unavailable for duration (e.g., 720h)")

	// Make flags mutually exclusive (either --config OR --github)
	cmd.MarkFlagsMutuallyExclusive("config", "github")

	// Note: We'll validate required flags in RunE since they're conditional

	return cmd
}

// syncFromConfig handles syncing from a configuration file.
func syncFromConfig(
	configPath string,
	manifestPath string,
	minProtocolVersion string,
	skipQuery bool,
	dryRun bool,
	verbose bool,
	prune bool,
	pruneRemoveAfter string,
) error {
	// Load config
	fmt.Printf("Loading sync configuration from: %s\n", configPath)
	config, err := repomanager.LoadSyncConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("Found %d sync source(s)\n", len(config.Sources))
	if minProtocolVersion != "" {
		fmt.Printf("Minimum protocol version: %s\n", minProtocolVersion)
	}
	fmt.Println()

	// Load manifest
	mgr, err := repomanager.LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Create GitHub client (reused for all GitHub sources)
	client := repomanager.NewGitHubClient()

	// Create protocol version tracker for cascade filtering
	tracker := NewProtocolVersionTracker()

	// Create metadata hydration cache for retroactive metadata application
	// This allows us to reuse metadata from one architecture for all others,
	// even if they failed before we successfully queried one
	hydrationCache := NewMetadataHydrationCache()

	totalAdded := 0
	totalSkipped := 0
	totalErrors := 0

	// Process each source
	for i, source := range config.Sources {
		fmt.Printf("[%d/%d] Processing %s source\n", i+1, len(config.Sources), source.Type)

		switch source.Type {
		case repomanager.SyncSourceGitHub:
			added, skipped, errors := ProcessGitHubSourceWithProtocol(
				&source, client, mgr, minProtocolVersion, tracker, hydrationCache,
				skipQuery, dryRun, verbose,
			)
			totalAdded += added
			totalSkipped += skipped
			totalErrors += errors

		case repomanager.SyncSourceURL:
			added, errors := ProcessURLSourceWithProtocol(
				&source, mgr, minProtocolVersion, tracker, hydrationCache,
				skipQuery, dryRun, verbose,
			)
			totalAdded += added
			totalErrors += errors

		default:
			fmt.Printf("  Error: unknown source type: %s\n", source.Type)
			totalErrors++
		}

		fmt.Println()
	}

	// Summary
	fmt.Printf("=== Sync Summary ===\n")
	fmt.Printf("Added: %d\n", totalAdded)
	if totalSkipped > 0 {
		fmt.Printf("Skipped: %d\n", totalSkipped)
	}
	if totalErrors > 0 {
		fmt.Printf("Errors: %d\n", totalErrors)
	}

	// Prune if requested
	var pruneStats *PruneStats

	if prune {
		fmt.Printf("\n=== Pruning ===\n")

		var removeAfterDuration time.Duration
		if pruneRemoveAfter != "" {
			removeAfterDuration, err = time.ParseDuration(pruneRemoveAfter)
			if err != nil {
				return fmt.Errorf("invalid prune-remove-after duration: %w", err)
			}
		}

		pruneStats = PruneManifest(mgr, removeAfterDuration, dryRun, verbose)

		fmt.Printf("\n=== Prune Summary ===\n")
		fmt.Printf("Checked: %d\n", pruneStats.Checked)
		fmt.Printf("Unavailable: %d\n", pruneStats.Unavailable)
		if pruneStats.FilterFailed > 0 {
			fmt.Printf("Filter failed: %d\n", pruneStats.FilterFailed)
		}
		if pruneStats.Removed > 0 {
			fmt.Printf("Removed: %d\n", pruneStats.Removed)
		}
	}

	// Save manifest
	saveNeeded := totalAdded > 0 || (prune && pruneStats != nil && (pruneStats.Unavailable > 0 || pruneStats.Removed > 0))

	if !dryRun && saveNeeded {
		// Only update LastPruned if we actually removed entries
		if prune && pruneStats != nil && pruneStats.Removed > 0 {
			now := time.Now()
			mgr.GetManifest().LastPruned = &now
		}

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
}
