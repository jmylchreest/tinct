// Package template provides template utilities and functions for tinct plugins.
package template

import (
	"bytes"
	"embed"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed testdata/*.tmpl templates/*/*.tmpl
var testEmbedFS embed.FS

func TestLoader_Load(t *testing.T) {
	tmpDir := t.TempDir()

	loader := &Loader{
		pluginName: "testplugin",
		templateFS: testEmbedFS,
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
		customDir := filepath.Join(tmpDir, "testplugin")
		if err := os.MkdirAll(customDir, 0o755); err != nil {
			t.Fatalf("failed to create custom dir: %v", err)
		}

		customContent := []byte("# This is a custom template\n")
		customPath := filepath.Join(customDir, "testdata", "test.tmpl")
		if err := os.MkdirAll(filepath.Dir(customPath), 0o755); err != nil {
			t.Fatalf("failed to create custom template dir: %v", err)
		}
		if err := os.WriteFile(customPath, customContent, 0o644); err != nil {
			t.Fatalf("failed to write custom template: %v", err)
		}

		content, fromCustom, err := loader.Load("testdata/test.tmpl")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !fromCustom {
			t.Error("expected custom template, got embedded")
		}
		if !bytes.Equal(content, customContent) {
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
		if err := os.MkdirAll(filepath.Dir(customPath), 0o755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(customPath, []byte("test"), 0o644); err != nil {
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
		templateFS: testEmbedFS,
	}

	templates, err := loader.ListEmbeddedTemplates()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(templates) == 0 {
		t.Error("expected at least one template")
	}

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
		templateFS: testEmbedFS,
		customBase: tmpDir,
	}

	t.Run("dumps template successfully", func(t *testing.T) {
		err := loader.DumpTemplate("testdata/test.tmpl", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		customPath := loader.CustomPath("testdata/test.tmpl")
		if _, err := os.Stat(customPath); err != nil {
			t.Errorf("custom template not created: %v", err)
		}
	})

	t.Run("fails without force when template exists", func(t *testing.T) {
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
		templateFS: testEmbedFS,
		customBase: tmpDir,
	}

	dumped, err := loader.DumpAllTemplates(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(dumped) == 0 {
		t.Error("expected at least one dumped template")
	}

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
		templateFS: testEmbedFS,
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

		if err == nil {
			t.Error("expected error when files already exist")
		}

		if err != nil && !strings.Contains(err.Error(), "already exists") {
			t.Errorf("expected 'already exists' in error, got: %v", err)
		}

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
		templateFS: testEmbedFS,
		customBase: tmpDir,
	}

	templates, err := loader.ListEmbeddedTemplates()
	if err != nil {
		t.Fatalf("failed to list templates: %v", err)
	}
	if len(templates) == 0 {
		t.Skip("no templates to test with")
	}

	if err := loader.DumpTemplate(templates[0], false); err != nil {
		t.Fatalf("failed to dump first template: %v", err)
	}

	dumped, err := loader.DumpAllTemplates(false)

	if len(templates) <= 1 {
		if err == nil {
			t.Error("expected error about existing file")
		}
		if len(dumped) != 0 {
			t.Error("expected no dumped templates")
		}
		return
	}

	if err == nil {
		t.Error("expected error about existing file")
	}
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' in error, got: %v", err)
	}
	if len(dumped) != len(templates)-1 {
		t.Errorf("expected %d dumped templates, got %d", len(templates)-1, len(dumped))
	}
}

func TestLoader_GetInfo(t *testing.T) {
	tmpDir := t.TempDir()

	loader := &Loader{
		pluginName: "testplugin",
		templateFS: testEmbedFS,
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
		customPath := loader.CustomPath("testdata/test.tmpl")
		if err := os.MkdirAll(filepath.Dir(customPath), 0o755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(customPath, []byte("test"), 0o644); err != nil {
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

func TestLoader_VersionedTemplates(t *testing.T) { //nolint:gocognit // comprehensive versioning test
	tmpDir := t.TempDir()

	loader := &Loader{
		pluginName: "testplugin",
		templateFS: testEmbedFS,
		customBase: tmpDir,
	}

	t.Run("loads default template when no version specified", func(t *testing.T) {
		content, fromCustom, err := loader.Load("testdata/test.tmpl")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fromCustom {
			t.Error("expected embedded template")
		}
		if !strings.Contains(string(content), "Test Template") {
			t.Errorf("expected default template content, got: %s", string(content))
		}
	})

	t.Run("loads versioned template when version matches exactly", func(t *testing.T) {
		loader.targetVersion = "0.53.0"
		content, _, err := loader.Load("test.tmpl")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(content), "Version 0.53") {
			t.Errorf("expected 0.53 template content, got: %s", string(content))
		}
		if !strings.Contains(string(content), "tinct-error-border") {
			t.Errorf("expected new syntax with descriptive name, got: %s", string(content))
		}
	})

	t.Run("loads older versioned template when target is between versions", func(t *testing.T) {
		loader.targetVersion = "0.52.5"
		content, _, err := loader.Load("test.tmpl")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(content), "Version 0.52") {
			t.Errorf("expected 0.52 template content, got: %s", string(content))
		}
	})

	t.Run("loads newer version template when target is higher", func(t *testing.T) {
		loader.targetVersion = "0.55.0"
		content, _, err := loader.Load("test.tmpl")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(content), "Version 0.53") {
			t.Errorf("expected 0.53 template content, got: %s", string(content))
		}
	})

	t.Run("falls back to default when target is older than all versions", func(t *testing.T) {
		loader.targetVersion = "0.50.0"
		content, _, err := loader.Load("testdata/test.tmpl")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(content), "Test Template") {
			t.Errorf("expected default template content, got: %s", string(content))
		}
	})

	t.Run("custom template takes priority over versioned", func(t *testing.T) {
		customPath := filepath.Join(tmpDir, "testplugin", "test.tmpl")
		if err := os.MkdirAll(filepath.Dir(customPath), 0o755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		customContent := "# Custom template\n"
		if err := os.WriteFile(customPath, []byte(customContent), 0o644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		loader.targetVersion = "0.53.0"
		content, fromCustom, err := loader.Load("test.tmpl")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !fromCustom {
			t.Error("expected custom template to take priority")
		}
		if !strings.Contains(string(content), "Custom template") {
			t.Errorf("expected custom template content, got: %s", string(content))
		}
	})
}

func TestLoader_FindVersionDirectories(t *testing.T) {
	loader := &Loader{
		pluginName: "testplugin",
		templateFS: testEmbedFS,
		customBase: t.TempDir(),
	}

	versions := loader.findVersionDirectories()

	if len(versions) != 2 {
		t.Errorf("expected 2 versions, got %d: %v", len(versions), versions)
	}

	if len(versions) >= 2 {
		if versions[0] != "0.52" {
			t.Errorf("expected first version to be 0.52, got %s", versions[0])
		}
		if versions[1] != "0.53" {
			t.Errorf("expected second version to be 0.53, got %s", versions[1])
		}
	}
}
