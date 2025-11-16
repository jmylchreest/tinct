package repocli

import (
	"fmt"
	"sort"
	"time"

	"github.com/jmylchreest/tinct/internal/repomanager"
)

// PruneStats tracks pruning statistics.
type PruneStats struct {
	Checked      int
	Unavailable  int
	Removed      int
	FilterFailed int
	Incompatible int
	OldVersions  int
}

// PruneOptions contains configuration for pruning operations.
type PruneOptions struct {
	RemoveAfterDuration time.Duration
	PruneIncompatible   bool
	KeepRecent          int
	DryRun              bool
	Verbose             bool
}

// PruneManifest performs comprehensive pruning on a manifest.
// It validates downloads against filters, marks unavailable URLs, and removes old entries.
func PruneManifest(
	mgr *repomanager.ManifestManager,
	removeAfterDuration time.Duration,
	dryRun bool,
	verbose bool,
) *PruneStats {
	return PruneManifestWithOptions(mgr, &PruneOptions{
		RemoveAfterDuration: removeAfterDuration,
		PruneIncompatible:   false,
		KeepRecent:          0,
		DryRun:              dryRun,
		Verbose:             verbose,
	})
}

// PruneManifestWithOptions performs comprehensive pruning on a manifest with advanced options.
// It validates downloads against filters, marks unavailable URLs, removes old entries,
// prunes incompatible plugins, and limits the number of versions kept per plugin.
func PruneManifestWithOptions(
	mgr *repomanager.ManifestManager,
	opts *PruneOptions,
) *PruneStats {

	manifest := mgr.GetManifest()
	verifier := repomanager.NewVerifier()
	validator := repomanager.NewPruneValidator()

	stats := &PruneStats{}

	// Iterate through all plugins
	for pluginName, plugin := range manifest.Plugins {
		// Track versions to remove for this plugin
		versionsToRemove := []int{}

		for vi := len(plugin.Versions) - 1; vi >= 0; vi-- {
			version := &plugin.Versions[vi]

			// Step 0: Check protocol compatibility if requested
			if opts.PruneIncompatible && version.Compatibility != "" {
				compatible, err := repomanager.IsProtocolCompatible(version.Compatibility)
				if err != nil || !compatible {
					if opts.Verbose {
						var reason string
						if err != nil {
							reason = err.Error()
						} else {
							reason = fmt.Sprintf("protocol version %s incompatible with tinct %s",
								version.Compatibility, "0.0.1")
						}
						fmt.Printf("  Removing incompatible version: %s %s - %s\n",
							pluginName, version.Version, reason)
					}
					versionsToRemove = append(versionsToRemove, vi)
					stats.Incompatible++
					stats.Removed++
					continue
				}
			}

			platformsToRemove := []string{}

			for platform, download := range version.Downloads {
				stats.Checked++

				// Step 1: Validate against filter patterns (remove invalid entries immediately)
				shouldKeep, filterReason := validator.ShouldKeepDownload(download.URL)
				if !shouldKeep {
					if opts.Verbose {
						fmt.Printf("  Filter validation failed: %s %s (%s) - %s\n",
							pluginName, version.Version, platform, filterReason)
					}
					platformsToRemove = append(platformsToRemove, platform)
					stats.FilterFailed++
					stats.Removed++
					continue
				}

				// Step 2: Check if should remove (already unavailable and past threshold)
				if opts.RemoveAfterDuration > 0 && download.UnavailableSince != nil {
					unavailableDuration := time.Since(*download.UnavailableSince)
					if unavailableDuration > opts.RemoveAfterDuration {
						if opts.Verbose {
							fmt.Printf("  Removing: %s %s (%s) - unavailable for %v\n",
								pluginName, version.Version, platform, unavailableDuration.Round(time.Hour))
						}
						platformsToRemove = append(platformsToRemove, platform)
						stats.Removed++
						continue
					}
				}

				// Step 3: Verify URL availability
				available, reason := verifier.VerifyURL(download.URL)

				if !available {
					stats.Unavailable++
					if opts.Verbose {
						fmt.Printf("  ✗ Unavailable: %s %s (%s) - %s\n",
							pluginName, version.Version, platform, reason)
					}

					// Mark as unavailable with timestamp
					if !opts.DryRun {
						// Only mark dirty if availability status changed
						wasAvailable := download.Available
						download.Available = false
						now := time.Now()
						if download.UnavailableSince == nil {
							download.UnavailableSince = &now
						}
						download.UnavailableReason = reason

						// Mark dirty only if status changed from available to unavailable
						if wasAvailable {
							mgr.MarkDirty()
						}
					}
				} else {
					// Mark as available and clear unavailable fields
					if !opts.DryRun {
						// Only mark dirty if availability status changed
						wasUnavailable := !download.Available || download.UnavailableSince != nil

						if wasUnavailable {
							// Status changed from unavailable to available
							download.Available = true
							download.UnavailableSince = nil
							download.UnavailableReason = ""
							mgr.MarkDirty()
						}
						// Don't update any timestamps if status hasn't changed
					}
				}
			}

			// Remove platforms marked for removal
			if !opts.DryRun && len(platformsToRemove) > 0 {
				for _, platform := range platformsToRemove {
					delete(version.Downloads, platform)
				}
				mgr.MarkDirty()

				// Clean up versions with no downloads
				if len(version.Downloads) == 0 {
					versionsToRemove = append(versionsToRemove, vi)
					if opts.Verbose {
						fmt.Printf("  Removed version %s (no downloads left)\n", version.Version)
					}
				}
			}
		}

		// Remove versions marked for removal
		if !opts.DryRun && len(versionsToRemove) > 0 {
			// Sort indices in descending order to avoid index issues when removing
			// We need to sort because versionsToRemove was built in reverse order
			sort.Sort(sort.Reverse(sort.IntSlice(versionsToRemove)))

			for _, vi := range versionsToRemove {
				if vi < len(plugin.Versions) {
					plugin.Versions = append(plugin.Versions[:vi], plugin.Versions[vi+1:]...)
				}
			}
			mgr.MarkDirty()
		}

		// Step 4: Limit to most recent versions if requested
		if opts.KeepRecent > 0 && len(plugin.Versions) > opts.KeepRecent {
			oldVersionCount := len(plugin.Versions) - opts.KeepRecent
			if opts.Verbose {
				fmt.Printf("  Keeping only %d most recent versions of %s (removing %d old versions)\n",
					opts.KeepRecent, pluginName, oldVersionCount)
			}

			if !opts.DryRun {
				plugin.Versions = plugin.Versions[:opts.KeepRecent]
				mgr.MarkDirty()
			}

			stats.OldVersions += oldVersionCount
			stats.Removed += oldVersionCount
		}

		// Clean up plugins with no versions
		if !opts.DryRun && len(plugin.Versions) == 0 {
			delete(manifest.Plugins, pluginName)
			mgr.MarkDirty()
			if opts.Verbose {
				fmt.Printf("  Removed plugin %s (no versions left)\n", pluginName)
			}
		}
	}

	return stats
}
