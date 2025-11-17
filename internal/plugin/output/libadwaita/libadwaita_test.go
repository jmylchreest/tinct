package libadwaita

import (
	"strings"
	"testing"

	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/testing"
)

// TestLibadwaitaPlugin runs all standard plugin tests using shared utilities.
func TestLibadwaitaPlugin(t *testing.T) {
	plugin := New()

	// Run basic tests (Name, Description, Validate).
	t.Run("Name", func(t *testing.T) {
		if plugin.Name() != "libadwaita" {
			t.Errorf("Name() = %s, want libadwaita", plugin.Name())
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
		ExpectedName:  "libadwaita",
		ExpectedFiles: []string{"tinct-libadwaita.css"},
	}

	// Run generation tests.
	plugintesting.TestGeneration(t, plugin, config.ExpectedFiles)
	plugintesting.TestFlags(t, plugin, "libadwaita")
}
