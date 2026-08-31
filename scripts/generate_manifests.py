#!/usr/bin/env python3
"""Generate package-manager manifests from a published release checksum file."""

import argparse
import json
from pathlib import Path

REPOSITORY = "akash-kamat/switchboard"


def checksum_map(path: Path) -> dict[str, str]:
    result = {}
    for line in path.read_text(encoding="utf-8").splitlines():
        parts = line.split()
        if len(parts) == 2:
            result[parts[1].lstrip("*")] = parts[0]
    return result


def require(checksums: dict[str, str], name: str) -> str:
    try:
        return checksums[name]
    except KeyError as error:
        raise SystemExit(f"missing checksum for {name}") from error


def write(path: Path, contents: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(contents.rstrip() + "\n", encoding="utf-8", newline="\n")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--version", required=True, help="release version, with or without v")
    parser.add_argument("--checksums", required=True, type=Path)
    parser.add_argument("--output", type=Path, default=Path("distribution"))
    args = parser.parse_args()
    version = args.version.removeprefix("v")
    tag = f"v{version}"
    base = f"https://github.com/{REPOSITORY}/releases/download/{tag}"
    sums = checksum_map(args.checksums)

    archives = {
        "windows_amd64": "switchboard_windows_amd64.zip",
        "windows_arm64": "switchboard_windows_arm64.zip",
        "darwin_amd64": "switchboard_darwin_amd64.tar.gz",
        "darwin_arm64": "switchboard_darwin_arm64.tar.gz",
        "linux_amd64": "switchboard_linux_amd64.tar.gz",
        "linux_arm64": "switchboard_linux_arm64.tar.gz",
        "linux_armv7": "switchboard_linux_armv7.tar.gz",
    }
    hashes = {key: require(sums, filename) for key, filename in archives.items()}

    formula = f'''class Switchboard < Formula
  desc "Lightweight dashboard for system services and Docker containers"
  homepage "https://github.com/{REPOSITORY}"
  version "{version}"
  license "MIT"

  on_macos do
    on_intel do
      url "{base}/{archives['darwin_amd64']}"
      sha256 "{hashes['darwin_amd64']}"
    end
    on_arm do
      url "{base}/{archives['darwin_arm64']}"
      sha256 "{hashes['darwin_arm64']}"
    end
  end
  on_linux do
    on_intel do
      url "{base}/{archives['linux_amd64']}"
      sha256 "{hashes['linux_amd64']}"
    end
    on_arm do
      url "{base}/{archives['linux_arm64']}"
      sha256 "{hashes['linux_arm64']}"
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
    assert_match version.to_s, shell_output("#{{bin}}/switchboard version")
  end
end'''
    write(args.output / "homebrew" / "Formula" / "switchboard.rb", formula)

    scoop = {
        "version": version,
        "description": "Lightweight dashboard for system services and Docker containers",
        "homepage": f"https://github.com/{REPOSITORY}",
        "license": "MIT",
        "architecture": {
            "64bit": {"url": f"{base}/{archives['windows_amd64']}", "hash": hashes["windows_amd64"]},
            "arm64": {"url": f"{base}/{archives['windows_arm64']}", "hash": hashes["windows_arm64"]},
        },
        "bin": "switchboard.exe",
        "checkver": {"github": f"https://github.com/{REPOSITORY}"},
        "autoupdate": {
            "architecture": {
                "64bit": {"url": f"https://github.com/{REPOSITORY}/releases/download/v$version/switchboard_windows_amd64.zip"},
                "arm64": {"url": f"https://github.com/{REPOSITORY}/releases/download/v$version/switchboard_windows_arm64.zip"},
            }
        },
    }
    write(args.output / "scoop" / "switchboard.json", json.dumps(scoop, indent=2))

    winget_version = f'''PackageIdentifier: AkashKamat.Switchboard
PackageVersion: {version}
DefaultLocale: en-US
ManifestType: version
ManifestVersion: 1.9.0'''
    winget_locale = f'''PackageIdentifier: AkashKamat.Switchboard
PackageVersion: {version}
PackageLocale: en-US
Publisher: Akash Kamat
PackageName: Switchboard
License: MIT
ShortDescription: Lightweight dashboard for system services and Docker containers
PackageUrl: https://github.com/{REPOSITORY}
ManifestType: defaultLocale
ManifestVersion: 1.9.0'''
    winget_installer = f'''PackageIdentifier: AkashKamat.Switchboard
PackageVersion: {version}
InstallerType: zip
NestedInstallerType: portable
NestedInstallerFiles:
  - RelativeFilePath: switchboard.exe
    PortableCommandAlias: switchboard
Installers:
  - Architecture: x64
    InstallerUrl: {base}/{archives['windows_amd64']}
    InstallerSha256: {hashes['windows_amd64']}
  - Architecture: arm64
    InstallerUrl: {base}/{archives['windows_arm64']}
    InstallerSha256: {hashes['windows_arm64']}
ManifestType: installer
ManifestVersion: 1.9.0'''
    winget = args.output / "winget" / "manifests" / "a" / "AkashKamat" / "Switchboard" / version
    write(winget / "AkashKamat.Switchboard.yaml", winget_version)
    write(winget / "AkashKamat.Switchboard.locale.en-US.yaml", winget_locale)
    write(winget / "AkashKamat.Switchboard.installer.yaml", winget_installer)

    nuspec = f'''<?xml version="1.0"?>
<package xmlns="http://schemas.microsoft.com/packaging/2015/06/nuspec.xsd">
  <metadata>
    <id>switchboard</id><version>{version}</version><title>Switchboard</title>
    <authors>Akash Kamat</authors><license type="expression">MIT</license>
    <projectUrl>https://github.com/{REPOSITORY}</projectUrl>
    <description>Lightweight dashboard for system services and Docker containers.</description>
    <tags>dashboard docker service raspberry-pi</tags>
  </metadata>
</package>'''
    chocolatey_script = f'''$ErrorActionPreference = 'Stop'
$arch = if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq 'Arm64') {{ 'arm64' }} else {{ 'amd64' }}
$urls = @{{ amd64 = '{base}/{archives['windows_amd64']}'; arm64 = '{base}/{archives['windows_arm64']}' }}
$hashes = @{{ amd64 = '{hashes['windows_amd64']}'; arm64 = '{hashes['windows_arm64']}' }}
Install-ChocolateyZipPackage -PackageName 'switchboard' -Url64bit $urls[$arch] -Checksum64 $hashes[$arch] -ChecksumType64 'sha256' -UnzipLocation $toolsDir'''
    write(args.output / "chocolatey" / "switchboard.nuspec", nuspec)
    write(args.output / "chocolatey" / "tools" / "chocolateyinstall.ps1", chocolatey_script)

    aur_archives = {
        "x86_64": f"switchboard_{version}_linux_amd64.tar.gz",
        "aarch64": f"switchboard_{version}_linux_arm64.tar.gz",
        "armv7h": f"switchboard_{version}_linux_armv7.tar.gz",
    }
    # GoReleaser's stable archive names do not contain a version; alias them in source names.
    pkgbuild = f'''pkgname=switchboard-bin
pkgver={version}
pkgrel=1
pkgdesc="Lightweight dashboard for system services and Docker containers"
arch=('x86_64' 'aarch64' 'armv7h')
url="https://github.com/{REPOSITORY}"
license=('MIT')
depends=('systemd')
source_x86_64=("{aur_archives['x86_64']}::{base}/{archives['linux_amd64']}")
source_aarch64=("{aur_archives['aarch64']}::{base}/{archives['linux_arm64']}")
source_armv7h=("{aur_archives['armv7h']}::{base}/{archives['linux_armv7']}")
sha256sums_x86_64=('{hashes['linux_amd64']}')
sha256sums_aarch64=('{hashes['linux_arm64']}')
sha256sums_armv7h=('{hashes['linux_armv7']}')

package() {{
  install -Dm755 "$srcdir/switchboard" "$pkgdir/usr/bin/switchboard"
  install -Dm644 "$srcdir/config.example.yaml" "$pkgdir/etc/switchboard/config.yaml"
  install -Dm644 "$srcdir/LICENSE" "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
}}'''
    write(args.output / "aur" / "PKGBUILD", pkgbuild)


if __name__ == "__main__":
    main()
