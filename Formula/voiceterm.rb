class Voiceterm < Formula
  desc "A voice-powered AI terminal assistant"
  homepage "https://github.com/devglyph1/homebrew-voiceterm"
  url "https://github.com/devglyph1/homebrew-voiceterm/releases/download/v0.1.5/voiceterm.tar.gz"
  sha256 "a10eeeb6d347055c133609b407cb096b39488f2a70039ec1f2c64437a53191a0"
  version "0.1.5"

  def install
    bin.install "voiceterm"
  end
end
