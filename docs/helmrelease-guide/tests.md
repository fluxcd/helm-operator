---
weight: 70
title: Tests
---

[Helm tests](https://helm.sh/docs/topics/chart_tests/) are a useful validation
mechanism for Helm Releases, and thus are supported by the Helm Operator.

## Enabling tests

When tests for a `HelmRelease` are enabled, the Helm Operator will run them
after any successful installation or upgrade attempt. In the case of a test
failure, the prior installation or upgrade will be treated as failed, resulting
in the release being purged or rolled back [if enabled](rollbacks.md#enabling-rollbacks).

Tests can be enabled by setting `.test.enable`:

```yaml
spec:
  test:
    enable: true
```

## Wait interaction

When tests are enabled, [resource waiting](release-configuration.md#wait-for-resources-to-be-ready)
defaults to `true` since this is likely needed for test pre-conditions to be satisfied.

## Uninstall or rollback release on test failure

The `spec.test.ignoreFailures` allows the `HelmRelease` to be left in a released state if the tests fail.
Setting `ignoreFailures` to `false` will automatically uninstall or rollback the `HelmRelease` if any of the tests fail.
If the tests are ignored, the `Released` condition will be left as `true` and `Tested` will be `false`.

```yaml
spec:
  test:
    enable: true
    ignoreFailures: false
```

## Test timeout

Test timeout can be set via the `.test.timeout` option.

```yaml
spec:
  test:
    enable: true
    timeout: 600
```

It is defined as the time to wait for any individual Kubernetes operation during
tests in seconds. Defaults to `300` when omitted.

## Helm 2 vs 3

In Helm 3, test-specific funcationality was removed in favor of a generic `test`
hook no different than any other hook. The Helm Operator takes this into account
as detailed below.

### Test cleanup

Helm 3 removed the `helm test --cleanup` flag in favor of [hook delete policies](https://helm.sh/docs/topics/chart_tests/#notes).
For `HelmRelease`s targeting Helm 2, test cleanup is enabled by default since
upgrades are highly likely to cause test pod name conflicts without it.
This flag currently deletes test pods immediately after they are run, but the only
guarantee is that tests are cleaned up before running a subsequent test for the
same `HelmRelease`, as delaying the deletion would allow time to debug failures,
and thus may be implemented in the future. Test cleanup can be disabled by setting
`.test.cleanup` to `false`.

```yaml
spec:
  helmVersion: v2
  test:
    enable: true
    cleanup: false
```

### Test parallelism

Helm 2 supported `helm test --parallel --max 10` to run tests in parallel. Helm 3 will
likely [expand this functionality to all hooks](https://github.com/helm/helm/issues/7763). Once the Helm 3 implementation is available
this can be integrated into the Helm Operator, and translated into the equivalent
Helm 2 options for test parallelism as well.
