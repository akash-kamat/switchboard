# Switchboard Roadmap

This document tracks the work required to turn Switchboard into a secure,
cross-platform application that can be installed as a native background service.

## Status legend

- [ ] Not started
- [x] Completed
- **Current phase** identifies the next phase to implement.

## Product decisions

- Switchboard is an application/service, not a public Go library.
- The dashboard, API, configuration model, and Docker support should remain shared
  across platforms.
- Metrics, native service management, paths, permissions, and service registration
  use platform-specific implementations.
- The CLI stays intentionally small:
  - `switchboard serve` runs the dashboard server.
  - `switchboard version` prints version and build information.
  - `switchboard validate-config` validates YAML without starting the server or
    changing any files.
- The existing invocation remains compatible during migration:
  `switchboard -config <path>` behaves like `switchboard serve -config <path>`.
- Native installers manage service installation, removal, starting, and stopping.
- Go is the only required runtime. Node.js will not be required and an npm package
  is not planned initially.
- User configuration must be preserved across upgrades and package removal unless
  the user explicitly requests a purge.

## Target support matrix

| Operating system | Architectures | Metrics | Docker | Native services | Packaging |
| --- | --- | --- | --- | --- | --- |
| Linux | amd64, arm64, armv7 | `/proc`, `/sys` | Docker Engine | systemd first | tar.gz, deb, rpm, Arch |
| Windows | amd64, arm64 | Windows APIs | Docker Desktop/Engine | Windows Services | zip, winget, Scoop, Chocolatey |
| macOS | amd64, arm64 | Darwin APIs | Docker Desktop | launchd | tar.gz, Homebrew |

OpenRC and other Linux init systems are possible future additions after systemd
support is stable.

## Phase 0 — Baseline and compatibility

Establish a reliable baseline before reorganizing code.

- [x] Record the current configuration format and API routes.
- [x] Add regression tests for existing configuration loading and saving.
- [x] Add regression tests for system, service, and configuration API behavior.
- [x] Document the current Linux installation and upgrade procedure.
- [x] Define supported Go version and minimum supported operating-system versions.
- [x] Add a configuration schema version, starting with `version: 1`.
- [x] Make missing schema version load as version 1 for backward compatibility.
- [x] Decide and document migration behavior for future schema versions.

### Exit criteria

- [x] Existing DietPi installations can use the reorganized build without changing
  their configuration.
- [x] The test suite detects accidental API or configuration breakage.

## Phase 1 — Portable foundation and minimal CLI

Reorganize the code without changing the user-facing dashboard.

### Project structure

- [x] Move executable startup code to `cmd/switchboard`.
- [x] Move configuration loading, validation, and atomic saving to
  `internal/config`.
- [x] Move HTTP routes and embedded frontend handling to `internal/server`.
- [x] Move Docker integration to `internal/docker`.
- [x] Introduce `internal/platform` interfaces for metrics and native services.
- [x] Place Linux implementations behind Linux build constraints.
- [x] Add compilable Windows and macOS platform implementations or explicit
  unsupported-capability responses.
- [x] Keep frontend assets embedded in the executable.

### CLI

- [x] Add `switchboard serve --config <path>`.
- [x] Add `switchboard version`.
- [x] Add `switchboard validate-config <path>`.
- [x] Preserve `switchboard -config <path>` compatibility.
- [x] Return useful non-zero exit codes and concise error messages.
- [x] Add command-level tests.

### Runtime quality

- [x] Add version, commit, build date, OS, and architecture build information.
- [x] Handle SIGINT/SIGTERM and shut down the HTTP server cleanly.
- [x] Apply sensible HTTP read, write, idle, and header timeouts.
- [x] Make all writable paths configurable and platform appropriate.
- [x] Ensure configuration writes remain validated and atomic.
- [x] Keep idle CPU and memory use appropriate for small Raspberry Pi systems.
- [x] Run `go test ./...`, `go vet ./...`, and frontend syntax checks in CI.
- [x] Cross-compile every target in the support matrix.

### Exit criteria

- [x] All three CLI commands work and are documented.
- [x] Linux, Windows, and macOS builds compile in CI.
- [x] Existing DietPi behavior and configuration remain compatible.

## Phase 2 — Secure Linux service and native packages

Deliver a production-quality Linux installation, beginning with systemd.

### Filesystem layout

- [x] Install the executable at `/usr/bin/switchboard` for native packages.
- [x] Store administrator configuration in `/etc/switchboard/config.yaml`.
- [x] Store persistent writable state in `/var/lib/switchboard`.
- [x] Confirm `/run/switchboard` is unnecessary while no runtime socket or
  transient file is created.
- [x] Log through stdout/stderr so systemd captures logs in the journal.
- [x] Define ownership and permissions for every installed path.

### Least-privilege security

- [x] Create a dedicated system user and group named `switchboard`.
- [x] Run the web server without root privileges.
- [x] Bind to loopback by default or clearly warn before exposing the dashboard.
- [x] Threat-model Docker socket access, which is effectively root-level access.
- [x] Do not silently add the service account to the Docker group.
- [x] Design an explicit opt-in mechanism for Docker control.
- [x] Design narrowly scoped privilege elevation for approved systemd operations.
- [x] Validate service/container identifiers before invoking control operations.
- [x] Avoid shell command construction; pass arguments directly to processes.
- [x] Add authentication guidance before supporting non-loopback access.
- [x] Harden the systemd unit where compatible, including filesystem and privilege
  restrictions.
- [x] Document the security consequences of every optional privileged feature.

### Packaging

- [x] Create a hardened systemd unit.
- [x] Create Debian packages for amd64, arm64, and armv7.
- [x] Create RPM packages for amd64, arm64, and armv7 where supported.
- [x] Create native Arch Linux packages for amd64, arm64, and armv7. The AUR
  `PKGBUILD` remains a Phase 5 distribution task.
- [x] Ensure installation creates required users, groups, directories, and files.
- [x] Ensure upgrades never overwrite an existing configuration.
- [x] Ensure uninstall stops and removes the service while preserving user data.
- [x] Provide an explicit purge path for configuration and state.
- [x] Test install, upgrade, uninstall, reinstall, and failed-upgrade behavior in
  clean virtual machines or containers.

### Exit criteria

- [x] DietPi/Raspberry Pi OS, Debian/Ubuntu, Fedora, and Arch installation paths are
  documented and tested.
- [x] Switchboard runs as a dedicated unprivileged account by default.
- [x] Docker and systemd control require deliberate, documented opt-in permissions.
- [x] Package upgrade preserves configuration and state.

## Phase 3 — Windows and macOS runtime support

Implement native behavior instead of merely producing compilable binaries.

### Windows

- [x] Collect CPU, memory, storage, temperature where available, uptime, network,
  hostname, OS, architecture, and local IP using supported Windows APIs.
- [x] Detect Docker Desktop/Engine and report a clear unavailable state otherwise.
- [x] Implement narrowly scoped Windows Service inspection/control.
- [x] Run Switchboard as a Windows Service under an appropriate service identity.
- [x] Store configuration and state in documented Windows locations.
- [x] Provide zip artifacts for amd64 and arm64.
- [x] Test installation and upgrades on supported Windows versions.

### macOS

- [x] Collect CPU, memory, storage, temperature where available, uptime, network,
  hostname, OS, architecture, and local IP using supported Darwin APIs.
- [x] Detect Docker Desktop and report a clear unavailable state otherwise.
- [x] Implement narrowly scoped launchd inspection/control.
- [x] Install Switchboard through launchd with appropriate permissions.
- [x] Store configuration and state in documented macOS locations.
- [x] Provide tar.gz artifacts for Intel and Apple Silicon.
- [x] Test installation and upgrades on supported macOS versions.

### Exit criteria

- [x] Windows and macOS show real native metrics.
- [x] Background-service installation, upgrade, and removal work on both platforms.
- [x] Missing Docker or unsupported metrics produce useful UI states, not failures.

## Phase 4 — Reproducible releases and one-line installers

Automate builds so users never need Go installed.

### Release automation

- [x] Build Linux amd64, arm64, and armv7 artifacts.
- [x] Build Windows amd64 and arm64 artifacts.
- [x] Build macOS amd64 and arm64 artifacts.
- [x] Build deb, rpm, and Arch package artifacts.
- [x] Produce SHA-256 checksums for every artifact.
- [x] Generate a software bill of materials where practical.
- [x] Sign release artifacts and document signature verification.
- [x] Attach artifacts and release notes to versioned GitHub releases.
- [x] Trigger releases from semantic-version tags such as `v1.0.0`.
- [x] Prevent publishing when tests or cross-platform builds fail.

### Install scripts

- [x] Create `install.sh` for Linux and macOS.
- [x] Create `install.ps1` for Windows.
- [x] Detect OS, architecture, available package format, and init system.
- [x] Download only from HTTPS release URLs.
- [x] Verify checksums/signatures before installing anything.
- [x] Install the native service using the platform-appropriate mechanism.
- [x] Preserve configuration during upgrades.
- [x] Support an explicit version argument for reproducible installation.
- [x] Provide uninstall instructions.
- [x] Keep scripts readable and safe to download and inspect before execution.
- [x] Publish both one-line and inspect-before-running instructions.

Planned commands:

```sh
curl -fsSL https://raw.githubusercontent.com/akash-kamat/switchboard/main/install.sh | sudo sh
```

```powershell
irm https://raw.githubusercontent.com/akash-kamat/switchboard/main/install.ps1 | iex
```

The GitHub raw URLs are canonical until a dedicated project domain is introduced.

### Exit criteria

- [x] A clean supported machine can install Switchboard without Go or Node.js.
- [x] Installers reject corrupt or unverifiable downloads.
- [x] Automated release artifacts produce identical behavior to native packages.

## Phase 5 — Package-manager distribution

**Current phase**

Publish existing signed release artifacts through native package ecosystems.

Recommended implementation order:

- [ ] Create a Homebrew tap for macOS and Linux.
- [ ] Create a Scoop bucket for Windows.
- [ ] Submit a Winget package.
- [ ] Publish Debian packages in an apt repository.
- [ ] Publish RPM packages in a yum/dnf repository.
- [ ] Publish and maintain an Arch AUR package.
- [ ] Publish a Chocolatey package.
- [ ] Automate package manifest updates after successful GitHub releases.
- [ ] Document package-manager-specific upgrades and uninstall behavior.

An npm package is intentionally deferred because Switchboard does not use Node.js.
It should only be reconsidered if there is a strong user need for npm as a binary
distribution wrapper.

### Exit criteria

- [ ] Homebrew, Scoop, and Winget installations are documented and tested.
- [ ] Linux repositories support normal package-manager upgrades.
- [ ] Package manifests are updated automatically or have a documented maintenance
  process.

## Phase 6 — Production readiness

- [x] Add `LICENSE`.
- [x] Add `CHANGELOG.md` and adopt semantic versioning.
- [x] Add contribution and security-reporting documentation.
- [x] Document every configuration field and environment override.
- [x] Document API routes and stability expectations.
- [x] Add backup, restore, migration, and disaster-recovery instructions.
- [x] Add dependency and vulnerability scanning.
- [x] Add integration tests for Docker and native service adapters.
- [x] Add package installation tests for every supported platform.
- [x] Establish a supported-version and security-update policy.
- [x] Complete a security review before declaring `v1.0.0` stable.

### Exit criteria

- [x] A new user can install, configure, update, troubleshoot, and uninstall
  Switchboard using only the documentation.
- [x] Release and security maintenance processes are documented and repeatable.

## Definition of done for each roadmap item

An item is complete only when:

- [ ] Its implementation is merged into the main branch.
- [ ] Automated tests cover the important behavior.
- [ ] Relevant platform builds pass.
- [ ] User-facing behavior is documented.
- [ ] Upgrade and backward-compatibility effects have been considered.
- [ ] Security implications have been reviewed.
