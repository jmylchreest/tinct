package utils

import "github.com/jmylchreest/tinct/pkg/util/subprocess"

// IsValidPath validates literal exec arguments, not shell source or containment.
func IsValidPath(path string) bool { return subprocess.ValidPath(path) }
