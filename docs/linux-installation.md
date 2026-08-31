# Manual Linux installation

These steps describe the current pre-package installation on a 64-bit ARM Linux
host such as DietPi. Native packages will automate them in Phase 2.

## Install

Copy the ARM64 executable, example configuration, and service unit to the host,
then run:

```sh
sudo install -m 0755 switchboard-linux-arm64 /usr/local/bin/switchboard
sudo install -d -m 0755 /etc/switchboard
sudo install -m 0644 config.example.yaml /etc/switchboard/config.yaml
sudo install -m 0644 switchboard.service /etc/systemd/system/switchboard.service
sudo systemctl daemon-reload
sudo systemctl enable --now switchboard
```

Do not replace `/etc/switchboard/config.yaml` when upgrading an existing install.
The current unit runs as root to access systemd and the Docker socket; read the
security warning in the main README before exposing the HTTP port.

## Verify

```sh
systemctl status switchboard
journalctl -u switchboard -n 100 --no-pager
curl --fail http://127.0.0.1:8080/api/system
curl --fail http://127.0.0.1:8080/api/services
```

## Upgrade and rollback

Keep the previous executable, stop the service, atomically replace the executable,
and start it again. If verification fails, restore the previous executable. Never
replace the configuration as part of an executable-only upgrade.

## Remove

```sh
sudo systemctl disable --now switchboard
sudo rm /etc/systemd/system/switchboard.service
sudo systemctl daemon-reload
sudo rm /usr/local/bin/switchboard
```

The removal steps intentionally preserve `/etc/switchboard/config.yaml`.
