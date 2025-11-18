package qt6

import (
	"strings"
	"testing"

	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/testing"
)

// TestQt6Plugin runs all standard plugin tests using shared utilities.
func TestQt6Plugin(t *testing.T) {
	plugin := New()

	// Run basic tests (Name, Description, Validate).
	t.Run("Name", func(t *testing.T) {
		if plugin.Name() != "qt6" {
			t.Errorf("Name() = %s, want qt6", plugin.Name())
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
		if !strings.Contains(dir, "qt6ct") {
			t.Errorf("DefaultOutputDir() = %s, should contain 'qt6ct'", dir)
		}
	})

	config := plugintesting.TestConfig{
		ExpectedName:  "qt6",
		ExpectedFiles: []string{"tinct.conf"},
	}

	// Run generation tests.
	plugintesting.TestGeneration(t, plugin, config.ExpectedFiles)
	plugintesting.TestFlags(t, plugin, "qt6")
}
