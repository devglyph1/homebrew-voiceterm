class Voiceterm < Formula
  desc "A voice-powered AI terminal assistant"
  homepage "https://github.com/devglyph1/homebrew-voiceterm"
  url "https://github.com/devglyph1/homebrew-voiceterm/releases/download/v0.1.2/voiceterm.tar.gz"
  sha256 "6174b82d3e8e382a21f818e60601f997aab42ff247e4409513dd309006b367a9"
  version "0.1.2"

  def install
    bin.install "voiceterm"
  end
end
