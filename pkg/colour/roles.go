// Package colour provides public types and utilities for working with color palettes.
package colour

// Role represents the semantic role of a colour in a theme.
type Role string

const (
	// Core roles.
	RoleBackground      Role = "background"
	RoleBackgroundMuted Role = "backgroundMuted"
	RoleForeground      Role = "foreground"
	RoleForegroundMuted Role = "foregroundMuted"

	// Accent roles.
	RoleAccent1      Role = "accent1"
	RoleAccent1Muted Role = "accent1Muted"
	RoleAccent2      Role = "accent2"
	RoleAccent2Muted Role = "accent2Muted"
	RoleAccent3      Role = "accent3"
	RoleAccent3Muted Role = "accent3Muted"
	RoleAccent4      Role = "accent4"
	RoleAccent4Muted Role = "accent4Muted"

	// Semantic roles.
	RoleDanger       Role = "danger"
	RoleWarning      Role = "warning"
	RoleSuccess      Role = "success"
	RoleInfo         Role = "info"
	RoleNotification Role = "notification"

	// Surface and container roles (Priority 1 - Material Design 3).
	RoleSurface   Role = "surface"   // Base surface for cards, sheets, dialogs
	RoleOnSurface Role = "onSurface" // Text/icons on surface
	RoleOutline   Role = "outline"   // Borders, dividers, outlines
	RoleBorder    Role = "border"    // Primary border color

	// Surface and border variants (Priority 2).
	RoleSurfaceVariant   Role = "surfaceVariant"   // Alternate surface color
	RoleOnSurfaceVariant Role = "onSurfaceVariant" // Text on surface variant
	RoleBorderMuted      Role = "borderMuted"      // Inactive/muted borders
	RoleOutlineVariant   Role = "outlineVariant"   // Secondary outline

	// On-colors for accents (Priority 2).
	RoleOnAccent1 Role = "onAccent1" // Text on accent1 background
	RoleOnAccent2 Role = "onAccent2" // Text on accent2 background
	RoleOnAccent3 Role = "onAccent3" // Text on accent3 background
	RoleOnAccent4 Role = "onAccent4" // Text on accent4 background

	// On-colors for semantic roles (Priority 2).
	RoleOnDanger  Role = "onDanger"  // Text on danger background
	RoleOnWarning Role = "onWarning" // Text on warning background
	RoleOnSuccess Role = "onSuccess" // Text on success background
	RoleOnInfo    Role = "onInfo"    // Text on info background

	// Inverse colors for overlays (Priority 3).
	RoleInverseSurface   Role = "inverseSurface"   // Inverse surface (tooltip backgrounds)
	RoleInverseOnSurface Role = "inverseOnSurface" // Text on inverse surface
	RoleInversePrimary   Role = "inversePrimary"   // Inverse accent color

	// Scrim and shadow with alpha (Priority 3).
	RoleScrim  Role = "scrim"  // Modal backdrop overlay (with alpha)
	RoleShadow Role = "shadow" // Elevation shadows (with alpha)

	// Surface container elevation variants (Priority 3 - Material Design 3).
	RoleSurfaceContainerLowest  Role = "surfaceContainerLowest"  // Lowest elevation
	RoleSurfaceContainerLow     Role = "surfaceContainerLow"     // Low elevation
	RoleSurfaceContainer        Role = "surfaceContainer"        // Default container
	RoleSurfaceContainerHigh    Role = "surfaceContainerHigh"    // High elevation
	RoleSurfaceContainerHighest Role = "surfaceContainerHighest" // Highest elevation

	// Positional roles for ambient lighting.
	// Core 8 positions (corners + mid-edges).
	RolePositionTopLeft     Role = "positionTopLeft"
	RolePositionTop         Role = "positionTop"
	RolePositionTopRight    Role = "positionTopRight"
	RolePositionRight       Role = "positionRight"
	RolePositionBottomRight Role = "positionBottomRight"
	RolePositionBottom      Role = "positionBottom"
	RolePositionBottomLeft  Role = "positionBottomLeft"
	RolePositionLeft        Role = "positionLeft"

	// Extended positions for 12+ region configurations.
	RolePositionTopLeftInner     Role = "positionTopLeftInner"
	RolePositionTopCenter        Role = "positionTopCenter"
	RolePositionTopRightInner    Role = "positionTopRightInner"
	RolePositionRightTop         Role = "positionRightTop"
	RolePositionRightBottom      Role = "positionRightBottom"
	RolePositionBottomRightInner Role = "positionBottomRightInner"
	RolePositionBottomCenter     Role = "positionBottomCenter"
	RolePositionBottomLeftInner  Role = "positionBottomLeftInner"
	RolePositionLeftBottom       Role = "positionLeftBottom"
	RolePositionLeftTop          Role = "positionLeftTop"

	// Ultra-extended positions for 16+ region configurations.
	RolePositionTopLeftOuter      Role = "positionTopLeftOuter"
	RolePositionTopLeftCenter     Role = "positionTopLeftCenter"
	RolePositionTopRightCenter    Role = "positionTopRightCenter"
	RolePositionTopRightOuter     Role = "positionTopRightOuter"
	RolePositionRightTopOuter     Role = "positionRightTopOuter"
	RolePositionRightBottomOuter  Role = "positionRightBottomOuter"
	RolePositionBottomRightOuter  Role = "positionBottomRightOuter"
	RolePositionBottomRightCenter Role = "positionBottomRightCenter"
	RolePositionBottomLeftCenter  Role = "positionBottomLeftCenter"
	RolePositionBottomLeftOuter   Role = "positionBottomLeftOuter"
	RolePositionLeftBottomOuter   Role = "positionLeftBottomOuter"
	RolePositionLeftTopOuter      Role = "positionLeftTopOuter"
)
