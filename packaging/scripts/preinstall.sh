#!/bin/sh
set -eu

if ! getent group switchboard >/dev/null 2>&1; then
    groupadd --system switchboard
fi

if ! id switchboard >/dev/null 2>&1; then
    useradd --system --gid switchboard --home-dir /var/lib/switchboard \
        --shell /usr/sbin/nologin --comment "Switchboard service" switchboard
fi

install -d -o root -g switchboard -m 0750 /etc/switchboard
install -d -o switchboard -g switchboard -m 0750 /var/lib/switchboard
