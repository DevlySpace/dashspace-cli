#!/bin/bash

set -e

VERSION=${1:-"1.0.0"}
BUILD_DIR="dist"
FORMULA_FILE="packaging/homebrew/dashspace.rb"

echo "🍺 Updating Homebrew formula for DashSpace CLI v$VERSION"

if [ ! -f "$BUILD_DIR/dashspace-darwin-amd64" ] || [ ! -f "$BUILD_DIR/dashspace-darwin-arm64" ]; then
    echo "❌ macOS binaries not found in $BUILD_DIR/"
    echo "Run 'make build-all' first to generate the binaries"
    exit 1
fi

if [ ! -f "$FORMULA_FILE" ]; then
    echo "⚠️  $FORMULA_FILE not found"
    echo "🔧 Creating directory and new formula file..."
    mkdir -p "packaging/homebrew"
else
    echo "📋 Found existing formula, creating backup..."
    cp "$FORMULA_FILE" "$FORMULA_FILE.backup"
fi

echo "🔐 Calculating checksums..."
AMD64_SHA=$(shasum -a 256 "$BUILD_DIR/dashspace-darwin-amd64" | cut -d' ' -f1)
ARM64_SHA=$(shasum -a 256 "$BUILD_DIR/dashspace-darwin-arm64" | cut -d' ' -f1)

echo "  AMD64 SHA256: $AMD64_SHA"
echo "  ARM64 SHA256: $ARM64_SHA"

echo "✏️  Updating formula..."

cp "$FORMULA_FILE" "$FORMULA_FILE.backup"

cat > "$FORMULA_FILE" << EOF
class Dashspace < Formula
  desc "Official DashSpace CLI for creating and publishing modules"
  homepage "https://dashspace.space"
  license "MIT"
  version "$VERSION"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/devlyspace/devly-cli/releases/download/$VERSION/dashspace-darwin-arm64"
      sha256 "$ARM64_SHA"
    else
      url "https://github.com/devlyspace/devly-cli/releases/download/$VERSION/dashspace-darwin-amd64"
      sha256 "$AMD64_SHA"
    end
  end

  def install
    if Hardware::CPU.arm?
      bin.install "dashspace-darwin-arm64" => "dashspace"
    else
      bin.install "dashspace-darwin-amd64" => "dashspace"
    end

    # Install auto-completion if supported
    begin
      generate_completions_from_executable(bin/"dashspace", "completion")
    rescue
      # Silently ignore if completion command doesn't exist
    end
  end

  test do
    system "#{bin}/dashspace", "--version"
    assert_match "dashspace version $VERSION", shell_output("#{bin}/dashspace --version")
  end
end
EOF

echo "✅ Formula updated successfully!"
echo ""
echo "📋 Summary:"
echo "  Version: $VERSION"
echo "  AMD64 binary SHA256: $AMD64_SHA"
echo "  ARM64 binary SHA256: $ARM64_SHA"
echo "  Formula file: $FORMULA_FILE"
echo ""
if [ -f "$FORMULA_FILE.backup" ]; then
    echo "🔍 Review changes:"
    echo "  diff $FORMULA_FILE.backup $FORMULA_FILE"
fi
echo ""
echo "🧪 Test locally (optional):"
echo "  brew install --build-from-source ./$FORMULA_FILE"
echo ""
echo "🚀 Formula updated at: $FORMULA_FILE"
echo ""
echo "📋 Installation options for users:"
echo ""
echo "  Option 1 - Direct install:"
echo "    brew install https://raw.githubusercontent.com/devlyspace/devly-cli/main/packaging/homebrew/dashspace.rb"
echo ""
echo "  Option 2 - Download then install:"
echo "    curl -O https://raw.githubusercontent.com/devlyspace/devly-cli/main/packaging/homebrew/dashspace.rb"
echo "    brew install ./dashspace.rb"
echo ""
echo "💡 Pro tip: Create a separate 'homebrew-dashspace' repo for a cleaner tap experience"