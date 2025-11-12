// Package main provides the tinct-repo-manager CLI tool.
package main

import (
	"fmt"
	"os"

	"github.com/jmylchreest/tinct/internal/repocli"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "tinct-repo-manager",
		Short: "Manage Tinct plugin repository manifests",
		Long: `Repository manager for Tinct plugin repositories.

Manage plugin repository manifests with support for:
  - Syncing from GitHub releases with integrated pruning
  - Adding/removing plugins and versions
  - Validating manifest structure
  - Listing plugins and versions

Used to maintain the official Tinct repository and can be used
by community repository maintainers.

Examples:
  # Sync from configuration file (recommended)
  tinct-repo-manager sync --config sync-config.jsonl --min-protocol-version 0.0.1 --prune

  # Sync from GitHub release with protocol filtering
  tinct-repo-manager sync --github jmylchreest/tinct --version latest \
    --plugin-filter "tinct-plugin-*" --min-protocol-version 0.0.1 --prune

  # Add plugin manually
  tinct-repo-manager add --plugin notify-send --type output --file plugin.py --platform any

  # List all plugins
  tinct-repo-manager list`,
		Version: version,
	}

	// Add commands from repocli package
	rootCmd.AddCommand(
		repocli.SyncCmd(),
		repocli.AddCmd(),
		repocli.RemoveCmd(),
		repocli.PruneCmd(),
		repocli.ValidateCmd(),
		repocli.ListCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
