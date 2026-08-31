# PROJECT.md — Lightweight Self-Hosted Dashboard (Homepage Replacement)

## Goal
Build a lightweight dashboard for DietPi that replaces Homepage. Same core idea — a single page listing services with links — but far lower resource usage, and with the ability to manage **both Docker containers and native systemd services** (something Homepage cannot do).

## Why Not Homepage
- Homepage (Node/Next.js) uses ~110MB+ RAM idle.
- Only supports start/stop/restart for Docker (and Kubernetes) — not native systemd services like Transmission.
- Target: a binary using ~15-25MB RAM idle.

## Tech Stack
- **Backend:** Go (compiles to single static binary, no runtime dependency, low memory footprint)
- **Frontend:** Plain HTML/CSS/vanilla JS — no React, no build step, no bundler
- **Config format:** Single YAML file
- **Deployment target:** DietPi on ARM64 (Raspberry Pi or similar SBC)

## Core Features

### 1. Service Listing
- Read services from a YAML config file.
- Group services under categories (e.g., Media, Dashboards, Automation).
- Display name, description, and link (href) for each service.

### 2. Service Types
Support two types of services:
- `docker` — controlled via the Docker Engine API (using Docker socket)
- `systemd` — controlled via shelling out to `systemctl`

### 3. Start / Stop / Restart
- For `docker` type: call Docker API to start/stop/restart the container.
- For `systemd` type: run `systemctl start|stop|restart <unit>` (requires root or sudoers permission for the running user).

### 4. Autostart / Startup Behavior Toggle
- For `docker` type: toggle container restart policy (`always` vs `no`) via Docker API.
- For `systemd` type: run `systemctl enable|disable <unit>`.

### 5. Resource Stats Display
- For `docker` type: fetch CPU% and memory usage via Docker stats API.
- For `systemd` type: fetch CPU% and memory via reading `/proc/<pid>/stat` and `/proc/<pid>/status`, using the PID from `systemctl show <unit> --property=MainPID`.

### 6. System-Wide Stats
- Overall CPU usage
- Overall RAM usage (used/free)
- Disk usage (used/free) for main partition

## Config File Format (Example)
```yaml
services:
  - name: Transmission
    type: systemd
    unit: transmission-daemon
    href: http://192.168.1.102:9091/
    description: Torrent client
    group: Media
    autostart: true

  - name: Dockge
    type: docker
    container: dockge-dockge-1
    href: http://192.168.1.102:5001/
    description: Compose stack manager
    group: Dashboards
    autostart: true

  - name: n8n
    type: docker
    container: n8n
    href: http://192.168.1.102:5678/
    description: Workflow automation
    group: Automation
    autostart: false
```

## API Endpoints (Backend)
- `GET /api/services` — list all services with current status and stats
- `POST /api/services/:name/start` — start a service
- `POST /api/services/:name/stop` — stop a service
- `POST /api/services/:name/restart` — restart a service
- `POST /api/services/:name/autostart` — toggle autostart on/off
- `GET /api/system` — system-wide CPU/RAM/disk stats

## Permissions Needed
- Access to `/var/run/docker.sock` for Docker control.
- Root privileges (or scoped sudoers rules) for `systemctl` control of native services.

## Non-Goals (Out of Scope for v1)
- No multi-user auth/login system (assume trusted local network use, matching current Homepage setup).
- No plugin/widget ecosystem like Homepage's weather/search widgets — just service management and stats.
- No mobile app — web page only, responsive layout is enough.

## Success Criteria
- Single compiled Go binary, no external runtime dependencies.
- Can list, start, stop, restart, and toggle autostart for both Docker containers and systemd services from one page.
- Runs reliably on DietPi ARM64.