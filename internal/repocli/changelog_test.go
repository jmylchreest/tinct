// Package repocli provides CLI commands for the repository manager tool.
package repocli

import (
	"strings"
	"testing"
)

func TestChangeLogConsolidatedFormat(t *testing.T) {
	changelog := NewChangeLog()

	// Add a new plugin
	changelog.AddPlugin("newplugin", "1.0.0", []string{"linux_x86", "linux_arm64"})

	// Add a new version to an existing plugin
	changelog.AddVersion("dunstify", "0.1.3", []string{"linux_arm64"})

	// Add platforms to the same version
	changelog.AddPlatform("dunstify", "0.1.3", "linux_arm")
	changelog.AddPlatform("dunstify", "0.1.3", "linux_x86")

	// Add another version
	changelog.AddVersion("dunstify", "0.1.2", []string{"linux_arm64"})
	changelog.AddPlatform("dunstify", "0.1.2", "linux_arm")
	changelog.AddPlatform("dunstify", "0.1.2", "linux_x86")

	// Remove the same versions (simulating pruning)
	changelog.RemoveEntry("dunstify", "0.1.3", "old version pruned")
	changelog.RemoveEntry("dunstify", "0.1.2", "old version pruned")

	// Get the formatted output
	output := changelog.String()

	// Verify the output structure
	if !strings.Contains(output, "New plugins:") {
		t.Error("Expected 'New plugins:' section")
	}

	if !strings.Contains(output, "newplugin v1.0.0") {
		t.Error("Expected newplugin to be listed in new plugins")
	}

	if !strings.Contains(output, "Updated plugins:") {
		t.Error("Expected 'Updated plugins:' section")
	}

	if !strings.Contains(output, "- dunstify:") {
		t.Error("Expected dunstify to be listed in updated plugins")
	}

	if !strings.Contains(output, "Added: v0.1.3") {
		t.Error("Expected version 0.1.3 to be shown as added")
	}

	if !strings.Contains(output, "Pruned: v0.1.3") {
		t.Error("Expected version 0.1.3 to be shown as pruned")
	}

	// Verify it doesn't have the old verbose format
	if strings.Contains(output, "New versions:") {
		t.Error("Should not have separate 'New versions:' section")
	}

	if strings.Contains(output, "Platform additions:") {
		t.Error("Should not have separate 'Platform additions:' section")
	}

	if strings.Contains(output, "Removed:") {
		t.Error("Should not have separate 'Removed:' section (should be under plugin)")
	}
}

func TestChangeLogOnlyPlatformAdditions(t *testing.T) {
	changelog := NewChangeLog()

	// Add platforms to existing versions
	changelog.AddPlatform("random", "0.1.3", "linux_arm")
	changelog.AddPlatform("random", "0.1.3", "linux_x86")
	changelog.AddPlatform("random", "0.1.2", "freebsd_x86")

	output := changelog.String()

	// Should show as updated plugin
	if !strings.Contains(output, "Updated plugins:") {
		t.Error("Expected 'Updated plugins:' section for platform additions")
	}

	if !strings.Contains(output, "- random:") {
		t.Error("Expected random to be listed in updated plugins")
	}

	if !strings.Contains(output, "Added: v0.1.3 (linux_arm, linux_x86)") {
		t.Error("Expected consolidated platform additions for v0.1.3")
	}
}

func TestChangeLogEmpty(t *testing.T) {
	changelog := NewChangeLog()

	output := changelog.String()

	if output != "No changes" {
		t.Errorf("Expected 'No changes', got: %s", output)
	}

	if !changelog.IsEmpty() {
		t.Error("Expected changelog to be empty")
	}
}

func TestChangeLogSummary(t *testing.T) {
	changelog := NewChangeLog()

	changelog.AddPlugin("plugin1", "1.0.0", []string{"linux_x86"})
	changelog.AddVersion("plugin2", "2.0.0", []string{"darwin_arm64"})
	changelog.AddPlatform("plugin2", "2.0.0", "darwin_x86")
	changelog.RemoveEntry("plugin3", "0.5.0", "old version pruned")

	output := changelog.String()

	// Check summary line
	if !strings.Contains(output, "Summary:") {
		t.Error("Expected summary section")
	}

	if !strings.Contains(output, "+2 versions") {
		t.Error("Expected +2 versions in summary")
	}

	if !strings.Contains(output, "+1 platforms") {
		t.Error("Expected +1 platforms in summary")
	}

	if !strings.Contains(output, "-1 removed") {
		t.Error("Expected -1 removed in summary")
	}
}

func TestChangeLogConsolidateByPlugin(t *testing.T) {
	changelog := NewChangeLog()

	// Add version with initial platforms
	changelog.AddVersion("test", "1.0.0", []string{"linux_x86"})

	// Add more platforms to same version
	changelog.AddPlatform("test", "1.0.0", "linux_arm64")
	changelog.AddPlatform("test", "1.0.0", "darwin_arm64")

	consolidated := changelog.consolidateByPlugin()

	if len(consolidated) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(consolidated))
	}

	testPlugin, exists := consolidated["test"]
	if !exists {
		t.Fatal("Expected 'test' plugin in consolidated changes")
	}

	if len(testPlugin.VersionsAdded) != 1 {
		t.Errorf("Expected 1 version, got %d", len(testPlugin.VersionsAdded))
	}

	// Should have all platforms merged
	platforms := testPlugin.VersionsAdded[0].PlatformsAdded
	if len(platforms) != 3 {
		t.Errorf("Expected 3 platforms, got %d: %v", len(platforms), platforms)
	}
}

func TestChangeLogDesiredFormat(t *testing.T) {
	// This test demonstrates the exact format requested:
	// - New plugins section
	// - Updated plugins section with consolidated Added/Pruned per version
	// - Clear diff-style view

	changelog := NewChangeLog()

	// Simulate adding new plugins
	changelog.AddPlugin("plugin12", "1.0.0", []string{"linux_x86"})
	changelog.AddPlugin("plugin11", "0.5.0", []string{"darwin_arm64"})

	// Simulate adding versions and platforms to existing plugins
	changelog.AddVersion("plugin1", "0.1.1", []string{"linux_arm64"})
	changelog.AddPlatform("plugin1", "0.1.1", "linux_x86")
	changelog.AddVersion("plugin1", "0.1.0", []string{"linux_arm64"})
	changelog.RemoveEntry("plugin1", "0.0.1", "old version pruned")

	changelog.AddVersion("plugin2", "0.2.0", []string{"linux_arm64"})
	changelog.AddPlatform("plugin2", "0.2.0", "linux_x86")
	changelog.RemoveEntry("plugin2", "0.1.0", "old version pruned")

	output := changelog.String()

	// Verify structure matches desired format
	if !strings.Contains(output, "New plugins:") {
		t.Error("Missing 'New plugins:' section")
	}

	if !strings.Contains(output, "Updated plugins:") {
		t.Error("Missing 'Updated plugins:' section")
	}

	// Check new plugins are listed
	if !strings.Contains(output, "plugin12") {
		t.Error("Missing plugin12 in new plugins")
	}
	if !strings.Contains(output, "plugin11") {
		t.Error("Missing plugin11 in new plugins")
	}

	// Check updated plugins show both additions and pruning
	if !strings.Contains(output, "- plugin1:") {
		t.Error("Missing plugin1 in updated plugins")
	}

	if !strings.Contains(output, "Added: v0.1.1") {
		t.Error("Missing 'Added: v0.1.1' for plugin1")
	}

	if !strings.Contains(output, "Pruned: v0.0.1") {
		t.Error("Missing 'Pruned: v0.0.1' for plugin1")
	}

	// Verify platform consolidation
	// When AddVersion is called with platforms AND AddPlatform is called,
	// they should be merged
	if !strings.Contains(output, "linux_arm64") && !strings.Contains(output, "linux_x86") {
		t.Error("Platforms should be shown for added versions")
	}

	// Print output for manual verification during development
	t.Logf("Changelog output:\n%s", output)
}
