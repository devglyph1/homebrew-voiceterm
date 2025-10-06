class Voiceterm < Formula
  desc "A voice-powered AI terminal assistant"
  homepage "https://github.com/devglyph1/homebrew-voiceterm"
  url "https://github.com/devglyph1/homebrew-voiceterm/releases/download/v0.1.1/voiceterm.tar.gz"
  sha256 "b66cbae652353ca8d820c60d92ce8f5c5fd6ae59a58f182be49a888bc4ebf423"
  version "0.1.1"

  def install
    bin.install "voiceterm"
  end
end
