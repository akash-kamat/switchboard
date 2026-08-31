# Security model

Switchboard exposes controls that can start and stop workloads. It has no built-in
login in the current release, so the HTTP listener must stay on a trusted network.
The server logs a prominent warning whenever `listen` is not loopback. For remote
access, bind to `127.0.0.1:8080` and use an authenticated HTTPS reverse proxy or a
private VPN.

## Default package privileges

Native packages run Switchboard as the dedicated `switchboard` system account.
That account owns `/var/lib/switchboard`; the `switchboard` group can update
`/etc/switchboard/config.yaml` so validated UI saves work. The service receives no
Linux capabilities and the systemd sandbox makes the rest of the filesystem
read-only.

Docker and systemd mutation are unavailable by default. Status queries may still
work when the operating system permits them.

## Docker opt-in

Access to `/var/run/docker.sock` is effectively root access: Docker can mount the
host filesystem and create privileged containers. Packages never add the service
account to the `docker` group automatically. An administrator may deliberately
enable it with:

```sh
sudo usermod -aG docker switchboard
sudo systemctl restart switchboard
```

Remove that access with `sudo gpasswd -d switchboard docker` and restart the
service. Do not enable Docker control for an Internet-exposed dashboard.

## systemd control

The packaged service has no general root or sudo access. A future privileged
helper will expose only configured units and start/stop/restart/enable/disable
operations. Do not grant passwordless unrestricted `systemctl` or shell access to
the service account. Until that helper exists, systemd mutations can fail with a
permission error while read-only status remains available.

Configuration validation restricts Docker identifiers and systemd unit names to
their safe identifier character sets. Native commands are executed with an
argument array, never through a shell. Docker names are URL-escaped before being
sent to the Engine API.
