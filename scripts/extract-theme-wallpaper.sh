#!/bin/bash
# extract-theme-wallpaper.sh - Extract wallpaper from tinct markdown theme files
#
# Usage: extract-theme-wallpaper.sh <theme-file.md> [output-directory]
#
# This script extracts embedded base64 wallpaper images from tinct markdown
# theme files. If no output directory is specified, the wallpaper is extracted
# to the same directory as the theme file.

set -e

if [ $# -lt 1 ]; then
    echo "Usage: $0 <theme-file.md> [output-directory]"
    echo ""
    echo "Extract wallpaper from a tinct markdown theme file."
    echo ""
    echo "Arguments:"
    echo "  theme-file.md     Path to the markdown theme file"
    echo "  output-directory  Directory to save extracted wallpaper (default: same as theme file)"
    exit 1
fi

THEME_FILE="$1"
OUTPUT_DIR="${2:-$(dirname "$THEME_FILE")}"

if [ ! -f "$THEME_FILE" ]; then
    echo "Error: Theme file not found: $THEME_FILE"
    exit 1
fi

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Extract the YAML front matter
FRONT_MATTER=$(sed -n '/^---$/,/^---$/p' "$THEME_FILE" | tail -n +2 | head -n -1)

# Check if wallpaper is embedded
IS_EMBEDDED=$(echo "$FRONT_MATTER" | grep -E '^\s*embedded:\s*true' || true)
if [ -z "$IS_EMBEDDED" ]; then
    echo "No embedded wallpaper found in theme file."

    # Check for external path
    EXTERNAL_PATH=$(echo "$FRONT_MATTER" | grep -E '^\s*path:' | sed 's/.*path:\s*//' | tr -d '"'"'" || true)
    if [ -n "$EXTERNAL_PATH" ]; then
        echo "Wallpaper is external reference: $EXTERNAL_PATH"
    fi
    exit 0
fi

# Extract format
FORMAT=$(echo "$FRONT_MATTER" | grep -E '^\s*format:' | sed 's/.*format:\s*//' | tr -d '"'"'" || echo "png")

# Extract base64 data - this is the tricky part since the data can span many lines
# We use awk to extract the data field value
DATA=$(echo "$FRONT_MATTER" | awk '
    /^\s*data:/ {
        # Remove the "data:" prefix and any leading quotes
        gsub(/^\s*data:\s*/, "")
        gsub(/^["'"'"']/, "")
        gsub(/["'"'"']$/, "")
        data = $0
        # If data is on same line, print it
        if (length(data) > 0) {
            print data
        }
        next
    }
' | tr -d '\n' | tr -d ' ')

if [ -z "$DATA" ]; then
    echo "Error: Could not extract wallpaper data from theme file."
    echo "The wallpaper data field may be in an unsupported format."
    exit 1
fi

# Generate output filename from theme name
THEME_NAME=$(basename "$THEME_FILE" .md)
OUTPUT_FILE="$OUTPUT_DIR/${THEME_NAME}-wallpaper.${FORMAT}"

# Decode and save
echo "$DATA" | base64 -d > "$OUTPUT_FILE"

echo "Extracted wallpaper to: $OUTPUT_FILE"

# Show file info
if command -v file &> /dev/null; then
    file "$OUTPUT_FILE"
fi

if command -v du &> /dev/null; then
    SIZE=$(du -h "$OUTPUT_FILE" | cut -f1)
    echo "Size: $SIZE"
fi
