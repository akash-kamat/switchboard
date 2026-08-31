# Support policy

## Current release baseline

- Build toolchain: Go 1.25.13 or newer. Patch releases below this baseline are
  not used because of fixed Go standard-library vulnerabilities.
- Runtime: Linux (`amd64`, `arm64`, ARMv7), Windows (`amd64`, `arm64`), and
  macOS (`amd64`, Apple Silicon).
- Init integration: systemd, Windows Service Control Manager, and launchd.
- Container integration: Docker Engine API v1.41 through a Unix socket or the
  Docker Desktop Windows named pipe.
- Tested project commands: `go test ./...`, `go vet ./...`, frontend syntax
  checks, and cross-builds for every target architecture.
- Package CI: Debian install/reinstall/remove and failed-install behavior on an
  Ubuntu runner, plus RPM and Arch install/reinstall/remove behavior in clean
  Fedora and Arch containers.

The current targets are DietPi 9 / Raspberry Pi OS Bookworm, contemporary
Debian, Ubuntu, Fedora, and Arch systems using systemd, supported Windows 10/11
and Server releases, and supported Intel/Apple Silicon macOS releases.

Switchboard embeds all frontend assets and requires no Node.js runtime. Docker is
optional unless Docker services are configured. Native service management is
required only for configured `native` (or legacy `systemd`) entries.
