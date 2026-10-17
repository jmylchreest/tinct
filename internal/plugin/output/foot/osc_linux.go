package foot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// Explicit targets only: never enumerate unrelated terminals or shell out.
func (p *Plugin) reloadOSC(ctx context.Context) error {
	if p.tty != "/dev/tty" && !terminalPath.MatchString(p.tty) {
		return fmt.Errorf("foot.tty must be /dev/tty or /dev/pts/N")
	}
	if len(p.osc) == 0 {
		return errors.New("no generated Foot palette")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return err
	}
	fd, err := unix.Open(p.tty, unix.O_WRONLY|unix.O_NOCTTY|unix.O_NONBLOCK|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	if !term.IsTerminal(fd) {
		return errors.New("foot target is not a terminal")
	}
	return writeOSC(ctx, fd, p.osc)
}

func writeOSC(ctx context.Context, fd int, remaining []byte) error {
	for len(remaining) > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, err := unix.Write(fd, remaining)
		if n > 0 {
			remaining = remaining[n:]
		}
		if err != nil && !errors.Is(err, unix.EAGAIN) && !errors.Is(err, unix.EINTR) {
			return err
		}
		if len(remaining) > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(10 * time.Millisecond):
			}
		}
	}
	return nil
}
