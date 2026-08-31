# Support policy

## Current release baseline

- Build toolchain: Go 1.23 or newer.
- Runtime: Linux on `amd64`, `arm64`, and ARMv7.
- Init integration: systemd.
- Container integration: Docker Engine API v1.41 through a Unix socket.
- Tested project commands: `go test ./...`, `go vet ./...`, frontend syntax
  checks, and cross-builds for every target architecture.
- Package CI: Debian install/reinstall/remove and failed-install behavior on an
  Ubuntu runner, plus RPM and Arch install/reinstall/remove behavior in clean
  Fedora and Arch containers.

The current target is DietPi 9 / Raspberry Pi OS Bookworm and contemporary
Debian, Ubuntu, Fedora, and Arch systems using systemd. Windows and macOS builds
compile, but their native runtime integrations remain roadmap work.

Switchboard embeds all frontend assets and requires no Node.js runtime. Docker is
optional unless Docker services are configured. systemd is required only for
configured systemd services.
