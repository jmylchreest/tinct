// Package cli provides the command-line interface for Tinct.
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/manager"
	"github.com/jmylchreest/tinct/internal/plugin/repository"
	"github.com/jmylchreest/tinct/pkg/plugin/paths"
)

const (
	// PluginManifestFile is the name of the plugin manifest file in the
	// config directory. The manifest tracks installed external plugins so
	// they auto-register on every CLI invocation.
	PluginManifestFile = "plugins.manifest.json"

	// legacyManifestFileNames holds prior names for the same file; loadPluginManifest
	// migrates from the first one that exists. Keep newest-to-oldest so a
	// stale legacy copy doesn't override a present new file.
	legacyManifestFileV1 = "plugins.lock.json"   // 0.2.x — pre-rename to "manifest"
	legacyManifestFileV0 = ".tinct-plugins.json" // pre-XDG hidden file in $HOME / CWD
)

// PluginManifest tracks the external plugins installed for tinct on this
// machine. Stored at ~/.config/tinct/plugins.manifest.json (or wherever
// --manifest-file points).
//
// This is a registry, not a versioned lock file: it records what's
// installed and where each entry originated, so commands like `tinct
// plugins list`, `tinct plugins update`, and `tinct plugins sync` work
// across shell sessions and machines. There is no checksum or version
// pin — that's why we call it a manifest, not a lock.
type PluginManifest struct {
	// ExternalPlugins maps plugin names to their metadata.
	ExternalPlugins map[string]*ExternalPluginMeta `json:"external_plugins,omitempty"`
}

// LivePlugins returns the subset of ExternalPlugins whose recorded
// binary still exists on disk. It's the right view for any read-only
// runtime path (auto-register on CLI invocation, `tinct plugins list`,
// generate / extract pipelines) so dangling entries don't trigger
// noisy warnings on every run.
//
// CRUD callers (add / delete / update / sync) should use the raw
// ExternalPlugins map instead — they need to see the dangling entry
// in order to clean it up.
//
// LivePlugins does not mutate the receiver or the on-disk file; it
// returns a fresh map each call.
func (m *PluginManifest) LivePlugins() map[string]*ExternalPluginMeta {
	if m == nil || m.ExternalPlugins == nil {
		return nil
	}
	live := make(map[string]*ExternalPluginMeta, len(m.ExternalPlugins))
	for k, meta := range m.ExternalPlugins {
		if missing, _ := isManifestEntryMissing(meta); missing {
			continue
		}
		live[k] = meta
	}
	return live
}

// ExternalPluginMeta contains metadata about an external plugin.
type ExternalPluginMeta struct {
	// Name is the plugin's actual name (from --plugin-info).
	Name string `json:"name"`

	// Path is the absolute path to the plugin executable.
	Path string `json:"path"`

	// Type is the plugin type (input or output).
	Type string `json:"type"`

	// Version is the plugin version if available.
	Version string `json:"version,omitempty"`

	// Description is the plugin description if available.
	Description string `json:"description,omitempty"`

	// Source contains structured information about where the plugin came from.
	Source *repository.PluginSource `json:"source,omitempty"`

	// SourceLegacy is the old string-based source field for backward compatibility.
	SourceLegacy string `json:"source_legacy,omitempty"`

	// Flags caches the plugin's GetFlagHelp() output so the generate
	// command can register --<plugin>.<arg> entries at startup without
	// spawning the plugin binary. Populated at install/update time;
	// stale entries are auto-refreshed by the runtime path if missing.
	Flags []input.FlagHelp `json:"flags,omitempty"`
}

var (
	// Plugin command flags.
	pluginManifestPath string
	pluginType         string
	pluginForce        bool
	pluginYes          bool
	pluginSourceType   string
	pluginNoCopy       bool
	pluginShowPath     bool
)

// pluginsCmd represents the plugins command.
var pluginsCmd = &cobra.Command{
	Use:   "plugins",
	Short: "Manage plugins",
	Long: `Manage Tinct plugins including listing and managing external plugins.

The plugin manifest (~/.config/tinct/plugins.manifest.json) tracks
installed external plugins and their sources. Use 'tinct plugins sync'
to reproduce your plugin setup on another machine.

Commands that modify the manifest:
  - add: Adds external plugin and updates the manifest
  - delete: Removes external plugin and updates the manifest
  - update: Updates external plugins from their sources

A manifest from an earlier version (named plugins.lock.json) is
migrated automatically on first read. The legacy --lock-file flag is
still accepted as an alias for --manifest-file.`,
}

// pluginListCmd lists all available plugins.
var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available plugins",
	Long: `List all available plugins.

Shows both built-in and external plugins with their type and description.`,
	RunE: runPluginList,
}

// pluginAddCmd adds an external plugin.
var pluginAddCmd = &cobra.Command{
	Use:   "add <source>",
	Short: "Add an external plugin",
	Long: `Add an external plugin from a local file, HTTP URL, or Git repository.

The plugin will be copied to the plugin directory and registered in
the plugin manifest. The plugin name is automatically detected from
the plugin's --plugin-info output.

WARNING: Only install plugins from trusted sources. Plugins execute with your
user permissions and can access your system. Review plugin source code before
installation to ensure it is safe.

The command will:
  1. Verify source and destination are not the same file
  2. Query plugin metadata (name, version, type, protocol)
  3. Check protocol compatibility
  4. Check for version conflicts (upgrades proceed automatically)
  5. Copy plugin to ~/.local/share/tinct/plugins/ (unless --no-copy is used)
  6. Register plugin in the manifest

Plugin upgrades (newer versions) proceed automatically.
Use --force to downgrade, reinstall same version, or overwrite.
Use --no-copy to reference the plugin at its current location without copying
(useful for system-installed packages that manage their own updates).

Examples:
  tinct plugins add ./contrib/notify-send.py
  tinct plugins add https://example.com/plugins/theme.sh
  tinct plugins add https://github.com/user/plugin.git
  tinct plugins add https://github.com/user/plugin.git:path/to/plugin.sh
  tinct plugins add ./my-plugin.sh --force  # Force overwrite
  tinct plugins add /usr/bin/tinct-plugin-random --no-copy  # System package`,
	Args: cobra.ExactArgs(1),
	RunE: runPluginAdd,
}

// pluginDeleteCmd removes an external plugin.
var pluginDeleteCmd = &cobra.Command{
	Use:   "delete <plugin-name>",
	Short: "Delete an external plugin",
	Long: `Delete an external plugin from the plugin directory and remove it from the manifest.

Built-in plugins cannot be deleted.

Examples:
  tinct plugins delete notify
  tinct plugins delete custom-theme`,
	Args: cobra.ExactArgs(1),
	RunE: runPluginDelete,
}

// pluginUpdateCmd updates external plugins from manifest sources.
var pluginUpdateCmd = &cobra.Command{
	Use:   "update [name...]",
	Short: "Update external plugins from manifest sources",
	Long: `Update external plugins by re-downloading/copying from their source locations.

This reads the plugin manifest and updates plugins based on their source field.
Useful for keeping plugins in sync across machines or after pulling manifest
changes from a shared dotfiles repo.

Named plugins are updated on their own; with no names, every external plugin in
the manifest is updated.

Examples:
  tinct plugins update
  tinct plugins update noctalia
  tinct plugins update noctalia zed
  tinct plugins update --manifest-file /path/to/plugins.manifest.json`,
	RunE: runPluginUpdate,
}

func init() {
	// Add plugins command flags.
	pluginsCmd.PersistentFlags().StringVar(&pluginManifestPath, "manifest-file", "",
		"path to plugin manifest file (default: ~/.config/tinct/plugins.manifest.json)")
	// Backward-compat alias for the pre-rename name; hidden to keep --help tidy
	// but still recognised so existing scripts/aliases keep working.
	pluginsCmd.PersistentFlags().StringVar(&pluginManifestPath, "lock-file", "",
		"deprecated: alias for --manifest-file")
	if f := pluginsCmd.PersistentFlags().Lookup("lock-file"); f != nil {
		f.Hidden = true
		f.Deprecated = "use --manifest-file"
	}

	// Plugin list flags.
	pluginListCmd.Flags().BoolVar(&pluginShowPath, "show-path", false, "show the actual file path used when loading each plugin")

	// Plugin add flags.
	pluginAddCmd.Flags().StringVar(&pluginType, "type", "output", "plugin type (input or output)")
	pluginAddCmd.Flags().BoolVarP(&pluginForce, "force", "f", false, "force overwrite if plugin already exists")
	pluginAddCmd.Flags().StringVar(&pluginSourceType, "source-type", "", "force source type (local, http, git) - auto-detected if not specified")
	pluginAddCmd.Flags().BoolVar(&pluginNoCopy, "no-copy", false, "register plugin at its current location without copying (useful for system packages)")

	// Plugin delete flags.
	pluginDeleteCmd.Flags().BoolVarP(&pluginForce, "force", "f", false, "force deletion without confirmation")

	// Add subcommands.
	pluginsCmd.AddCommand(pluginListCmd)
	pluginsCmd.AddCommand(pluginAddCmd)
	pluginsCmd.AddCommand(pluginDeleteCmd)
	pluginsCmd.AddCommand(pluginUpdateCmd)
}

// runPluginList lists all available plugins.
func runPluginList(cmd *cobra.Command, _ []string) (err error) {
	items := 0
	defer func() { sendPluginSubcommandResult("list", "", items, err) }()

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	// Load plugin lock and create manager.
	lock, manifestPath, err := loadPluginManifest()
	if err != nil && verbose {
		fmt.Fprintf(os.Stderr, "Note: %v\n", err)
	}

	mgr := createManagerFromManifest(lock)

	if verbose && manifestPath != "" {
		fmt.Fprintf(os.Stderr, "Using manifest: %s\n\n", manifestPath)
	}

	allPlugins := collectAllPlugins(mgr, lock)
	items = len(allPlugins)

	displayPluginTable(allPlugins, pluginShowPath)

	return nil
}

// runPluginAdd adds an external plugin with comprehensive safety checks.
func runPluginAdd(cmd *cobra.Command, args []string) (err error) { //nolint:gocyclo,gocognit // multi-source plugin installation with validation
	defer func() { sendPluginSubcommandResult("add", "", 1, err) }()

	source := args[0]
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	// Load or create plugin lock.
	lock, manifestPath := loadOrCreatePluginManifest()

	if verbose {
		fmt.Fprintf(os.Stderr, "Using manifest: %s\n", manifestPath)
	}

	// Initialize external plugins map if needed.
	if lock.ExternalPlugins == nil {
		lock.ExternalPlugins = make(map[string]*ExternalPluginMeta)
	}

	// Get plugin directory.
	pluginDir, err := getPluginDir()
	if err != nil {
		return fmt.Errorf("failed to get plugin directory: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Plugin directory: %s\n", pluginDir)
	}

	// Stage 1: Resolve source path and check if it's already in the plugin directory
	sourcePath, isAlreadyInstalled, err := resolvePluginSource(source, pluginDir, pluginSourceType, verbose)
	if err != nil {
		return err
	}

	if isAlreadyInstalled {
		if verbose {
			fmt.Fprintf(os.Stderr, "Note: Plugin source is already in plugin directory\n")
		}
	}

	// Stage 2: Query plugin metadata (before copying to avoid executing untrusted code from final location)
	pluginInfo, err := queryFullPluginMetadata(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to query plugin metadata: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Plugin name: %s\n", pluginInfo.Name)
		fmt.Fprintf(os.Stderr, "Plugin type: %s\n", pluginInfo.Type)
		fmt.Fprintf(os.Stderr, "Plugin version: %s\n", pluginInfo.Version)
		fmt.Fprintf(os.Stderr, "Protocol version: %s\n", pluginInfo.ProtocolVersion)
	}

	// Stage 3: Check protocol compatibility
	if err := checkProtocolCompatibility(pluginInfo.ProtocolVersion, verbose); err != nil {
		return err
	}

	// Stage 4: Check for conflicts and version comparisons
	action, existingMeta, err := determinePluginAction(lock, pluginInfo, pluginForce)
	if err != nil {
		return err
	}

	// Stage 5: Install plugin to final location (if not already there)
	var finalPath string
	if pluginNoCopy { //nolint:nestif
		// Use the source path directly without copying
		finalPath = sourcePath
		if verbose {
			fmt.Fprintf(os.Stderr, "Using plugin at: %s (no-copy mode)\n", finalPath)
		}

		// If we're overwriting an existing plugin that was copied to the plugin dir, clean it up
		if existingMeta != nil && existingMeta.Path != "" {
			existingPathDir := filepath.Dir(existingMeta.Path)
			if existingPathDir == pluginDir {
				// Old plugin was in the plugin directory, remove it
				if err := os.Remove(existingMeta.Path); err != nil && !os.IsNotExist(err) {
					if verbose {
						fmt.Fprintf(os.Stderr, "Warning: failed to remove old plugin file: %v\n", err)
					}
				} else if verbose {
					fmt.Fprintf(os.Stderr, "Removed old plugin file: %s\n", existingMeta.Path)
				}
			}
		}
	} else {
		finalPath = filepath.Join(pluginDir, filepath.Base(sourcePath))
		if !isAlreadyInstalled {
			if err := installPlugin(sourcePath, finalPath, verbose); err != nil {
				return fmt.Errorf("failed to install plugin: %w", err)
			}
		}
	}

	// Stage 6: Update manifest
	// Build structured source metadata.
	pluginSource := buildPluginSource(source, pluginSourceType, finalPath)

	meta := &ExternalPluginMeta{
		Name:         pluginInfo.Name,
		Path:         finalPath,
		Type:         pluginInfo.Type,
		Source:       pluginSource,
		SourceLegacy: source,
		Version:      pluginInfo.Version,
		Description:  pluginInfo.Description,
	}
	PopulateFlagsCache(meta)
	lock.ExternalPlugins[pluginInfo.Name] = meta

	if err := savePluginManifest(manifestPath, lock); err != nil {
		return fmt.Errorf("failed to save plugin lock: %w", err)
	}

	// Stage 7: Display success message
	printPluginAddSuccess(pluginInfo, action, existingMeta, finalPath)
	return nil
}

// runPluginDelete removes an external plugin.
func runPluginDelete(cmd *cobra.Command, args []string) (err error) { //nolint:gocyclo
	defer func() { sendPluginSubcommandResult("delete", "", 1, err) }()

	pluginName := args[0]
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	// Load plugin lock.
	lock, manifestPath, err := loadPluginManifest()
	if err != nil {
		return fmt.Errorf("failed to load plugin lock: %w", err)
	}

	if lock == nil || lock.ExternalPlugins == nil {
		return fmt.Errorf("no external plugins found")
	}

	// Check if plugin exists.
	meta, exists := lock.ExternalPlugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin '%s' not found", pluginName)
	}

	// Confirm deletion if not forced.
	if !pluginForce {
		fmt.Printf("Are you sure you want to delete plugin '%s'? (y/N): ", pluginName)
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}
		if !strings.EqualFold(response, "y") {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	// Delete plugin file only if it's in the plugin directory.
	// Plugins added with --no-copy are not deleted from disk.
	pluginDir, err := getPluginDir()
	if err != nil {
		return fmt.Errorf("failed to get plugin directory: %w", err)
	}

	// Check if plugin is in the plugin directory.
	pluginPathDir := filepath.Dir(meta.Path)
	if pluginPathDir == pluginDir {
		// Plugin is in plugin directory, safe to delete.
		if err := os.Remove(meta.Path); err != nil && !os.IsNotExist(err) {
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to delete plugin file: %v\n", err)
			}
		} else if verbose {
			fmt.Fprintf(os.Stderr, "Deleted plugin file: %s\n", meta.Path)
		}
	} else {
		// Plugin is outside plugin directory (--no-copy), don't delete.
		if verbose {
			fmt.Fprintf(os.Stderr, "Plugin file not deleted (outside plugin directory): %s\n", meta.Path)
		}
	}

	// Remove from manifest.
	delete(lock.ExternalPlugins, pluginName)

	// Save manifest.
	if err := savePluginManifest(manifestPath, lock); err != nil {
		return fmt.Errorf("failed to save plugin lock: %w", err)
	}

	fmt.Printf("Plugin '%s' deleted successfully\n", pluginName)
	return nil
}

// updatePluginFromRepository updates a plugin from a repository source.
func updatePluginFromRepository(meta *ExternalPluginMeta, pluginDir string, verbose bool) (string, error) {
	// Get repository manager.
	mgr, err := getRepoManager()
	if err != nil {
		return "", fmt.Errorf("failed to get repository manager: %w", err)
	}

	// Find the plugin in the repository.
	// If Source.Version is empty, update to latest. Otherwise respect the pinned version.
	version := "latest"
	if meta.Source.Version != "" {
		version = meta.Source.Version
	}

	result, err := mgr.FindPluginInRepository(
		meta.Source.Repository,
		meta.Source.Plugin,
		version,
	)
	if err != nil {
		return "", fmt.Errorf("failed to find plugin in repository: %w", err)
	}

	// Determine current platform.
	platform := repository.NormalizePlatform(runtime.GOOS, runtime.GOARCH)

	// Find download for current platform.
	download, ok := result.Version.Downloads[platform]
	if !ok {
		return "", fmt.Errorf("plugin not available for platform %s", platform)
	}

	if !download.Available {
		reason := "unknown reason"
		if download.UnavailableReason != "" {
			reason = download.UnavailableReason
		}
		return "", fmt.Errorf("plugin unavailable: %s", reason)
	}

	// Check if we need to update by comparing repository versions.
	// Use the source's checksum to detect if we already have this exact version.
	if meta.Source.Checksum != "" && download.Checksum == meta.Source.Checksum {
		// Same checksum = same file, no update needed.
		return "", fmt.Errorf("already up to date (version %s)", result.Version.Version)
	}

	// If pinned to a specific version and it matches, no update needed.
	if meta.Source.Version != "" && result.Version.Version == meta.Source.Version {
		return "", fmt.Errorf("already up to date (version %s)", meta.Source.Version)
	}

	// Download and install from URL (quiet=true since caller shows table status).
	pluginPath, err := installPluginFromSource(download.URL, "", pluginDir, sourceTypeHTTP, verbose, true)
	if err != nil {
		return "", fmt.Errorf("failed to download plugin: %w", err)
	}

	// Update the source metadata with the new checksum.
	// This is critical so next update can detect we already have this version.
	meta.Source.Checksum = download.Checksum

	return pluginPath, nil
}

// selectPluginsToUpdate resolves the plugin names to update, sorted.
//
// With no names, every external plugin in the manifest is returned —
// the historical behaviour. With names, only those are returned, and an
// unknown one is an error rather than a silent no-op.
//
// This exists because the command previously declared no positional
// arguments and never read them: cobra accepted `tinct plugins update
// noctalia` and then updated everything, which is both surprising and
// slow when a user wants one plugin refreshed.
func selectPluginsToUpdate(lock *PluginManifest, names []string) ([]string, error) {
	if len(names) == 0 {
		all := make([]string, 0, len(lock.ExternalPlugins))
		for name := range lock.ExternalPlugins {
			all = append(all, name)
		}
		sort.Strings(all)
		return all, nil
	}

	known := make([]string, 0, len(lock.ExternalPlugins))
	for name := range lock.ExternalPlugins {
		known = append(known, name)
	}
	sort.Strings(known)

	seen := make(map[string]bool, len(names))
	selected := make([]string, 0, len(names))
	for _, name := range names {
		if _, ok := lock.ExternalPlugins[name]; !ok {
			return nil, fmt.Errorf(
				"unknown plugin %q (installed: %s)",
				name, strings.Join(known, ", "),
			)
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		selected = append(selected, name)
	}

	sort.Strings(selected)
	return selected, nil
}

// runPluginUpdate updates external plugins from manifest sources.
func runPluginUpdate(cmd *cobra.Command, args []string) (err error) { //nolint:gocyclo,gocognit // batch plugin update with per-plugin error handling
	items := 0
	defer func() { sendPluginSubcommandResult("update", "", items, err) }()

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	// Load plugin lock.
	lock, manifestPath, err := loadPluginManifest()
	if err != nil {
		return fmt.Errorf("failed to load plugin manifest: %w", err)
	}

	if lock == nil || lock.ExternalPlugins == nil || len(lock.ExternalPlugins) == 0 {
		fmt.Println("No external plugins found in manifest")
		return nil
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Using manifest: %s\n", manifestPath)
	}

	// Get plugin directory.
	pluginDir, err := getPluginDir()
	if err != nil {
		return fmt.Errorf("failed to get plugin directory: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Plugin directory: %s\n", pluginDir)
	}

	// Create live table for updates
	fmt.Fprintln(os.Stderr, "\nUpdating plugins...")
	fmt.Fprintln(os.Stderr)
	table := NewTable([]string{"PLUGIN", "STATUS"}).SetLive(true).WithWriter(os.Stderr)

	// Get sorted plugin names, restricted to the ones named on the
	// command line when any were given.
	pluginNames, err := selectPluginsToUpdate(lock, args)
	if err != nil {
		return err
	}

	// Initialize all rows with "Queued" status
	// In live mode, rendering is deferred until Finish() is called
	for _, name := range pluginNames {
		table.AddRowWithID(name, []string{name, "Queued"})
	}

	items = len(pluginNames)

	// Update each external plugin
	successCount := 0
	failCount := 0

	for _, name := range pluginNames {
		meta := lock.ExternalPlugins[name]
		oldVersion := meta.Version

		table.UpdateRow(name, map[string]string{"STATUS": "Checking..."})

		// Handle repository sources specially.
		var pluginPath string
		var err error
		var skipUpdate bool
		if meta.Source != nil && meta.Source.Type == sourceTypeRepository {
			table.UpdateRow(name, map[string]string{"STATUS": "Downloading..."})
			pluginPath, err = updatePluginFromRepository(meta, pluginDir, verbose)
			// Check if error is "already up to date"
			if err != nil && strings.Contains(err.Error(), "already up to date") {
				skipUpdate = true
				err = nil
			}
		} else {
			// Install plugin from other source types (HTTP, local, git).
			var sourceForInstall string
			switch {
			case meta.Source != nil && meta.Source.URL != "":
				sourceForInstall = meta.Source.URL
			case meta.Source != nil && meta.Source.OriginalPath != "":
				// Local sources record where they were installed from in
				// OriginalPath, not URL. Without this, every plugin added
				// with `tinct plugins add ./some/binary` reported "No
				// source information" and could never be updated — even
				// though the manifest knew exactly where it came from.
				sourceForInstall = meta.Source.OriginalPath
			case meta.SourceLegacy != "":
				sourceForInstall = meta.SourceLegacy
			default:
				table.UpdateRow(name, map[string]string{"STATUS": "✗ No source information"})
				failCount++
				continue
			}
			table.UpdateRow(name, map[string]string{"STATUS": "Downloading..."})
			pluginPath, err = installPluginFromSource(sourceForInstall, name, pluginDir, "", verbose, true)
		}

		if err != nil {
			table.UpdateRow(name, map[string]string{"STATUS": fmt.Sprintf("✗ %v", err)})
			failCount++
			continue
		}

		if skipUpdate {
			status := "Already up to date"
			if oldVersion != "" {
				status = fmt.Sprintf("Already up to date (%s)", oldVersion)
			}
			table.UpdateRow(name, map[string]string{"STATUS": status})
			successCount++
			continue
		}

		table.UpdateRow(name, map[string]string{"STATUS": "Installing..."})

		// Query plugin for updated metadata.
		actualName, pluginDescription, pluginType, version, _ := queryPluginMetadata(pluginPath)
		if actualName == "" {
			actualName = meta.Name // Keep existing name if query fails
			if actualName == "" {
				actualName = name // Fallback to manifest key
			}
		}
		if pluginType == "" {
			pluginType = meta.Type // Keep existing type if query fails
		}

		// Update metadata in manifest.
		updatedMeta := &ExternalPluginMeta{
			Name:        actualName,
			Path:        pluginPath,
			Type:        pluginType,
			Source:      meta.Source,
			Version:     version,
			Description: pluginDescription,
		}
		PopulateFlagsCache(updatedMeta)
		lock.ExternalPlugins[name] = updatedMeta

		// Show version change or just new version
		var status string
		switch {
		case oldVersion != "" && version != "" && oldVersion != version:
			status = fmt.Sprintf("%s → %s", oldVersion, version)
		case version != "":
			status = fmt.Sprintf("Updated (%s)", version)
		default:
			status = "Updated"
		}
		table.UpdateRow(name, map[string]string{"STATUS": status})
		successCount++
	}

	table.Finish()

	// Save updated manifest.
	if successCount > 0 {
		if err := savePluginManifest(manifestPath, lock); err != nil {
			return fmt.Errorf("failed to save plugin lock: %w", err)
		}
	}

	if failCount > 0 {
		return fmt.Errorf("some plugins failed to update")
	}

	return nil
}

// getDefaultManifestFilePath returns the default manifest path at
// ~/.config/tinct/plugins.manifest.json on Linux/macOS, %AppData%/tinct/...
// on Windows. Uses paths.XDGConfigDir so macOS lands on ~/.config/tinct
// rather than ~/Library/Application Support/tinct (matching docs).
func getDefaultManifestFilePath() string {
	return filepath.Join(paths.XDGConfigDir(), "tinct", PluginManifestFile)
}

// migrateLegacyManifestFile checks for the manifest at legacy locations
// and migrates them to the new default path. Returns the new path if a
// legacy file was found and migrated, or empty string. Searched in
// priority order — newer locations first so a stale older copy doesn't
// override a more recent state.
func migrateLegacyManifestFile(newPath string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}

	configDir := filepath.Dir(newPath) // ~/.config/tinct

	legacyPaths := []string{
		filepath.Join(configDir, legacyManifestFileV1), // ~/.config/tinct/plugins.manifest.json (0.2.x)
		filepath.Join(home, legacyManifestFileV0),      // ~/.tinct-plugins.json (pre-XDG)
		legacyManifestFileV0,                           // ./.tinct-plugins.json (CWD fallback)
	}

	for _, legacyPath := range legacyPaths {
		if legacyPath == "" || legacyPath == newPath {
			continue
		}
		data, err := os.ReadFile(legacyPath) // #nosec G304 - Legacy manifest path under user-controlled config dir
		if err != nil {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(newPath), 0o750); err != nil {
			return ""
		}

		if err := os.WriteFile(newPath, data, 0o600); err != nil { // #nosec G703 -- newPath derives from XDGConfigDir
			return ""
		}

		_ = os.Remove(legacyPath)

		fmt.Fprintf(os.Stderr, "Migrated plugin manifest: %s → %s\n", legacyPath, newPath)
		return newPath
	}

	return ""
}

// loadPluginManifest loads the plugin manifest.
func loadPluginManifest() (*PluginManifest, string, error) {
	manifestPath := pluginManifestPath

	if manifestPath == "" {
		defaultPath := getDefaultManifestFilePath()

		if _, err := os.Stat(defaultPath); err == nil {
			manifestPath = defaultPath
		} else {
			// Try migrating from legacy locations.
			migratedPath := migrateLegacyManifestFile(defaultPath)
			if migratedPath != "" {
				manifestPath = migratedPath
			} else {
				return nil, "", fmt.Errorf("no plugin manifest found")
			}
		}
	}

	data, err := os.ReadFile(manifestPath) // #nosec G304 - Manifest path controlled by application
	if err != nil {
		return nil, "", fmt.Errorf("failed to read plugin manifest: %w", err)
	}

	var lock PluginManifest
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, "", fmt.Errorf("failed to parse plugin manifest: %w", err)
	}

	return &lock, manifestPath, nil
}

// loadOrCreatePluginManifest loads or creates a plugin manifest.
// Always succeeds by creating a new lock if one doesn't exist.
func loadOrCreatePluginManifest() (lock *PluginManifest, manifestPath string) {
	lock, manifestPath, err := loadPluginManifest()
	if err == nil {
		return lock, manifestPath
	}

	// Create new manifest at the default path.
	manifestPath = pluginManifestPath
	if manifestPath == "" {
		manifestPath = getDefaultManifestFilePath()
	}

	lock = &PluginManifest{
		ExternalPlugins: make(map[string]*ExternalPluginMeta),
	}

	return lock, manifestPath
}

// savePluginManifest saves the plugin manifest.
func savePluginManifest(path string, lock *PluginManifest) error {
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plugin lock: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	return nil
}

// createManagerFromManifest creates a plugin manager from a manifest.
func createManagerFromManifest(lock *PluginManifest) *manager.Manager {
	mgr := manager.NewBuilder().Build()

	if lock == nil {
		return mgr
	}

	// Register external plugins using their actual names.
	if lock.ExternalPlugins != nil {
		for _, meta := range lock.ExternalPlugins {
			// Use the plugin's actual name (from metadata) not the manifest key.
			pluginName := meta.Name
			if pluginName == "" {
				// Fallback: query the plugin if name is missing.
				pluginName, _, _, _, _ = queryPluginMetadata(meta.Path)
			}

			// Use plugin's description if available.
			desc := meta.Description
			if desc == "" {
				desc = fmt.Sprintf("External plugin (source: %s)", meta.Source)
			}

			if err := mgr.RegisterExternalPlugin(pluginName, meta.Type, meta.Path, desc); err != nil {
				// Silently ignore registration errors.
				continue
			}
		}
	}

	return mgr
}

// loadAndApplyPluginManifest loads the plugin manifest and registers any external
// plugins into the shared plugin manager.
func loadAndApplyPluginManifest() error {
	lock, _, err := loadPluginManifest()
	if err != nil {
		return err
	}

	if lock != nil {
		registerExternalPluginsFromManifest(lock, true, false)
	}

	return nil
}

// registerExternalPluginsFromManifest registers all external plugins from
// the manifest into the shared plugin manager. Optionally resolves
// relative paths to absolute.
//
// Manifest entries whose binary no longer exists on disk (typical when
// a user `rm`'d the binary without `tinct plugins delete`) are silently
// skipped at runtime — the dangling entry is not registered, so it
// doesn't pollute `tinct plugins list` and doesn't trigger noisy
// "Failed to register external plugin 'X'" warnings on every CLI
// invocation. The manifest file on disk is not rewritten here: explicit
// removal is `tinct plugins delete <name>`. Pass verbose=true to log
// what was pruned.
func registerExternalPluginsFromManifest(manifest *PluginManifest, resolveAbsolutePaths, verbose bool) {
	if manifest == nil || manifest.ExternalPlugins == nil {
		return
	}

	for _, meta := range manifest.ExternalPlugins {
		if missing, why := isManifestEntryMissing(meta); missing {
			if verbose {
				fmt.Fprintf(os.Stderr, "Skipping stale manifest entry '%s': %s\n", meta.Name, why)
			}
			continue
		}
		if err := registerExternalPlugin(meta, resolveAbsolutePaths, verbose); err != nil {
			// An outdated plugin is always worth reporting: it silently
			// disappears from the registry, so the user would otherwise
			// only see "unknown plugin" and chase the wrong problem.
			var incompatible *manager.IncompatiblePluginError
			if errors.As(err, &incompatible) {
				fmt.Fprintf(os.Stderr, "⊘ Skipping plugin %q: %v\n", meta.Name, err)
				fmt.Fprintf(os.Stderr, "   Update it with: tinct plugins update %s\n", meta.Name)
				continue
			}
			if verbose {
				fmt.Fprintf(os.Stderr, " Failed to register external plugin '%s': %v\n", meta.Name, err)
			}
		}
	}
}

// isManifestEntryMissing reports whether the binary a manifest entry
// points at no longer exists on disk (or has an empty path). Used to
// auto-prune dangling entries from the runtime registration path.
//
// Returns (missing=true, reason) when the entry should be skipped.
// URL-based paths (http://, https://) are trusted at registration
// time and not stat'd — a non-existent URL surfaces later when the
// plugin is actually invoked.
func isManifestEntryMissing(meta *ExternalPluginMeta) (missing bool, reason string) {
	if meta == nil || meta.Path == "" {
		return true, "empty path"
	}
	if strings.HasPrefix(meta.Path, "http://") || strings.HasPrefix(meta.Path, "https://") {
		return false, ""
	}
	resolved := meta.Path
	if !filepath.IsAbs(resolved) {
		if abs, err := filepath.Abs(resolved); err == nil {
			resolved = abs
		}
	}
	if _, err := os.Stat(resolved); err != nil {
		if os.IsNotExist(err) {
			return true, fmt.Sprintf("binary not found at %s", resolved)
		}
		return true, fmt.Sprintf("cannot stat %s: %v", resolved, err)
	}
	return false, ""
}

// registerExternalPlugin registers a single external plugin into the shared manager.
func registerExternalPlugin(meta *ExternalPluginMeta, resolveAbsolutePaths, _ bool) error {
	// Use the plugin's actual name from metadata.
	pluginName := meta.Name
	if pluginName == "" {
		// Fallback: query the plugin if name is missing.
		pluginName, _, _, _, _ = queryPluginMetadata(meta.Path)
		if pluginName == "" {
			return fmt.Errorf("unable to determine plugin name")
		}
	}

	// Use plugin's description if available.
	desc := meta.Description
	if desc == "" {
		desc = fmt.Sprintf("External plugin (source: %s)", meta.Source)
	}

	// Convert relative paths to absolute if requested.
	pluginPath := meta.Path
	if resolveAbsolutePaths && !filepath.IsAbs(pluginPath) {
		absPath, err := filepath.Abs(pluginPath)
		if err != nil {
			return fmt.Errorf("failed to resolve absolute path: %w", err)
		}
		pluginPath = absPath
	}

	// Register the plugin.
	return sharedPluginManager.RegisterExternalPlugin(pluginName, meta.Type, pluginPath, desc)
}

// configureExternalPlugin applies additional configuration to an external plugin
// (dry-run mode, plugin-specific arguments, etc.)
// applyExternalPluginArgs forwards plugin args to every external plugin
// registered from the manifest. Unlike configureExternalPlugin this only
// sets arguments — extract writes no files, so there is no dry-run state
// to apply. Call it before validating plugins so an optional Validator
// sees the same args Generate() will receive.
func applyExternalPluginArgs(pluginArgs map[string]string, verbose bool) {
	if len(pluginArgs) == 0 {
		return
	}

	lock, _, err := loadPluginManifest()
	if err != nil || lock == nil {
		return
	}

	for _, meta := range lock.ExternalPlugins {
		pluginName := meta.Name
		if pluginName == "" {
			pluginName, _, _, _, _ = queryPluginMetadata(meta.Path)
		}
		if pluginName == "" {
			continue
		}

		argsJSON, ok := pluginArgs[pluginName]
		if !ok {
			continue
		}

		if err := setPluginArgs(sharedPluginManager, pluginName, meta.Type, argsJSON); err != nil && verbose {
			fmt.Fprintf(os.Stderr, " Failed to set args for plugin '%s': %v\n", pluginName, err)
		}
	}
}

func configureExternalPlugin(pluginName, pluginType string, dryRun bool, pluginArgs map[string]string, verbose bool) {
	// Set dry-run mode if applicable.
	if err := setPluginDryRun(sharedPluginManager, pluginName, pluginType, dryRun); err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, " Failed to set dry-run for plugin '%s': %v\n", pluginName, err)
		}
	}

	// Set plugin args if provided.
	if argsJSON, ok := pluginArgs[pluginName]; ok {
		if err := setPluginArgs(sharedPluginManager, pluginName, pluginType, argsJSON); err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, " Failed to set args for plugin '%s': %v\n", pluginName, err)
			}
		}
	}
}
