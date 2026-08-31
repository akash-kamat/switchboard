# Backup, restore, and recovery

The executable is replaceable; configuration and optional state are the data to
protect. Stop Switchboard before taking a consistent backup.

| Platform | Configuration | Persistent state |
| --- | --- | --- |
| Linux | `/etc/switchboard/config.yaml` | `/var/lib/switchboard` |
| Windows | `%ProgramData%\Switchboard\config.yaml` | `%ProgramData%\Switchboard` |
| macOS | `/Library/Application Support/Switchboard/config.yaml` | same directory |

Copy the relevant paths to protected storage and retain their permissions. The
configuration can also be checked offline with `switchboard validate-config
<path>`.

To restore, reinstall the same or newer Switchboard version without starting it,
restore the saved files, validate the configuration, and start the native
service. Schema migrations are performed only when explicitly documented; schema
version 1 is never rewritten merely by starting the program.

If an upgrade fails, reinstall the previous version with `install.sh --version
vX.Y.Z` or `install.ps1 -Version vX.Y.Z`, restore the backup if necessary, and
inspect systemd journal, Windows Event/Service logs, or the macOS log files. A
normal uninstall intentionally preserves configuration. Delete it only as an
explicit purge after verifying the backup.
