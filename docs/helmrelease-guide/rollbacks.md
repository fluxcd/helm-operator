---
weight: 60
title: Rollbacks
---

From time to time a release made by the Helm Operator may fail, this section
of the guide will explain how you can recover from a failed release by enabling
rollbacks.

{{% alert color="warning" title="Caution" %}}
Rollbacks of Helm charts containing `StatefulSet` resources can be a
tricky operation, and are one of the main reasons automated rollbacks are not
enabled by default. Verify a manual rollback (using `helm`) of your Helm
chart does not cause any problems before enabling it.
{{% /alert %}}

## Enabling rollbacks

When rollbacks for a `HelmRelease` are enabled, the Helm Operator will detect
a faulty upgrade, including post-upgrade helm test [if enabled](tests.md#enabling-tests)
failures, and instruct Helm to perform a rollback, it will not attempt a new
upgrade unless it detects a change in values and/or the chart, or
[retries have been enabled](#enabling-retries-of-rolled-back-releases). Changes
are detected by comparing the failed release to a fresh dry-run release.

Rollbacks can be enabled by setting `.rollback.enable`:

```yaml
spec:
  rollback:
    enable: true
```

## Wait interaction

When rollbacks are enabled, [resource waiting](release-configuration.md#wait-for-resources-to-be-ready)
defaults to `true` since this is necessary to validate whether the release should
be rolled back or not.

## Tweaking the rollback configuration

To get more fine-grained control over how the rollback is performed by Helm,
the `.rollback` of the `HelmRelease` resources offers a couple of additional
settings.

```yaml
spec:
  rollback:
    enable: true
    disableHooks: false
    force: false
    recreate: false
    timeout: 300
```

The definition of the listed keys is as follows:

* `enable`: Enables the performance of a rollback when a release fails.
* `disableHooks` _(Optional)_: When set to `true`, prevent hooks from running
  during rollback. Defaults to `false` when omitted.
* `force` _(Optional)_: When set to `true`, force resource update through
  delete/recreate if needed. Defaults to `false` when omitted.
* `recreate` _(Optional)_: When set to `true`, performs pods restart for the
  resource if applicable. Defaults to `false` when omitted.
* `timeout` _(Optional)_: Time to wait for any individual Kubernetes operation
  during rollback in seconds. Defaults to `300` when omitted.

{{% alert color="warning" title="Warning" %}}
When your chart requires a high non-default `timeout` value it is advised
to increase the `terminationGracePeriod` on the Helm Operator pod to not
end up with a release in a faulty state due to the operator receiving a
`SIGKILL` signal during an upgrade.
{{% /alert %}}

## Enabling retries of rolled back releases

Sometimes the cause of an upgrade failure may be transient. To guard yourself
against this it is possible to instruct the Helm Operator to retry the upgrade
of a rolled back release by setting `.rollback.retry` to `true`. This will
cause the Helm Operator to retry the upgrade until the `.rollback.maxRetries`
is reached:

```yaml
spec:
  rollback:
    enable: true
    retry: true
    maxRetries: 5
```

The definition of the listed keys is as follows:

* `enable`: Enables the performance of a rollback when a release fails.
* `retry` _(Optional)_: When set to `true`, retries the upgrade of a failed
  release until `maxRetries` is reached. Defaults to `false` when omitted.
* `maxRetries` _(Optional)_: The maximum amount of retries that should be
  attempted for a rolled back release. Defaults to `5` when omitted, use `0`
  for an unlimited amount of retries.
