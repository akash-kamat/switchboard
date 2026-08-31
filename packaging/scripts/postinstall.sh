#!/bin/sh
set -eu

if [ -f /etc/switchboard/config.yaml ]; then
    chown root:switchboard /etc/switchboard/config.yaml
    chmod 0660 /etc/switchboard/config.yaml
fi

if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload || true
    systemctl enable switchboard.service || true
    if systemctl is-active --quiet switchboard.service; then
        systemctl restart switchboard.service || true
    else
        systemctl start switchboard.service || true
    fi
fi
