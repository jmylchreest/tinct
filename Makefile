.PHONY: bump-patch bump-minor bump-major bump help get-version get-latest-release

help:
	@echo "Usage:"
	@echo "  make bump               - Bump patch version (e.g., 1.0.0 -> 1.0.1)"
	@echo "  make bump-patch         - Bump patch version (e.g., 1.0.0 -> 1.0.1)"
	@echo "  make bump-minor         - Bump minor version (e.g., 0.0.22 -> 0.1.0)"
	@echo "  make bump-major         - Bump major version (e.g., 0.1.0 -> 1.0.0)"
	@echo "  make get-version        - Get current version (like goreleaser)"
	@echo "  make get-latest-release - Get latest release tag"

# Get the latest tag
CURRENT_VERSION := $(shell git tag --sort=-v:refname | head -1)

get-version:
	@if [ -z "$(CURRENT_VERSION)" ]; then \
		git describe --always --dirty 2>/dev/null || echo "unknown"; \
	else \
		git describe --tags --always --dirty 2>/dev/null || echo $(CURRENT_VERSION); \
	fi

get-latest-release:
	@if [ -z "$(CURRENT_VERSION)" ]; then \
		echo "No release tags found"; \
	else \
		echo $(CURRENT_VERSION); \
	fi

# Default to patch if no target specified
bump: bump-patch

bump-patch:
	@if [ -z "$(CURRENT_VERSION)" ]; then \
		echo "No existing tags found, creating v0.0.1"; \
		git tag -a v0.0.1 -m "Release v0.0.1"; \
		echo "Created tag: v0.0.1"; \
	else \
		CURRENT_COMMIT=$$(git rev-parse HEAD); \
		TAG_COMMIT=$$(git rev-parse $(CURRENT_VERSION)^{commit} 2>/dev/null || echo ""); \
		if [ "$$CURRENT_COMMIT" = "$$TAG_COMMIT" ] && [ "$(FORCE)" != "1" ]; then \
			echo "Error: Current commit is already tagged as $(CURRENT_VERSION)"; \
			echo "No changes since last release. Use 'make bump-patch FORCE=1' to force."; \
			exit 1; \
		fi; \
		NEW_VERSION=$$(echo $(CURRENT_VERSION) | sed 's/^v//' | awk -F. '{print "v"$$1"."$$2"."$$3+1}'); \
		echo "Bumping $(CURRENT_VERSION) -> $$NEW_VERSION"; \
		git tag -a $$NEW_VERSION -m "Release $$NEW_VERSION"; \
		echo "Created tag: $$NEW_VERSION"; \
	fi

bump-minor:
	@if [ -z "$(CURRENT_VERSION)" ]; then \
		echo "No existing tags found, creating v0.1.0"; \
		git tag -a v0.1.0 -m "Release v0.1.0"; \
		echo "Created tag: v0.1.0"; \
	else \
		CURRENT_COMMIT=$$(git rev-parse HEAD); \
		TAG_COMMIT=$$(git rev-parse $(CURRENT_VERSION)^{commit} 2>/dev/null || echo ""); \
		if [ "$$CURRENT_COMMIT" = "$$TAG_COMMIT" ] && [ "$(FORCE)" != "1" ]; then \
			echo "Error: Current commit is already tagged as $(CURRENT_VERSION)"; \
			echo "No changes since last release. Use 'make bump-minor FORCE=1' to force."; \
			exit 1; \
		fi; \
		NEW_VERSION=$$(echo $(CURRENT_VERSION) | sed 's/^v//' | awk -F. '{print "v"$$1"."$$2+1".0"}'); \
		echo "Bumping $(CURRENT_VERSION) -> $$NEW_VERSION"; \
		git tag -a $$NEW_VERSION -m "Release $$NEW_VERSION"; \
		echo "Created tag: $$NEW_VERSION"; \
	fi

bump-major:
	@if [ -z "$(CURRENT_VERSION)" ]; then \
		echo "No existing tags found, creating v1.0.0"; \
		git tag -a v1.0.0 -m "Release v1.0.0"; \
		echo "Created tag: v1.0.0"; \
	else \
		CURRENT_COMMIT=$$(git rev-parse HEAD); \
		TAG_COMMIT=$$(git rev-parse $(CURRENT_VERSION)^{commit} 2>/dev/null || echo ""); \
		if [ "$$CURRENT_COMMIT" = "$$TAG_COMMIT" ] && [ "$(FORCE)" != "1" ]; then \
			echo "Error: Current commit is already tagged as $(CURRENT_VERSION)"; \
			echo "No changes since last release. Use 'make bump-major FORCE=1' to force."; \
			exit 1; \
		fi; \
		NEW_VERSION=$$(echo $(CURRENT_VERSION) | sed 's/^v//' | awk -F. '{print "v"$$1+1".0.0"}'); \
		echo "Bumping $(CURRENT_VERSION) -> $$NEW_VERSION"; \
		git tag -a $$NEW_VERSION -m "Release $$NEW_VERSION"; \
		echo "Created tag: $$NEW_VERSION"; \
	fi
