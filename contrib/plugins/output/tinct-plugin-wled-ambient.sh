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
  "protocol_version": "0.2.0",
  "plugin_protocol": "json-stdio",
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

echo "===============================================" >&2
echo "WLED Ambient Monitor Lighting" >&2
echo "===============================================" >&2
echo "" >&2
echo "Configuration:" >&2
echo "  WLED Host:    $WLED_HOST" >&2
echo "  Segments:     $WLED_SEGMENTS" >&2
echo "  Brightness:   $BRIGHTNESS" >&2
echo "  Dry Run:      $DRY_RUN" >&2
echo "" >&2

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
  echo '{"success":false,"files_written":[],"message":"Positional colours not found in palette. Enable ambient extraction with --image.extractAmbience"}' 
  exit 0
fi

echo "Extracted Positional Colours:" >&2
echo "  Top Left:     $TOP_LEFT" >&2
echo "  Top:          $TOP" >&2
echo "  Top Right:    $TOP_RIGHT" >&2
echo "  Right:        $RIGHT" >&2
echo "  Bottom Right: $BOTTOM_RIGHT" >&2
echo "  Bottom:       $BOTTOM" >&2
echo "  Bottom Left:  $BOTTOM_LEFT" >&2
echo "  Left:         $LEFT" >&2
echo "" >&2

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
    echo "Would POST to: http://$WLED_HOST/json/state" >&2
    echo "Payload:" >&2
    echo "$payload" | jq '.' >&2
  else
    echo "Sending to segment $segment..." >&2
    response=$(curl -s -X POST \
      -H "Content-Type: application/json" \
      -d "$payload" \
      "http://$WLED_HOST/json/state")

    if [ $? -eq 0 ]; then
      echo "  ✓ Success (segment $segment)" >&2
    else
      echo "  ✗ Failed (segment $segment)" >&2
      return 1
    fi
  fi
}

# Process each segment
echo "Updating WLED segments..." >&2
echo "" >&2

SEGMENT_IDS=$(echo "$WLED_SEGMENTS" | jq -r '.[]')
for seg_id in $SEGMENT_IDS; do
  send_to_wled $seg_id
done

echo "" >&2
echo "===============================================" >&2
echo "Status: SUCCESS" >&2
echo "Updated: $(echo "$WLED_SEGMENTS" | jq 'length') segment(s)" >&2
echo "Mode: $([ "$DRY_RUN" = "true" ] && echo "DRY-RUN" || echo "APPLIED")" >&2
echo "===============================================" >&2

# Protocol 0.2.0: write structured JSON response to stdout
SEGMENT_COUNT=$(echo "$WLED_SEGMENTS" | jq 'length')
echo "{\"success\":true,\"files_written\":[],\"message\":\"Updated $SEGMENT_COUNT WLED segment(s) on $WLED_HOST\"}"

exit 0
