//go:build linux

package swaync

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
)

// reloadCSS avoids swaync-client's wait loop when its initial proxy connection
// fails. NoAutoStart prevents a reload from launching a stopped daemon.
func reloadCSS(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	conn, err := dbus.ConnectSessionBus(dbus.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("connect to SwayNC session bus: %w", err)
	}
	defer conn.Close()
	var reloaded bool
	err = conn.Object("org.erikreider.swaync.cc", "/org/erikreider/swaync/cc").CallWithContext(ctx, "org.erikreider.swaync.cc.ReloadCss", dbus.FlagNoAutoStart).Store(&reloaded)
	if err != nil {
		var busErr dbus.Error
		if errors.As(err, &busErr) && (busErr.Name == "org.freedesktop.DBus.Error.ServiceUnknown" || busErr.Name == "org.freedesktop.DBus.Error.NameHasNoOwner") {
			return nil
		}
		return fmt.Errorf("reload SwayNC CSS: %w", err)
	}
	if !reloaded {
		return fmt.Errorf("SwayNC rejected the stylesheet")
	}
	return nil
}
