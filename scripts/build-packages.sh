#!/bin/bash

set -e

VERSION=${1:-"1.0.0"}
BUILD_DIR="dist"
PACKAGE_DIR="packages"

echo "📦 Building DashSpace CLI v$VERSION packages"

# Clean and create directories
rm -rf $PACKAGE_DIR
mkdir -p $PACKAGE_DIR

# Build binaries (assumes Makefile exists)
echo "🔨 Compiling binaries..."
make build-all

# Function to build Debian package
build_debian_package() {
    echo "🐧 Building Debian package..."

    # Create package structure
    DEB_DIR="$PACKAGE_DIR/dashspace-deb"
    mkdir -p "$DEB_DIR/usr/local/bin"
    mkdir -p "$DEB_DIR/DEBIAN"
    mkdir -p "$DEB_DIR/usr/share/doc/dashspace-cli"
    mkdir -p "$DEB_DIR/usr/share/man/man1"

    # Copy binary
    cp "$BUILD_DIR/dashspace-linux-amd64" "$DEB_DIR/usr/local/bin/dashspace"
    chmod 755 "$DEB_DIR/usr/local/bin/dashspace"

    # Copy packaging files
    cp packaging/debian/control "$DEB_DIR/DEBIAN/"
    cp packaging/debian/postinst "$DEB_DIR/DEBIAN/"
    cp packaging/debian/prerm "$DEB_DIR/DEBIAN/"
    chmod 755 "$DEB_DIR/DEBIAN/postinst"
    chmod 755 "$DEB_DIR/DEBIAN/prerm"

    # Copy and compress man page
    cp docs/man/dashspace.1 "$DEB_DIR/usr/share/man/man1/"
    gzip -9 "$DEB_DIR/usr/share/man/man1/dashspace.1"

    # Create package documentation
    cat > "$DEB_DIR/usr/share/doc/dashspace-cli/README.md" << EOF
# DashSpace CLI v$VERSION

Official CLI for creating and publishing DashSpace modules.

## Documentation
- Website: https://dashspace.space
- Documentation: https://docs.dashspace.dev/cli
- Support: https://discord.gg/dashspace
EOF

    # Build DEB package
    dpkg-deb --build "$DEB_DIR" "$PACKAGE_DIR/dashspace_${VERSION}_amd64.deb"
    echo "✅ Debian package created"
}

# Function to build Homebrew formula
build_homebrew_formula() {
    echo "🍺 Creating Homebrew formula..."

    # Calculate checksums
    AMD64_SHA=$(shasum -a 256 "$BUILD_DIR/dashspace-darwin-amd64" | cut -d' ' -f1)
    ARM64_SHA=$(shasum -a 256 "$BUILD_DIR/dashspace-darwin-arm64" | cut -d' ' -f1)

    # Update formula with real checksums
    sed -e "s/ARM64_SHA256_HERE/$ARM64_SHA/" \
        -e "s/AMD64_SHA256_HERE/$AMD64_SHA/" \
        -e "s/1\.0\.0/$VERSION/g" \
        packaging/homebrew/dashspace.rb > "$PACKAGE_DIR/dashspace.rb"

    echo "✅ Homebrew formula created"
}

# Function to create archives
create_archives() {
    echo "📁 Creating release archives..."

    # Linux archives
    tar -czf "$PACKAGE_DIR/dashspace-$VERSION-linux-amd64.tar.gz" -C "$BUILD_DIR" dashspace-linux-amd64
    tar -czf "$PACKAGE_DIR/dashspace-$VERSION-linux-arm64.tar.gz" -C "$BUILD_DIR" dashspace-linux-arm64

    # macOS archives
    tar -czf "$PACKAGE_DIR/dashspace-$VERSION-darwin-amd64.tar.gz" -C "$BUILD_DIR" dashspace-darwin-amd64
    tar -czf "$PACKAGE_DIR/dashspace-$VERSION-darwin-arm64.tar.gz" -C "$BUILD_DIR" dashspace-darwin-arm64

    # Windows archive
    zip -j "$PACKAGE_DIR/dashspace-$VERSION-windows-amd64.zip" "$BUILD_DIR/dashspace-windows-amd64.exe"

    echo "✅ Archives created"
}

# Function to generate checksums
generate_checksums() {
    echo "🔐 Generating checksums..."
    cd "$PACKAGE_DIR"
    shasum -a 256 * > checksums.txt
    cd ..
    echo "✅ Checksums generated"
}

# Main execution
main() {
    build_debian_package
    build_homebrew_formula
    create_archives
    generate_checksums

    echo ""
    echo "✅ All packages created successfully in $PACKAGE_DIR/"
    echo ""
    echo "📦 Available packages:"
    ls -la "$PACKAGE_DIR"
    echo ""
    echo "🚀 Next steps:"
    echo "  1. Test packages locally"
    echo "  2. Create GitHub release with archives"
    echo "  3. Submit Homebrew formula"
    echo "  4. Publish Debian package to your repository"
}

# Run if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi