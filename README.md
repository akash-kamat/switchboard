# Switchboard

A lightweight, single-page dashboard for viewing and controlling Docker containers and native services on Linux, Windows, and macOS. The backend and embedded frontend ship as one Go executable; there is no Node.js runtime, bundler, or frontend build step.

## Features

- Start, stop, restart, and control autostart for Docker and systemd services.
- Live service CPU/RAM usage and configurable system metrics.
- Add, edit, and remove dashboard entries in the web UI.
- GUI settings plus a validated YAML editor. Invalid YAML is never written.
- Configurable refresh interval, overview cards, sidebar details, theme, and background.
- Responsive fixed-window layout with scrolling contained inside the dashboard.

## Repository layout

```text
cmd/switchboard/       CLI entry point
internal/config/       configuration parsing, validation, and atomic saving
internal/docker/       Docker Engine adapter
internal/platform/     host capability contracts
internal/server/       HTTP API and embedded web assets
deploy/                systemd unit
packaging/             native package lifecycle scripts
dist/                  local build output (not committed)
.goreleaser.yaml       reproducible release and package matrix
config.example.yaml    documented starter configuration
PROJECT.md             original project specification
docs/                  installation, configuration, API, and support contracts
```

## Build

Go 1.25.13 or newer is required on the build machine. Published binaries do
not require Go to be installed.

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -trimpath -ldflags="-s -w" \
  -o dist/switchboard-linux-arm64 ./cmd/switchboard
```

The resulting `dist/switchboard-linux-arm64` file is the complete ARM64 Linux application, including the website and images.

## Install on DietPi

See [docs/linux-installation.md](docs/linux-installation.md) for installation,
upgrade, rollback, and removal steps.

```sh
sudo groupadd --system switchboard 2>/dev/null || true
sudo useradd --system --gid switchboard --home-dir /var/lib/switchboard --shell /usr/sbin/nologin switchboard 2>/dev/null || true
sudo install -m 0755 dist/switchboard-linux-arm64 /usr/bin/switchboard
sudo install -d -o root -g switchboard -m 0750 /etc/switchboard
sudo install -d -o switchboard -g switchboard -m 0750 /var/lib/switchboard
sudo install -o root -g switchboard -m 0660 config.example.yaml /etc/switchboard/config.yaml
sudo install -m 0644 deploy/switchboard.service /etc/systemd/system/switchboard.service
sudo systemctl daemon-reload
sudo systemctl enable --now switchboard
```

Open `http://dietpi-hostname:8080`. Check operation with:

```sh
systemctl status switchboard
journalctl -u switchboard -n 100 --no-pager
```

The supplied service runs as the unprivileged `switchboard` account. Docker and
systemd mutations require explicit additional permissions; they are not granted by
the installer.

Tagged releases provide `.deb`, `.rpm`, and `.pkg.tar.zst` packages for amd64,
arm64, and ARMv7, plus portable archives. The native-package commands and manual
fallback are documented in [docs/linux-installation.md](docs/linux-installation.md).
One-line installers, reproducible version selection, checksums, SBOMs, and
Sigstore verification are documented in [docs/releases.md](docs/releases.md).

## Package managers

Homebrew and Scoop channels are live:

```sh
brew tap akash-kamat/switchboard
brew install switchboard
```

```powershell
scoop bucket add switchboard https://github.com/akash-kamat/scoop-switchboard
scoop install switchboard
```

See [docs/package-manager-distribution.md](docs/package-manager-distribution.md)
for upgrade, service, and uninstall commands and the status of other registries.

Switchboard has no login by design. Keep it on a trusted LAN, bind it to `127.0.0.1:8080` behind an authenticated reverse proxy, or firewall port 8080 from untrusted networks.

## Configuration

Start with [config.example.yaml](config.example.yaml) and see
[docs/configuration.md](docs/configuration.md). Existing configurations without a
`version` field remain valid and are interpreted as schema version 1.

Service icons are detected automatically for common applications. Set `icon` to a bundled name such as `jellyfin`, `transmission`, `dockge`, `n8n`, `plex`, `qbittorrent`, `nextcloud`, `paperless-ngx`, `portainer`, `uptime-kuma`, or `adguard-home`. You can also use an `http` or `https` image URL. Leave it empty or use `auto` for name-based detection.

The six available overview metrics are `cpu`, `memory`, `storage`, `temperature`, `load`, and `swap`. CPU, memory, and storage are enabled by default. Sidebar details can include `hostname`, `local_ip`, `os`, `uptime`, `kernel`, and `architecture`.

The Settings page can edit the same options. Saves use a temporary file and atomic rename, so a validation or write failure leaves the existing configuration untouched. Changing `listen` requires a service restart; other changes apply immediately.

Docker actions use `/var/run/docker.sock` directly. Systemd actions use `systemctl`. Linux host and process statistics come from `/proc` and `/sys`.

## API

The current routes and response contracts are documented in [docs/api.md](docs/api.md).

Current platform and toolchain requirements are in [docs/support.md](docs/support.md).
Windows and macOS installation is documented in
[docs/windows-macos-installation.md](docs/windows-macos-installation.md).
The lightweight runtime budget and measurement commands are in
[docs/performance.md](docs/performance.md).
Read [docs/security.md](docs/security.md) before enabling Docker access or exposing
Switchboard beyond the local machine.
Backup and recovery procedures are in [docs/backup-recovery.md](docs/backup-recovery.md).
Common failures are covered in [docs/troubleshooting.md](docs/troubleshooting.md).

## Development

```sh
go test ./...
go vet ./...
node --check internal/server/web/app.js
```

Run locally with `go run ./cmd/switchboard -config config.example.yaml`. Live system metrics and service integrations require Linux.

The explicit CLI form is:

```sh
switchboard serve --config config.example.yaml
switchboard validate-config config.example.yaml
switchboard version
```

The original `switchboard -config ...` invocation remains supported.

Packaged installs default to `/etc/switchboard/config.yaml` on Linux,
`%ProgramData%\Switchboard\config.yaml` on Windows, and
`/Library/Application Support/Switchboard/config.yaml` on macOS. Override runtime
integration paths with `--docker-socket` and `--disk-path` when needed.
