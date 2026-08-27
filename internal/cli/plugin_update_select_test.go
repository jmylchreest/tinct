package cli

import (
	"strings"
	"testing"
)

func testManifest(names ...string) *PluginManifest {
	m := &PluginManifest{ExternalPlugins: map[string]*ExternalPluginMeta{}}
	for _, n := range names {
		m.ExternalPlugins[n] = &ExternalPluginMeta{Name: n}
	}
	return m
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// The regression: `tinct plugins update <name>` previously ignored the
// name and updated every plugin in the manifest.
func TestSelectPluginsToUpdate(t *testing.T) {
	lock := testManifest("zed", "noctalia", "random", "wob")

	tests := []struct {
		name  string
		names []string
		want  []string
	}{
		{
			name:  "no names updates everything, sorted",
			names: nil,
			want:  []string{"noctalia", "random", "wob", "zed"},
		},
		{
			name:  "one name updates only that plugin",
			names: []string{"random"},
			want:  []string{"random"},
		},
		{
			name:  "several names, sorted",
			names: []string{"zed", "noctalia"},
			want:  []string{"noctalia", "zed"},
		},
		{
			name:  "duplicates collapse",
			names: []string{"zed", "zed"},
			want:  []string{"zed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectPluginsToUpdate(lock, tt.names)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !equal(got, tt.want) {
				t.Errorf("selected %v, want %v", got, tt.want)
			}
		})
	}
}

// An unknown name must fail loudly rather than silently updating
// everything, which is what made the old behaviour so surprising.
func TestSelectPluginsToUpdateRejectsUnknown(t *testing.T) {
	lock := testManifest("zed", "noctalia")

	got, err := selectPluginsToUpdate(lock, []string{"nope"})
	if err == nil {
		t.Fatalf("expected an error, got selection %v", got)
	}
	for _, want := range []string{"nope", "noctalia", "zed"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q, want it to mention %q", err.Error(), want)
		}
	}
}

// A known name alongside an unknown one must still fail — partially
// applying the request would be worse than refusing it.
func TestSelectPluginsToUpdateRejectsMixedUnknown(t *testing.T) {
	lock := testManifest("zed")

	if _, err := selectPluginsToUpdate(lock, []string{"zed", "nope"}); err == nil {
		t.Fatal("expected an error when one of several names is unknown")
	}
}
