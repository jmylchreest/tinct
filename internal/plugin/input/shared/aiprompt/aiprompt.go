// Package aiprompt provides shared prompt-construction helpers for AI image
// generation plugins (e.g. googlegenai, openrouter). The contents — wallpaper
// enhancement suffix, default negative prompt, and the combiners that respect
// aiflags.NoExtendedPrompt / aiflags.NegativePrompt — are not specific to any
// one model provider, so all AI input plugins should use them to keep the
// generated images visually consistent.
package aiprompt

import (
	"fmt"

	"github.com/jmylchreest/tinct/internal/plugin/input/shared/aiflags"
)

// WallpaperEnhancement is appended to user prompts to bias image generators
// toward wallpaper-suitable outputs (edge-to-edge, no borders, etc.). Kept in
// one place so updates apply uniformly across providers.
const WallpaperEnhancement = ", high quality desktop wallpaper suitable for widescreen and ultrawidescreen computer monitors, edge-to-edge composition, full bleed, seamless edges, vibrant colors, no borders, no frames, no padding"

// DefaultNegativePrompt enumerates artefacts (borders, letterboxing, vignettes,
// matting, ...) that we never want in a generated wallpaper. Used as the base
// negative prompt by every AI input plugin.
const DefaultNegativePrompt = "white borders, white edges, black borders, black edges, gray borders, padding, margins, letterbox, pillarbox, widescreen bars, black bars, frames, picture frames, border around image, vignette edges, faded edges, cropped edges, incomplete edges, cut off edges, canvas texture, matting, mounting"

// EnhanceForWallpaper returns basePrompt with WallpaperEnhancement appended,
// unless aiflags.NoExtendedPrompt is set (in which case basePrompt is returned
// unchanged so power users can drive the model directly).
func EnhanceForWallpaper(basePrompt string) string {
	if aiflags.NoExtendedPrompt {
		return basePrompt
	}
	return basePrompt + WallpaperEnhancement
}

// BuildNegative returns a negative prompt for image generation. If suppress is
// true the result is empty (some users want zero filtering). Otherwise the
// caller-supplied userExtra is prepended to DefaultNegativePrompt; pass "" for
// userExtra to use only the default. This matches the historic behaviour of
// both googlegenai (caller passes aiflags.NegativePrompt) and openrouter
// (which read aiflags.NegativePrompt directly).
func BuildNegative(userExtra string, suppress bool) string {
	if suppress {
		return ""
	}
	if userExtra == "" {
		return DefaultNegativePrompt
	}
	return fmt.Sprintf("%s, %s", userExtra, DefaultNegativePrompt)
}
