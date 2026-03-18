#!/bin/bash
# openrgb-peripheral.sh - OpenRGB peripheral lighting plugin
#
# Sets keyboard/mouse/peripheral RGB lighting to match theme background colour.
# Requires OpenRGB SDK server running (openrgb --server)
#
# Example usage:
#   tinct generate -i image -p wallpaper.jpg \
#     -o openrgb-peripheral \
#     --plugin-args 'openrgb-peripheral={"host":"localhost","port":6742,"devices":["keyboard","mouse"]}'
#
# Author: Tinct Contributors
# License: MIT

set -e

# Handle --plugin-info flag
if [ "$1" = "--plugin-info" ]; then
  cat <<'EOF'
{
  "name": "openrgb-peripheral",
  "type": "output",
  "version": "0.0.1",
  "protocol_version": "0.2.0",
  "plugin_protocol": "json-stdio",
  "description": "OpenRGB peripheral lighting using theme background colour",
  "enabled": false,
  "author": "Tinct Contributors"
}
EOF
  exit 0
fi

# Check if openrgb CLI is available
if ! command -v openrgb &> /dev/null; then
  echo '{"success":false,"files_written":[],"message":"openrgb command not found. Please install OpenRGB: https://openrgb.org"}'
  exit 0
fi

# Read JSON palette from stdin
PALETTE=$(cat)

# Extract dry-run flag
DRY_RUN=$(echo "$PALETTE" | jq -r '.dry_run // false')

# Extract plugin args
PLUGIN_ARGS=$(echo "$PALETTE" | jq -r '.plugin_args // {}')
OPENRGB_HOST=$(echo "$PLUGIN_ARGS" | jq -r '.host // "localhost"')
OPENRGB_PORT=$(echo "$PLUGIN_ARGS" | jq -r '.port // 6742')
DEVICE_FILTER=$(echo "$PLUGIN_ARGS" | jq -r '.devices // [] | @json')
MODE=$(echo "$PLUGIN_ARGS" | jq -r '.mode // "static"')
BRIGHTNESS=$(echo "$PLUGIN_ARGS" | jq -r '.brightness // 100')

echo "===============================================" >&2
echo "OpenRGB Peripheral Lighting" >&2
echo "===============================================" >&2
echo "" >&2
echo "Configuration:" >&2
echo "  OpenRGB Host:   $OPENRGB_HOST:$OPENRGB_PORT" >&2
echo "  Device Filter:  $DEVICE_FILTER" >&2
echo "  Mode:           $MODE" >&2
echo "  Brightness:     $BRIGHTNESS%" >&2
echo "  Dry Run:        $DRY_RUN" >&2
echo "" >&2

# Extract background colour (dominant theme colour)
BG_HEX=$(echo "$PALETTE" | jq -r '.colours.background.hex // ""')
if [ -z "$BG_HEX" ]; then
  echo '{"success":false,"files_written":[],"message":"Background colour not found in palette"}'
  exit 0
fi

# Also extract accent for optional effects
ACCENT1_HEX=$(echo "$PALETTE" | jq -r '.colours.accent1.hex // ""')

echo "Theme Colours:" >&2
echo "  Background: $BG_HEX (primary)" >&2
echo "  Accent 1:   $ACCENT1_HEX (optional)" >&2
echo "" >&2

# Convert hex to RGB
hex_to_rgb() {
  local hex=$1
  hex=${hex#\#}
  echo "${hex:0:2} ${hex:2:2} ${hex:4:2}"
}

BG_RGB=($(hex_to_rgb $BG_HEX))
BG_R=$((16#${BG_RGB[0]}))
BG_G=$((16#${BG_RGB[1]}))
BG_B=$((16#${BG_RGB[2]}))

# Apply brightness scaling
BG_R=$(($BG_R * $BRIGHTNESS / 100))
BG_G=$(($BG_G * $BRIGHTNESS / 100))
BG_B=$(($BG_B * $BRIGHTNESS / 100))

echo "RGB Values (with brightness):" >&2
echo "  R: $BG_R, G: $BG_G, B: $BG_B" >&2
echo "" >&2

# List available devices
list_devices() {
  openrgb --noautoconnect --list-devices 2>/dev/null || echo "Could not list devices"
}

# Set colour for a device
set_device_colour() {
  local device=$1
  local r=$2
  local g=$3
  local b=$4

  if [ "$DRY_RUN" = "true" ]; then
    echo "Would execute: openrgb --server $OPENRGB_HOST:$OPENRGB_PORT --device $device --mode $MODE --color $r,$g,$b" >&2
  else
    openrgb --server "$OPENRGB_HOST:$OPENRGB_PORT" \
      --device "$device" \
      --mode "$MODE" \
      --color "$r,$g,$b" 2>&1 >&2

    if [ $? -eq 0 ]; then
      echo "  ✓ Updated: $device" >&2
      return 0
    else
      echo "  ✗ Failed: $device" >&2
      return 1
    fi
  fi
}

# Get device count
echo "Detecting OpenRGB devices..." >&2
if [ "$DRY_RUN" = "false" ]; then
  DEVICE_LIST=$(openrgb --server "$OPENRGB_HOST:$OPENRGB_PORT" --list-devices 2>&1)
  DEVICE_COUNT=$(echo "$DEVICE_LIST" | grep -c "Device [0-9]:" || echo "0")
  echo "Found $DEVICE_COUNT device(s)" >&2
  echo "" >&2
else
  echo "(Skipped in dry-run mode)" >&2
  echo "" >&2
fi

# Apply colours to devices
echo "Updating devices..." >&2
echo "" >&2

# If device filter is provided, only update matching devices
FILTER_ARRAY=$(echo "$DEVICE_FILTER" | jq -r '.[]' 2>/dev/null)
if [ -n "$FILTER_ARRAY" ]; then
  # Filter by device name/type
  for filter in $FILTER_ARRAY; do
    echo "Searching for devices matching: $filter" >&2

    if [ "$DRY_RUN" = "false" ]; then
      # Find matching device indices
      MATCHING_DEVICES=$(echo "$DEVICE_LIST" | grep -i "$filter" | grep -oP "Device \K[0-9]+" || echo "")

      if [ -z "$MATCHING_DEVICES" ]; then
        echo "  No devices found matching '$filter'" >&2
      else
        for device_id in $MATCHING_DEVICES; do
          set_device_colour "$device_id" "$BG_R" "$BG_G" "$BG_B"
        done
      fi
    else
      echo "  Would search for: $filter" >&2
    fi
  done
else
  # No filter - update all devices
  echo "Updating all devices..." >&2

  if [ "$DRY_RUN" = "false" ] && [ "$DEVICE_COUNT" -gt 0 ]; then
    for ((i=0; i<$DEVICE_COUNT; i++)); do
      set_device_colour "$i" "$BG_R" "$BG_G" "$BG_B"
    done
  else
    echo "  Would update all $DEVICE_COUNT device(s)" >&2
  fi
fi

echo ""  >&2
echo "===============================================" >&2
echo "Status: SUCCESS" >&2
echo "Colour: $BG_HEX → RGB($BG_R, $BG_G, $BG_B)" >&2
echo "Mode: $([ "$DRY_RUN" = "true" ] && echo "DRY-RUN" || echo "APPLIED")" >&2
echo "===============================================" >&2
echo "" >&2
echo "Tip: To see available devices, run:" >&2
echo "  openrgb --list-devices" >&2
echo "" >&2
echo "To enable effects like breathing or rainbow:" >&2
echo "  --plugin-args 'openrgb-peripheral={\"mode\":\"breathing\"}'" >&2

# Protocol 0.2.0: write structured JSON response to stdout
echo '{"success":true,"files_written":[],"message":"OpenRGB peripheral lighting updated"}'

exit 0
