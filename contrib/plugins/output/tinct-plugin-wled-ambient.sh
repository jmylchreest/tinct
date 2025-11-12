#!/bin/bash
# wled-ambient.sh - WLED ambient monitor lighting plugin
#
# Controls WLED LED strips using positional colours for monitor bias lighting.
# Requires ambient extraction enabled: --image.extractAmbience
#
# Example usage:
#   tinct generate -i image -p wallpaper.jpg \
#     --image.extractAmbience \
#     --image.ambienceRegions 8 \
#     -o wled-ambient \
#     --plugin-args 'wled-ambient={"host":"192.168.1.100","segments":[0,1,2,3]}'
#
# Author: Tinct Contributors
# License: MIT

set -e

# Handle --plugin-info flag
if [ "$1" = "--plugin-info" ]; then
  cat <<'EOF'
{
  "name": "wled-ambient",
  "type": "output",
  "version": "0.0.1",
  "protocol_version": "0.0.1",
  "description": "WLED ambient monitor lighting using positional colours",
  "enabled": false,
  "author": "Tinct Contributors"
}
EOF
  exit 0
fi

# Read JSON palette from stdin
PALETTE=$(cat)

# Extract dry-run flag
DRY_RUN=$(echo "$PALETTE" | jq -r '.dry_run // false')

# Extract plugin args
PLUGIN_ARGS=$(echo "$PALETTE" | jq -r '.plugin_args // {}')
WLED_HOST=$(echo "$PLUGIN_ARGS" | jq -r '.host // "192.168.1.100"')
WLED_SEGMENTS=$(echo "$PLUGIN_ARGS" | jq -r '.segments // [0] | @json')
BRIGHTNESS=$(echo "$PLUGIN_ARGS" | jq -r '.brightness // 128')

echo "==============================================="
echo "WLED Ambient Monitor Lighting"
echo "==============================================="
echo ""
echo "Configuration:"
echo "  WLED Host:    $WLED_HOST"
echo "  Segments:     $WLED_SEGMENTS"
echo "  Brightness:   $BRIGHTNESS"
echo "  Dry Run:      $DRY_RUN"
echo ""

# Extract positional colours for monitor edges
# Typical monitor layout (8 positions):
#   topLeft - top - topRight
#   left           right
#   bottomLeft - bottom - bottomRight

TOP_LEFT=$(echo "$PALETTE" | jq -r '.colours.topLeft.hex // ""')
TOP=$(echo "$PALETTE" | jq -r '.colours.top.hex // ""')
TOP_RIGHT=$(echo "$PALETTE" | jq -r '.colours.topRight.hex // ""')
RIGHT=$(echo "$PALETTE" | jq -r '.colours.right.hex // ""')
BOTTOM_RIGHT=$(echo "$PALETTE" | jq -r '.colours.bottomRight.hex // ""')
BOTTOM=$(echo "$PALETTE" | jq -r '.colours.bottom.hex // ""')
BOTTOM_LEFT=$(echo "$PALETTE" | jq -r '.colours.bottomLeft.hex // ""')
LEFT=$(echo "$PALETTE" | jq -r '.colours.left.hex // ""')

# Check if positional colours are available
if [ -z "$TOP_LEFT" ] || [ -z "$TOP" ] || [ -z "$TOP_RIGHT" ]; then
  echo "ERROR: Positional colours not found in palette"
  echo ""
  echo "Please enable ambient extraction:"
  echo "  tinct generate -i image -p wallpaper.jpg \\"
  echo "    --image.extractAmbience \\"
  echo "    --image.ambienceRegions 8 \\"
  echo "    -o wled-ambient"
  echo ""
  exit 1
fi

echo "Extracted Positional Colours:"
echo "  Top Left:     $TOP_LEFT"
echo "  Top:          $TOP"
echo "  Top Right:    $TOP_RIGHT"
echo "  Right:        $RIGHT"
echo "  Bottom Right: $BOTTOM_RIGHT"
echo "  Bottom:       $BOTTOM"
echo "  Bottom Left:  $BOTTOM_LEFT"
echo "  Left:         $LEFT"
echo ""

# Convert hex to RGB for WLED API
hex_to_rgb() {
  local hex=$1
  hex=${hex#\#}
  printf "%d,%d,%d" 0x${hex:0:2} 0x${hex:2:2} 0x${hex:4:2}
}

# Build LED colour array (clockwise from top-left)
LED_COLOURS=(
  "$(hex_to_rgb $TOP_LEFT)"
  "$(hex_to_rgb $TOP)"
  "$(hex_to_rgb $TOP_RIGHT)"
  "$(hex_to_rgb $RIGHT)"
  "$(hex_to_rgb $BOTTOM_RIGHT)"
  "$(hex_to_rgb $BOTTOM)"
  "$(hex_to_rgb $BOTTOM_LEFT)"
  "$(hex_to_rgb $LEFT)"
)

# Build WLED JSON payload
# WLED API: http://[host]/json/state
build_wled_payload() {
  local segment=$1
  cat <<EOF
{
  "on": true,
  "bri": $BRIGHTNESS,
  "seg": [{
    "id": $segment,
    "i": [
EOF

  # Add individual LED colours
  for i in "${!LED_COLOURS[@]}"; do
    echo "      [${LED_COLOURS[$i]}]"
    if [ $i -lt $((${#LED_COLOURS[@]} - 1)) ]; then
      echo "      ,"
    fi
  done

  cat <<EOF
    ]
  }]
}
EOF
}

# Send to WLED
send_to_wled() {
  local segment=$1
  local payload=$(build_wled_payload $segment)

  if [ "$DRY_RUN" = "true" ]; then
    echo "Would POST to: http://$WLED_HOST/json/state"
    echo "Payload:"
    echo "$payload" | jq '.'
  else
    echo "Sending to segment $segment..."
    response=$(curl -s -X POST \
      -H "Content-Type: application/json" \
      -d "$payload" \
      "http://$WLED_HOST/json/state")

    if [ $? -eq 0 ]; then
      echo "  ✓ Success (segment $segment)"
    else
      echo "  ✗ Failed (segment $segment)"
      return 1
    fi
  fi
}

# Process each segment
echo "Updating WLED segments..."
echo ""

SEGMENT_IDS=$(echo "$WLED_SEGMENTS" | jq -r '.[]')
for seg_id in $SEGMENT_IDS; do
  send_to_wled $seg_id
done

echo ""
echo "==============================================="
echo "Status: SUCCESS"
echo "Updated: $(echo "$WLED_SEGMENTS" | jq 'length') segment(s)"
echo "Mode: $([ "$DRY_RUN" = "true" ] && echo "DRY-RUN" || echo "APPLIED")"
echo "==============================================="

exit 0
