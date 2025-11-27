// Package dbus_kde provides D-Bus helpers for KDE Plasma configuration reload.
package dbus_kde

import (
	"context"
	"fmt"
	"slices"

	"github.com/jmylchreest/tinct/pkg/dbus"
)

const (
	kwinService = "org.kde.KWin"
	kwinPath    = "/KWin"
	kwinIface   = "org.kde.KWin"

	plasmashellService = "org.kde.plasmashell"
	plasmashellPath    = "/PlasmaShell"
	plasmashellIface   = "org.kde.PlasmaShell"
)

// ReconfigureKWin triggers KWin to reload its configuration via D-Bus.
// This is equivalent to running: qdbus org.kde.KWin /KWin reconfigure.
func ReconfigureKWin(ctx context.Context) (bool, error) {
	if !dbus.IsAvailable() {
		return false, nil
	}

	conn, err := dbus.SessionBus(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to connect to session bus: %w", err)
	}
	defer conn.Close()

	// Check if KWin is running
	names, err := conn.ListNames(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to list bus names: %w", err)
	}

	kwinRunning := slices.Contains(names, kwinService)

	if !kwinRunning {
		return false, nil // KWin not running, not an error
	}

	obj := conn.Object(kwinService, kwinPath)

	// Call reconfigure method
	err = obj.Call(ctx, kwinIface+".reconfigure")
	if err != nil {
		return false, fmt.Errorf("failed to reconfigure KWin: %w", err)
	}

	return true, nil
}

// ReconfigurePlasmaShell triggers Plasma Shell to reload its configuration via D-Bus.
func ReconfigurePlasmaShell(ctx context.Context) (bool, error) {
	if !dbus.IsAvailable() {
		return false, nil
	}

	conn, err := dbus.SessionBus(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to connect to session bus: %w", err)
	}
	defer conn.Close()

	// Check if plasmashell is running
	names, err := conn.ListNames(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to list bus names: %w", err)
	}

	plasmaRunning := slices.Contains(names, plasmashellService)

	if !plasmaRunning {
		return false, nil // Plasma Shell not running, not an error
	}

	obj := conn.Object(plasmashellService, plasmashellPath)

	// Evaluate JavaScript to reload theme (similar to how plasma-apply-colorscheme works)
	script := "for (var i in panels()) { panels()[i].reload(); }"
	_, err = obj.CallWithReturn(ctx, "org.kde.PlasmaShell.evaluateScript", script)
	if err != nil {
		return false, fmt.Errorf("failed to reload Plasma Shell: %w", err)
	}

	return true, nil
}

// ReloadTheme triggers both KWin and Plasma Shell to reload their configurations.
// This ensures the color scheme is applied across all KDE components.
func ReloadTheme(ctx context.Context) (bool, error) {
	if !dbus.IsAvailable() {
		return false, nil
	}

	kwinReloaded, kwinErr := ReconfigureKWin(ctx)
	plasmaReloaded, plasmaErr := ReconfigurePlasmaShell(ctx)

	// Return success if at least one component reloaded
	if kwinReloaded || plasmaReloaded {
		return true, nil
	}

	// If both failed, return the error
	if kwinErr != nil {
		return false, kwinErr
	}
	if plasmaErr != nil {
		return false, plasmaErr
	}

	return false, nil
}
