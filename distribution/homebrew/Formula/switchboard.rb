class Switchboard < Formula
  desc "Lightweight dashboard for system services and Docker containers"
  homepage "https://github.com/akash-kamat/switchboard"
  version "0.1.2"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/akash-kamat/switchboard/releases/download/v0.1.2/switchboard_darwin_amd64.tar.gz"
      sha256 "8605ade6109322e73935d9e974f1731808aa12d5c7feaf7cdd4f6fd370850b3f"
    end
    on_arm do
      url "https://github.com/akash-kamat/switchboard/releases/download/v0.1.2/switchboard_darwin_arm64.tar.gz"
      sha256 "2a34d97e001060207d35db1d98fa85b9a1ae70143225ef3051b725320c4ed24b"
    end
  end
  on_linux do
    on_intel do
      url "https://github.com/akash-kamat/switchboard/releases/download/v0.1.2/switchboard_linux_amd64.tar.gz"
      sha256 "45ed08622f9734e629c645dec9bf79a727298cb6b793b26fa597c27bc19b216e"
    end
    on_arm do
      url "https://github.com/akash-kamat/switchboard/releases/download/v0.1.2/switchboard_linux_arm64.tar.gz"
      sha256 "e3c0ef8d10228ee76961b424a6cdbccb9da8f616d7bf276a417a5de0b036a3a7"
    end
  end

  def install
    bin.install "switchboard"
    (etc/"switchboard").install "config.example.yaml" => "config.yaml"
  end

  service do
    run [opt_bin/"switchboard", "serve", "--config", etc/"switchboard/config.yaml"]
    keep_alive true
    log_path var/"log/switchboard.log"
    error_log_path var/"log/switchboard.error.log"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/switchboard version")
  end
end
