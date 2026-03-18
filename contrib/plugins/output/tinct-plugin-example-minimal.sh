#!/bin/bash
# example-minimal.sh - Minimal Tinct output plugin example
#
# This is the simplest possible Tinct plugin. It demonstrates the
# basic interface requirements: handle --plugin-info and process JSON.
#
# Author: Tinct Contributors
# License: MIT

# Handle --plugin-info flag
# This is called by Tinct to discover plugin metadata
if [ "$1" = "--plugin-info" ]; then
  echo "# When called with --plugin-info, plugins must return JSON metadata:" >&2
  echo "# This tells Tinct about the plugin's name, type, version, etc." >&2
  echo "" >&2
  cat <<'EOF'
{
  "name": "example-minimal",
  "type": "output",
  "version": "0.0.1",
  "protocol_version": "0.2.0",
  "plugin_protocol": "json-stdio",
  "description": "Minimal example plugin that prints colour info",
  "enabled": false,
  "author": "Tinct Contributors"
}
EOF
  exit 0
fi

# Read JSON palette from stdin
PALETTE=$(cat)

# Show the input payload received from Tinct (all informational output goes to stderr)
cat >&2 <<EOF
=================================
Tinct Minimal Example Plugin
=================================

PLUGIN INTERFACE OVERVIEW:
---------------------------------
Plugins communicate with Tinct via:

1. --plugin-info flag:
   Tinct calls: plugin.sh --plugin-info
   Plugin returns: JSON metadata (see above)
   Sent to: stdout

2. Colour palette input:
   Tinct sends: JSON palette via stdin
   Plugin receives: Complete palette data
   Plugin processes and responds

=================================

INPUT PAYLOAD RECEIVED:
---------------------------------
This is the JSON payload that Tinct
sends to all output plugins via stdin:

EOF

# Pretty print the JSON if jq is available
if command -v jq &> /dev/null; then
    echo "$PALETTE" | jq '.' >&2
else
    echo "$PALETTE" >&2
fi

cat >&2 <<EOF

=================================

PROCESSING PAYLOAD:
---------------------------------
EOF

# Extract theme type using jq (install: apt-get install jq)
THEME=$(echo "$PALETTE" | jq -r '.theme_type // "unknown"')

# Extract some colours
BG=$(echo "$PALETTE" | jq -r '.colours.background.hex // "N/A"')
FG=$(echo "$PALETTE" | jq -r '.colours.foreground.hex // "N/A"')
ACCENT=$(echo "$PALETTE" | jq -r '.colours.accent1.hex // "N/A"')

# Count total colours
COLOR_COUNT=$(echo "$PALETTE" | jq '.all_colours | length')

# Extract plugin args if provided
PLUGIN_ARGS=$(echo "$PALETTE" | jq -r '.plugin_args // empty')
HAS_ARGS=""
if [ -n "$PLUGIN_ARGS" ]; then
  HAS_ARGS="yes"
fi

# Extract dry-run flag
DRY_RUN=$(echo "$PALETTE" | jq -r '.dry_run // false')

# Output results to stderr
cat >&2 <<EOF

Extracted Information:
  Theme Type:    $THEME
  Colour Count:  $COLOR_COUNT
  Background:    $BG
  Foreground:    $FG
  Accent 1:      $ACCENT
  Plugin Args:   ${HAS_ARGS:-no custom args provided}
  Dry Run:       $DRY_RUN

=================================

CUSTOM PLUGIN ARGUMENTS:
---------------------------------
EOF

if [ -n "$PLUGIN_ARGS" ]; then
  cat >&2 <<EOF
Plugins can receive custom arguments via
the --plugin-args flag. For example:

  tinct generate -i file -p palette.json \\
    -o example-minimal \\
    --plugin-args 'example-minimal={"format":"json","verbose":true}'

The plugin_args field in the JSON payload:
EOF
  echo "$PLUGIN_ARGS" | jq '.' >&2
else
  cat >&2 <<EOF
No custom arguments were provided.
To pass custom arguments to this plugin:

  tinct generate -i file -p palette.json \\
    -o example-minimal \\
    --plugin-args 'example-minimal={"key":"value"}'

This allows plugins to accept configuration
specific to their functionality.
EOF
fi

cat >&2 <<EOF

=================================

DRY-RUN MODE:
---------------------------------
EOF

if [ "$DRY_RUN" = "true" ]; then
  cat >&2 <<EOF
Dry-run mode is ENABLED. The plugin should:
  - NOT write any files to disk
  - NOT modify system settings
  - NOT send notifications
  - Show what WOULD be done instead

Example dry-run output:
  Would write: ~/.config/myapp/theme.conf
  Would execute: notify-send "Theme updated"
  Would modify: 3 configuration files

To run in normal mode, omit the --dry-run flag.
EOF
else
  cat >&2 <<EOF
Dry-run mode is DISABLED. The plugin can:
  - Write files to disk
  - Modify system settings
  - Send notifications
  - Execute commands

In this example, we're not actually writing files,
but a real plugin would do so now.
EOF
fi

cat >&2 <<EOF

=================================

Status: SUCCESS
Processed: $COLOR_COUNT colours
Theme: $THEME
Mode: $([ "$DRY_RUN" = "true" ] && echo "DRY-RUN" || echo "NORMAL")

=================================
EOF

# Protocol 0.2.0: write structured JSON response to stdout
cat <<EOF
{"success":true,"files_written":[],"message":"Processed $COLOR_COUNT colours from $THEME theme"}
EOF

exit 0
