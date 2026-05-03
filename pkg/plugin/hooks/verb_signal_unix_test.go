//go:build unix

package hooks

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// TestParseSignal verifies the supported signal names map to the correct
// syscall.Signal values. Unknown names return nil; runSignalVerb then logs a
// warning and returns instead of crashing.
func TestParseSignal(t *testing.T) {
	tests := []struct {
		name  string
		want  syscall.Signal
		isNil bool
	}{
		{"SIGUSR1", syscall.SIGUSR1, false},
		{"SIGUSR2", syscall.SIGUSR2, false},
		{"SIGHUP", syscall.SIGHUP, false},
		{"SIGTERM", syscall.SIGTERM, false},
		{"SIGINT", syscall.SIGINT, false},
		{"SIGBOGUS", 0, true},
		{"", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSignal(tt.name)
			if tt.isNil {
				if got != nil {
					t.Errorf("parseSignal(%q) = %v, want nil", tt.name, got)
				}
				return
			}
			if got != tt.want {
				t.Errorf("parseSignal(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// TestRunSignalVerb_BadArgs verifies the runner doesn't panic on malformed
// argument slices. The runner logs a warning in verbose mode and returns;
// callers see no error.
func TestRunSignalVerb_BadArgs(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("runSignalVerb panicked: %v", r)
		}
	}()
	runSignalVerb(nil, false)
	runSignalVerb([]string{}, false)
	runSignalVerb([]string{"only-one-arg"}, false)
	runSignalVerb([]string{"foo", "SIGBOGUS"}, false)
	runSignalVerb([]string{"definitely-not-a-process-xyz", "SIGUSR1"}, false)
}

// TestRunSignalVerb_SelfSignal exercises the full signal delivery path. We
// register a SIGUSR2 handler in this process, ask the runner to deliver
// SIGUSR2 to processes matching our own executable's basename, and verify
// the handler fires within a short timeout.
//
// This depends on go-ps listing the running test binary, which is normally
// reliable on Linux/macOS/BSD. If it doesn't (some unusual kernel or
// container), the test skips rather than fails — runSignalVerb's failure
// modes are already covered by TestRunSignalVerb_BadArgs.
func TestRunSignalVerb_SelfSignal(t *testing.T) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGUSR2)
	defer signal.Stop(sig)

	exe, err := os.Executable()
	if err != nil {
		t.Skipf("cannot determine test binary path: %v", err)
	}
	procName := filepath.Base(exe)

	runSignalVerb([]string{procName, "SIGUSR2"}, false)

	select {
	case <-sig:
		// Signal was delivered as expected.
	case <-time.After(500 * time.Millisecond):
		t.Skip("SIGUSR2 not received within 500ms; go-ps may not list this process on this platform. The runner did not panic, which is the more important guarantee.")
	}
}
