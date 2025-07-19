#!/bin/bash

set -e

VERSION=${1}
HOMEBREW_TAP_REPO=${2:-"devlyspace/homebrew-dashspace"}
BUILD_DIR="dist"

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version> [homebrew-tap-repo]"
    echo "Example: $0 1.0.0"
    echo "Example: $0 1.0.0 myorg/homebrew-mytool"
    exit 1
fi

echo "🍺 Auto-updating Homebrew tap for $VERSION"

if [ ! -f "$BUILD_DIR/dashspace-darwin-amd64" ] || [ ! -f "$BUILD_DIR/dashspace-darwin-arm64" ]; then
    echo "❌ macOS binaries not found. Run 'make build-all' first"
    exit 1
fi

echo "🔐 Calculating checksums..."
AMD64_SHA=$(shasum -a 256 "packages/dashspace-$VERSION-darwin-amd64.tar.gz" | cut -d' ' -f1)
ARM64_SHA=$(shasum -a 256 "packages/dashspace-$VERSION-darwin-arm64.tar.gz" | cut -d' ' -f1)

echo "📝 Generating updated formula..."

FORMULA_CONTENT=$(cat << EOF
class Dashspace < Formula
  desc "Official DashSpace CLI for creating and publishing modules"
  homepage "https://dashspace.space"
  license "MIT"
  version "$VERSION"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/devlyspace/devly-cli/releases/download/$VERSION/dashspace-$VERSION-darwin-arm64.tar.gz"
      sha256 "$ARM64_SHA"
    else
      url "https://github.com/devlyspace/devly-cli/releases/download/$VERSION/dashspace-$VERSION-darwin-amd64.tar.gz"
      sha256 "$AMD64_SHA"
    end
  end

  def install
    bin.install "dashspace" => "dashspace"

    begin
      generate_completions_from_executable(bin/"dashspace", "completion")
    rescue
    end
  end

  test do
    system "#{bin}/dashspace", "--version"
    assert_match "dashspace version $VERSION", shell_output("#{bin}/dashspace --version")
  end
end
EOF
)

if [ -n "$GITHUB_TOKEN" ]; then
    echo "🚀 Updating Homebrew tap via GitHub API..."

    CURRENT_SHA=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
        "https://api.github.com/repos/$HOMEBREW_TAP_REPO/contents/Formula/dashspace.rb" \
        | jq -r '.sha // empty')

    ENCODED_CONTENT=$(echo "$FORMULA_CONTENT" | base64 -w 0)

    if [ -n "$CURRENT_SHA" ]; then
        UPDATE_PAYLOAD=$(jq -n \
            --arg message "Update DashSpace CLI to $VERSION" \
            --arg content "$ENCODED_CONTENT" \
            --arg sha "$CURRENT_SHA" \
            '{message: $message, content: $content, sha: $sha}')
    else
        UPDATE_PAYLOAD=$(jq -n \
            --arg message "Add DashSpace CLI $VERSION" \
            --arg content "$ENCODED_CONTENT" \
            '{message: $message, content: $content}')
    fi

    RESPONSE=$(curl -s -X PUT \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$UPDATE_PAYLOAD" \
        "https://api.github.com/repos/$HOMEBREW_TAP_REPO/contents/Formula/dashspace.rb")

    if echo "$RESPONSE" | grep -q '"commit"'; then
        echo "✅ Homebrew tap updated successfully!"
        echo "🔗 View at: https://github.com/$HOMEBREW_TAP_REPO/blob/main/Formula/dashspace.rb"
    else
        echo "❌ Failed to update Homebrew tap"
        echo "Response: $RESPONSE"
        exit 1
    fi

else
    echo "⚠️  GITHUB_TOKEN not set. Manual update required."
    echo ""
    echo "📋 Save this formula to your homebrew tap repo:"
    echo "----------------------------------------"
    echo "$FORMULA_CONTENT"
    echo "----------------------------------------"
    echo ""
    echo "🚀 Manual steps:"
    echo "  1. Go to: https://github.com/$HOMEBREW_TAP_REPO"
    echo "  2. Edit Formula/dashspace.rb"
    echo "  3. Replace content with the formula above"
    echo "  4. Commit with message: 'Update DashSpace CLI to v$VERSION'"
fi

echo ""
echo "🍺 Users can now install with:"
echo "   brew tap devlyspace/dashspace"
echo "   brew install dashspace"