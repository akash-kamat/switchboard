# Runtime resource budget

Switchboard is intended for small single-board computers. The Phase 1 budget is:

- Idle resident memory (RSS): less than 25 MiB on a 64-bit Raspberry Pi host.
- Idle CPU: less than 1% averaged over a minute when no dashboard client is polling.
- No background metrics polling in the server; collection is request-driven.

The current ARM64 DietPi deployment was observed at approximately 11–13 MiB RSS
and 0.2–0.7% CPU while serving the dashboard. Results vary with Go version,
configured services, requests, and kernel accounting.

Measure the actual Switchboard process rather than a parent `sudo` process:

```sh
pid=$(systemctl show switchboard --property=MainPID --value)
ps -p "$pid" -o pid,%cpu,%mem,rss,vsz,etime,cmd
```

For a longer CPU sample:

```sh
pidstat -p "$pid" 5 12
```
