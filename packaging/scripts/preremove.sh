#!/bin/sh
set -eu

# Debian passes "upgrade" during an upgrade and RPM passes 1. Arch passes the
# installed package version during final removal. Treat every other value as a
# final removal and preserve the administrator configuration in package-owned
# state so every package manager leaves it at its documented path.
case "${1:-}" in
    upgrade|1)
        ;;
    *)
        if [ -f /etc/switchboard/config.yaml ]; then
            install -d -o root -g root -m 0755 /var/lib/switchboard
            cp -p /etc/switchboard/config.yaml /var/lib/switchboard/.config-uninstall-backup
        fi
        if command -v systemctl >/dev/null 2>&1; then
            systemctl disable --now switchboard.service || true
        fi
        ;;
esac
