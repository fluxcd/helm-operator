# Monitoring

The Helm Operator exposes a metrics endpoint at `/metrics`  on the configured
[`--listen`](operator.md#general-flags) address (defaults to `:3030`) with data
in Prometheus format.

## Metrics

| Metric | Description
|--------|---
| `release_count` | Count of releases managed by the operator. |
| `release_duration_seconds` | Release synchronization duration in seconds. This duration includes one or many `release_phase_durations`. |
| `release_phase_duration_seconds` | Release phase synchronization duration in seconds. |
| `release_phase_info` | The (negative) integer equaling the current phase of a release. Negative values are failed phases, `0` equals to unknown. See [release phases](#release-phases).
| `release_queue_length_count` | Count of release jobs waiting in the queue to be processed. |


### Release phases

The following is a table of the values the `release_phase_info` metric exposes,
and the phase they represent:

| Value | Phase |
|-------|---
| `-4`  | `ChartFetchFailed`
| `-3`  | `Failed`
| `-2`  | `RollbackFailed`
| `-1 ` | `RolledBack`
| `0`   | `Unknown`
| `1`   | `RollingBack`
| `2`   | `Installing`
| `3`   | `Upgrading`
| `4`   | `ChartFetched`
| `5`   | `Succeeded`

## Prometheus alert rules examples

The following is a list of Prometheus alert rules examples possible
with the exposed metrics. We are open to [pull requests](
https://github.com/fluxcd/helm-operator/pulls) adding additional rules.

### Low queue throughput

```yaml
alert: HelmOperatorLowThroughput
expr: flux_helm_operator_release_queue_length_count > 0
for: 30m
```

### Automatic rollback of `HelmRelease`

```yaml
alert: HelmReleaseRolledBack
expr: flux_helm_operator_release_phase_info == -1
```

### `HelmRelease` subject to an error

```yaml
alert: HelmReleaseError
expr: flux_helm_operator_release_phase_info < -1
```
