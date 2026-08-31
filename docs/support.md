# Support policy

## Current release baseline

- Build toolchain: Go 1.23 or newer.
- Runtime: Linux on `arm64`; `amd64` works from source but is not yet packaged.
- Init integration: systemd.
- Container integration: Docker Engine API v1.41 through a Unix socket.
- Tested project commands: `go test ./...` and `go vet ./...` on Windows; Linux
  behavior is covered by unit tests and an ARM64 cross-build.

The current target is DietPi 9 / Raspberry Pi OS Bookworm and contemporary Debian
or Ubuntu systems using systemd. Other Linux distributions are best-effort until
the Phase 2 package test matrix is automated. Windows and macOS are roadmap targets,
not supported runtimes yet.

Switchboard embeds all frontend assets and requires no Node.js runtime. Docker is
optional unless Docker services are configured. systemd is required only for
configured systemd services.
