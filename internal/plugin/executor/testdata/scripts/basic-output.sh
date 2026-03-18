#!/bin/sh
# Basic output plugin for testing (legacy protocol < 0.2.0)
if [ "$1" = "--plugin-info" ]; then
  echo '{"name":"test","type":"output","version":"1.0.0","protocol_version":"0.1.0","plugin_protocol":"json-stdio"}'
  exit 0
fi

# Read JSON input from stdin
read -r input

# Legacy behavior: freeform text to stdout
echo "theme configuration content"
