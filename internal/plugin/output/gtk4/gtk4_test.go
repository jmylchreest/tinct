package gtk4

import (
	"strings"
	"testing"

	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/shared/testing"
)

// TestGTK4Plugin runs all standard plugin tests using shared utilities.
func TestGTK4Plugin(t *testing.T) {
	plugin := New()

	// Run basic tests (Name, Description, Validate).
	t.Run("Name", func(t *testing.T) {
		if plugin.Name() != "gtk4" {
			t.Errorf("Name() = %s, want gtk4", plugin.Name())
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
		if !strings.Contains(dir, "gtk-4.0") {
			t.Errorf("DefaultOutputDir() = %s, should contain 'gtk-4.0'", dir)
		}
	})

	config := plugintesting.TestConfig{
		ExpectedName:  "gtk4",
		ExpectedFiles: []string{"tinct-gtk4.css"},
	}

	// Run generation tests.
	plugintesting.TestGeneration(t, plugin, config.ExpectedFiles)
	plugintesting.TestFlags(t, plugin, "gtk4")
}
