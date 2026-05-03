package hooks

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestRunReloadVerb_Exec verifies the VerbExec branch actually runs the
// command. We use a test-controlled exec target (touching a file via /bin/sh)
// because we need an exit-code-zero command that's portable in CI containers.
func TestRunReloadVerb_Exec(t *testing.T) {
	tmp := t.TempDir()
	marker := filepath.Join(tmp, "marker")

	spec := ReloadSpec{
		Verb: VerbExec,
		Args: []string{"/bin/sh", "-c", "touch " + marker},
	}
	runReloadVerb(context.Background(), spec, "", false)

	if _, err := os.Stat(marker); err != nil {
		t.Errorf("VerbExec did not invoke the command (marker missing): %v", err)
	}
}

// TestRunReloadVerb_ExecBareShorthand verifies that {Args: [...]} without a
// Verb is treated as exec — the shorthand form plugins use most often.
func TestRunReloadVerb_ExecBareShorthand(t *testing.T) {
	tmp := t.TempDir()
	marker := filepath.Join(tmp, "marker")

	spec := ReloadSpec{
		// No Verb set — empty string ("") falls through to VerbExec.
		Args: []string{"/bin/sh", "-c", "touch " + marker},
	}
	runReloadVerb(context.Background(), spec, "", false)

	if _, err := os.Stat(marker); err != nil {
		t.Errorf("bare-args exec shorthand did not run: %v", err)
	}
}

// TestRunExecVerb_NoArgs verifies the runner doesn't panic on an empty Args
// slice — plugins with malformed specs should fail gracefully.
func TestRunExecVerb_NoArgs(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("runExecVerb panicked on empty args: %v", r)
		}
	}()
	runExecVerb(context.Background(), nil, "", false)
	runExecVerb(context.Background(), []string{}, "", false)
}

// TestRunExecVerb_BadCommand verifies the runner does not panic or block on
// a command that doesn't exist; failure is non-fatal (just verbose-mode warning).
func TestRunExecVerb_BadCommand(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("runExecVerb panicked on bad command: %v", r)
		}
	}()
	runExecVerb(context.Background(), []string{"definitely-not-a-real-binary-xyz"}, "", false)
}

// TestRunReloadVerb_UnknownVerb verifies unknown verb names don't crash —
// the runner emits a warning in verbose mode and returns.
func TestRunReloadVerb_UnknownVerb(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("runReloadVerb panicked on unknown verb: %v", r)
		}
	}()
	runReloadVerb(context.Background(), ReloadSpec{Verb: "totally-bogus", Args: []string{"x"}}, "", false)
}

// TestPrintInstructions_TemplateRendering covers the InstructionsFn path.
func TestPrintInstructions_TemplateRendering(t *testing.T) {
	called := false
	spec := Spec{
		InstructionsFn: func(ctx Context) string {
			called = true
			if len(ctx.WrittenFiles) == 0 {
				t.Error("InstructionsFn did not receive WrittenFiles")
			}
			return "instructions invoked"
		},
	}
	hctx := Context{
		Verbose:      true,
		WrittenFiles: []string{"/tmp/foo"},
	}
	printInstructions(spec, hctx)
	if !called {
		t.Error("InstructionsFn was not invoked")
	}
}
