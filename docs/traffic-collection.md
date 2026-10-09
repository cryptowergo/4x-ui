# Traffic collection limits and health

QueryStats has a bounded 32 MiB receive budget. Xray keeps user counters after
RemoveUser; retained counter names can exceed gRPC's default 4 MiB response limit
with few current clients. The higher budget restores collection, not counter GC.

Fetch failures are logged at warning level without peer identities. Optional
`XUI_TRAFFIC_METRICS_FILE` writes an atomic node-exporter textfile with:

- `v_agent_xray_stats_collection_success`;
- `v_agent_xray_stats_collection_last_attempt_timestamp_seconds`;
- `v_agent_xray_stats_collection_last_success_timestamp_seconds`;
- `v_agent_xray_stats_collection_failures_total` (process lifetime).

Set the variable to a file in the existing node-exporter textfile directory;
create the directory first. Without it, no metrics file is written. A successful
value older than 120 seconds is stale. These describe collection, not tunnel
availability. Existing database APIs may return stored values after a failed
collection; successful HTTP does not prove fresh traffic or online counts.

The current Xray StatsService offers read/reset, not counter removal. Reset only
zeros a value. PostgreSQL cleanup does not free Xray memory. Expired-peer GC is
still needed, but safe counter GC requires core support that accounts for live
references and in-flight traffic. Do not delete counters based only on age, zero
values or absence from the client table. No automatic restart or purge is added.

A controlled Xray restart clears retained counters and interrupts tunnels.
PostgreSQL totals remain. Failed reset-and-fetch responses may already have reset
core values; missing historical traffic cannot be reliably reconstructed.
Replacing x-ui normally requires restarting its child Xray too.

Run the focused checks with:

```sh
go test -tags integration ./xray -run TestStatsReceiveBudget
go test ./web/job -run TestTrafficCollection
```
