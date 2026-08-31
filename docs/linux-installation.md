# Linux installation

Switchboard release builds provide native Debian, RPM, and Arch packages for
amd64, arm64, and 32-bit ARMv7. DietPi and Raspberry Pi OS use the Debian
package. Download the package matching the machine from the GitHub release.

Check the architecture first:

```sh
uname -m
dpkg --print-architecture 2>/dev/null || true
```

Typical mappings are `x86_64` to `amd64`, `aarch64` to `arm64`, and `armv7l` to
`armv7` (`armhf` in Debian metadata).

## Native packages

On DietPi, Raspberry Pi OS, Debian, or Ubuntu:

```sh
sudo apt install ./switchboard_VERSION_linux_arm64.deb
```

On Fedora or another RPM-based distribution:

```sh
sudo dnf install ./switchboard_VERSION_linux_arm64.rpm
```

On Arch Linux:

```sh
sudo pacman -U ./switchboard_VERSION_linux_arm64.pkg.tar.zst
```

Replace `arm64` with `amd64` or `armv7` as appropriate. The package creates the
unprivileged `switchboard` account, installs and enables the hardened systemd
unit, creates the configuration and state directories, and starts the service.
An existing configuration is preserved during upgrades, package removal, and
reinstallation.

Release packages are generated from `.goreleaser.yaml`. Before the first tagged
release, maintainers can download the `switchboard-snapshot` artifact from a
successful GitHub Actions CI run.

## Manual installation

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

Upgrade a native package with the same `apt install`, `dnf install`, or
`pacman -U` command used for installation. Package upgrades retain the existing
configuration and state.

Keep the previous executable, stop the service, atomically replace the executable,
and start it again when performing a manual upgrade. If verification fails,
restore the previous executable. Never replace the configuration as part of an
executable-only upgrade.

## Remove

Use the native package manager where applicable:

```sh
sudo apt remove switchboard       # Debian, Ubuntu, DietPi, Raspberry Pi OS
sudo dnf remove switchboard       # Fedora/RPM
sudo pacman -R switchboard        # Arch Linux
```

For a manual installation, disable the service and remove the unit and binary.

```sh
sudo systemctl disable --now switchboard
sudo rm /etc/systemd/system/switchboard.service
sudo systemctl daemon-reload
sudo rm /usr/bin/switchboard
```

The removal steps intentionally preserve `/etc/switchboard/config.yaml`.

Package removal preserves the `switchboard` account, configuration, and
`/var/lib/switchboard`. To explicitly purge that retained data after removal:

```sh
sudo rm -r /etc/switchboard /var/lib/switchboard
sudo userdel switchboard
sudo groupdel switchboard 2>/dev/null || true
```

Back up the configuration before purging; these commands are irreversible.
