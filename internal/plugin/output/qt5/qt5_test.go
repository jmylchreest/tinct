package qt5

import (
	"strings"
	"testing"

	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/shared/testing"
)

// TestQt5Plugin runs all standard plugin tests using shared utilities.
func TestQt5Plugin(t *testing.T) {
	plugin := New()

	// Run basic tests (Name, Description, Validate).
	t.Run("Name", func(t *testing.T) {
		if plugin.Name() != "qt5" {
			t.Errorf("Name() = %s, want qt5", plugin.Name())
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
		if !strings.Contains(dir, "qt5ct") {
			t.Errorf("DefaultOutputDir() = %s, should contain 'qt5ct'", dir)
		}
	})

	config := plugintesting.TestConfig{
		ExpectedName:  "qt5",
		ExpectedFiles: []string{"tinct.conf"},
	}

	// Run generation tests.
	plugintesting.TestGeneration(t, plugin, config.ExpectedFiles)
	plugintesting.TestFlags(t, plugin, "qt5")
}
