// Package hyprland provides an output plugin for Hyprland window manager colour themes.
package hyprland

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jmylchreest/tinct/internal/plugin/output"
)

// wallpaperManager represents different wallpaper management tools
type wallpaperManager string

const (
	wallpaperManagerHyprpaper wallpaperManager = "hyprpaper"
	wallpaperManagerSwaybg    wallpaperManager = "swaybg"
	wallpaperManagerSwww      wallpaperManager = "swww"
	wallpaperManagerNone      wallpaperManager = "none"
)

// detectWallpaperManager detects which wallpaper manager is running
func (p *Plugin) detectWallpaperManager(ctx context.Context) wallpaperManager {
	// Check for hyprpaper via hyprctl
	cmd := exec.CommandContext(ctx, "hyprctl", "hyprpaper", "listloaded")
	if err := cmd.Run(); err == nil {
		return wallpaperManagerHyprpaper
	}

	// Check for swww daemon
	cmd = exec.CommandContext(ctx, "pgrep", "-x", "swww-daemon")
	if err := cmd.Run(); err == nil {
		return wallpaperManagerSwww
	}

	// Check for swaybg
	cmd = exec.CommandContext(ctx, "pgrep", "-x", "swaybg")
	if err := cmd.Run(); err == nil {
		return wallpaperManagerSwaybg
	}

	return wallpaperManagerNone
}

// setWallpaper sets the wallpaper using the detected wallpaper manager
func (p *Plugin) setWallpaper(ctx context.Context, execCtx output.ExecutionContext) error {
	manager := p.detectWallpaperManager(ctx)

	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Detected wallpaper manager: %s\n", manager)
	}

	switch manager {
	case wallpaperManagerHyprpaper:
		return p.setWallpaperHyprpaper(ctx, execCtx.WallpaperPath)
	case wallpaperManagerSwww:
		return p.setWallpaperSwww(ctx, execCtx.WallpaperPath)
	case wallpaperManagerSwaybg:
		return p.setWallpaperSwaybg(ctx, execCtx.WallpaperPath)
	default:
		return fmt.Errorf("no supported wallpaper manager detected (tried: hyprpaper, swww, swaybg)")
	}
}

// setWallpaperHyprpaper sets the wallpaper using hyprpaper
// Reuses existing monitor assignments from listactive to avoid querying all monitors
func (p *Plugin) setWallpaperHyprpaper(ctx context.Context, wallpaperPath string) error {
	// Make the path absolute
	absPath, err := filepath.Abs(wallpaperPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get current wallpaper assignments (monitors and/or wildcard)
	assignments, err := p.getActiveWallpaperAssignments(ctx)
	if err != nil {
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Warning: failed to get active assignments, using wildcard: %v\n", err)
		}
		// Fallback to wildcard if we can't get assignments
		assignments = []string{""}
	}

	// If no assignments found, use wildcard
	if len(assignments) == 0 {
		assignments = []string{""}
	}

	// First, unload all existing wallpapers to clear any cached images
	// This is crucial for symlinks - hyprpaper may cache the resolved target
	cmd := exec.CommandContext(ctx, "hyprctl", "hyprpaper", "unload", "all")
	cmd.Run() // Ignore errors - wallpapers might not be loaded

	// Preload the new wallpaper
	cmd = exec.CommandContext(ctx, "hyprctl", "hyprpaper", "preload", absPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to preload wallpaper: %w (output: %s)", err, string(output))
	}

	// Set the wallpaper using the same assignments (monitors or wildcard)
	successCount := 0
	for _, monitor := range assignments {
		cmd = exec.CommandContext(ctx, "hyprctl", "hyprpaper", "wallpaper", monitor+","+absPath)
		if err := cmd.Run(); err != nil {
			if p.verbose {
				monitorName := monitor
				if monitorName == "" {
					monitorName = "wildcard"
				}
				fmt.Fprintf(os.Stderr, "   Warning: failed to set wallpaper on %s: %v\n", monitorName, err)
			}
			// Continue to try other monitors even if one fails
			continue
		}
		successCount++
		if p.verbose {
			monitorName := monitor
			if monitorName == "" {
				monitorName = "wildcard"
			}
			fmt.Fprintf(os.Stderr, "   Set wallpaper on %s\n", monitorName)
		}
	}

	if p.verbose {
		if len(assignments) == 1 && assignments[0] == "" {
			fmt.Fprintf(os.Stderr, "   Set wallpaper using hyprpaper (wildcard): %s\n", absPath)
		} else {
			fmt.Fprintf(os.Stderr, "   Set wallpaper using hyprpaper on %d monitor(s): %s\n", len(assignments), absPath)
		}
	}

	return nil
}

// getActiveWallpaperAssignments retrieves the current monitor assignments from listactive
// Returns a list of monitor names (or empty string for wildcard)
// Example output from hyprctl: "DP-2 = /path", " = /path" (wildcard)
func (p *Plugin) getActiveWallpaperAssignments(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "hyprctl", "hyprpaper", "listactive")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to query active wallpapers: %w", err)
	}

	// Parse the output line by line
	// Format: "MONITOR = /path/to/wallpaper"
	// Wildcard format: " = /path/to/wallpaper" (empty monitor name)
	lines := strings.Split(string(output), "\n")
	assignments := make([]string, 0, len(lines))

	for _, line := range lines {
		// Don't trim yet - we need to preserve leading space for wildcard detection
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Split on " = " to get monitor name
		parts := strings.SplitN(line, " = ", 2)
		if len(parts) != 2 {
			continue
		}

		// Now trim the monitor name (will be empty string for wildcard " = /path")
		monitor := strings.TrimSpace(parts[0])
		// Empty string means wildcard assignment
		assignments = append(assignments, monitor)
	}

	return assignments, nil
}

// setWallpaperSwww sets the wallpaper using swww
func (p *Plugin) setWallpaperSwww(ctx context.Context, wallpaperPath string) error {
	// Make the path absolute
	absPath, err := filepath.Abs(wallpaperPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Set the wallpaper with a nice transition
	cmd := exec.CommandContext(ctx, "swww", "img", absPath, "--transition-type", "fade")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set wallpaper: %w (output: %s)", err, string(output))
	}

	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Set wallpaper using swww: %s\n", absPath)
	}

	return nil
}

// setWallpaperSwaybg sets the wallpaper using swaybg
func (p *Plugin) setWallpaperSwaybg(ctx context.Context, wallpaperPath string) error {
	// Make the path absolute
	absPath, err := filepath.Abs(wallpaperPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Kill existing swaybg instances
	exec.CommandContext(ctx, "pkill", "swaybg").Run()

	// Start new swaybg instance
	cmd := exec.CommandContext(ctx, "swaybg", "-i", absPath, "-m", "fill")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start swaybg: %w", err)
	}

	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Set wallpaper using swaybg: %s\n", absPath)
	}

	return nil
}
