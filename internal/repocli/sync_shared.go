package repocli

import (
	"fmt"
	"time"

	"github.com/jmylchreest/tinct/internal/plugin/repository"
	"github.com/jmylchreest/tinct/internal/repomanager"
)

// PendingPlugin represents a plugin that was processed but couldn't be queried for metadata.
type PendingPlugin struct {
	PluginName    string
	PluginVersion string
	Platform      string
	DownloadURL   string
	Checksum      string
	Size          int64
	Released      time.Time
	ChangelogURL  string
}

// MetadataHydrationCache tracks plugins waiting for metadata and enables retroactive hydration.
// When one architecture succeeds in querying metadata, all other architectures of the same
// plugin+version can reuse that metadata.
type MetadataHydrationCache struct {
	// Cached metadata by "pluginName:version"
	metadata map[string]*repomanager.PluginMetadata

	// Pending plugins that failed metadata query, by "pluginName:version"
	pending map[string][]*PendingPlugin
}

// NewMetadataHydrationCache creates a new cache.
func NewMetadataHydrationCache() *MetadataHydrationCache {
	return &MetadataHydrationCache{
		metadata: make(map[string]*repomanager.PluginMetadata),
		pending:  make(map[string][]*PendingPlugin),
	}
}

// GetMetadata retrieves cached metadata for a plugin+version.
func (c *MetadataHydrationCache) GetMetadata(pluginName, version string) (*repomanager.PluginMetadata, bool) {
	key := fmt.Sprintf("%s:%s", pluginName, version)
	metadata, exists := c.metadata[key]
	return metadata, exists
}

// SetMetadata stores metadata and hydrates any pending plugins.
func (c *MetadataHydrationCache) SetMetadata(
	pluginName string,
	version string,
	metadata *repomanager.PluginMetadata,
	mgr *repomanager.ManifestManager,
	dryRun bool,
	verbose bool,
) int {

	key := fmt.Sprintf("%s:%s", pluginName, version)
	c.metadata[key] = metadata

	// Check if there are pending plugins for this key
	pendingList, hasPending := c.pending[key]
	if !hasPending {
		return 0
	}

	// Hydrate all pending plugins with this metadata
	hydrated := 0
	compatibility := repomanager.CalculateCompatibility(metadata.ProtocolVersion)

	for _, pending := range pendingList {
		if verbose {
			fmt.Printf("      Hydrating %s %s (%s) with cached metadata\n",
				pending.PluginName, pending.PluginVersion, pending.Platform)
		}

		// Create version entry with the metadata
		version := &repository.Version{
			Version:       pending.PluginVersion,
			Released:      pending.Released,
			Compatibility: compatibility,
			ChangelogURL:  pending.ChangelogURL,
			Downloads: map[string]*repository.Download{
				pending.Platform: {
					URL:       pending.DownloadURL,
					Checksum:  pending.Checksum,
					Size:      pending.Size,
					Available: true,
				},
			},
		}

		// Add to manifest
		if !dryRun {
			if err := mgr.AddOrUpdatePluginVersion(pending.PluginName, version); err != nil {
				if verbose {
					fmt.Printf("      Error hydrating: %v\n", err)
				}
				continue
			}

			// Update plugin metadata
			mgr.SetPluginMetadata(pending.PluginName, metadata)
		}

		hydrated++
	}

	// Clear pending list for this key
	delete(c.pending, key)

	return hydrated
}

// AddPending adds a plugin that failed metadata query to the pending list.
func (c *MetadataHydrationCache) AddPending(pending *PendingPlugin) {
	key := fmt.Sprintf("%s:%s", pending.PluginName, pending.PluginVersion)
	c.pending[key] = append(c.pending[key], pending)
}

// ProtocolVersionTracker tracks which plugins have failed protocol checks and at what version.
// This enables cascade filtering: if version X fails, all versions < X are automatically skipped.
type ProtocolVersionTracker struct {
	failures map[string]string // plugin name -> highest failing version
}

// NewProtocolVersionTracker creates a new tracker.
func NewProtocolVersionTracker() *ProtocolVersionTracker {
	return &ProtocolVersionTracker{
		failures: make(map[string]string),
	}
}

// ShouldSkip checks if a plugin version should be skipped based on protocol filtering.
// Returns true if this version should be skipped (either failed directly or older than a failure).
func (t *ProtocolVersionTracker) ShouldSkip(pluginName, pluginVersion, protocolVersion, minProtocol string) (bool, string) {
	// Check direct protocol version
	if minProtocol != "" && !repomanager.CheckProtocolVersion(protocolVersion, minProtocol) {
		// Record this failure
		t.RecordFailure(pluginName, pluginVersion)
		return true, fmt.Sprintf("protocol %s < minimum %s", protocolVersion, minProtocol)
	}

	// Check if there's a recorded failure for this plugin
	if failedVersion, exists := t.failures[pluginName]; exists {
		// If this version is older than or equal to the failed version, skip it
		if repomanager.CompareVersions(pluginVersion, failedVersion) <= 0 {
			return true, fmt.Sprintf("older than or equal to failed version %s", failedVersion)
		}
	}

	return false, ""
}

// RecordFailure records that a plugin version failed protocol check.
func (t *ProtocolVersionTracker) RecordFailure(pluginName, version string) {
	// Update if this is the highest failing version we've seen
	if existingVersion, exists := t.failures[pluginName]; exists {
		if repomanager.CompareVersions(version, existingVersion) > 0 {
			t.failures[pluginName] = version
		}
	} else {
		t.failures[pluginName] = version
	}
}

// ProcessGitHubSourceWithProtocol processes a GitHub sync source with protocol filtering.
func ProcessGitHubSourceWithProtocol(
	source *repomanager.SyncSource,
	client *repomanager.GitHubClient,
	mgr *repomanager.ManifestManager,
	minProtocolVersion string,
	tracker *ProtocolVersionTracker,
	hydrationCache *MetadataHydrationCache,
	skipQuery bool,
	dryRun bool,
	verbose bool,
) (added int, skipped int, errors int) {
	// Parse GitHub repo
	owner, repo, err := repomanager.ParseGitHubRepo(source.Repo)
	if err != nil {
		fmt.Printf("  Error: invalid github repo: %v\n", err)
		return 0, 0, 1
	}

	fmt.Printf("  Repository: %s\n", source.Repo)
	fmt.Printf("  Version: %s\n", source.Version)

	// Resolve version specifier
	releases, err := client.GetReleases(owner, repo, source.Version)
	if err != nil {
		fmt.Printf("  Error: failed to fetch releases: %v\n", err)
		return 0, 0, 1
	}

	fmt.Printf("  Found %d release(s)\n", len(releases))

	// Create filter
	filter := repomanager.NewFilter(source.Filter, source.Exclude)

	// Process each release (newest first due to sorting)
	for ri, release := range releases {
		if verbose {
			fmt.Printf("  [%d/%d] Release %s (%d assets)\n",
				ri+1, len(releases), release.TagName, len(release.Assets))
		}

		for _, asset := range release.Assets {
			match, pluginName := filter.Match(asset.Name)
			if !match {
				skipped++
				if verbose {
					fmt.Printf("    Skipped: %s (no match)\n", asset.Name)
				}
				continue
			}

			if verbose {
				fmt.Printf("    Processing: %s (plugin: %s)\n", asset.Name, pluginName)
			}

			// Parse version and platform from asset name
			pluginVersion, platform := repomanager.ParseAssetName(asset.Name)
			if pluginVersion == "" || platform == "" {
				if verbose {
					fmt.Printf("      Warning: could not parse version/platform from asset name\n")
				}
				errors++
				continue
			}

			// Calculate checksum
			if verbose {
				fmt.Printf("      Calculating checksum...\n")
			}
			checksum, size, err := repomanager.CalculateChecksum(asset.DownloadURL)
			if err != nil {
				fmt.Printf("      Error: failed to calculate checksum: %v\n", err)
				errors++
				continue
			}

			// Query plugin metadata (unless skipped)
			var metadata *repomanager.PluginMetadata
			var compatibility string

			if !skipQuery {
				// Check cache first for this plugin+version combination
				cachedMetadata, hasCached := hydrationCache.GetMetadata(pluginName, pluginVersion)

				if hasCached {
					// Reuse cached metadata from another architecture
					metadata = cachedMetadata
					compatibility = repomanager.CalculateCompatibility(metadata.ProtocolVersion)
					if verbose {
						fmt.Printf("      Using cached metadata from another architecture\n")
						fmt.Printf("      Protocol: %s, Compatibility: %s\n",
							metadata.ProtocolVersion, compatibility)
					}
				} else {
					if verbose {
						fmt.Printf("      Querying plugin metadata...\n")
					}
					metadata, err = repomanager.QueryPlugin(asset.DownloadURL)
					if err != nil {
						fmt.Printf("      Warning: query failed: %v\n", err)
						fmt.Printf("      Continuing without metadata...\n")

						// Add to pending list for retroactive hydration
						hydrationCache.AddPending(&PendingPlugin{
							PluginName:    pluginName,
							PluginVersion: pluginVersion,
							Platform:      platform,
							DownloadURL:   asset.DownloadURL,
							Checksum:      fmt.Sprintf("sha256:%s", checksum),
							Size:          size,
							Released:      release.PublishedAt,
							ChangelogURL:  release.URL,
						})
					} else {
						compatibility = repomanager.CalculateCompatibility(metadata.ProtocolVersion)
						if verbose {
							fmt.Printf("      Protocol: %s, Compatibility: %s\n",
								metadata.ProtocolVersion, compatibility)
						}

						// Store metadata and hydrate any pending plugins
						hydratedCount := hydrationCache.SetMetadata(pluginName, pluginVersion, metadata, mgr, dryRun, verbose)
						if hydratedCount > 0 {
							added += hydratedCount
							if verbose {
								fmt.Printf("      Hydrated %d pending plugin(s)\n", hydratedCount)
							}
						}
					}
				}

				// Check protocol version with cascade if we have metadata
				if metadata != nil {
					if shouldSkip, reason := tracker.ShouldSkip(pluginName, pluginVersion, metadata.ProtocolVersion, minProtocolVersion); shouldSkip {
						if verbose {
							fmt.Printf("      Skipped: %s\n", reason)
						}
						skipped++
						continue
					}

					// Update plugin metadata
					if !dryRun {
						mgr.SetPluginMetadata(pluginName, metadata)
					}
				}
			}

			// Create version entry
			version := &repository.Version{
				Version:       pluginVersion,
				Released:      release.PublishedAt,
				Compatibility: compatibility,
				ChangelogURL:  release.URL,
				Downloads: map[string]*repository.Download{
					platform: {
						URL:       asset.DownloadURL,
						Checksum:  fmt.Sprintf("sha256:%s", checksum),
						Size:      size,
						Available: true,
					},
				},
			}

			// Add to manifest
			if !dryRun {
				if err := mgr.AddOrUpdatePluginVersion(pluginName, version); err != nil {
					fmt.Printf("      Error: %v\n", err)
					errors++
					continue
				}
			}

			if verbose {
				fmt.Printf("      ✓ Added: %s %s (%s)\n", pluginName, pluginVersion, platform)
			}
			added++
		}
	}

	if !verbose {
		fmt.Printf("  Added %d plugins from %d release(s)\n", added, len(releases))
	}

	return added, skipped, errors
}

// ProcessURLSourceWithProtocol processes a URL sync source with protocol filtering.
func ProcessURLSourceWithProtocol(
	source *repomanager.SyncSource,
	mgr *repomanager.ManifestManager,
	minProtocolVersion string,
	tracker *ProtocolVersionTracker,
	hydrationCache *MetadataHydrationCache,
	skipQuery bool,
	dryRun bool,
	verbose bool,
) (added int, errors int) {

	fmt.Printf("  URL: %s\n", source.URL)
	fmt.Printf("  Plugin: %s (%s)\n", source.Plugin, source.PluginType)
	fmt.Printf("  Platform: %s\n", source.Platform)

	// Calculate checksum
	if verbose {
		fmt.Printf("  Calculating checksum...\n")
	}
	checksum, size, err := repomanager.CalculateChecksum(source.URL)
	if err != nil {
		fmt.Printf("  Error: failed to calculate checksum: %v\n", err)
		return 0, 1
	}

	// Determine version
	pluginVersion := source.Version
	var metadata *repomanager.PluginMetadata
	var compatibility string

	// Auto-detect version if "-" or query for metadata
	if pluginVersion == "-" || !skipQuery {
		// For URL sources, check cache if we have a version
		var hasCached bool
		var cachedMetadata *repomanager.PluginMetadata

		if pluginVersion != "-" {
			cachedMetadata, hasCached = hydrationCache.GetMetadata(source.Plugin, pluginVersion)
		}

		if hasCached {
			// Reuse cached metadata from another platform
			metadata = cachedMetadata
			compatibility = repomanager.CalculateCompatibility(metadata.ProtocolVersion)
			if verbose {
				fmt.Printf("  Using cached metadata from another platform\n")
				fmt.Printf("  Protocol: %s, Compatibility: %s\n",
					metadata.ProtocolVersion, compatibility)
			}
		} else {
			if verbose {
				fmt.Printf("  Querying plugin metadata...\n")
			}
			metadata, err = repomanager.QueryPlugin(source.URL)
			if err != nil {
				if pluginVersion == "-" {
					fmt.Printf("  Error: version auto-detection failed: %v\n", err)
					return 0, 1
				}
				fmt.Printf("  Warning: query failed: %v\n", err)
				fmt.Printf("  Continuing without metadata...\n")

				// Add to pending list for retroactive hydration (if we have a version)
				if pluginVersion != "" && pluginVersion != "-" {
					hydrationCache.AddPending(&PendingPlugin{
						PluginName:    source.Plugin,
						PluginVersion: pluginVersion,
						Platform:      source.Platform,
						DownloadURL:   source.URL,
						Checksum:      fmt.Sprintf("sha256:%s", checksum),
						Size:          size,
						Released:      time.Now(),
					})
				}
			} else {
				// Use detected version if auto-detect
				if pluginVersion == "-" {
					pluginVersion = metadata.Version
					if verbose {
						fmt.Printf("  Detected version: %s\n", pluginVersion)
					}
				}

				compatibility = repomanager.CalculateCompatibility(metadata.ProtocolVersion)
				if verbose {
					fmt.Printf("  Protocol: %s, Compatibility: %s\n",
						metadata.ProtocolVersion, compatibility)
				}

				// Store metadata and hydrate any pending plugins
				hydratedCount := hydrationCache.SetMetadata(source.Plugin, pluginVersion, metadata, mgr, dryRun, verbose)
				if hydratedCount > 0 && verbose {
					fmt.Printf("  Hydrated %d pending plugin(s)\n", hydratedCount)
				}
			}
		}

		// Check protocol version with cascade if we have metadata
		if metadata != nil {
			if shouldSkip, reason := tracker.ShouldSkip(source.Plugin, pluginVersion, metadata.ProtocolVersion, minProtocolVersion); shouldSkip {
				fmt.Printf("  Skipped: %s\n", reason)
				return 0, 0
			}

			// Update plugin metadata
			if !dryRun {
				mgr.SetPluginMetadata(source.Plugin, metadata)
			}
		}
	}

	// Validate we have a version
	if pluginVersion == "" || pluginVersion == "-" {
		fmt.Printf("  Error: could not determine plugin version\n")
		return 0, 1
	}

	// Create version entry
	version := &repository.Version{
		Version:       pluginVersion,
		Released:      time.Now(),
		Compatibility: compatibility,
		Downloads: map[string]*repository.Download{
			source.Platform: {
				URL:       source.URL,
				Checksum:  fmt.Sprintf("sha256:%s", checksum),
				Size:      size,
				Runtime:   source.Runtime,
				Available: true,
			},
		},
	}

	// Add to manifest
	if !dryRun {
		if err := mgr.AddOrUpdatePluginVersion(source.Plugin, version); err != nil {
			fmt.Printf("  Error: %v\n", err)
			return 0, 1
		}
	}

	fmt.Printf("  ✓ Added: %s %s (%s)\n", source.Plugin, pluginVersion, source.Platform)
	return 1, 0
}
