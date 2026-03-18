#!/bin/sh
# Input plugin that returns color data
if [ "$1" = "--plugin-info" ]; then
  echo '{"name":"test","type":"input","version":"1.0.0","protocol_version":"0.1.0","plugin_protocol":"json-stdio"}'
  exit 0
fi

# Read JSON input from stdin
read -r input

# Return colors
cat <<'EOF'
{
  "colors": [
    {"r": 255, "g": 0, "b": 0, "a": 255},
    {"r": 0, "g": 255, "b": 0, "a": 255},
    {"r": 0, "g": 0, "b": 255, "a": 255}
  ],
  "wallpaper_path": "/tmp/test.jpg"
}
EOF
