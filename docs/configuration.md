# Configuration contract

Switchboard uses one YAML document. The current schema version is `1`. A file
without `version` is treated as version 1 for compatibility with installations
created before schema versioning was introduced. An explicit unsupported version
is rejected instead of being guessed or partially loaded.

Future schema changes follow these rules:

- Additive changes keep the current version when old and new readers remain safe.
- Breaking changes increment `version` and include an explicit migration.
- Migrations create a backup, validate the complete result, and replace the active
  file atomically only after validation succeeds.
- Switchboard never silently downgrades a newer configuration.

## Top-level fields

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `version` | integer | `1` | Configuration schema version. |
| `listen` | string | `:8080` | HTTP listen address. |
| `dashboard` | object | see below | Dashboard presentation and refresh settings. |
| `services` | list | empty | Managed Docker containers and systemd units. |

`dashboard.refresh_seconds` accepts 5 through 3600. `theme` accepts `light`,
`dark`, or `system`. `background` is a six-digit CSS hex color. `overview` accepts
`cpu`, `memory`, `storage`, `temperature`, `load`, and `swap`. `system_details`
accepts `hostname`, `local_ip`, `os`, `uptime`, `kernel`, and `architecture`.

Each service requires a unique `name` and a `type` of `docker` or `native`.
The older `systemd` type remains accepted for compatibility. Docker entries
require `container`; native entries require `unit` (a systemd unit, Windows
Service name, or launchd label). Optional
fields are `icon`, `href`, `description`, `group`, and `autostart`.

Unknown fields, duplicate service names, unsafe URLs, invalid choices, and missing
required fields are rejected. UI saves are validated and written using a temporary
file plus atomic rename; a failed validation leaves the original file untouched.
