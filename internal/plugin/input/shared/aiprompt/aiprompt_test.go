package aiprompt

import (
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/plugin/input/shared/aiflags"
)

func TestEnhanceForWallpaper(t *testing.T) {
	original := aiflags.NoExtendedPrompt
	t.Cleanup(func() { aiflags.NoExtendedPrompt = original })

	base := "a mountain landscape"

	aiflags.NoExtendedPrompt = false
	enhanced := EnhanceForWallpaper(base)
	if enhanced == base {
		t.Error("expected prompt to be enhanced when NoExtendedPrompt is false")
	}
	if !strings.HasPrefix(enhanced, base) {
		t.Errorf("expected enhanced prompt to start with base; got %q", enhanced)
	}
	if !strings.Contains(enhanced, WallpaperEnhancement) {
		t.Error("enhanced prompt should contain WallpaperEnhancement")
	}

	aiflags.NoExtendedPrompt = true
	if got := EnhanceForWallpaper(base); got != base {
		t.Errorf("expected base prompt unchanged when NoExtendedPrompt is true; got %q", got)
	}
}

func TestBuildNegative(t *testing.T) {
	if got := BuildNegative("", false); got != DefaultNegativePrompt {
		t.Errorf("BuildNegative(\"\", false) = %q; want DefaultNegativePrompt", got)
	}

	user := "ugly, bad quality"
	want := user + ", " + DefaultNegativePrompt
	if got := BuildNegative(user, false); got != want {
		t.Errorf("BuildNegative(user, false) = %q; want %q", got, want)
	}

	if got := BuildNegative("", true); got != "" {
		t.Errorf("BuildNegative(\"\", true) = %q; want empty", got)
	}

	if got := BuildNegative(user, true); got != "" {
		t.Errorf("BuildNegative(user, true) = %q; want empty (suppress wins)", got)
	}
}
