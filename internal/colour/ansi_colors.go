package colour

import (
	pkgcolour "github.com/jmylchreest/tinct/pkg/colour"
)

// Re-export ANSI color functions from pkg/colour for backward compatibility.
var (
	GetANSIColorNames = pkgcolour.GetANSIColorNames
	ANSIColorInfo     = pkgcolour.ANSIColorInfo
)
