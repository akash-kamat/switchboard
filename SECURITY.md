# Security policy

## Supported versions

Until v1.0, only the latest tagged release receives security fixes. After v1.0,
the latest minor release and the previous minor release receive fixes for at
least six months. Operating-system support follows each vendor's supported
release lifecycle.

## Reporting a vulnerability

Do not open a public issue. Use GitHub's **Security > Report a vulnerability**
private reporting form for this repository. Include the affected version,
platform, impact, reproduction steps, and any suggested mitigation. Expect an
acknowledgement within seven days and a status update within fourteen days.

## Security boundaries

Switchboard does not provide authentication. Its HTTP API can start and stop
configured workloads, so it must not be exposed directly to an untrusted
network. Docker socket access is equivalent to host-root authority. Native
service actions are limited by the identity and permissions assigned by the
operating system. See [docs/security.md](docs/security.md) for deployment rules.
