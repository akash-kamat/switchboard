#!/bin/sh
set -eu

if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload || true
fi

backup=/var/lib/switchboard/.config-uninstall-backup
if [ -f "$backup" ]; then
    if [ ! -f /etc/switchboard/config.yaml ]; then
        install -d -o root -g root -m 0755 /etc/switchboard
        cp -p "$backup" /etc/switchboard/config.yaml
    fi
    rm -f "$backup"
fi

# Configuration and state are intentionally preserved.
# See docs/linux-installation.md for the explicit purge procedure.
