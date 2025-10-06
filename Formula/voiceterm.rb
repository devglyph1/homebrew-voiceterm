class Voiceterm < Formula
  desc "A voice-powered AI terminal assistant"
  homepage "https://github.com/devglyph1/homebrew-voiceterm"
  url "https://github.com/devglyph1/homebrew-voiceterm/releases/download/v0.1.3/voiceterm.tar.gz"
  sha256 "b8716f4c258f7276b9e4be080db671ace584c748db5403f107c4cee492557d3b"
  version "0.1.3"

  def install
    bin.install "voiceterm"
  end
end
