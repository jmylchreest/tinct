// Package template provides utilities for loading plugin templates with custom override support.
package template

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed testdata/*.tmpl
var testEmbedFS embed.FS

func TestLoader_Load(t *testing.T) {
	// Create a temporary directory for custom templates.
	tmpDir := t.TempDir()

	loader := &Loader{
		pluginName: "testplugin",
		embedFS:    testEmbedFS,
		customBase: tmpDir,
	}

	t.Run("loads embedded template when no custom exists", func(t *testing.T) {
		content, fromCustom, err := loader.Load("testdata/test.tmpl")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fromCustom {
			t.Error("expected embedded template, got custom")
		}
		if len(content) == 0 {
			t.Error("expected content, got empty")
		}
	})

	t.Run("loads custom template when it exists", func(t *testing.T) {
		// Create a custom template.
		customDir := filepath.Join(tmpDir, "testplugin")
		if err := os.MkdirAll(customDir, 0755); err != nil {
			t.Fatalf("failed to create custom dir: %v", err)
		}

		customContent := []byte("# This is a custom template\n")
		customPath := filepath.Join(customDir, "testdata", "test.tmpl")
		if err := os.MkdirAll(filepath.Dir(customPath), 0755); err != nil {
			t.Fatalf("failed to create custom template dir: %v", err)
		}
		if err := os.WriteFile(customPath, customContent, 0644); err != nil {
			t.Fatalf("failed to write custom template: %v", err)
		}

		content, fromCustom, err := loader.Load("testdata/test.tmpl")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !fromCustom {
			t.Error("expected custom template, got embedded")
		}
		if string(content) != string(customContent) {
			t.Errorf("expected custom content %q, got %q", customContent, content)
		}
	})

	t.Run("returns error for non-existent template", func(t *testing.T) {
		_, _, err := loader.Load("nonexistent.tmpl")
		if err == nil {
			t.Error("expected error for non-existent template")
		}
	})
}

func TestLoader_CustomPath(t *testing.T) {
	loader := &Loader{
		pluginName: "testplugin",
		customBase: "/home/user/.config/tinct/templates",
	}

	expected := "/home/user/.config/tinct/templates/testplugin/test.tmpl"
	got := loader.CustomPath("test.tmpl")
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestLoader_CustomDir(t *testing.T) {
	loader := &Loader{
		pluginName: "testplugin",
		customBase: "/home/user/.config/tinct/templates",
	}

	expected := "/home/user/.config/tinct/templates/testplugin"
	got := loader.CustomDir()
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestLoader_HasCustomTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	loader := &Loader{
		pluginName: "testplugin",
		customBase: tmpDir,
	}

	t.Run("returns false when custom template doesn't exist", func(t *testing.T) {
		if loader.HasCustomTemplate("test.tmpl") {
			t.Error("expected false for non-existent custom template")
		}
	})

	t.Run("returns true when custom template exists", func(t *testing.T) {
		customPath := loader.CustomPath("test.tmpl")
		if err := os.MkdirAll(filepath.Dir(customPath), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(customPath, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		if !loader.HasCustomTemplate("test.tmpl") {
			t.Error("expected true for existing custom template")
		}
	})
}

func TestLoader_ListEmbeddedTemplates(t *testing.T) {
	loader := &Loader{
		pluginName: "testplugin",
		embedFS:    testEmbedFS,
	}

	templates, err := loader.ListEmbeddedTemplates()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(templates) == 0 {
		t.Error("expected at least one template")
	}

	// Check that all returned files have .tmpl extension.
	for _, tmpl := range templates {
		if filepath.Ext(tmpl) != ".tmpl" {
			t.Errorf("expected .tmpl extension, got %q", tmpl)
		}
	}
}

func TestLoader_DumpTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	loader := &Loader{
		pluginName: "testplugin",
		embedFS:    testEmbedFS,
		customBase: tmpDir,
	}

	t.Run("dumps template successfully", func(t *testing.T) {
		err := loader.DumpTemplate("testdata/test.tmpl", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check that file was created.
		customPath := loader.CustomPath("testdata/test.tmpl")
		if _, err := os.Stat(customPath); err != nil {
			t.Errorf("custom template not created: %v", err)
		}
	})

	t.Run("fails without force when template exists", func(t *testing.T) {
		// Try to dump again without force.
		err := loader.DumpTemplate("testdata/test.tmpl", false)
		if err == nil {
			t.Error("expected error when dumping existing template without force")
		}
	})

	t.Run("overwrites with force flag", func(t *testing.T) {
		err := loader.DumpTemplate("testdata/test.tmpl", true)
		if err != nil {
			t.Fatalf("unexpected error with force flag: %v", err)
		}
	})

	t.Run("returns error for non-existent template", func(t *testing.T) {
		err := loader.DumpTemplate("nonexistent.tmpl", false)
		if err == nil {
			t.Error("expected error for non-existent template")
		}
	})
}

func TestLoader_DumpAllTemplates(t *testing.T) {
	tmpDir := t.TempDir()

	loader := &Loader{
		pluginName: "testplugin",
		embedFS:    testEmbedFS,
		customBase: tmpDir,
	}

	dumped, err := loader.DumpAllTemplates(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(dumped) == 0 {
		t.Error("expected at least one dumped template")
	}

	// Verify all files were created.
	for _, path := range dumped {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("dumped file not found: %s (%v)", path, err)
		}
	}
}

func TestLoader_DumpAllTemplates_WithExisting(t *testing.T) {
	tmpDir := t.TempDir()

	loader := &Loader{
		pluginName: "testplugin",
		embedFS:    testEmbedFS,
		customBase: tmpDir,
	}

	t.Run("first dump succeeds", func(t *testing.T) {
		dumped, err := loader.DumpAllTemplates(false)
		if err != nil {
			t.Fatalf("unexpected error on first dump: %v", err)
		}
		if len(dumped) == 0 {
			t.Error("expected at least one dumped template")
		}
	})

	t.Run("second dump without force returns error but lists all files", func(t *testing.T) {
		dumped, err := loader.DumpAllTemplates(false)

		// Should return error about existing files.
		if err == nil {
			t.Error("expected error when files already exist")
		}

		// Should contain "already exists" in error.
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			t.Errorf("expected 'already exists' in error, got: %v", err)
		}

		// Should have empty dumped list (nothing was dumped).
		if len(dumped) != 0 {
			t.Errorf("expected empty dumped list, got %d items", len(dumped))
		}
	})

	t.Run("dump with force overwrites files", func(t *testing.T) {
		dumped, err := loader.DumpAllTemplates(true)
		if err != nil {
			t.Fatalf("unexpected error with force flag: %v", err)
		}
		if len(dumped) == 0 {
			t.Error("expected at least one dumped template")
		}
	})
}

func TestLoader_DumpAllTemplates_PartialExisting(t *testing.T) {
	tmpDir := t.TempDir()

	loader := &Loader{
		pluginName: "testplugin",
		embedFS:    testEmbedFS,
		customBase: tmpDir,
	}

	// Get list of templates.
	templates, err := loader.ListEmbeddedTemplates()
	if err != nil {
		t.Fatalf("failed to list templates: %v", err)
	}
	if len(templates) == 0 {
		t.Skip("no templates to test with")
	}

	// Dump only the first template manually.
	if err := loader.DumpTemplate(templates[0], false); err != nil {
		t.Fatalf("failed to dump first template: %v", err)
	}

	// Now try to dump all - should skip existing and dump others.
	dumped, err := loader.DumpAllTemplates(false)

	// If we have more than one template, should get an error about existing file.
	// but should have dumped the others
	if len(templates) > 1 {
		if err == nil {
			t.Error("expected error about existing file")
		}
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			t.Errorf("expected 'already exists' in error, got: %v", err)
		}
		// Should have dumped the remaining templates.
		if len(dumped) != len(templates)-1 {
			t.Errorf("expected %d dumped templates, got %d", len(templates)-1, len(dumped))
		}
	} else {
		// Only one template, should just get the error.
		if err == nil {
			t.Error("expected error about existing file")
		}
		if len(dumped) != 0 {
			t.Error("expected no dumped templates")
		}
	}
}

func TestLoader_GetInfo(t *testing.T) {
	tmpDir := t.TempDir()

	loader := &Loader{
		pluginName: "testplugin",
		embedFS:    testEmbedFS,
		customBase: tmpDir,
	}

	t.Run("info for embedded-only template", func(t *testing.T) {
		info := loader.GetInfo("testdata/test.tmpl")
		if !info.EmbeddedExists {
			t.Error("expected embedded template to exist")
		}
		if info.CustomExists {
			t.Error("expected custom template not to exist")
		}
		if info.UsingCustom {
			t.Error("expected to use embedded template")
		}
	})

	t.Run("info for custom template", func(t *testing.T) {
		// Create a custom template.
		customPath := loader.CustomPath("testdata/test.tmpl")
		if err := os.MkdirAll(filepath.Dir(customPath), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(customPath, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		info := loader.GetInfo("testdata/test.tmpl")
		if !info.EmbeddedExists {
			t.Error("expected embedded template to exist")
		}
		if !info.CustomExists {
			t.Error("expected custom template to exist")
		}
		if !info.UsingCustom {
			t.Error("expected to use custom template")
		}
	})
}
