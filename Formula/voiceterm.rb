class Voiceterm < Formula
  desc "A voice-powered AI terminal assistant"
  homepage "https://github.com/devglyph1/homebrew-voiceterm"
  url "https://github.com/devglyph1/homebrew-voiceterm/releases/download/v0.1.5/voiceterm.tar.gz"
  sha256 "2f52fcb65ae82cfe49cea5aa01c2e6e3f0e84542a5068e4ae07e4332ac50091c"
  version "0.1.5"

  def install
    bin.install "voiceterm"
  end
end
