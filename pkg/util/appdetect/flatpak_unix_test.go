//go:build unix

package appdetect

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func fakeFlatpak(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	count := filepath.Join(dir, "calls")
	t.Setenv("TINCT_TEST_FLATPAK_CALLS", count)
	t.Setenv("PATH", dir)
	t.Setenv("HOME", t.TempDir())
	script := "#!/bin/sh\necho call >> \"$TINCT_TEST_FLATPAK_CALLS\"\n" + body
	if err := os.WriteFile(filepath.Join(dir, "flatpak"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return count
}

func TestFlatpakInventoryQueriedOnce(t *testing.T) {
	for _, tc := range []struct {
		name   string
		detect func() bool
	}{
		{"any", func() bool { return IsPresentAny([]string{"tinctmissingone", "tinctmissingtwo"}, nil) }},
		{"all", func() bool { return !IsPresentAll([]string{"first", "second"}, nil) }},
		{"type", func() bool { return GetInstallationType([]string{"tinctmissingone", "tinctmissingtwo"}, nil) != "" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			count := fakeFlatpak(t, "printf '%s\\n' org.gnome.First com.github.example.Second\n")
			if tc.detect() {
				t.Fatal("unexpected detection result")
			}
			calls, err := os.ReadFile(count)
			if err != nil {
				t.Fatal(err)
			}
			if string(calls) != "call\n" {
				t.Fatalf("expected one probe, got %q", calls)
			}
		})
	}
}

func TestFlatpakInventoryRefreshesBetweenCalls(t *testing.T) {
	count := fakeFlatpak(t, "echo org.gnome.First\n")
	if !IsPresentAny([]string{"first"}, nil) {
		t.Fatal("first not detected")
	}
	if err := os.WriteFile(filepath.Join(os.Getenv("PATH"), "flatpak"), []byte("#!/bin/sh\necho call >> \"$TINCT_TEST_FLATPAK_CALLS\"\necho org.gnome.Second\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	if IsPresentAny([]string{"first"}, nil) {
		t.Fatal("stale inventory")
	}
	calls, err := os.ReadFile(count)
	if err != nil {
		t.Fatal(err)
	}
	if string(calls) != "call\ncall\n" {
		t.Fatalf("calls = %q", calls)
	}
}

func TestFlatpakFailureIsNotRetriedPerPattern(t *testing.T) {
	count := fakeFlatpak(t, "exit 1\n")
	if IsPresentAny([]string{"tinctmissingone", "tinctmissingtwo"}, nil) {
		t.Fatal("unexpected match")
	}
	calls, err := os.ReadFile(count)
	if err != nil {
		t.Fatal(err)
	}
	if string(calls) != "call\n" {
		t.Fatalf("calls = %q", calls)
	}
}

func TestFlatpakQueryTimeout(t *testing.T) {
	fakeFlatpak(t, "exec /bin/sleep 5\n")
	start := time.Now()
	if IsPresentAny([]string{"tinctmissingone"}, nil) {
		t.Fatal("unexpected match")
	}
	if elapsed := time.Since(start); elapsed > 4*time.Second {
		t.Fatalf("unbounded probe: %v", elapsed)
	}
}
