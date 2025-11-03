package colour

// ThemeData is the standard data structure passed to all plugin templates.
// It embeds PaletteHelper to provide all color access methods and includes
// additional optional fields that plugins can populate.
type ThemeData struct {
	*PaletteHelper

	// WallpaperPath is the path to the wallpaper file, populated by plugins
	// that implement the WallpaperContextProvider interface
	WallpaperPath string

	// ThemeName is an optional theme name that can be set by plugins
	ThemeName string
}

// NewThemeData creates a new ThemeData instance with the given palette.
// Optional wallpaperPath and themeName can be provided (pass empty strings if not needed).
func NewThemeData(palette *CategorisedPalette, wallpaperPath, themeName string) *ThemeData {
	return &ThemeData{
		PaletteHelper: NewPaletteHelper(palette),
		WallpaperPath: wallpaperPath,
		ThemeName:     themeName,
	}
}
