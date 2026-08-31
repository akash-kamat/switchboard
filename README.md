# Switchboard

A lightweight, single-page dashboard for viewing and controlling Docker containers and systemd services on DietPi. The backend and embedded frontend ship as one Go executable; there is no Node.js runtime, bundler, or frontend build step.

## Features

- Start, stop, restart, and control autostart for Docker and systemd services.
- Live service CPU/RAM usage and configurable system metrics.
- Add, edit, and remove dashboard entries in the web UI.
- GUI settings plus a validated YAML editor. Invalid YAML is never written.
- Configurable refresh interval, overview cards, sidebar details, theme, and background.
- Responsive fixed-window layout with scrolling contained inside the dashboard.

## Repository layout

```text
cmd/switchboard/       Go application and embedded web assets
deploy/                systemd unit
dist/                  local build output (not committed)
config.example.yaml    documented starter configuration
PROJECT.md             original project specification
```

## Build

Go 1.23 or newer is required on the build machine.

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -trimpath -ldflags="-s -w" \
  -o dist/switchboard-linux-arm64 ./cmd/switchboard
```

The resulting `dist/switchboard-linux-arm64` file is the complete ARM64 Linux application, including the website and images.

## Install on DietPi

```sh
sudo install -m 0755 dist/switchboard-linux-arm64 /usr/local/bin/switchboard
sudo install -d -m 0755 /etc/switchboard
sudo install -m 0644 config.example.yaml /etc/switchboard/config.yaml
sudo install -m 0644 deploy/switchboard.service /etc/systemd/system/switchboard.service
sudo systemctl daemon-reload
sudo systemctl enable --now switchboard
```

Open `http://dietpi-hostname:8080`. Check operation with:

```sh
systemctl status switchboard
journalctl -u switchboard -n 100 --no-pager
```

The supplied service runs as root because systemd control requires privilege and access to Docker's socket is effectively root-level access. It permits writes to `/etc/switchboard` so settings can be saved from the UI.

Switchboard has no login by design. Keep it on a trusted LAN, bind it to `127.0.0.1:8080` behind an authenticated reverse proxy, or firewall port 8080 from untrusted networks.

## Configuration

Start with [config.example.yaml](config.example.yaml). Existing configurations containing only `listen` and `services` remain valid; dashboard defaults are filled automatically.

Service icons are detected automatically for common applications. Set `icon` to a bundled name such as `jellyfin`, `transmission`, `dockge`, `n8n`, `plex`, `qbittorrent`, `nextcloud`, `paperless-ngx`, `portainer`, `uptime-kuma`, or `adguard-home`. You can also use an `http` or `https` image URL. Leave it empty or use `auto` for name-based detection.

The six available overview metrics are `cpu`, `memory`, `storage`, `temperature`, `load`, and `swap`. CPU, memory, and storage are enabled by default. Sidebar details can include `hostname`, `local_ip`, `os`, `uptime`, `kernel`, and `architecture`.

The Settings page can edit the same options. Saves use a temporary file and atomic rename, so a validation or write failure leaves the existing configuration untouched. Changing `listen` requires a service restart; other changes apply immediately.

Docker actions use `/var/run/docker.sock` directly. Systemd actions use `systemctl`. Linux host and process statistics come from `/proc` and `/sys`.

## API

- `GET /api/services`
- `POST /api/services/{name}/start`
- `POST /api/services/{name}/stop`
- `POST /api/services/{name}/restart`
- `POST /api/services/{name}/autostart` with JSON `{ "enabled": true }`
- `GET /api/system`
- `GET /api/config`
- `POST /api/config/validate`
- `PUT /api/config`

## Development

```sh
go test ./...
go vet ./...
node --check cmd/switchboard/web/app.js
```

Run locally with `go run ./cmd/switchboard -config config.example.yaml`. Live system metrics and service integrations require Linux.
