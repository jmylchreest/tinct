//go:build !linux

package foot

import (
	"context"
	"errors"
)

func (p *Plugin) reloadOSC(_ context.Context) error {
	return errors.New("Foot live updates require Linux")
}
