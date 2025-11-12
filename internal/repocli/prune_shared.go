package repocli

import (
	"fmt"
	"time"

	"github.com/jmylchreest/tinct/internal/repomanager"
)

// PruneStats tracks pruning statistics.
type PruneStats struct {
	Checked      int
	Unavailable  int
	Removed      int
	FilterFailed int
}

// PruneManifest performs comprehensive pruning on a manifest.
// It validates downloads against filters, marks unavailable URLs, and removes old entries.
func PruneManifest(
	mgr *repomanager.ManifestManager,
	removeAfterDuration time.Duration,
	dryRun bool,
	verbose bool,
) *PruneStats {
	manifest := mgr.GetManifest()
	verifier := repomanager.NewVerifier()
	validator := repomanager.NewPruneValidator()

	stats := &PruneStats{}

	// Iterate through all plugins
	for pluginName, plugin := range manifest.Plugins {
		for vi := len(plugin.Versions) - 1; vi >= 0; vi-- {
			version := &plugin.Versions[vi]

			platformsToRemove := []string{}

			for platform, download := range version.Downloads {
				stats.Checked++

				// Step 1: Validate against filter patterns (remove invalid entries immediately)
				shouldKeep, filterReason := validator.ShouldKeepDownload(download.URL)
				if !shouldKeep {
					if verbose {
						fmt.Printf("  Filter validation failed: %s %s (%s) - %s\n",
							pluginName, version.Version, platform, filterReason)
					}
					platformsToRemove = append(platformsToRemove, platform)
					stats.FilterFailed++
					stats.Removed++
					continue
				}

				// Step 2: Check if should remove (already unavailable and past threshold)
				if removeAfterDuration > 0 && download.UnavailableSince != nil {
					unavailableDuration := time.Since(*download.UnavailableSince)
					if unavailableDuration > removeAfterDuration {
						if verbose {
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
					if verbose {
						fmt.Printf("  ✗ Unavailable: %s %s (%s) - %s\n",
							pluginName, version.Version, platform, reason)
					}

					// Mark as unavailable with timestamp
					if !dryRun {
						download.Available = false
						now := time.Now()
						if download.UnavailableSince == nil {
							download.UnavailableSince = &now
						}
						download.UnavailableReason = reason
					}
				} else {
					// Mark as available and clear unavailable fields
					if !dryRun {
						download.Available = true
						now := time.Now()
						download.LastVerified = &now
						download.UnavailableSince = nil
						download.UnavailableReason = ""
					}
				}
			}

			// Remove platforms marked for removal
			if !dryRun {
				for _, platform := range platformsToRemove {
					delete(version.Downloads, platform)
				}

				// Clean up versions with no downloads
				if len(version.Downloads) == 0 {
					plugin.Versions = append(plugin.Versions[:vi], plugin.Versions[vi+1:]...)
					if verbose {
						fmt.Printf("  Removed version %s (no downloads left)\n", version.Version)
					}
				}
			}
		}

		// Clean up plugins with no versions
		if !dryRun && len(plugin.Versions) == 0 {
			delete(manifest.Plugins, pluginName)
			if verbose {
				fmt.Printf("  Removed plugin %s (no versions left)\n", pluginName)
			}
		}
	}

	return stats
}
