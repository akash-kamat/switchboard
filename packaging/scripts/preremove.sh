#!/bin/sh
set -eu

# Debian passes "remove" for final removal and RPM passes 0. Do not stop the
# service during an in-place package upgrade.
case "${1:-}" in
    remove|0)
        if command -v systemctl >/dev/null 2>&1; then
            systemctl disable --now switchboard.service || true
        fi
        ;;
esac
