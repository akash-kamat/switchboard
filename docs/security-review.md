# Security review checklist

This checklist is required before declaring v1.0 stable.

- [x] Configuration writes are size-limited, strictly parsed, validated, and atomic.
- [x] Service and container identifiers reject shell syntax and unsafe characters.
- [x] Child processes receive separate arguments; no user value is evaluated by a shell.
- [x] Cross-origin mutating API calls are rejected and browser security headers are set.
- [x] Linux packages run as an unprivileged dedicated identity with systemd hardening.
- [x] Windows uses `LocalService`; macOS uses a controlled system LaunchDaemon.
- [x] Docker access is opt-in and documented as root-equivalent.
- [x] Release checksums, keyless signatures, and an SBOM are generated automatically.
- [x] Dependency and CodeQL scans run on a schedule and relevant pull requests.
- [ ] Independent review of authentication/network exposure before a v1.0 release.
- [ ] Independent review of installer and update trust boundaries before v1.0.

The unchecked independent reviews deliberately block a stable v1.0 declaration;
they do not block preview releases.
