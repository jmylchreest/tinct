// Package hyprland provides an output plugin for Hyprland window manager colour themes.
package hyprland

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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
// Uses the same pattern as hyprpaper.conf: wallpaper = , /path/to/image
// This sets the wallpaper on all monitors at once
func (p *Plugin) setWallpaperHyprpaper(ctx context.Context, wallpaperPath string) error {
	// Make the path absolute
	absPath, err := filepath.Abs(wallpaperPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Unload the wallpaper first if it's already loaded
	// This is important for symlinks that may point to different files
	// hyprpaper caches by path, so if ~/.wallpaper symlink changes,
	// we need to unload the old cached image first
	cmd := exec.CommandContext(ctx, "hyprctl", "hyprpaper", "unload", absPath)
	// Ignore errors - the wallpaper might not be loaded yet
	cmd.Run()

	// Preload the wallpaper (now gets the fresh image)
	cmd = exec.CommandContext(ctx, "hyprctl", "hyprpaper", "preload", absPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to preload wallpaper: %w (output: %s)", err, string(output))
	}

	// Set the wallpaper on all monitors using the ", path" syntax
	// This matches the hyprpaper.conf pattern: wallpaper = , ~/.wallpaper
	cmd = exec.CommandContext(ctx, "hyprctl", "hyprpaper", "wallpaper", ","+absPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set wallpaper: %w (output: %s)", err, string(output))
	}

	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Set wallpaper using hyprpaper on all monitors: %s\n", absPath)
	}

	return nil
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
