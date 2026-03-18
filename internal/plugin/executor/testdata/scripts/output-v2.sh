#!/bin/sh
# Output plugin for testing protocol 0.2.0 structured response
if [ "$1" = "--plugin-info" ]; then
  echo '{"name":"test-v2","type":"output","version":"1.0.0","protocol_version":"0.2.0","plugin_protocol":"json-stdio"}'
  exit 0
fi

# Read JSON input from stdin
read -r input

# Create a temp file to simulate writing
TMPFILE=$(mktemp /tmp/tinct-test-XXXXXX.conf)
echo "theme content" > "$TMPFILE"

# Return structured JSON response (protocol 0.2.0)
cat <<EOF
{"success": true, "files_written": ["$TMPFILE"], "message": "Generated 1 file"}
EOF
