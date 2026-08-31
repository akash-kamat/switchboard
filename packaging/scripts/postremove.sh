#!/bin/sh
set -eu

if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload || true
fi

# Configuration, state, and the service account are intentionally preserved.
# See docs/linux-installation.md for the explicit purge procedure.
