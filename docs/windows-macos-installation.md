# Windows and macOS installation

Switchboard uses native host metrics and native background-service managers on
both platforms. Docker is optional. A missing Docker Desktop pipe/socket is
reported on the affected card without stopping the dashboard.

## Windows

Open PowerShell as Administrator and inspect the installer before running it:

```powershell
irm https://raw.githubusercontent.com/akash-kamat/switchboard/main/install.ps1 -OutFile install.ps1
Get-Content .\install.ps1
.\install.ps1
```

The installer verifies the release SHA-256 checksum, installs to
`%ProgramFiles%\Switchboard`, stores persistent configuration in
`%ProgramData%\Switchboard\config.yaml`, and registers an automatic Windows
Service running as `LocalService`. Upgrades preserve the configuration.

```powershell
Get-Service Switchboard
Get-Content "$env:ProgramData\Switchboard\config.yaml"
.\install.ps1 -Version v1.2.3
.\install.ps1 -Uninstall
```

Windows service entries use `type: native` and the Service Control Manager name
in `unit`. Managing another service requires the Switchboard service identity to
have the corresponding SCM permission. Docker Desktop is reached through
`//./pipe/docker_engine`.

## macOS

Download and inspect the shell installer, then run it as root:

```sh
curl -fsSL https://raw.githubusercontent.com/akash-kamat/switchboard/main/install.sh -o install.sh
less install.sh
sudo sh install.sh
```

The binary is installed at `/usr/local/bin/switchboard`, configuration at
`/Library/Application Support/Switchboard/config.yaml`, and the launch daemon at
`/Library/LaunchDaemons/com.akash-kamat.switchboard.plist`.

```sh
sudo launchctl print system/com.akash-kamat.switchboard
sudo sh install.sh --version v1.2.3
sudo sh install.sh --uninstall
```

launchd entries use `type: native` and the launchd label in `unit`. Docker
Desktop uses `/var/run/docker.sock`; override `--docker-socket` if your Docker
installation exposes another socket.

CPU, memory, storage, uptime, OS, kernel, hostname, architecture, and local IP
are collected natively. Temperature is shown as unavailable when the operating
system does not expose a stable, permission-safe API for it.
