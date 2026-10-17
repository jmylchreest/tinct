//go:build unix

package subprocess

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestCancellationKillsDescendants(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "survived")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := CommandContext(ctx, "sh", "-c", `(sleep 1; echo survived > "$1") & echo ready; wait`, "sh", marker)
	ready, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	// The child has been started before the parent announces readiness.
	buf := make([]byte, 6)
	if _, err := ready.Read(buf); err != nil {
		t.Fatal(err)
	}
	cancel()
	if err := cmd.Wait(); err == nil {
		t.Fatal("cancelled command succeeded")
	}
	time.Sleep(1200 * time.Millisecond)
	if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("descendant survived cancellation: %v", err)
	}
}

func TestOutputDoesNotWaitForInheritedPipe(t *testing.T) {
	cmd := CommandContext(context.Background(), "sh", "-c", "sleep 2 & echo done")
	cmd.WaitDelay = 50 * time.Millisecond
	start := time.Now()
	output, err := cmd.Output()
	if !errors.Is(err, exec.ErrWaitDelay) {
		t.Fatalf("error = %v; want ErrWaitDelay", err)
	}
	if string(output) != "done\n" {
		t.Fatalf("output = %q", output)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("waited for descendant pipe: %v", elapsed)
	}
}

func TestCommandPreservesOutputAndArguments(t *testing.T) {
	cmd := CommandContext(context.Background(), "printf", "%s", "spaces;$(literal)")
	output, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	if string(output) != "spaces;$(literal)" {
		t.Fatalf("output = %q", output)
	}
}
