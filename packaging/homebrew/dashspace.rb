class Dashspace < Formula
  desc "Official DashSpace CLI for creating and publishing modules"
  homepage "https://dashspace.space"
  license "MIT"
  version "1.0.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/devlyspace/dashspace-cli/releases/download/1.0.0/dashspace-darwin-arm64"
      sha256 "f11f43f9e073884face5a1bbec0051197b1657dc6500a45576a6e09cdc6c420b"
    else
      url "https://github.com/devlyspace/dashspace-cli/releases/download/1.0.0/dashspace-darwin-amd64"
      sha256 "feb2c94b28c94372d670e78d0df67a467d62e60471b10900bb9096748d1c2cad"
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
    assert_match "dashspace version 1.0.0", shell_output("#{bin}/dashspace --version")
  end
end
