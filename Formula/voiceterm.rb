class Voiceterm < Formula
  desc "A voice-powered AI terminal assistant"
  homepage "https://github.com/devglyph1/homebrew-voiceterm"
  url "https://github.com/devglyph1/homebrew-voiceterm/releases/download/v0.1.0/voiceterm.tar.gz"
  sha256 "98411b151bbebd803f53400aae914612e50ea2bc903239dd1a695d86b0f28292"
  version "0.1.0"

  def install
    bin.install "voiceterm"
  end
end
