#!/bin/sh
set -eu

repository="akash-kamat/switchboard"
version="${SWITCHBOARD_VERSION:-latest}"

usage() {
  echo "Usage: install.sh [--version v1.2.3] [--uninstall]" >&2
}

uninstall=false
while [ "$#" -gt 0 ]; do
  case "$1" in
    --version) version="${2:?missing version}"; shift 2 ;;
    --uninstall) uninstall=true; shift ;;
    -h|--help) usage; exit 0 ;;
    *) usage; exit 2 ;;
  esac
done

if [ "$(id -u)" -ne 0 ]; then
  echo "Run this installer as root (for example: curl ... | sudo sh)." >&2
  exit 1
fi

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in linux|darwin) ;; *) echo "Unsupported operating system: $os" >&2; exit 1 ;; esac
machine=$(uname -m)
case "$machine" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  armv7l|armv7) arch=armv7 ;;
  *) echo "Unsupported architecture: $machine" >&2; exit 1 ;;
esac
if [ "$os" = darwin ] && [ "$arch" = armv7 ]; then echo "Unsupported macOS architecture" >&2; exit 1; fi

if [ "$uninstall" = true ]; then
  if [ "$os" = linux ]; then
    systemctl disable --now switchboard 2>/dev/null || true
    rm -f /etc/systemd/system/switchboard.service /usr/lib/systemd/system/switchboard.service /usr/local/bin/switchboard
    systemctl daemon-reload
    echo "Removed Switchboard; /etc/switchboard and /var/lib/switchboard were preserved."
  else
    launchctl bootout system/com.akash-kamat.switchboard 2>/dev/null || true
    rm -f /Library/LaunchDaemons/com.akash-kamat.switchboard.plist /usr/local/bin/switchboard
    echo "Removed Switchboard; /Library/Application Support/Switchboard was preserved."
  fi
  exit 0
fi

command -v curl >/dev/null 2>&1 || { echo "curl is required" >&2; exit 1; }
command -v sha256sum >/dev/null 2>&1 || command -v shasum >/dev/null 2>&1 || { echo "sha256sum or shasum is required" >&2; exit 1; }

base="https://github.com/$repository/releases"
if [ "$version" = latest ]; then base="$base/latest/download"; else base="$base/download/$version"; fi
archive="switchboard_${os}_${arch}.tar.gz"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
curl --fail --silent --show-error --location "$base/$archive" --output "$tmp/$archive"
curl --fail --silent --show-error --location "$base/checksums.txt" --output "$tmp/checksums.txt"
expected=$(awk -v file="$archive" '$2 == file {print $1}' "$tmp/checksums.txt")
[ -n "$expected" ] || { echo "Checksum for $archive is missing" >&2; exit 1; }
if command -v sha256sum >/dev/null 2>&1; then actual=$(sha256sum "$tmp/$archive" | awk '{print $1}'); else actual=$(shasum -a 256 "$tmp/$archive" | awk '{print $1}'); fi
[ "$actual" = "$expected" ] || { echo "Checksum verification failed" >&2; exit 1; }
tar -xzf "$tmp/$archive" -C "$tmp"
install -m 0755 "$tmp/switchboard" /usr/local/bin/switchboard

if [ "$os" = linux ]; then
  getent group switchboard >/dev/null 2>&1 || groupadd --system switchboard
  id switchboard >/dev/null 2>&1 || useradd --system --gid switchboard --home-dir /var/lib/switchboard --shell "$(command -v nologin || echo /usr/sbin/nologin)" switchboard
  install -d -o root -g switchboard -m 0750 /etc/switchboard
  install -d -o switchboard -g switchboard -m 0750 /var/lib/switchboard
  if [ ! -e /etc/switchboard/config.yaml ]; then install -o root -g switchboard -m 0660 "$tmp/config.example.yaml" /etc/switchboard/config.yaml; fi
  curl --fail --silent --show-error --location "https://raw.githubusercontent.com/$repository/main/deploy/switchboard.service" --output /etc/systemd/system/switchboard.service
  systemctl daemon-reload
  systemctl enable --now switchboard
else
  install -d -m 0755 "/Library/Application Support/Switchboard"
  if [ ! -e "/Library/Application Support/Switchboard/config.yaml" ]; then install -m 0644 "$tmp/config.example.yaml" "/Library/Application Support/Switchboard/config.yaml"; fi
  curl --fail --silent --show-error --location "https://raw.githubusercontent.com/$repository/main/deploy/com.akash-kamat.switchboard.plist" --output /Library/LaunchDaemons/com.akash-kamat.switchboard.plist
  chmod 0644 /Library/LaunchDaemons/com.akash-kamat.switchboard.plist
  launchctl bootout system/com.akash-kamat.switchboard 2>/dev/null || true
  launchctl bootstrap system /Library/LaunchDaemons/com.akash-kamat.switchboard.plist
fi
echo "Switchboard installed. Configuration was preserved if it already existed."
