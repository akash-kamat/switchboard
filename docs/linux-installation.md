# Manual Linux installation

These steps describe the current pre-package installation on a 64-bit ARM Linux
host such as DietPi. Native packages will automate them in Phase 2.

## Install

Copy the ARM64 executable, example configuration, and service unit to the host,
then run:

```sh
sudo groupadd --system switchboard 2>/dev/null || true
sudo useradd --system --gid switchboard --home-dir /var/lib/switchboard --shell /usr/sbin/nologin switchboard 2>/dev/null || true
sudo install -m 0755 switchboard-linux-arm64 /usr/bin/switchboard
sudo install -d -o root -g switchboard -m 0750 /etc/switchboard
sudo install -d -o switchboard -g switchboard -m 0750 /var/lib/switchboard
sudo install -o root -g switchboard -m 0660 config.example.yaml /etc/switchboard/config.yaml
sudo install -m 0644 switchboard.service /etc/systemd/system/switchboard.service
sudo systemctl daemon-reload
sudo systemctl enable --now switchboard
```

Do not replace `/etc/switchboard/config.yaml` when upgrading an existing install.
The unit runs without root privileges. Read the security guide before deliberately
granting Docker access or exposing the HTTP port.

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
sudo rm /usr/bin/switchboard
```

The removal steps intentionally preserve `/etc/switchboard/config.yaml`.

Native package removal also preserves the `switchboard` account, configuration,
and `/var/lib/switchboard`. To explicitly purge that retained data after removal:

```sh
sudo rm -r /etc/switchboard /var/lib/switchboard
sudo userdel switchboard
sudo groupdel switchboard 2>/dev/null || true
```

Back up the configuration before purging; these commands are irreversible.
