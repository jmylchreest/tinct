// Package metadata queries and caches external plugin metadata for this process.
package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

const queryTimeout = 5 * time.Second

type entry struct {
	info os.FileInfo
	data []byte
}

type cache struct {
	mu      sync.Mutex
	gates   map[string]chan struct{}
	entries map[string]entry
}

var queries = cache{gates: make(map[string]chan struct{}), entries: make(map[string]entry)}

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
	resolved, err := exec.LookPath(path)
	if err != nil {
		return nil, fmt.Errorf("resolve plugin executable: %w", err)
	}
	path, err = filepath.Abs(resolved)
	if err != nil {
		return nil, fmt.Errorf("resolve plugin path: %w", err)
	}
	// Serialize fills for this executable without blocking unrelated plugins.
	gate := c.gateFor(path)
	select {
	case gate <- struct{}{}:
		defer func() { <-gate }()
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
	c.mu.Lock()
	old, ok := c.entries[path]
	c.mu.Unlock()
	if ok && unchanged(old.info, info) {
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
		c.mu.Lock()
		c.entries[path] = entry{info: info, data: append([]byte(nil), data...)}
		c.mu.Unlock()
	}
	return data, nil
}

func unchanged(a, b os.FileInfo) bool {
	return os.SameFile(a, b) && a.Size() == b.Size() && a.Mode() == b.Mode() && a.ModTime().Equal(b.ModTime())
}

// gateFor returns a per-executable lock; map access remains safe across probes.
func (c *cache) gateFor(path string) chan struct{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	gate := c.gates[path]
	if gate == nil {
		gate = make(chan struct{}, 1)
		c.gates[path] = gate
	}
	return gate
}
