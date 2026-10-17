//go:build !linux

package swaync

import "context"

func reloadCSS(_ context.Context) error { return nil }
