---
title: Monitoring
weight: 40
---

The Helm Operator exposes a metrics endpoint at `/metrics`  on the configured
[`--listen`](operator.md#general-flags) address (defaults to `:3030`) with data
in Prometheus format.

## Metrics

| Metric | Description
|--------|---
| `release_count` | Count of releases managed by the operator. |
| `release_action_duration_seconds` | Duration of release sync actions in seconds. See [release actions](#release-actions). |
| `release_condition_info` | Release condition status gauge, see [release conditions](#release-conditions).
| `release_queue_length_count` | Count of release jobs waiting in the queue to be processed. |

### Release actions

`release_action_duration_seconds` supports the following labels and label values.

#### Labels

| Label              | Label Value |
|--------------------|---
| `target_namespace` | `targetNamespace` of `HelmRelease`
| `release_name`     | `releaseName` of `HelmRelease`
| `success`          | Whether the action was successful (`true` or `false`)
| `action`           | The release action, see below.

#### Actions

| Action            | Description
|-------------------|---
| `sync`            | One entire release sync attempt, as configured to occur once every [--charts-sync-interval](operator.md#reconciliation-configuration)
| `install`         | Installation attempt
| `upgrade`         | Upgrade attempt
| `rollback`        | Rollback attempt
| `uninstall`       | Uninstallation attempt
| `dry-run-compare` | Dry run compare attempt to [determine whether to upgrade](../helmrelease-guide/reconciliation-and-upgrades.md#what-triggers-an-upgrade)
| `annotate`        | [Annotation](../helmrelease-guide/reconciliation-and-upgrades.md#the-antecedent-annotation) attempt

### Release conditions

`release_condition_info` supports the following labels and label values.

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
expr: flux_helm_operator_release_condition_info{condition="Released"} == -1
```
