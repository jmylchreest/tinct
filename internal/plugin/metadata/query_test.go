//go:build unix

package metadata

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func newCache() *cache {
	return &cache{gates: make(map[string]chan struct{}), entries: make(map[string]entry)}
}

func TestQueryCachesAndInvalidates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plugin")
	count := filepath.Join(dir, "calls")
	t.Setenv("TINCT_TEST_CALLS", count)
	script := "#!/bin/sh\necho call >> \"$TINCT_TEST_CALLS\"\nprintf '%s' '{\"name\":\"first\"}'\n"
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	c := newCache()
	for range 3 {
		data, err := c.query(context.Background(), path)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `{"name":"first"}` {
			t.Fatalf("data = %s", data)
		}
		data[0] = 'x' // A caller must not mutate the cached value.
	}
	calls, err := os.ReadFile(count)
	if err != nil {
		t.Fatal(err)
	}
	if string(calls) != "call\n" {
		t.Fatalf("expected one probe, got %q", calls)
	}
	// Atomic replacement must invalidate even when size and timestamp match.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	replacement := path + ".new"
	if err := os.WriteFile(replacement, []byte(strings.ReplaceAll(script, "first", "other")), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(replacement, info.ModTime(), info.ModTime()); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(replacement, path); err != nil {
		t.Fatal(err)
	}
	data, err := c.query(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"name":"other"}` {
		t.Fatalf("stale data = %s", data)
	}
}

func TestQueryDeadline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plugin")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexec sleep 5\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := newCache().query(ctx, path)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v", err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("query did not respect caller deadline")
	}
}

func TestQueryRetriesInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plugin")
	marker := filepath.Join(t.TempDir(), "ready")
	t.Setenv("TINCT_TEST_READY", marker)
	script := `#!/bin/sh
if [ -f "$TINCT_TEST_READY" ]; then echo '{"name":"ok"}'; else echo invalid; fi
`
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	c := newCache()
	if _, err := c.query(context.Background(), path); err == nil {
		t.Fatal("accepted invalid JSON")
	}
	if err := os.WriteFile(marker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := c.query(context.Background(), path); err != nil {
		t.Fatal(err)
	}
}

func TestQueryDeadlineWhileWaitingForProbe(t *testing.T) {
	c := newCache()
	path := filepath.Join(t.TempDir(), "plugin")
	if err := os.WriteFile(path, nil, 0o700); err != nil {
		t.Fatal(err)
	}
	c.gateFor(path) <- struct{}{}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := c.query(ctx, path); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v", err)
	}
}

func TestQueryResolvesPATH(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tinct-test-plugin")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho '{\"name\":\"path\"}'\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	data, err := newCache().query(context.Background(), "tinct-test-plugin")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != `{"name":"path"}` {
		t.Fatalf("data = %s", data)
	}
}

func TestQueryIndependentPlugins(t *testing.T) {
	c := newCache()
	dir := t.TempDir()
	// A probe already in progress must not block a different executable.
	c.gateFor(filepath.Join(dir, "slow")) <- struct{}{}
	path := filepath.Join(dir, "fast")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho '{}'\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for range 2 { // Cover both a cache miss and the subsequent cache hit.
		if _, err := c.query(ctx, path); err != nil {
			t.Fatal(err)
		}
	}
}

func TestQueryConcurrentConsumers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plugin")
	calls := filepath.Join(dir, "calls")
	t.Setenv("TINCT_TEST_CALLS", calls)
	script := "#!/bin/sh\necho call >> \"$TINCT_TEST_CALLS\"\nsleep 0.05\necho '{}'\n"
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	c := newCache()
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			if _, err := c.query(context.Background(), path); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	data, err := os.ReadFile(calls)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "call\n" {
		t.Fatalf("expected one probe, got %q", data)
	}
}
