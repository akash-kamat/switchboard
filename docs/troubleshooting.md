# Troubleshooting

Start by validating the active configuration and checking the native service:

```sh
switchboard validate-config /etc/switchboard/config.yaml
systemctl status switchboard
journalctl -u switchboard -n 100 --no-pager
curl http://127.0.0.1:8080/api/system
```

On Windows use `Get-Service Switchboard` and inspect
`%ProgramData%\Switchboard\config.yaml`. On macOS use `sudo launchctl print
system/com.akash-kamat.switchboard` and inspect `/Library/Logs/Switchboard*.log`.

- A Docker card error normally means the socket/pipe is absent, Docker is not
  running, the configured container name is wrong, or the service identity lacks
  access. Check the actual name with `docker ps -a --format '{{.Names}}'`.
- A native-service permission error means Switchboard's unprivileged identity is
  not authorized to control that unit/service/label. Do not grant unrestricted
  administrator or root access.
- If the page loads but data does not, request `/api/system` and `/api/services`
  directly and inspect their JSON error messages.
- If a settings save fails, use the YAML validation endpoint or
  `validate-config`; the original file remains untouched.
- If port 8080 is occupied, change `listen` and restart Switchboard.

For rollback and restoration see [backup-recovery.md](backup-recovery.md). Include
the output of `switchboard version`, the operating system, architecture, relevant
service logs, and a redacted configuration in bug reports.
