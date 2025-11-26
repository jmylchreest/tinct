//go:build !linux

// Package dbus provides a cross-platform abstraction for D-Bus communication.
// On non-Linux platforms, it provides no-op stubs.
package dbus

import (
	"context"
	"errors"
)

var errNotSupported = errors.New("D-Bus is not supported on this platform")

// Connection represents a D-Bus connection.
type Connection struct{}

// SessionBus connects to the session bus (not supported on non-Linux).
func SessionBus(ctx context.Context) (*Connection, error) {
	return nil, errNotSupported
}

// SystemBus connects to the system bus (not supported on non-Linux).
func SystemBus(ctx context.Context) (*Connection, error) {
	return nil, errNotSupported
}

// Close closes the D-Bus connection.
func (c *Connection) Close() error {
	return nil
}

// Object represents a D-Bus object.
type Object struct{}

// Object gets a D-Bus object (not supported on non-Linux).
func (c *Connection) Object(dest string, path string) *Object {
	return &Object{}
}

// Call invokes a D-Bus method (not supported on non-Linux).
func (o *Object) Call(ctx context.Context, method string, args ...interface{}) error {
	return errNotSupported
}

// CallWithReturn invokes a D-Bus method and returns the result (not supported on non-Linux).
func (o *Object) CallWithReturn(ctx context.Context, method string, args ...interface{}) ([]interface{}, error) {
	return nil, errNotSupported
}

// GetProperty gets a D-Bus property (not supported on non-Linux).
func (o *Object) GetProperty(ctx context.Context, property string) (interface{}, error) {
	return nil, errNotSupported
}

// SetProperty sets a D-Bus property (not supported on non-Linux).
func (o *Object) SetProperty(ctx context.Context, property string, value interface{}) error {
	return errNotSupported
}

// ListNames lists all available names on the bus (not supported on non-Linux).
func (c *Connection) ListNames(ctx context.Context) ([]string, error) {
	return nil, errNotSupported
}

// IsAvailable checks if D-Bus is available on this platform.
func IsAvailable() bool {
	return false
}
