package colour

import (
	"maps"

	pkgcolour "github.com/jmylchreest/tinct/pkg/colour"
)

// ThemeData is the standard data structure passed to all plugin templates.
// It embeds the public ThemeData and adds internal-specific fields.
type ThemeData struct {
	*pkgcolour.ThemeData

	// OutputDir is the directory where output files will be written.
	// This allows templates to reference other generated files with correct paths.
	OutputDir string

	// ColorFileName is the name of the primary color palette file being generated.
	// This allows stub/config templates to reference the correct color file.
	ColorFileName string

	// internalPalette stores the original internal palette for backward compatibility.
	internalPalette *CategorisedPalette
}

// Palette returns the internal CategorisedPalette for backward compatibility.
func (td *ThemeData) Palette() *CategorisedPalette {
	return td.internalPalette
}

// GetPaletteHelper returns the embedded public PaletteHelper.
// This method allows internal ThemeData to satisfy the pkg/template.PaletteProvider interface,
// enabling the internal template functions to delegate to the public ones.
func (td *ThemeData) GetPaletteHelper() *pkgcolour.PaletteHelper {
	return td.ThemeData.PaletteHelper
}

// NewThemeData creates a new ThemeData instance with the given palette.
// Optional wallpaperPath and themeName can be provided (pass empty strings if not needed).
// This is the internal factory that works with internal CategorisedPalette.
func NewThemeData(palette *CategorisedPalette, wallpaperPath, themeName string) *ThemeData {
	// Convert internal CategorisedPalette to pkg version.
	pkgPalette := &pkgcolour.CategorisedPalette{
		Colours:    make(map[pkgcolour.Role]pkgcolour.CategorisedColour),
		ThemeType:  palette.ThemeType,
		AllColours: make([]pkgcolour.CategorisedColour, len(palette.AllColours)),
		ANSI:       palette.ANSI,
	}

	// Copy colours map.
	maps.Copy(pkgPalette.Colours, palette.Colours)

	// Copy AllColours slice.
	copy(pkgPalette.AllColours, palette.AllColours)

	return &ThemeData{
		ThemeData:       pkgcolour.NewThemeData(pkgPalette, wallpaperPath, themeName),
		internalPalette: palette,
	}
}
