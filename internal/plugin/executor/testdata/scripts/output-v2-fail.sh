#!/bin/sh
# Output plugin for testing protocol 0.2.0 failure response
if [ "$1" = "--plugin-info" ]; then
  echo '{"name":"test-v2-fail","type":"output","version":"1.0.0","protocol_version":"0.2.0","plugin_protocol":"json-stdio"}'
  exit 0
fi

# Read JSON input from stdin
read -r input

# Return structured failure response (protocol 0.2.0)
echo '{"success": false, "files_written": [], "message": "configuration validation failed"}'
