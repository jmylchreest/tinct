//go:build linux

// Package dbus provides a cross-platform abstraction for D-Bus communication.
// On Linux, it uses the godbus library. On other platforms, it provides no-op stubs.
package dbus

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
)

// Connection represents a D-Bus connection.
type Connection struct {
	conn *dbus.Conn
}

// SessionBus connects to the session bus.
func SessionBus(ctx context.Context) (*Connection, error) {
	_ = ctx
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to session bus: %w", err)
	}
	return &Connection{conn: conn}, nil
}

// SystemBus connects to the system bus.
func SystemBus(ctx context.Context) (*Connection, error) {
	_ = ctx
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to system bus: %w", err)
	}
	return &Connection{conn: conn}, nil
}

// Close closes the D-Bus connection.
func (c *Connection) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("closing D-Bus connection: %w", err)
		}
	}
	return nil
}

// Object represents a D-Bus object.
type Object struct {
	obj dbus.BusObject
}

// Object gets a D-Bus object.
func (c *Connection) Object(dest, path string) *Object {
	return &Object{
		obj: c.conn.Object(dest, dbus.ObjectPath(path)),
	}
}

// Call invokes a D-Bus method.
func (o *Object) Call(ctx context.Context, method string, args ...any) error {
	call := o.obj.CallWithContext(ctx, method, 0, args...)
	if call.Err != nil {
		return fmt.Errorf("D-Bus call failed: %w", call.Err)
	}
	return nil
}

// CallWithReturn invokes a D-Bus method and returns the result.
func (o *Object) CallWithReturn(ctx context.Context, method string, args ...any) ([]any, error) {
	call := o.obj.CallWithContext(ctx, method, 0, args...)
	if call.Err != nil {
		return nil, fmt.Errorf("D-Bus call failed: %w", call.Err)
	}
	return call.Body, nil
}

// GetProperty gets a D-Bus property.
func (o *Object) GetProperty(ctx context.Context, property string) (any, error) {
	_ = ctx
	variant, err := o.obj.GetProperty(property)
	if err != nil {
		return nil, fmt.Errorf("failed to get property: %w", err)
	}
	return variant.Value(), nil
}

// SetProperty sets a D-Bus property.
func (o *Object) SetProperty(ctx context.Context, property string, value any) error {
	_ = ctx
	err := o.obj.SetProperty(property, dbus.MakeVariant(value))
	if err != nil {
		return fmt.Errorf("failed to set property: %w", err)
	}
	return nil
}

// ListNames lists all available names on the bus.
func (c *Connection) ListNames(ctx context.Context) ([]string, error) {
	var names []string
	err := c.conn.BusObject().CallWithContext(
		ctx,
		"org.freedesktop.DBus.ListNames",
		0,
	).Store(&names)
	if err != nil {
		return nil, fmt.Errorf("failed to list bus names: %w", err)
	}
	return names, nil
}

// IsAvailable checks if D-Bus is available on this platform.
func IsAvailable() bool {
	return true
}
