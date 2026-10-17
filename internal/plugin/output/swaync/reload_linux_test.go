//go:build linux

package swaync

import (
	"bufio"
	"context"
	"net"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
)

func TestReloadPrivateBus(t *testing.T) {
	daemon, err := exec.LookPath("dbus-daemon")
	if err != nil {
		t.Skip("dbus-daemon unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, daemon, "--session", "--nofork", "--print-address=1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { cancel(); _ = cmd.Wait() }()
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() {
		t.Fatal("private bus did not provide an address")
	}
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", scanner.Text())
	// An absent service is a successful no-op; no activation or wait loop.
	if err = reloadCSS(ctx); err != nil {
		t.Fatalf("absent service: %v", err)
	}
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_, err = conn.RequestName("org.erikreider.swaync.cc", dbus.NameFlagDoNotQueue)
	if err != nil {
		t.Fatal(err)
	}
	for _, accepted := range []bool{true, false} {
		methods := map[string]any{"ReloadCss": func() (bool, *dbus.Error) { return accepted, nil }}
		if err = conn.ExportMethodTable(methods, "/org/erikreider/swaync/cc", "org.erikreider.swaync.cc"); err != nil {
			t.Fatal(err)
		}
		err = reloadCSS(ctx)
		if accepted && err != nil {
			t.Fatalf("successful reload: %v", err)
		}
		if !accepted && (err == nil || !strings.Contains(err.Error(), "rejected")) {
			t.Fatalf("expected stylesheet rejection, got %v", err)
		}
	}
}

func TestReloadCancelledDuringAuthentication(t *testing.T) {
	socket := filepath.Join(t.TempDir(), "bus")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	release := make(chan struct{})
	defer close(release)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		defer conn.Close()
		<-release
	}()
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "unix:path="+socket)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- reloadCSS(ctx) }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected cancelled authentication error")
		}
	case <-time.After(time.Second):
		t.Fatal("reload ignored cancellation during authentication")
	}
}
