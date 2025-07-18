class Dashspace < Formula
  desc "Official DashSpace CLI for creating and publishing modules"
  homepage "https://dashspace.space"
  license "MIT"
  version "1.0.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/devlyspace/devly-cli/releases/download/#{version}/dashspace-darwin-arm64"
      sha256 "PLACEHOLDER_ARM64_SHA"
    else
      url "https://github.com/devlyspace/devly-cli/releases/download/#{version}/dashspace-darwin-amd64"
      sha256 "PLACEHOLDER_AMD64_SHA"
    end
  end

  def install
    if Hardware::CPU.arm?
      bin.install "dashspace-darwin-arm64" => "dashspace"
    else
      bin.install "dashspace-darwin-amd64" => "dashspace"
    end

    begin
      generate_completions_from_executable(bin/"dashspace", "completion")
    rescue
      # Ignore if completion not supported
    end
  end

  test do
    system "#{bin}/dashspace", "--version"
    assert_match version.to_s, shell_output("#{bin}/dashspace --version")
  end
end