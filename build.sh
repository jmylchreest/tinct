#!/usr/bin/env bash
# Build script for tinct and all plugins using GoReleaser

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get absolute path to script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo -e "${GREEN}Building tinct and all plugins using GoReleaser...${NC}\n"

# Run GoReleaser in snapshot mode for local builds
echo -e "${YELLOW}Running GoReleaser snapshot build...${NC}"
if ! goreleaser build --snapshot --clean --single-target; then
    echo -e "${RED}✗ GoReleaser build failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ GoReleaser build complete${NC}\n"

# Detect the current platform directory name
PLATFORM_ARCH=$(uname -m)
case "$PLATFORM_ARCH" in
    x86_64)
        GOARCH="amd64"
        ;;
    aarch64|arm64)
        GOARCH="arm64"
        ;;
    armv7l)
        GOARCH="arm"
        ;;
    *)
        echo -e "${RED}Unsupported architecture: $PLATFORM_ARCH${NC}"
        exit 1
        ;;
esac

GOOS=$(uname -s | tr '[:upper:]' '[:lower:]')
DIST_PLATFORM_SUFFIX="${GOOS}_${GOARCH}"

# Add v1 suffix for amd64
if [ "$GOARCH" = "amd64" ]; then
    DIST_PLATFORM_SUFFIX="${DIST_PLATFORM_SUFFIX}_v1"
fi

TINCT_DIST_DIR="$SCRIPT_DIR/dist/tinct_${DIST_PLATFORM_SUFFIX}"
PLUGINS_DIST_DIR="$TINCT_DIST_DIR/plugins"
INPUT_PLUGINS_DIST_DIR="$PLUGINS_DIST_DIR/input"
OUTPUT_PLUGINS_DIST_DIR="$PLUGINS_DIST_DIR/output"

# Create plugin directories in dist
mkdir -p "$INPUT_PLUGINS_DIST_DIR"
mkdir -p "$OUTPUT_PLUGINS_DIST_DIR"

echo -e "${YELLOW}Organizing plugin binaries into dist structure...${NC}"

# Copy input plugins from their build directories
for plugin_dist_dir in "$SCRIPT_DIR"/dist/tinct-plugin-*_${DIST_PLATFORM_SUFFIX}/; do
    if [ -d "$plugin_dist_dir" ]; then
        # Extract plugin name from directory
        dir_name=$(basename "$plugin_dist_dir")
        plugin_name=$(echo "$dir_name" | sed "s/tinct-plugin-//;s/_${DIST_PLATFORM_SUFFIX}//")

        # Find the binary in the directory
        binary_path=$(find "$plugin_dist_dir" -maxdepth 1 -type f -executable | head -1)

        if [ -n "$binary_path" ]; then
            binary_name=$(basename "$binary_path")

            # Determine if it's an input or output plugin
            if [ -d "contrib/plugins/input/$plugin_name" ]; then
                cp "$binary_path" "$INPUT_PLUGINS_DIST_DIR/$binary_name"
                echo -e "${GREEN}✓ $binary_name (input, Go)${NC}"
            elif [ -d "contrib/plugins/output/$plugin_name" ]; then
                cp "$binary_path" "$OUTPUT_PLUGINS_DIST_DIR/$binary_name"
                echo -e "${GREEN}✓ $binary_name (output, Go)${NC}"
            fi
        fi
    fi
done

# Copy script-based plugins
echo -e "\n${YELLOW}Copying script-based plugins...${NC}"

# Output plugins (scripts)
for script in contrib/plugins/output/*.sh contrib/plugins/output/*.py contrib/plugins/output/*.rb; do
    if [ -f "$script" ]; then
        # Extract plugin name (remove tinct-plugin- prefix if present)
        filename=$(basename "$script")
        plugin_name="${filename}"

        cp "$script" "$OUTPUT_PLUGINS_DIST_DIR/$plugin_name"
        chmod +x "$OUTPUT_PLUGINS_DIST_DIR/$plugin_name"

        # Determine language for display
        case "$script" in
            *.sh) lang="shell" ;;
            *.py) lang="Python" ;;
            *.rb) lang="Ruby" ;;
            *) lang="script" ;;
        esac

        echo -e "${GREEN}✓ $plugin_name (output, $lang)${NC}"
    fi
done

# Input plugins (scripts) - if any exist
for script in contrib/plugins/input/*.sh contrib/plugins/input/*.py contrib/plugins/input/*.rb; do
    if [ -f "$script" ]; then
        filename=$(basename "$script")
        plugin_name="${filename}"

        cp "$script" "$INPUT_PLUGINS_DIST_DIR/$plugin_name"
        chmod +x "$INPUT_PLUGINS_DIST_DIR/$plugin_name"

        case "$script" in
            *.sh) lang="shell" ;;
            *.py) lang="Python" ;;
            *.rb) lang="Ruby" ;;
            *) lang="script" ;;
        esac

        echo -e "${GREEN}✓ $plugin_name (input, $lang)${NC}"
    fi
done

# Summary
echo -e "\n${GREEN}Build complete!${NC}"
echo -e "Output structure:"
echo -e "  $TINCT_DIST_DIR/"
echo -e "  ├── tinct"
echo -e "  └── plugins/"
echo -e "      ├── input/"
if [ -n "$(ls -A "$INPUT_PLUGINS_DIST_DIR" 2>/dev/null)" ]; then
    ls -1 "$INPUT_PLUGINS_DIST_DIR" 2>/dev/null | while read -r plugin; do
        echo -e "      │   └── $plugin"
    done
else
    echo -e "      │   (none)"
fi
echo -e "      └── output/"
if [ -n "$(ls -A "$OUTPUT_PLUGINS_DIST_DIR" 2>/dev/null)" ]; then
    ls -1 "$OUTPUT_PLUGINS_DIST_DIR" 2>/dev/null | while read -r plugin; do
        echo -e "          └── $plugin"
    done
else
    echo -e "          (none)"
fi

# Generate force-install-plugins.sh script in dist root
cat > "$SCRIPT_DIR/dist/force-install-plugins.sh" <<'SCRIPT_EOF'
#!/bin/bash
# Auto-generated script to force-install all built plugins
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Detect platform
PLATFORM_ARCH=$(uname -m)
case "$PLATFORM_ARCH" in
    x86_64)
        GOARCH="amd64"
        ;;
    aarch64|arm64)
        GOARCH="arm64"
        ;;
    armv7l)
        GOARCH="arm"
        ;;
    *)
        echo "Unsupported architecture: $PLATFORM_ARCH"
        exit 1
        ;;
esac

GOOS=$(uname -s | tr '[:upper:]' '[:lower:]')
DIST_PLATFORM_SUFFIX="${GOOS}_${GOARCH}"

# Add v1 suffix for amd64
if [ "$GOARCH" = "amd64" ]; then
    DIST_PLATFORM_SUFFIX="${DIST_PLATFORM_SUFFIX}_v1"
fi

TINCT_DIR="$SCRIPT_DIR/tinct_${DIST_PLATFORM_SUFFIX}"
TINCT_BIN="$TINCT_DIR/tinct"
PLUGINS_DIR="$TINCT_DIR/plugins"

if [ ! -f "$TINCT_BIN" ]; then
    echo "Error: tinct binary not found at $TINCT_BIN"
    exit 1
fi

if [ ! -d "$PLUGINS_DIR" ]; then
    echo "Error: plugins directory not found at $PLUGINS_DIR"
    exit 1
fi

echo "Installing plugins from $PLUGINS_DIR..."
echo ""

for plugin in "$PLUGINS_DIR"/input/* "$PLUGINS_DIR"/output/*; do
  if [ -f "$plugin" ]; then
    echo "Installing $(basename "$plugin")..."
    "$TINCT_BIN" plugins add "$plugin" --force || true
  fi
done

echo ""
echo "Plugin installation complete!"
SCRIPT_EOF

chmod +x "$SCRIPT_DIR/dist/force-install-plugins.sh"
echo -e "${GREEN}✓ Generated force-install-plugins.sh in dist/${NC}"

echo -e "\n${YELLOW}To install locally:${NC}"
echo -e "  cp $TINCT_DIST_DIR/tinct ~/.local/bin/"
echo -e "  mkdir -p ~/.local/share/tinct/plugins/{input,output}"
echo -e "  cp -r $PLUGINS_DIST_DIR/* ~/.local/share/tinct/plugins/"
echo -e "\n${YELLOW}Or use the force-install script:${NC}"
echo -e "  ./dist/force-install-plugins.sh"
