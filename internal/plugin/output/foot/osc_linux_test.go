package foot

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLiveTargetRejectsFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(path, []byte("unchanged"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{path, "/dev/null", "/dev/pts/../null", "/dev/pts/1/extra"} {
		p := &Plugin{tty: target, osc: []byte("payload")}
		if err := p.reloadOSC(context.Background()); err == nil {
			t.Errorf("accepted %s", target)
		}
	}
	data, _ := os.ReadFile(path)
	if string(data) != "unchanged" {
		t.Fatal("file modified")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := &Plugin{tty: "/dev/tty", osc: []byte("payload")}
	if err := p.reloadOSC(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled: %v", err)
	}
}
