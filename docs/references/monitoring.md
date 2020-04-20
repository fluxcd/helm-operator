# Monitoring

The Helm Operator exposes a metrics endpoint at `/metrics`  on the configured
[`--listen`](operator.md#general-flags) address (defaults to `:3030`) with data
in Prometheus format.

The following metrics are exposed:

| Metric | Description
|--------|---
| `release_duration_seconds` | Release synchronization duration in seconds. This duration includes one or many `release_phase_durations`. |
| `release_phase_duration_seconds` | Release phase synchronization duration in seconds. |
| `release_queue_length_count` | Count of release jobs waiting in the queue to be processed. |
