package subprocess

import "testing"

func TestValidPath(t *testing.T) {
	for _, path := range []string{"/tmp/my ($wallpaper)..png", "../wallpaper.png", "quote'and;pipe|.png"} {
		if !ValidPath(path) {
			t.Errorf("rejected literal path %q", path)
		}
	}
	for _, path := range []string{"", "bad\x00path"} {
		if ValidPath(path) {
			t.Errorf("accepted invalid path %q", path)
		}
	}
}
