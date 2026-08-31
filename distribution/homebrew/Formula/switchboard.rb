class Switchboard < Formula
  desc "Lightweight dashboard for system services and Docker containers"
  homepage "https://github.com/akash-kamat/switchboard"
  version "0.1.0"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/akash-kamat/switchboard/releases/download/v0.1.0/switchboard_darwin_amd64.tar.gz"
      sha256 "e3d5f3a6b0126707b4d95066641332cbc36c40f5141a7ad96ce36de46504160b"
    end
    on_arm do
      url "https://github.com/akash-kamat/switchboard/releases/download/v0.1.0/switchboard_darwin_arm64.tar.gz"
      sha256 "964b5168805de0676a19d65c75ad8cdb027af00960d6342f3d8e52f59a67b324"
    end
  end
  on_linux do
    on_intel do
      url "https://github.com/akash-kamat/switchboard/releases/download/v0.1.0/switchboard_linux_amd64.tar.gz"
      sha256 "03465dc6b337501ac9096f253892b6309f4c5f8ede48dd002d330d4cbd1e59a3"
    end
    on_arm do
      url "https://github.com/akash-kamat/switchboard/releases/download/v0.1.0/switchboard_linux_arm64.tar.gz"
      sha256 "2cf36f9c68eb08c88b50605a9de707307eca8ade476453c7c41fe34bec897ad9"
    end
  end

  def install
    bin.install "switchboard"
    etc.install "config.example.yaml" => "switchboard/config.yaml"
  end

  service do
    run [opt_bin/"switchboard", "serve", "--config", etc/"switchboard/config.yaml"]
    keep_alive true
    log_path var/"log/switchboard.log"
    error_log_path var/"log/switchboard.error.log"
  end

  test do
    assert_match version.to_s, shell_output("#<built-in function bin>/switchboard version")
  end
end
