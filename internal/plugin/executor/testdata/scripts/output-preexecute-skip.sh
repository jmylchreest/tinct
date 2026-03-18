#!/bin/sh
# Output plugin with pre-execute hook that skips
if [ "$1" = "--plugin-info" ]; then
  echo '{"name":"test","type":"output","version":"1.0.0","protocol_version":"0.1.0","plugin_protocol":"json-stdio"}'
  exit 0
fi
if [ "$1" = "--pre-execute" ]; then
  echo "test skip reason"
  exit 1
fi
