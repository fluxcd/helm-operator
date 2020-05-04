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
| `release_condition_info` | Release condition status gauge, see [release conditions](#release-conditions).
| `release_queue_length_count` | Count of release jobs waiting in the queue to be processed. |

### Release conditions

#### Labels

| Label              | Label Value |
|--------------------|---
| `target_namespace` | `targetNamespace` of `HelmRelease`
| `release_name`     | `releaseName` of `HelmRelease`
| `condition`        | [condition type](helmrelease-custom-resource.md#helm.fluxcd.io/v1.HelmReleaseConditionType)

#### Values

Values represent the [condition status](helmrelease-custom-resource.md#helm.fluxcd.io/v1.ConditionStatus).

| Value | Condition Status |
|-------|---
| `-1`  | `False`
| `0`   | `Unknown`
| `1`   | `True`

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
expr: flux_helm_operator_release_condition_info{condition="RolledBack"} == 1
```

### `HelmRelease` subject to an error

```yaml
alert: HelmReleaseError
expr: flux_helm_operator_release_phase_info{condition="Released"} == -1
```
