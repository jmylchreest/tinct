#!/bin/sh
# Basic input plugin for testing
if [ "$1" = "--plugin-info" ]; then
  echo '{"name":"test","type":"input","version":"1.0.0","protocol_version":"1.0.0"}'
  exit 0
fi
if [ "$1" = "--detect-protocol" ]; then
  echo "json-stdio"
  exit 0
fi
