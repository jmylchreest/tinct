// Package themeformat provides types and utilities for the markdown theme format.
package themeformat

import (
	"time"
)

// Theme represents a complete theme file with metadata, colors, and wallpaper.
type Theme struct {
	Name      string    `yaml:"name"`
	Version   string    `yaml:"version"`
	Generator string    `yaml:"generator"`
	Generated time.Time `yaml:"generated"`

	Source    *Source    `yaml:"source,omitempty"`
	Wallpaper *Wallpaper `yaml:"wallpaper,omitempty"`
	Colors    []Color    `yaml:"colors"`
	Ambience  []Ambience `yaml:"ambience,omitempty"`
}

// Source contains information about how the theme was generated.
type Source struct {
	Plugin string `yaml:"plugin"`
	Model  string `yaml:"model,omitempty"`
	Prompt string `yaml:"prompt,omitempty"`
}

// Wallpaper contains wallpaper image data or references.
type Wallpaper struct {
	Embedded bool   `yaml:"embedded"`
	Format   string `yaml:"format,omitempty"`
	Width    int    `yaml:"width,omitempty"`
	Height   int    `yaml:"height,omitempty"`
	Data     string `yaml:"data,omitempty"` // base64, only if embedded
	Path     string `yaml:"path,omitempty"` // relative/absolute path
	URL      string `yaml:"url,omitempty"`  // remote URL
}

// Color represents a color in the palette.
type Color struct {
	Hex    string  `yaml:"hex"`
	Weight float64 `yaml:"weight"`
	Role   string  `yaml:"role,omitempty"`
}

// Ambience represents an ambient/edge color.
type Ambience struct {
	Hex    string  `yaml:"hex"`
	Region string  `yaml:"region"`
	Weight float64 `yaml:"weight"`
}

// Thumbnail contains a smaller preview image for the markdown body.
type Thumbnail struct {
	Format string `yaml:"format"`
	Width  int    `yaml:"width"`
	Height int    `yaml:"height"`
	Data   string `yaml:"data"` // base64
}
