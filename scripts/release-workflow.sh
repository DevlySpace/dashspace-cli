#!/bin/bash

set -e

VERSION=${1}
HOMEBREW_TAP_REPO=${2:-"../homebrew-dashspace"}

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version> [homebrew-tap-repo-path]"
    echo "Example: $0 1.0.0"
    echo "Example: $0 1.0.1 ../homebrew-dashspace"
    exit 1
fi

echo "🚀 Complete release workflow for DashSpace CLI v$VERSION"
echo "=============================================="

echo ""
echo "📦 Step 1: Building all packages..."
make clean
make package VERSION="$VERSION"

echo ""
echo "🧪 Step 2: Testing packages..."
make test-packages

echo ""
echo "🍺 Step 3: Updating Homebrew formula..."
make update-homebrew VERSION="$VERSION"

echo ""
echo "📤 Step 4: Creating GitHub release..."
if command -v gh &> /dev/null; then
    echo "Creating GitHub release with gh CLI..."
    gh release create "v$VERSION" packages/* \
        --title "DashSpace CLI v$VERSION" \
        --notes "Release notes for v$VERSION" \
        --verify-tag

    echo "✅ GitHub release created!"
else
    echo "⚠️  GitHub CLI not found. Please create the release manually:"
    echo "   1. Go to https://github.com/dashspace/cli/releases/new"
    echo "   2. Tag: v$VERSION"
    echo "   3. Upload files from packages/ directory"
fi

echo ""
echo "🏠 Step 5: Updating Homebrew tap..."
if [ -d "$HOMEBREW_TAP_REPO" ]; then
    echo "Copying formula to Homebrew tap repository..."
    mkdir -p "$HOMEBREW_TAP_REPO/Formula"
    cp dashspace.rb "$HOMEBREW_TAP_REPO/Formula/"

    cd "$HOMEBREW_TAP_REPO"
    git add Formula/dashspace.rb
    git commit -m "Update DashSpace CLI to v$VERSION" || echo "No changes to commit"

    echo "📋 Homebrew tap updated. Don't forget to push:"
    echo "   cd $HOMEBREW_TAP_REPO && git push origin main"
    cd - > /dev/null
else
    echo "⚠️  Homebrew tap repository not found at $HOMEBREW_TAP_REPO"
    echo "   Please copy dashspace.rb to your homebrew-dashspace repository manually"
fi

echo ""
echo "🎉 Release workflow completed!"
echo ""
echo "📦 What was created:"
echo "  ✅ All platform packages in packages/"
echo "  ✅ Updated Homebrew formula (dashspace.rb)"
if command -v gh &> /dev/null; then
    echo "  ✅ GitHub release v$VERSION"
else
    echo "  ⚠️  GitHub release (manual step needed)"
fi
echo ""
echo "🚀 Next steps:"
echo "  1. Push Homebrew tap changes if not done automatically"
echo "  2. Test installation: brew tap dashspace/dashspace && brew install dashspace"
echo "  3. Update documentation"
echo "  4. Announce the release"
echo ""
echo "📋 Installation commands for users:"
echo "  macOS: brew tap dashspace/dashspace && brew install dashspace"
echo "  Linux: wget https://github.com/dashspace/cli/releases/download/v$VERSION/dashspace_${VERSION}_amd64.deb && sudo dpkg -i dashspace_${VERSION}_amd64.deb"
echo "  Universal: curl -fsSL https://get.dashspace.dev | sh"