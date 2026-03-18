#!/bin/sh
# Basic input plugin for testing
if [ "$1" = "--plugin-info" ]; then
  echo '{"name":"test","type":"input","version":"1.0.0","protocol_version":"0.1.0","plugin_protocol":"json-stdio"}'
  exit 0
fi
