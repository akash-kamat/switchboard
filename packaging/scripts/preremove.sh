#!/bin/sh
set -eu

# Debian passes "remove" for final removal, RPM passes 0, and Arch package
# hooks do not pass an argument. Preserve the administrator configuration in
# package-owned state so every package manager leaves it at its documented path.
case "${1:-}" in
    remove|0|"")
        if [ -f /etc/switchboard/config.yaml ]; then
            install -d -o switchboard -g switchboard -m 0750 /var/lib/switchboard
            cp -p /etc/switchboard/config.yaml /var/lib/switchboard/.config-uninstall-backup
        fi
        if command -v systemctl >/dev/null 2>&1; then
            systemctl disable --now switchboard.service || true
        fi
        ;;
esac
