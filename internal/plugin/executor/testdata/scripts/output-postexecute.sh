#!/bin/sh
# Output plugin with post-execute hook
if [ "$1" = "--plugin-info" ]; then
  echo '{"name":"test","type":"output","version":"1.0.0","protocol_version":"0.1.0","plugin_protocol":"json-stdio"}'
  exit 0
fi
if [ "$1" = "--post-execute" ]; then
  # Read and echo the request for verification
  read -r input
  echo '{"success": true}'
  exit 0
fi
