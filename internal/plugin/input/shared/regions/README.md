# Regions Package

Internal utility package for extracting colors from specific regions of 2D images for ambient lighting applications.

## Overview

This package provides functionality to sample colors from edges and corners of images in various configurations (4, 8, 12, or 16 regions). It's designed to support ambient LED lighting setups where lights are positioned around a monitor and should match colors from the wallpaper edges.

## Features

- **Multiple Configurations:** Support for 4, 8, 12, and 16 region layouts
- **Flexible Sampling:** Configurable sample size and extraction method
- **RoleHints Integration:** Automatically maps extracted colors to positional semantic roles
- **Two Extraction Methods:** 
  - Average: Calculate average color of all pixels in region (default)
  - Dominant: Find most frequent color in region (more vibrant)

## Configurations

### 4 Regions - Corners Only
```
[TL]              [TR]
       Monitor
[BL]              [BR]
```
- **Use Case:** Budget setups, small monitors
- **Positions:** 4 corners

### 8 Regions - Corners + Mid-Edges (Default)
```
     [T]      
[L]       [R]
    Monitor
[L]       [R]
     [B]      
```
- **Use Case:** Standard setups, 24-32" monitors
- **Positions:** 4 corners + 4 mid-edges
- **Recommended for most users**

### 12 Regions - Corners + Double Edges
```
  [TL] [T] [TR]
[L]           [R]
     Monitor
[L]           [R]
  [BL] [B] [BR]
```
- **Use Case:** Large monitors, ultrawide displays
- **Positions:** 4 corners + 8 edge positions (2 per edge)

### 16 Regions - Corners + Triple Edges
```
[TL][TL][T][TR][TR]
[L]            [R]
     Monitor
[L]            [R]
[BL][BL][B][BR][BR]
```
- **Use Case:** Enthusiast setups, massive displays
- **Positions:** 4 corners + 12 edge positions (3 per edge)

## Usage

### Basic Example

```go
package main

import (
    "image"
    _ "image/jpeg"
    "os"
    
    "github.com/jmylchreest/tinct/internal/plugin/input/shared/regions"
)

func main() {
    // Load an image
    file, _ := os.Open("wallpaper.jpg")
    defer file.Close()
    img, _, _ := image.Decode(file)
    
    // Create sampler
    sampler := regions.NewSampler()
    
    // Extract 8 regions (default)
    palette, err := sampler.Extract(img, regions.Config8Regions)
    if err != nil {
        panic(err)
    }
    
    // palette.Colors contains the extracted colors
    // palette.RoleHints maps positional roles to color indices
}
```

### Custom Configuration

```go
// Create sampler with custom settings
sampler := &regions.Sampler{
    SamplePercent: 15,      // Sample 15% of edges (default: 10%)
    Method:        "dominant", // Use dominant color (default: "average")
}

// Extract 12 regions
palette, err := sampler.Extract(img, regions.Config12Regions)
```

### Accessing Results

```go
palette, _ := sampler.Extract(img, regions.Config8Regions)

// Colors are in the palette
for i, color := range palette.Colors {
    rgb := colour.ToRGB(color)
    fmt.Printf("Color %d: %s\n", i, rgb.Hex())
}

// RoleHints map positions to color indices
for role, index := range palette.RoleHints {
    color := palette.Colors[index]
    fmt.Printf("%s -> Color %d: %s\n", role, index, colour.ToRGB(color).Hex())
}
```

## Positional Roles

The package uses semantic roles from `internal/colour` to identify positions:

### Core 8 Positions
- `RolePositionTopLeft`
- `RolePositionTop`
- `RolePositionTopRight`
- `RolePositionRight`
- `RolePositionBottomRight`
- `RolePositionBottom`
- `RolePositionBottomLeft`
- `RolePositionLeft`

### Extended Positions (12 regions)
- `RolePositionTopLeftInner`
- `RolePositionTopRightInner`
- `RolePositionRightTop`
- `RolePositionRightBottom`
- `RolePositionBottomRightInner`
- `RolePositionBottomLeftInner`
- `RolePositionLeftBottom`
- `RolePositionLeftTop`

### Ultra Positions (16 regions)
- `RolePositionTopLeftCenter`
- `RolePositionTopCenter`
- `RolePositionTopRightCenter`
- `RolePositionRightTopOuter`
- `RolePositionRightBottomOuter`
- `RolePositionBottomRightCenter`
- `RolePositionBottomCenter`
- `RolePositionBottomLeftCenter`
- `RolePositionLeftBottomOuter`
- `RolePositionLeftTopOuter`

## Sampling Strategy

### Sample Size
- **Default:** 10% of edge dimensions
- **Adjustable:** 1-50% via `SamplePercent`
- **Automatic Scaling:** Reduces for higher region counts to avoid overlap
  - 8 regions: 10% (default)
  - 12 regions: 8%
  - 16 regions: 6%

### Extraction Methods

**Average (default):**
- Calculates arithmetic mean of all pixels in region
- Produces natural, smooth colors
- Best for most use cases

**Dominant:**
- Finds most frequent color in region (quantized)
- Produces more vibrant, saturated colors
- Better for images with distinct color blocks

## Integration with Image Plugin

This package is designed to be used by the image input plugin:

```go
// In image plugin
if mode == "ambient" {
    sampler := regions.NewSampler()
    return sampler.Extract(img, regions.Config8Regions)
}
```

## Performance

Extraction is very fast:
- 4 regions: ~0.02s
- 8 regions: ~0.04s
- 12 regions: ~0.06s
- 16 regions: ~0.08s

Performance scales linearly with region count and sample size.

## Design Decisions

### Why Not More Regions?
- Beyond 16 regions, diminishing visual returns
- Most LED setups have 4-16 lights
- Complexity increases significantly
- Sample overlap becomes an issue

### Why Positional Roles?
- RoleHints system allows explicit mapping
- Output plugins can map roles to physical light IDs
- Separates extraction from output
- Extensible to other LED systems

### Why Sample Percentage?
- Allows users to control precision vs averaging
- Larger samples = smoother, more representative colors
- Smaller samples = more precise, position-specific colors
- Default (10%) works well for most images

## Testing

```bash
# Run tests
go test ./internal/plugin/input/shared/regions

# Run with coverage
go test -cover ./internal/plugin/input/shared/regions
```

## See Also

- [Ambient Lighting Research](../../../../../docs/AMBIENT-LIGHTING-RESEARCH.md)
- [Region Selection Guide](../../../../../docs/AMBIENT-LIGHTING-REGIONS.md)
- [RoleHints Documentation](../../../../../docs/ROLE-HINTS-USAGE.md)