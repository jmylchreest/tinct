.PHONY: build install clean test help

BINARY_NAME=wob-tinct
BUILD_DIR=.
INSTALL_DIR=$(HOME)/.config/tinct/plugins

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) -ldflags="-s -w"
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

install: build
	@echo "Installing to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	@chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed: $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo ""
	@echo "To use as a Tinct plugin:"
	@echo "  tinct generate -i image -p wallpaper.jpg -o wob"

clean:
	@echo "Cleaning..."
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@echo "Cleaned"

test:
	@echo "Testing plugin mode..."
	@echo '{"background":"#1e1e2e","foreground":"#cdd6f4","accent1":"#89b4fa","success":"#a6e3a1","warning":"#f9e2af","danger":"#f38ba8"}' | \
		./$(BINARY_NAME) plugin
	@echo ""
	@echo "Testing wrapper mode..."
	@./$(BINARY_NAME) version
	@./$(BINARY_NAME) status || true

help:
	@echo "wob-tinct Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build    - Build the wob-tinct binary"
	@echo "  install  - Build and install to ~/.config/tinct/plugins/"
	@echo "  clean    - Remove built binary"
	@echo "  test     - Test plugin and wrapper modes"
	@echo "  help     - Show this help"
	@echo ""
	@echo "Usage:"
	@echo "  make build && ./wob-tinct help"
	@echo "  make install"
