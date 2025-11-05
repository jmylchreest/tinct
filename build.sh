#!/usr/bin/env bash
# Build script for tinct and all plugins

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get absolute path to script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Output directories (absolute paths)
OUT_DIR="$SCRIPT_DIR/out"
PLUGINS_DIR="$OUT_DIR/plugins"
INPUT_PLUGINS_DIR="$PLUGINS_DIR/input"
OUTPUT_PLUGINS_DIR="$PLUGINS_DIR/output"

# Create output directories
mkdir -p "$OUT_DIR"
mkdir -p "$INPUT_PLUGINS_DIR"
mkdir -p "$OUTPUT_PLUGINS_DIR"

echo -e "${GREEN}Building tinct and all plugins...${NC}\n"

# Build main tinct binary
echo -e "${YELLOW}Building tinct...${NC}"
go build -o "$OUT_DIR/tinct" ./cmd/tinct
echo -e "${GREEN}✓ Built tinct${NC}\n"

# Build input plugins
echo -e "${YELLOW}Building input plugins...${NC}"
for plugin_dir in contrib/plugins/input/*/; do
    if [ -d "$plugin_dir" ]; then
        plugin_name=$(basename "$plugin_dir")
        echo "  Building input plugin: $plugin_name"

        # cd into plugin directory to build
        if (cd "$plugin_dir" && go build -o "$INPUT_PLUGINS_DIR/$plugin_name" . 2>&1) >/dev/null 2>&1; then
            echo -e "  ${GREEN}✓ $plugin_name${NC}"
        else
            echo -e "  ${RED}✗ $plugin_name (build failed)${NC}"
        fi
    fi
done
echo ""

# Build output plugins
echo -e "${YELLOW}Building output plugins...${NC}"
for plugin_dir in contrib/plugins/output/*/; do
    if [ -d "$plugin_dir" ]; then
        plugin_name=$(basename "$plugin_dir")
        echo "  Building output plugin: $plugin_name"

        # cd into plugin directory to build
        if (cd "$plugin_dir" && go build -o "$OUTPUT_PLUGINS_DIR/$plugin_name" . 2>&1) >/dev/null 2>&1; then
            echo -e "  ${GREEN}✓ $plugin_name${NC}"
        else
            echo -e "  ${RED}✗ $plugin_name (build failed)${NC}"
        fi
    fi
done
echo ""

# Summary
echo -e "${GREEN}Build complete!${NC}"
echo -e "Output structure:"
echo -e "  $OUT_DIR/"
echo -e "  ├── tinct"
echo -e "  └── plugins/"
echo -e "      ├── input/"
echo -e "      │   $(ls -1 "$INPUT_PLUGINS_DIR" 2>/dev/null | sed 's/^/└── /' | tr '\n' ' ' || echo '(none)')"
echo -e "      └── output/"
echo -e "          $(ls -1 "$OUTPUT_PLUGINS_DIR" 2>/dev/null | sed 's/^/└── /' | tr '\n' ' ' || echo '(none)')"
