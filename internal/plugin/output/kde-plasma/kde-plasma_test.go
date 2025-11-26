package kdeplasma

import (
	"strings"
	"testing"

	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/testing"
)

// TestKdePlasmaPlugin runs all standard plugin tests using shared utilities.
func TestKdePlasmaPlugin(t *testing.T) {
	plugin := New()

	// Run basic tests (Name, Description, Validate).
	t.Run("Name", func(t *testing.T) {
		if plugin.Name() != "kde-plasma" {
			t.Errorf("Name() = %s, want kde-plasma", plugin.Name())
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
		if !strings.Contains(dir, "color-schemes") {
			t.Errorf("DefaultOutputDir() = %s, should contain 'color-schemes'", dir)
		}
	})

	config := plugintesting.TestConfig{
		ExpectedName:  "kde-plasma",
		ExpectedFiles: []string{"TinctDark1.colors", "TinctDark2.colors"}, // Now generates both variants
	}

	// Run generation tests.
	plugintesting.TestGeneration(t, plugin, config.ExpectedFiles)
	plugintesting.TestFlags(t, plugin, "kde-plasma")
}
