.PHONY: build build-all clean test package install

# Variables
VERSION = 1.0.0
BINARY_NAME = dashspace
BUILD_DIR = dist

# Build pour toutes les plateformes
build-all: clean
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 main.go
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 main.go
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go

# Build local
build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) main.go

# Test
test:
	go test ./...

# Package tous les formats
package: build-all
	chmod +x scripts/build-packages.sh
	./scripts/build-packages.sh $(VERSION)

# Installation locale
install: build
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Nettoyage
clean:
	rm -rf $(BUILD_DIR) packages/

# Test des packages
test-packages: package
	chmod +x scripts/test-installation.sh
	./scripts/test-installation.sh

# Release GitHub (nécessite gh CLI)
release: package
	@echo "Creating GitHub release v$(VERSION)..."
	gh release create v$(VERSION) packages/* --title "DashSpace CLI v$(VERSION)" --notes "Release notes for v$(VERSION)"

# Aide
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