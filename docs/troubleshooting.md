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
- A native-service permission error on a current native installation usually
  means an older service definition is still installed. Reinstall Switchboard
  and confirm the service uses `root` on Linux or `LocalSystem` on Windows.
- If the page loads but data does not, request `/api/system` and `/api/services`
  directly and inspect their JSON error messages.
- If a settings save fails, use the YAML validation endpoint or
  `validate-config`; the original file remains untouched.
- If default port 8080 is occupied, Switchboard automatically selects and saves
  the first free port from 8081 through 8099. Check `listen` in the active YAML
  or the service log for the selected address. An occupied custom port fails
  visibly so an administrator's explicit choice is never silently replaced.

For rollback and restoration see [backup-recovery.md](backup-recovery.md). Include
the output of `switchboard version`, the operating system, architecture, relevant
service logs, and a redacted configuration in bug reports.
