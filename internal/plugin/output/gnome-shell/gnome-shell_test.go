package gnomeshell

import (
	"strings"
	"testing"

	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/testing"
)

// TestGNOMEShellPlugin runs all standard plugin tests using shared utilities.
func TestGNOMEShellPlugin(t *testing.T) {
	plugin := New()

	// Run basic tests (Name, Description, Validate).
	t.Run("Name", func(t *testing.T) {
		if plugin.Name() != "gnome-shell" {
			t.Errorf("Name() = %s, want gnome-shell", plugin.Name())
		}
	})

	t.Run("Description", func(t *testing.T) {
		if plugin.Description() == "" {
			t.Error("Description() should not be empty")
		}
	})

	t.Run("DefaultOutputDir", func(t *testing.T) {
		dir := plugin.DefaultOutputDir()
		if dir == "" {
			t.Error("DefaultOutputDir() should not be empty")
		}
		if !strings.Contains(dir, "gnome-shell") {
			t.Errorf("DefaultOutputDir() = %s, should contain 'gnome-shell'", dir)
		}
	})

	config := plugintesting.TestConfig{
		ExpectedName:       "gnome-shell",
		ExpectedFiles:      []string{"gnome-shell.css"},
		ExpectedBinaryName: "gnome-shell",
	}

	// Run generation tests.
	plugintesting.TestGeneration(t, plugin, config.ExpectedFiles)
	plugintesting.TestPreExecuteHook(t, plugin, config.ExpectedBinaryName)
	plugintesting.TestFlags(t, plugin, "gnome-shell")
}
