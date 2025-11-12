.PHONY: all build clean install test help plugins main

# Build configuration
BINARY_NAME=tinct
OUT_DIR=./out
PLUGINS_DIR=$(OUT_DIR)/plugins
INPUT_PLUGINS_DIR=$(PLUGINS_DIR)/input
OUTPUT_PLUGINS_DIR=$(PLUGINS_DIR)/output

# Installation directories
INSTALL_BIN_DIR=$(HOME)/.local/bin
INSTALL_PLUGINS_DIR=$(HOME)/.local/share/tinct/plugins

# Go plugins that need building
GO_INPUT_PLUGINS=$(wildcard contrib/plugins/input/*)
GO_OUTPUT_PLUGINS=$(wildcard contrib/plugins/output/*/go.mod)

all: build

build: main plugins

main:
	@echo "Building $(BINARY_NAME)..."
	@go build -ldflags="-s -w" -o $(OUT_DIR)/$(BINARY_NAME) ./cmd/tinct
	@echo "✓ Built: $(OUT_DIR)/$(BINARY_NAME)"

plugins: plugins-input plugins-output plugins-scripts force-install-script

plugins-input:
	@echo "Building input plugins..."
	@mkdir -p $(INPUT_PLUGINS_DIR)
	@for dir in contrib/plugins/input/*/; do \
		if [ -f "$$dir/go.mod" ]; then \
			name=$$(basename $$dir); \
			echo "  Building input plugin: $$name"; \
			(cd $$dir && go build -ldflags="-s -w" -o $(CURDIR)/$(INPUT_PLUGINS_DIR)/tinct-plugin-$$name .) && \
				echo "  ✓ tinct-plugin-$$name" || \
				echo "  ✗ tinct-plugin-$$name (failed)"; \
		fi \
	done

plugins-output:
	@echo "Building output plugins..."
	@mkdir -p $(OUTPUT_PLUGINS_DIR)
	@for dir in contrib/plugins/output/*/; do \
		if [ -f "$$dir/go.mod" ]; then \
			name=$$(basename $$dir); \
			echo "  Building output plugin: $$name"; \
			(cd $$dir && go build -ldflags="-s -w" -o $(CURDIR)/$(OUTPUT_PLUGINS_DIR)/tinct-plugin-$$name .) && \
				echo "  ✓ tinct-plugin-$$name" || \
				echo "  ✗ tinct-plugin-$$name (failed)"; \
		fi \
	done

plugins-scripts:
	@echo "Copying script plugins..."
	@mkdir -p $(OUTPUT_PLUGINS_DIR)
	@for script in contrib/plugins/output/tinct-plugin-*.sh contrib/plugins/output/tinct-plugin-*.py; do \
		if [ -f "$$script" ]; then \
			name=$$(basename $$script); \
			cp "$$script" $(OUTPUT_PLUGINS_DIR)/$$name; \
			chmod +x $(OUTPUT_PLUGINS_DIR)/$$name; \
			echo "  ✓ $$name"; \
		fi \
	done

force-install-script:
	@echo "Generating force-install-plugins.sh..."
	@echo '#!/bin/bash' > $(OUT_DIR)/force-install-plugins.sh
	@echo '# Auto-generated script to force-install all built plugins' >> $(OUT_DIR)/force-install-plugins.sh
	@echo 'set -e' >> $(OUT_DIR)/force-install-plugins.sh
	@echo '' >> $(OUT_DIR)/force-install-plugins.sh
	@echo 'cd "$$(dirname "$$0")/.."' >> $(OUT_DIR)/force-install-plugins.sh
	@echo '' >> $(OUT_DIR)/force-install-plugins.sh
	@echo 'for plugin in ./out/plugins/input/* ./out/plugins/output/*; do' >> $(OUT_DIR)/force-install-plugins.sh
	@echo '  if [ -f "$$plugin" ]; then' >> $(OUT_DIR)/force-install-plugins.sh
	@echo '    echo "Installing $$plugin..."' >> $(OUT_DIR)/force-install-plugins.sh
	@echo '    go run ./cmd/tinct plugins add "$$plugin" --force || true' >> $(OUT_DIR)/force-install-plugins.sh
	@echo '  fi' >> $(OUT_DIR)/force-install-plugins.sh
	@echo 'done' >> $(OUT_DIR)/force-install-plugins.sh
	@chmod +x $(OUT_DIR)/force-install-plugins.sh
	@echo "✓ Generated: $(OUT_DIR)/force-install-plugins.sh"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(OUT_DIR)
	@echo "Cleaned: $(OUT_DIR)"

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_BIN_DIR)..."
	@mkdir -p $(INSTALL_BIN_DIR)
	@mkdir -p $(INSTALL_PLUGINS_DIR)
	@cp $(OUT_DIR)/$(BINARY_NAME) $(INSTALL_BIN_DIR)/
	@echo "✓ Installed: $(INSTALL_BIN_DIR)/$(BINARY_NAME)"
	@echo "Installing plugins to $(INSTALL_PLUGINS_DIR)..."
	@if [ -d $(INPUT_PLUGINS_DIR) ] && [ -n "$$(ls -A $(INPUT_PLUGINS_DIR) 2>/dev/null)" ]; then \
		cp $(INPUT_PLUGINS_DIR)/* $(INSTALL_PLUGINS_DIR)/; \
		echo "✓ Installed input plugins"; \
	fi
	@if [ -d $(OUTPUT_PLUGINS_DIR) ] && [ -n "$$(ls -A $(OUTPUT_PLUGINS_DIR) 2>/dev/null)" ]; then \
		cp $(OUTPUT_PLUGINS_DIR)/* $(INSTALL_PLUGINS_DIR)/; \
		chmod +x $(INSTALL_PLUGINS_DIR)/*; \
		echo "✓ Installed output plugins"; \
	fi

test:
	@echo "Running tests..."
	@go test ./...

test-race:
	@echo "Running tests with race detector..."
	@go test -race ./...

test-cover:
	@echo "Running tests with coverage..."
	@go test -cover ./...

help:
	@echo "Tinct Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build        - Build tinct and all plugins (default)"
	@echo "  main         - Build only tinct binary"
	@echo "  plugins      - Build all plugins"
	@echo "  clean        - Remove build artifacts"
	@echo "  install      - Build and install to ~/.local/bin and ~/.local/share/tinct/plugins"
	@echo "  test         - Run all tests"
	@echo "  test-race    - Run tests with race detector"
	@echo "  test-cover   - Run tests with coverage"
	@echo "  help         - Show this help"
	@echo ""
	@echo "Output structure:"
	@echo "  $(OUT_DIR)/"
	@echo "  ├── tinct"
	@echo "  └── plugins/"
	@echo "      ├── input/  (tinct-plugin-*)"
	@echo "      └── output/ (tinct-plugin-*)"
