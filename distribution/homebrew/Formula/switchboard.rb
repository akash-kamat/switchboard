class Switchboard < Formula
  desc "Lightweight dashboard for system services and Docker containers"
  homepage "https://github.com/akash-kamat/switchboard"
  version "0.1.1"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/akash-kamat/switchboard/releases/download/v0.1.1/switchboard_darwin_amd64.tar.gz"
      sha256 "e562e2f213dfd002b174217dc718979afbf1dcf5b416a9547a8afd6aae4c72bf"
    end
    on_arm do
      url "https://github.com/akash-kamat/switchboard/releases/download/v0.1.1/switchboard_darwin_arm64.tar.gz"
      sha256 "01ff37328a755a59271b76722c1ab5d2da4446320e046a1b82177d8c1bb911c1"
    end
  end
  on_linux do
    on_intel do
      url "https://github.com/akash-kamat/switchboard/releases/download/v0.1.1/switchboard_linux_amd64.tar.gz"
      sha256 "475ff9b987b29116d3faad516800d51d6cbbd30360e883b396f986974e509ede"
    end
    on_arm do
      url "https://github.com/akash-kamat/switchboard/releases/download/v0.1.1/switchboard_linux_arm64.tar.gz"
      sha256 "5239d50c25e01197d5638e0356586bfe843ecd434e0ba59129e32282350bd375"
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
    assert_match version.to_s, shell_output("#{bin}/switchboard version")
  end
end
