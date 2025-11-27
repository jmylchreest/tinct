// Package dbus_gtk provides helpers for GTK theme reloading via GNOME Settings Daemon D-Bus.
package dbus_gtk

import (
	"context"
	"fmt"
	"slices"

	"github.com/jmylchreest/tinct/pkg/dbus"
)

const (
	xsettingsService = "org.gnome.SettingsDaemon.XSettings"
	xsettingsPath    = "/org/gnome/SettingsDaemon/XSettings"
	xsettingsIface   = "org.gnome.SettingsDaemon.XSettings"
)

// ReloadTheme triggers GTK applications to reload their theme via GNOME Settings Daemon.
// This only works on GNOME desktop with gsd-xsettings running.
// On other desktops (Hyprland, Sway, etc.), GTK apps must be restarted to pick up theme changes.
// Returns true if D-Bus reload succeeded, false otherwise.
func ReloadTheme(ctx context.Context) (bool, error) {
	if !dbus.IsAvailable() {
		return false, nil
	}

	conn, err := dbus.SessionBus(ctx)
	if err != nil {
		return false, nil // D-Bus not available
	}
	defer conn.Close()

	// Check if gsd-xsettings is running (GNOME desktop only)
	names, err := conn.ListNames(ctx)
	if err != nil {
		return false, nil
	}

	gsdRunning := slices.Contains(names, xsettingsService)

	if !gsdRunning {
		return false, nil // Not running GNOME, no error
	}

	obj := conn.Object(xsettingsService, xsettingsPath)

	// Force reload by calling GetAll to trigger XSettings re-broadcast
	err = obj.Call(ctx, "org.freedesktop.DBus.Properties.GetAll", xsettingsIface)
	if err != nil {
		return false, fmt.Errorf("failed to trigger XSettings refresh: %w", err)
	}

	return true, nil
}

// GetThemeName gets the current GTK theme name from GNOME Settings Daemon.
func GetThemeName(ctx context.Context) (string, error) {
	if !dbus.IsAvailable() {
		return "", fmt.Errorf("D-Bus is not available")
	}

	conn, err := dbus.SessionBus(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to connect to session bus: %w", err)
	}
	defer conn.Close()

	obj := conn.Object(xsettingsService, xsettingsPath)

	theme, err := obj.GetProperty(ctx, xsettingsIface+".ThemeName")
	if err != nil {
		return "", fmt.Errorf("failed to get theme name: %w", err)
	}

	if themeStr, ok := theme.(string); ok {
		return themeStr, nil
	}

	return "", fmt.Errorf("unexpected theme type: %T", theme)
}

// SetThemeName sets the GTK theme name via GNOME Settings Daemon.
func SetThemeName(ctx context.Context, themeName string) error {
	if !dbus.IsAvailable() {
		return fmt.Errorf("D-Bus is not available")
	}

	conn, err := dbus.SessionBus(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to session bus: %w", err)
	}
	defer conn.Close()

	obj := conn.Object(xsettingsService, xsettingsPath)

	err = obj.SetProperty(ctx, xsettingsIface+".ThemeName", themeName)
	if err != nil {
		return fmt.Errorf("failed to set theme name: %w", err)
	}

	return nil
}
