class Dashspace < Formula
  desc "Official DashSpace CLI for creating and publishing modules"
  homepage "https://dashspace.space"
  license "MIT"
  version "1.0.5"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/devlyspace/dashspace-cli/releases/download/1.0.5/dashspace-darwin-arm64"
      sha256 "a1f1fba483fefb14151b8bdb1a07666c9ed29bb6af42f3348cbb4afd06da9d32"
    else
      url "https://github.com/devlyspace/dashspace-cli/releases/download/1.0.5/dashspace-darwin-amd64"
      sha256 "bc5ffe2a6f4043c40d8559283c358f9fbf1c01f110dac28858fc333db91e033a"
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
    assert_match "dashspace version 1.0.5", shell_output("#{bin}/dashspace --version")
  end
end
