// Package repocli provides CLI commands for the repository manager tool.
package repocli

import (
	"fmt"
	"os"
	"time"

	"github.com/jmylchreest/tinct/internal/repomanager"
	"github.com/spf13/cobra"
)

// syncOptions bundles the flags accepted by the `sync` cobra command. The
// same set is consumed by both the GitHub-mode RunE body and syncFromConfig,
// so the bundle replaces what would otherwise be a duplicated long parameter
// list. GitHub-only fields (githubRepo, version, pluginFilter, exclude) are
// unused by config-mode sync.
type syncOptions struct {
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
	pruneIncompatible  bool
	keepRecent         int
	changelogOutput    string
	changelogFormat    string
}

// SyncCmd returns the sync command.
func SyncCmd() *cobra.Command {
	var opts syncOptions

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
  Use --prune-incompatible to remove plugin versions incompatible with current protocol.
  Use --keep-recent to keep only the N most recent versions per plugin.

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
		RunE: func(_ *cobra.Command, _ []string) error {
			// Validate mode selection
			if opts.configPath == "" && opts.githubRepo == "" {
				return fmt.Errorf("must specify either --config or --github")
			}

			// If config is specified, delegate to config-based sync
			if opts.configPath != "" {
				return syncFromConfig(opts)
			}

			// GitHub mode - validate required flags
			if opts.version == "" {
				return fmt.Errorf("--version is required when using --github")
			}
			if len(opts.pluginFilter) == 0 {
				return fmt.Errorf("--plugin-filter is required when using --github")
			}

			// Create GitHub client
			client := repomanager.NewGitHubClient()

			// Load manifest
			mgr, err := repomanager.LoadManifest(opts.manifestPath)
			if err != nil {
				return fmt.Errorf("failed to load manifest: %w", err)
			}

			// Create protocol version tracker for cascade filtering
			tracker := NewProtocolVersionTracker()

			// Create metadata hydration cache for retroactive metadata application
			// This allows us to reuse metadata from one architecture for all others,
			// even if they failed before we successfully queried one
			hydrationCache := NewMetadataHydrationCache()

			// Create changelog tracker
			changelog := NewChangeLog()

			// Create a synthetic source for the shared processing function
			source := &repomanager.SyncSource{
				Type:    repomanager.SyncSourceGitHub,
				Repo:    opts.githubRepo,
				Version: opts.version,
				Filter:  opts.pluginFilter,
				Exclude: opts.exclude,
			}

			// Process using the shared function
			totalAdded, totalSkipped, totalErrors := ProcessGitHubSourceWithProtocol(
				source, client, mgr, opts.minProtocolVersion, tracker, hydrationCache,
				opts.skipQuery, opts.dryRun, opts.verbose, changelog,
			)

			// Summary
			fmt.Printf("\n=== Sync Summary ===\n")
			fmt.Printf("Added: %d\n", totalAdded)
			fmt.Printf("Skipped: %d\n", totalSkipped)
			if totalErrors > 0 {
				fmt.Printf("Errors: %d\n", totalErrors)
			}

			// Prune if requested
			pruneStats, err := runPruneIfRequested(mgr, opts.prune, opts.pruneRemoveAfter, opts.pruneIncompatible, opts.keepRecent, opts.dryRun, opts.verbose, changelog)
			if err != nil {
				return err
			}

			return finalizeSyncResults(mgr, changelog, pruneStats, opts.prune, opts.dryRun, opts.manifestPath, opts.changelogOutput, opts.changelogFormat)
		},
	}

	// Mode selection flags
	cmd.Flags().StringVarP(&opts.configPath, "config", "c", "", "Path to sync configuration file (JSONL)")
	cmd.Flags().StringVar(&opts.githubRepo, "github", "", "GitHub repository (owner/repo)")

	// GitHub mode flags
	cmd.Flags().StringVar(&opts.version, "version", "", "Release version/tag (or 'latest'/'all')")
	cmd.Flags().StringSliceVar(&opts.pluginFilter, "plugin-filter", []string{}, "Plugin patterns to include")
	cmd.Flags().StringSliceVar(&opts.exclude, "exclude", []string{}, "Patterns to exclude")

	// Common flags
	cmd.Flags().StringVar(&opts.minProtocolVersion, "min-protocol-version", "", "Minimum plugin protocol version (e.g., 0.0.1)")
	cmd.Flags().StringVar(&opts.manifestPath, "manifest", "repository.json", "Path to manifest")
	cmd.Flags().BoolVar(&opts.skipQuery, "skip-query", false, "Skip querying plugin metadata")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Preview changes without saving")
	cmd.Flags().BoolVarP(&opts.verbose, "verbose", "v", false, "Verbose output")
	cmd.Flags().BoolVar(&opts.prune, "prune", false, "Verify and prune unavailable entries after sync")
	cmd.Flags().StringVar(&opts.pruneRemoveAfter, "prune-remove-after", "720h", "Remove entries unavailable for duration (e.g., 720h)")
	cmd.Flags().BoolVar(&opts.pruneIncompatible, "prune-incompatible", false, "Remove plugin versions incompatible with current protocol")
	cmd.Flags().IntVar(&opts.keepRecent, "keep-recent", 0, "Keep only the N most recent versions per plugin (0 = keep all)")
	cmd.Flags().StringVar(&opts.changelogOutput, "changelog-output", "", "Write changelog to file (use '-' for stdout)")
	cmd.Flags().StringVar(&opts.changelogFormat, "changelog-format", "text", "Changelog format: text, short, json")

	// Make flags mutually exclusive (either --config OR --github)
	cmd.MarkFlagsMutuallyExclusive("config", "github")

	// Note: We'll validate required flags in RunE since they're conditional

	return cmd
}

// syncFromConfig handles syncing from a configuration file.
func syncFromConfig(opts syncOptions) error {
	// Load config
	fmt.Printf("Loading sync configuration from: %s\n", opts.configPath)
	config, err := repomanager.LoadSyncConfig(opts.configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("Found %d sync source(s)\n", len(config.Sources))
	if opts.minProtocolVersion != "" {
		fmt.Printf("Minimum protocol version: %s\n", opts.minProtocolVersion)
	}
	fmt.Println()

	// Load manifest
	mgr, err := repomanager.LoadManifest(opts.manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Create snapshot of manifest before sync for diff calculation
	mgr.CreateSnapshot()

	// Create GitHub client (reused for all GitHub sources)
	client := repomanager.NewGitHubClient()

	// Create protocol version tracker for cascade filtering
	tracker := NewProtocolVersionTracker()

	// Create metadata hydration cache for retroactive metadata application
	// This allows us to reuse metadata from one architecture for all others,
	// even if they failed before we successfully queried one
	hydrationCache := NewMetadataHydrationCache()

	// Create temporary changelog tracker for operations (not used for final output)
	// We'll build the real changelog from the manifest diff
	changelog := NewChangeLog()

	totalAdded := 0
	totalSkipped := 0
	totalErrors := 0

	// Process each source
	for i := range config.Sources {
		source := &config.Sources[i]
		fmt.Printf("[%d/%d] Processing %s source\n", i+1, len(config.Sources), source.Type)

		switch source.Type {
		case repomanager.SyncSourceGitHub:
			added, skipped, errors := ProcessGitHubSourceWithProtocol(
				source, client, mgr, opts.minProtocolVersion, tracker, hydrationCache,
				opts.skipQuery, opts.dryRun, opts.verbose, changelog,
			)
			totalAdded += added
			totalSkipped += skipped
			totalErrors += errors

		case repomanager.SyncSourceURL:
			added, errors := ProcessURLSourceWithProtocol(
				source, mgr, opts.minProtocolVersion, tracker, hydrationCache,
				opts.skipQuery, opts.dryRun, opts.verbose, changelog,
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
	pruneStats, err := runPruneIfRequested(mgr, opts.prune, opts.pruneRemoveAfter, opts.pruneIncompatible, opts.keepRecent, opts.dryRun, opts.verbose, changelog)
	if err != nil {
		return err
	}

	// Compute the actual diff between snapshot and current state to get
	// the authoritative changelog (syncFromConfig uses snapshots).
	realChangelog := BuildFromManifestDiff(mgr.ComputeDiff())

	return finalizeSyncResults(mgr, realChangelog, pruneStats, opts.prune, opts.dryRun, opts.manifestPath, opts.changelogOutput, opts.changelogFormat)
}

// runPruneIfRequested executes pruning when enabled and prints the prune summary.
// It returns the prune stats (nil if pruning was not requested) and any error.
func runPruneIfRequested(
	mgr *repomanager.ManifestManager,
	prune bool,
	pruneRemoveAfter string,
	pruneIncompatible bool,
	keepRecent int,
	dryRun bool,
	verbose bool,
	changelog *ChangeLog,
) (*PruneStats, error) {

	if !prune {
		return nil, nil
	}

	fmt.Printf("\n=== Pruning ===\n")

	var removeAfterDuration time.Duration
	if pruneRemoveAfter != "" {
		var err error
		removeAfterDuration, err = time.ParseDuration(pruneRemoveAfter)
		if err != nil {
			return nil, fmt.Errorf("invalid prune-remove-after duration: %w", err)
		}
	}

	stats := PruneManifestWithOptions(mgr, &PruneOptions{
		RemoveAfterDuration: removeAfterDuration,
		PruneIncompatible:   pruneIncompatible,
		KeepRecent:          keepRecent,
		DryRun:              dryRun,
		Verbose:             verbose,
	}, changelog)

	fmt.Printf("\n=== Prune Summary ===\n")
	fmt.Printf("Checked: %d\n", stats.Checked)
	fmt.Printf("Unavailable: %d\n", stats.Unavailable)
	if stats.FilterFailed > 0 {
		fmt.Printf("Filter failed: %d\n", stats.FilterFailed)
	}
	if stats.Incompatible > 0 {
		fmt.Printf("Incompatible: %d\n", stats.Incompatible)
	}
	if stats.OldVersions > 0 {
		fmt.Printf("Old versions: %d\n", stats.OldVersions)
	}
	if stats.Removed > 0 {
		fmt.Printf("Removed: %d\n", stats.Removed)
	}

	return stats, nil
}

// finalizeSyncResults saves the manifest (if there are material changes) and
// writes the changelog to the configured output. Both the RunE and
// syncFromConfig paths call this after sync + optional prune.
func finalizeSyncResults(
	mgr *repomanager.ManifestManager,
	changelog *ChangeLog,
	pruneStats *PruneStats,
	prune bool,
	dryRun bool,
	manifestPath string,
	changelogOutput string,
	changelogFormat string,
) error {

	saveNeeded := changelog.HasMaterialChanges()

	switch {
	case !dryRun && saveNeeded:
		if prune && pruneStats != nil && pruneStats.Removed > 0 {
			now := time.Now()
			mgr.GetManifest().LastPruned = &now
			mgr.MarkDirty()
		}

		if err := mgr.Save(); err != nil {
			return fmt.Errorf("failed to save manifest: %w", err)
		}
		fmt.Printf("\n✓ Manifest saved: %s\n", manifestPath)
	case dryRun:
		fmt.Println("\n(Dry run - no changes saved)")
	default:
		fmt.Println("\n(No changes to save)")
	}

	if err := writeChangelog(changelog, changelogOutput, changelogFormat); err != nil {
		return fmt.Errorf("failed to write changelog: %w", err)
	}

	return nil
}

// writeChangelog writes the changelog to the specified output.
func writeChangelog(changelog *ChangeLog, output, format string) error {
	// Skip if no output specified
	if output == "" {
		return nil
	}

	// Format the changelog
	content := changelog.Format(format)

	// Write to stdout
	if output == "-" {
		fmt.Print(content)
		return nil
	}

	// Write to file
	return os.WriteFile(output, []byte(content), 0o600)
}
