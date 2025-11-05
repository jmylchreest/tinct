#!/usr/bin/env bash
# Prepare all Go plugins by downloading their dependencies

set -e

echo "Preparing plugin dependencies..."

# Find all plugin directories with go.mod files
PLUGIN_DIRS=$(find contrib/plugins -name "go.mod" -exec dirname {} \;)

for dir in $PLUGIN_DIRS; do
    echo "  Processing: $dir"
    (cd "$dir" && go mod download && go mod verify)
done

echo "✓ All plugin dependencies ready"
