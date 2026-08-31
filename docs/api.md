# HTTP API contract

The API is currently unversioned and intended for the bundled dashboard. Breaking
changes will be called out in release notes until a versioned public API is added.
Every response is JSON except embedded frontend assets. Errors use
`{"error":"message"}`.

| Method and path | Success | Purpose |
| --- | --- | --- |
| `GET /api/services` | `200` array | Service definitions, state, autostart, CPU, RAM, and optional error. |
| `POST /api/services/{name}/start` | `200` | Start a configured service. |
| `POST /api/services/{name}/stop` | `200` | Stop a configured service. |
| `POST /api/services/{name}/restart` | `200` | Restart a configured service. |
| `POST /api/services/{name}/autostart` | `200` | Set startup behavior using `{"enabled":true}`. |
| `GET /api/system` | `200` object | System CPU, memory, swap, disk, temperature, load, uptime, and identity. |
| `GET /api/config` | `200` object | Parsed configuration and normalized YAML. |
| `POST /api/config/validate` | `200` object | Validate YAML without writing it. |
| `PUT /api/config` | `200` object | Validate, atomically save, and activate a config object or YAML string. |

Service names are URL path encoded. A missing service returns `404`; a backend
failure returns `502`; malformed input returns `400`; valid but unacceptable
configuration returns `422`. Mutating cross-origin browser requests return `403`.

`PUT /api/config` accepts exactly one of `{"config":{...}}` or
`{"yaml":"..."}`. Request bodies are size-limited and unknown JSON fields are
rejected.
