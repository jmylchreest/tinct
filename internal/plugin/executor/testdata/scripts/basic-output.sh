#!/bin/sh
# Basic output plugin for testing
if [ "$1" = "--plugin-info" ]; then
  echo '{"name":"test","type":"output","version":"1.0.0","protocol_version":"1.0.0"}'
  exit 0
fi
if [ "$1" = "--detect-protocol" ]; then
  echo "json-stdio"
  exit 0
fi

# Read JSON input from stdin
read -r input

# Return output to stdout - this becomes "output.txt"
echo "theme configuration content"
