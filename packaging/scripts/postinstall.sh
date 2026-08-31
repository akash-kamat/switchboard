#!/bin/sh
set -eu

if [ -f /etc/switchboard/config.yaml ]; then
    chown root:root /etc/switchboard/config.yaml
    chmod 0644 /etc/switchboard/config.yaml
fi

chown root:root /var/lib/switchboard
chmod 0755 /var/lib/switchboard

if command -v systemctl >/dev/null 2>&1 && [ -d /run/systemd/system ]; then
    systemctl daemon-reload
    systemctl enable switchboard.service
    if systemctl is-active --quiet switchboard.service; then
        systemctl restart switchboard.service
    else
        systemctl start switchboard.service
    fi
    sleep 1
    listen_address=$(awk '/^[[:space:]]*listen:/ {sub(/^[^:]*:[[:space:]]*/, ""); gsub(/"/, ""); print; exit}' /etc/switchboard/config.yaml)
    echo "Switchboard listen address: ${listen_address:-:8080}"
fi
