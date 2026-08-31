# Changelog

Switchboard follows [Semantic Versioning](https://semver.org/). Changes that are
not yet released are recorded under `Unreleased`; tagged versions are immutable.

## Unreleased

## 0.1.1 - 2026-08-31

- Build releases with Go 1.25.13 or newer to include current standard-library
  security fixes.
- Wait for the Windows service to stop before replacing or removing its binary.
- Allow Docker's configured graceful restart period to finish.
- Run distribution generation and installer validation independently after a
  successful release.

## 0.1.0 - 2026-08-31

- Added native Windows and macOS metrics and service management.
- Added verified one-line installers and reproducible signed release automation.
- Added Debian, RPM, Arch, Windows, and macOS packaging support.
- Added the portable CLI, validated configuration editor, and dashboard UI.
