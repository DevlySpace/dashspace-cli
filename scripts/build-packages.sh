#!/bin/bash

set -e

VERSION=${1:-"1.0.0"}
BUILD_DIR="dist"
PACKAGE_DIR="packages"

echo "📦 Building DashSpace CLI v$VERSION packages"

rm -rf $PACKAGE_DIR
mkdir -p $PACKAGE_DIR

echo "🔨 Compiling binaries..."
make build-all

detect_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux";;
        Darwin*)    echo "darwin";;
        MINGW*)     echo "windows";;
        *)          echo "unknown";;
    esac
}

build_debian_package() {
    echo "🐧 Building Debian package..."

    DEB_DIR="$PACKAGE_DIR/dashspace-deb"
    mkdir -p "$DEB_DIR/usr/local/bin"
    mkdir -p "$DEB_DIR/DEBIAN"
    mkdir -p "$DEB_DIR/usr/share/doc/dashspace-cli"
    mkdir -p "$DEB_DIR/usr/share/man/man1"

    cp "$BUILD_DIR/dashspace-linux-amd64" "$DEB_DIR/usr/local/bin/dashspace"
    chmod 755 "$DEB_DIR/usr/local/bin/dashspace"

    if [ -d "packaging/debian" ]; then
        cp packaging/debian/control "$DEB_DIR/DEBIAN/" 2>/dev/null || true
        cp packaging/debian/postinst "$DEB_DIR/DEBIAN/" 2>/dev/null || true
        cp packaging/debian/prerm "$DEB_DIR/DEBIAN/" 2>/dev/null || true
        chmod 755 "$DEB_DIR/DEBIAN/postinst" 2>/dev/null || true
        chmod 755 "$DEB_DIR/DEBIAN/prerm" 2>/dev/null || true
    fi

    if [ -f "docs/man/dashspace.1" ]; then
        cp docs/man/dashspace.1 "$DEB_DIR/usr/share/man/man1/"
        gzip -9 "$DEB_DIR/usr/share/man/man1/dashspace.1"
    fi

    cat > "$DEB_DIR/usr/share/doc/dashspace-cli/README.md" << EOF
# DashSpace CLI v$VERSION

Official CLI for creating and publishing DashSpace modules.

## Documentation
- Website: https://dashspace.space
- Documentation: https://docs.dashspace.dev/cli
- Support: https://discord.gg/dashspace
EOF

    cat > "$DEB_DIR/DEBIAN/control" << EOF
Package: dashspace-cli
Version: $VERSION
Section: devel
Priority: optional
Architecture: amd64
Depends: libc6
Maintainer: DashSpace Team <support@dashspace.space>
Description: Official CLI for creating and publishing DashSpace modules
 DashSpace CLI is a command-line tool that allows developers to create,
 build, and publish DashSpace modules easily.
EOF

    cat > "$DEB_DIR/DEBIAN/postinst" << 'EOF'
#!/bin/bash
set -e
if [ "$1" = "configure" ]; then
    echo "DashSpace CLI installed successfully!"
    echo "Run 'dashspace --help' to get started."
fi
EOF

    cat > "$DEB_DIR/DEBIAN/prerm" << 'EOF'
#!/bin/bash
set -e
if [ "$1" = "remove" ]; then
    echo "Removing DashSpace CLI..."
fi
EOF

    chmod 755 "$DEB_DIR/DEBIAN/postinst"
    chmod 755 "$DEB_DIR/DEBIAN/prerm"

    OS=$(detect_os)
    if [ "$OS" = "darwin" ]; then
        if command -v docker &> /dev/null; then
            echo "🐳 Using Docker to build Debian package on macOS..."

            TEMP_DIR="$(mktemp -d)"
            cp -r "$DEB_DIR"/* "$TEMP_DIR/"

            docker run --rm -v "$TEMP_DIR":/package ubuntu:20.04 sh -c "
                apt-get update -qq && apt-get install -y dpkg-dev >/dev/null 2>&1
                cd /package
                dpkg-deb --build . dashspace_${VERSION}_amd64.deb
            "

            if [ -f "$TEMP_DIR/dashspace_${VERSION}_amd64.deb" ]; then
                mv "$TEMP_DIR/dashspace_${VERSION}_amd64.deb" "$PACKAGE_DIR/"
                echo "✅ Debian package created"
            else
                echo "❌ Failed to create Debian package"
                return 1
            fi

            rm -rf "$TEMP_DIR"
        else
            echo "⚠️  Docker not available. Creating tar.gz instead of .deb package..."
            tar -czf "$PACKAGE_DIR/dashspace_${VERSION}_amd64.tar.gz" -C "$DEB_DIR" .
        fi
    else
        dpkg-deb --build "$DEB_DIR" "$PACKAGE_DIR/dashspace_${VERSION}_amd64.deb"
        echo "✅ Debian package created"
    fi
}

build_homebrew_formula() {
    echo "🍺 Creating Homebrew formula..."

    if [ -f "$BUILD_DIR/dashspace-darwin-amd64" ] && [ -f "$BUILD_DIR/dashspace-darwin-arm64" ]; then
        AMD64_SHA=$(shasum -a 256 "$BUILD_DIR/dashspace-darwin-amd64" | cut -d' ' -f1)
        ARM64_SHA=$(shasum -a 256 "$BUILD_DIR/dashspace-darwin-arm64" | cut -d' ' -f1)
    else
        echo "⚠️  macOS binaries not found, using placeholder checksums"
        AMD64_SHA="AMD64_SHA256_HERE"
        ARM64_SHA="ARM64_SHA256_HERE"
    fi

    cat > "$PACKAGE_DIR/dashspace.rb" << EOF
class Dashspace < Formula
  desc "Official CLI for creating and publishing DashSpace modules"
  homepage "https://dashspace.space"
  version "$VERSION"

  if Hardware::CPU.arm?
    url "https://github.com/your-org/dashspace-cli/releases/download/v#{version}/dashspace-#{version}-darwin-arm64.tar.gz"
    sha256 "$ARM64_SHA"
  else
    url "https://github.com/your-org/dashspace-cli/releases/download/v#{version}/dashspace-#{version}-darwin-amd64.tar.gz"
    sha256 "$AMD64_SHA"
  fi

  def install
    bin.install "dashspace-darwin-arm64" => "dashspace" if Hardware::CPU.arm?
    bin.install "dashspace-darwin-amd64" => "dashspace" if Hardware::CPU.intel?
  end

  test do
    system "#{bin}/dashspace", "--version"
  end
end
EOF

    echo "✅ Homebrew formula created"
}

create_archives() {
    echo "📁 Creating release archives..."

    [ -f "$BUILD_DIR/dashspace-linux-amd64" ] && tar -czf "$PACKAGE_DIR/dashspace-$VERSION-linux-amd64.tar.gz" -C "$BUILD_DIR" dashspace-linux-amd64
    [ -f "$BUILD_DIR/dashspace-linux-arm64" ] && tar -czf "$PACKAGE_DIR/dashspace-$VERSION-linux-arm64.tar.gz" -C "$BUILD_DIR" dashspace-linux-arm64
    [ -f "$BUILD_DIR/dashspace-darwin-amd64" ] && tar -czf "$PACKAGE_DIR/dashspace-$VERSION-darwin-amd64.tar.gz" -C "$BUILD_DIR" dashspace-darwin-amd64
    [ -f "$BUILD_DIR/dashspace-darwin-arm64" ] && tar -czf "$PACKAGE_DIR/dashspace-$VERSION-darwin-arm64.tar.gz" -C "$BUILD_DIR" dashspace-darwin-arm64
    [ -f "$BUILD_DIR/dashspace-windows-amd64.exe" ] && zip -j "$PACKAGE_DIR/dashspace-$VERSION-windows-amd64.zip" "$BUILD_DIR/dashspace-windows-amd64.exe"

    echo "✅ Archives created"
}

generate_checksums() {
    echo "🔐 Generating checksums..."
    cd "$PACKAGE_DIR"
    shasum -a 256 * > checksums.txt 2>/dev/null || true
    cd ..
    echo "✅ Checksums generated"
}

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

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi