class Dashspace < Formula
  desc "Official DashSpace CLI for creating and publishing modules"
  homepage "https://dashspace.space"
  license "MIT"
  version "1.0.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/dashspace/cli/releases/download/v1.0.0/dashspace-darwin-arm64"
      sha256 "ARM64_SHA256_HERE"
    else
      url "https://github.com/dashspace/cli/releases/download/v1.0.0/dashspace-darwin-amd64"
      sha256 "AMD64_SHA256_HERE"
    end
  end

  def install
    if Hardware::CPU.arm?
      bin.install "dashspace-darwin-arm64" => "dashspace"
    else
      bin.install "dashspace-darwin-amd64" => "dashspace"
    end

    # Install auto-completion
    generate_completions_from_executable(bin/"dashspace", "completion")
  end

  test do
    system "#{bin}/dashspace", "--version"
    assert_match "dashspace version 1.0.0", shell_output("#{bin}/dashspace --version")
  end
end