# Security model

Switchboard is designed as a trusted-machine appliance. Native installers run it
with the platform's system administrator identity so Docker and native-service
controls work immediately: `root` on Linux and macOS launchd installs, and
`LocalSystem` on Windows. Homebrew services run as the user who starts them.

This convenience is intentionally privileged. Anyone who can reach Switchboard's
HTTP API can operate configured services and containers. Docker control is
effectively root access because Docker can mount the host filesystem and create
privileged containers.

Switchboard has no built-in login in the current release. Do not port-forward it
to the Internet. Keep it on a trusted LAN or private VPN, or bind it to loopback
and place it behind an authenticated HTTPS reverse proxy.

Configuration validation restricts Docker identifiers and native-service names
to safe identifier character sets. Native commands use argument arrays rather
than a shell, Docker names are URL-escaped, YAML writes are validated and atomic,
and the Linux systemd unit retains filesystem and process hardening compatible
with its administrator-level integrations.
