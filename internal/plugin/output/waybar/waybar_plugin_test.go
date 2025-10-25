package waybar

import (
	"testing"

	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/testing"
)

// TestWaybarPlugin runs all standard plugin tests using shared utilities.
func TestWaybarPlugin(t *testing.T) {
	plugin := New()

	config := plugintesting.TestConfig{
		ExpectedName:       "waybar",
		ExpectedFiles:      []string{"tinct-colours.css", "tinct.css"},
		ExpectedBinaryName: "waybar",
	}

	plugintesting.RunAllTests(t, plugin, config)
}
