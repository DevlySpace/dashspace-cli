.PHONY: build build-all clean test package install update-homebrew update-tap complete-release

VERSION ?= 1.0.0
BINARY_NAME = dashspace
BUILD_DIR = dist
SCRIPTS_DIR = scripts

build-all: clean copy-templates
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 main.go
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 main.go
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go

build: copy-templates
	mkdir -p $(BUILD_DIR)
	go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) main.go

copy-templates:
	mkdir -p $(BUILD_DIR)/templates
	cp -r internal/templates/* $(BUILD_DIR)/templates/ || echo "No templates found"


test:
	go test ./...

setup-scripts:
	mkdir -p $(SCRIPTS_DIR)
	@if [ ! -f $(SCRIPTS_DIR)/build-packages.sh ]; then \
		echo "Creating build-packages.sh script..."; \
		cp build-packages.sh $(SCRIPTS_DIR)/build-packages.sh 2>/dev/null || echo "build-packages.sh not found in current directory"; \
	fi
	@if [ ! -f $(SCRIPTS_DIR)/test-installation.sh ]; then \
		echo "Creating test-installation.sh script..."; \
		cp test-installation.sh $(SCRIPTS_DIR)/test-installation.sh 2>/dev/null || echo "test-installation.sh not found in current directory"; \
	fi
	@if [ ! -f $(SCRIPTS_DIR)/update-homebrew.sh ]; then \
		echo "Creating update-homebrew.sh script..."; \
		cp update-homebrew.sh $(SCRIPTS_DIR)/update-homebrew.sh 2>/dev/null || echo "update-homebrew.sh not found in current directory"; \
	fi
	@if [ ! -f $(SCRIPTS_DIR)/auto-update-tap.sh ]; then \
		echo "Creating auto-update-tap.sh script..."; \
		cp auto-update-tap.sh $(SCRIPTS_DIR)/auto-update-tap.sh 2>/dev/null || echo "auto-update-tap.sh not found in current directory"; \
	fi
	@if [ ! -f $(SCRIPTS_DIR)/release-workflow.sh ]; then \
		echo "Creating release-workflow.sh script..."; \
		cp release-workflow.sh $(SCRIPTS_DIR)/release-workflow.sh 2>/dev/null || echo "release-workflow.sh not found in current directory"; \
	fi
	chmod +x $(SCRIPTS_DIR)/*.sh

package: build-all setup-scripts
	chmod +x $(SCRIPTS_DIR)/build-packages.sh
	$(SCRIPTS_DIR)/build-packages.sh $(VERSION)

install: build
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	sudo mkdir -p /usr/local/share/dashspace/templates
	sudo cp -r internal/templates/* /usr/local/share/dashspace/templates/

clean:
	rm -rf $(BUILD_DIR) packages/

test-packages: package
	chmod +x $(SCRIPTS_DIR)/test-installation.sh
	$(SCRIPTS_DIR)/test-installation.sh $(VERSION)

release: package
	@echo "Creating GitHub release $(VERSION)..."
	gh release create $(VERSION) packages/* --title "DashSpace CLI $(VERSION)" --notes "Release notes for $(VERSION)"

update-homebrew: build-all setup-scripts
	chmod +x $(SCRIPTS_DIR)/update-homebrew.sh
	$(SCRIPTS_DIR)/update-homebrew.sh $(VERSION)

update-tap: build-all setup-scripts
	chmod +x $(SCRIPTS_DIR)/auto-update-tap.sh
	$(SCRIPTS_DIR)/auto-update-tap.sh $(VERSION)

complete-release: setup-scripts
	@read -p "Enter version (e.g., 1.0.0): " version; \
	chmod +x $(SCRIPTS_DIR)/release-workflow.sh; \
	$(SCRIPTS_DIR)/release-workflow.sh $$version

help:
	@echo "Available targets:"
	@echo "  build       - Build binary for current platform"
	@echo "  build-all   - Build binaries for all platforms"
	@echo "  test        - Run Go tests"
	@echo "  package     - Create all distribution packages"
	@echo "  install     - Install locally"
	@echo "  clean       - Clean build artifacts"
	@echo "  test-packages - Test all packages"
	@echo "  release     - Create GitHub release"
	@echo "  update-homebrew - Update local Homebrew formula"
	@echo "  update-tap  - Auto-update Homebrew tap repository"
	@echo "  complete-release - Full release workflow (interactive)"