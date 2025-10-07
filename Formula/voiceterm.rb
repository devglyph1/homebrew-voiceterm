class Voiceterm < Formula
  desc "A voice-powered AI terminal assistant"
  homepage "https://github.com/devglyph1/homebrew-voiceterm"
  url "https://github.com/devglyph1/homebrew-voiceterm/releases/download/v0.1.4/voiceterm.tar.gz"
  sha256 "8b9dec590719d50c6e107b84ae35a1864e61c8bda17a5b9729dde29b0e23521f"
  version "0.1.4"

  def install
    bin.install "voiceterm"
  end
end
