package hooks

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPre_NoSpec(t *testing.T) {
	skip, reason, err := RunPre(Spec{}, Context{})
	if skip || reason != "" || err != nil {
		t.Errorf("empty Spec should be a no-op; got skip=%v reason=%q err=%v", skip, reason, err)
	}
}

func TestRunPre_RequiredBinaryMissing(t *testing.T) {
	skip, reason, err := RunPre(Spec{
		RequiredBinaries: []string{"definitely-not-a-real-binary-xyzzy"},
	}, Context{})
	if !skip {
		t.Error("expected skip=true when required binary missing")
	}
	if !strings.Contains(reason, "definitely-not-a-real-binary-xyzzy") {
		t.Errorf("reason should mention the missing binary; got %q", reason)
	}
	if err != nil {
		t.Errorf("err should be nil, not %v", err)
	}
}

func TestRunPre_RequiredDirMissing(t *testing.T) {
	skip, reason, _ := RunPre(Spec{
		RequiredDirs: []string{"/definitely/not/a/real/dir/xyzzy"},
	}, Context{})
	if !skip {
		t.Error("expected skip=true when required dir missing")
	}
	if !strings.Contains(reason, "xyzzy") {
		t.Errorf("reason should mention the missing dir; got %q", reason)
	}
}

// TestRunPre_RequiredDirTildeExpansion verifies plugins can declare
// "~/.config/foo" in RequiredDirs without doing their own os.UserHomeDir
// dance. We rebind HOME to a tmp dir, make the dir under it exist, and
// confirm the runner doesn't skip.
func TestRunPre_RequiredDirTildeExpansion(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	target := filepath.Join(tmp, ".config", "tinct-tilde-test")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	skip, reason, err := RunPre(Spec{
		RequiredDirs: []string{"~/.config/tinct-tilde-test"},
	}, Context{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if skip {
		t.Errorf("expected skip=false for existing tilde-prefixed dir, got skip=%v reason=%q", skip, reason)
	}
}

// TestRunPre_RequiredDirTildeMissingStillSkips verifies the expansion
// path doesn't accidentally swallow missing-dir cases.
func TestRunPre_RequiredDirTildeMissingStillSkips(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	skip, reason, _ := RunPre(Spec{
		RequiredDirs: []string{"~/definitely-missing-xyzzy"},
	}, Context{})
	if !skip {
		t.Error("expected skip=true for missing tilde-prefixed dir")
	}
	// The reason should preserve the original (un-expanded) path so the
	// user sees what they wrote.
	if !strings.Contains(reason, "~/definitely-missing-xyzzy") {
		t.Errorf("reason should mention the original path; got %q", reason)
	}
}

func TestRunPre_AutoCreateDir(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "nested", "subdir")

	skip, _, err := RunPre(Spec{AutoCreateDir: true}, Context{OutputDir: target})
	if skip || err != nil {
		t.Fatalf("AutoCreateDir should not skip; got skip=%v err=%v", skip, err)
	}
	if _, statErr := os.Stat(target); statErr != nil {
		t.Errorf("AutoCreateDir did not create %s: %v", target, statErr)
	}
}

func TestRunPre_AutoCreateDirIdempotent(t *testing.T) {
	tmp := t.TempDir()
	skip, _, err := RunPre(Spec{AutoCreateDir: true}, Context{OutputDir: tmp})
	if skip || err != nil {
		t.Errorf("AutoCreateDir on existing dir should not skip; got skip=%v err=%v", skip, err)
	}
}

func TestRunPost_DryRunSkipsEverything(t *testing.T) {
	called := false
	spec := Spec{
		ReloadFn: func(_ context.Context) error {
			called = true
			return nil
		},
	}
	_ = RunPost(context.Background(), spec, Context{DryRun: true})
	if called {
		t.Error("ReloadFn should not be invoked in DryRun mode")
	}
}

func TestRunPost_ReloadFnInvoked(t *testing.T) {
	called := false
	spec := Spec{
		ReloadFn: func(_ context.Context) error {
			called = true
			return nil
		},
	}
	_ = RunPost(context.Background(), spec, Context{
		WrittenFiles: []string{"/tmp/something"},
	})
	if !called {
		t.Error("ReloadFn should be invoked")
	}
}

func TestRunPost_WallpaperHandlerInvoked(t *testing.T) {
	gotPath := ""
	spec := Spec{
		SupportsWallpaper: true,
		Wallpaper: func(_ context.Context, p string) error {
			gotPath = p
			return nil
		},
	}
	_ = RunPost(context.Background(), spec, Context{
		WallpaperPath: "/tmp/wp.jpg",
		WrittenFiles:  []string{"/tmp/written"},
	})
	if gotPath != "/tmp/wp.jpg" {
		t.Errorf("Wallpaper handler not invoked with correct path; got %q", gotPath)
	}
}

func TestRunPost_WallpaperSkippedWhenPathEmpty(t *testing.T) {
	called := false
	spec := Spec{
		SupportsWallpaper: true,
		Wallpaper: func(_ context.Context, _ string) error {
			called = true
			return nil
		},
	}
	_ = RunPost(context.Background(), spec, Context{
		WrittenFiles: []string{"/tmp/written"},
	})
	if called {
		t.Error("Wallpaper handler should not be invoked when WallpaperPath is empty")
	}
}

func TestRunPost_MakeExecutable(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "tinct-helper.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	spec := Spec{MakeExecutable: []string{"tinct-helper.sh"}}
	_ = RunPost(context.Background(), spec, Context{
		WrittenFiles: []string{scriptPath},
	})

	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("expected owner-execute bit set on %s; got %v", scriptPath, info.Mode().Perm())
	}
}

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name, in, want string
		hctx           Context
	}{
		{
			name: "outputDir",
			in:   "wrote to {{.OutputDir}}",
			hctx: Context{OutputDir: "/foo/bar"},
			want: "wrote to /foo/bar",
		},
		{
			name: "writtenFiles index",
			in:   "first: {{index .WrittenFiles 0}}",
			hctx: Context{WrittenFiles: []string{"a", "b"}},
			want: "first: a",
		},
		{
			name: "broken template returns original",
			in:   "{{.OutputDir",
			hctx: Context{OutputDir: "/foo"},
			want: "{{.OutputDir",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderTemplate(tt.in, tt.hctx)
			if got != tt.want {
				t.Errorf("renderTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}
