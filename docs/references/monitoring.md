# Monitoring

The Helm Operator exposes a metrics endpoint at `/metrics`  on the configured
[`--listen`](operator.md#general-flags) address (defaults to `:3030`) with data
in Prometheus format.

The following metrics are exposed:

| Metric | Description
|--------|---
| `release_duration_seconds` | Release duration in seconds.
| `release_queue_length_count` | Count of releases waiting in the queue to be processed.
