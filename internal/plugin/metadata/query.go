// Package metadata queries and caches external plugin metadata for this process.
package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const queryTimeout = 5 * time.Second

type entry struct {
	info os.FileInfo
	data []byte
}

type cache struct {
	gate    chan struct{}
	entries map[string]entry
}

var queries = cache{gate: make(chan struct{}, 1), entries: make(map[string]entry)}

// Query returns valid JSON from --plugin-info within five seconds, plus at most
// one second to drain inherited pipes. Successful results are reused while the
// executable's identity, size, mode and modification time remain unchanged.
// Failures are not cached, and callers receive their own copy of the JSON.
func Query(ctx context.Context, path string) ([]byte, error) {
	return queries.query(ctx, path)
}

func (c *cache) query(ctx context.Context, path string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve plugin path: %w", err)
	}
	// Serialize cache fills so concurrent consumers do not launch duplicate probes.
	select {
	case c.gate <- struct{}{}:
		defer func() { <-c.gate }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat plugin: %w", err)
	}
	if old, ok := c.entries[path]; ok && unchanged(old.info, info) {
		return append([]byte(nil), old.data...), nil
	}
	cmd := exec.CommandContext(ctx, path, "--plugin-info") // #nosec G204 -- the caller explicitly selects an external plugin executable.
	cmd.WaitDelay = time.Second
	data, err := cmd.Output()
	if ctx.Err() != nil {
		return nil, fmt.Errorf("query plugin %q: %w", path, ctx.Err())
	}
	if err != nil {
		return nil, fmt.Errorf("query plugin %q: %w", path, err)
	}
	if !json.Valid(data) {
		return nil, fmt.Errorf("plugin %q returned invalid metadata JSON", path)
	}
	// Do not cache an executable that was replaced while it was being queried.
	current, err := os.Stat(path)
	if err == nil && unchanged(info, current) {
		c.entries[path] = entry{info: info, data: append([]byte(nil), data...)}
	}
	return data, nil
}

func unchanged(a, b os.FileInfo) bool {
	return os.SameFile(a, b) && a.Size() == b.Size() && a.Mode() == b.Mode() && a.ModTime().Equal(b.ModTime())
}
