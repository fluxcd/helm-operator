---
title: Release configuration
weight: 40
---

When making use of the `helm` binary you can pass along various flags while
making a release that influence how a release is performed. Most of these flags
are also available as parameters in the `spec` of the `HelmRelease` resource,
others are not available for `helm` but fine-tune how the Helm Operator itself
functions.

This section of the guide documents the parameters, and the effect they have
when the Helm Operator performs a release for the `HelmRelease` they are
configured on.

## Targeting a Helm version

When you install the Helm Operator with multiple Helm versions enabled, the
lowest enabled version is targeted by default. To target another version or to
ensure it does not accidentally target the wrong version due to a configuration
mistake, you can set the `.helmVersion`:

```yaml
spec:
  helmVersion: v3
```

## Migrating from Helm v2 to v3

Helm Operator uses the [helm-2to3 plugin](https://github.com/helm/helm-2to3)
under the hood to migrate HelmRelease objects.
In order to perform a release conversion you have to:

1. Set `.spec.helmVersion` to `v3`
2. Add a migrate annotation `helm.fluxcd.io/migrate: "true"` (For testing, you
can set the value to "dry-run" instead of "true")

```yaml
apiVersion: helm.fluxcd.io/v1
kind: HelmRelease
metadata:
  name: redis
  annotations:
    helm.fluxcd.io/migrate: "true" # add annotation
spec:
  helmVersion: v3 # set helmVersion to v3
  releaseName: redis
  chart:
    repository: https://kubernetes-charts.storage.googleapis.com
    name: redis
```

3. Assuming you will be deleting the tiller deployment using gitops, consider
setting the operator flags `--convert-tiller-out-cluster=true` and
`--convert-release-storage=configmaps`. If tiller is in a custom namespace, make
sure you set `--tiller-namespace=` flag as well.

After applying the new HelmRelease, the operator will take care of deleting the
old v2 release that Tiller managed and converting it to the v3 format. Once
you're satisfied with the migration, you can go ahead and remove the
annotation. This approach allows teams to migrate their charts at scale to
the v3 format without stopping the world.

## Configuring a target namespace

It is possible to target a different namespace for the release than the
`HelmRelease` lives in using `.targetNamespace`. This can come in handy when
you have to deploy releases into a namespace that is also managed by another
team running their own namespace-restricted Helm Operator in this same
namespace:

```yaml
spec:
  targetNamespace: team-a
```

## Specifying a release name

The default release name used by the Helm Operator is a composition of the
following values:

```text
<namespace>-[<target namespace>-]<name>
```
This format was invented for Helm 2 to ensure release names from
`HelmRelease` resources in different namespaces would never accidentally
collide with each if they would have the same name, as release names
[were not scoped in this version like they are in Helm 3](
https://helm.sh/docs/faq/#release-names-are-now-scoped-to-the-namespace),
and it still serves this purpose when a target namespace is defined.

In some situations you may want to overwrite this generated release name, for
example because you want to take over a release made with `helm`. This is
possible by declaring a `.releaseName` which will replace the generated format:

```yaml
spec:
  releaseName: podinfo
```

## Forcing resource updates

When a chart contains a breaking upgrade, you may need to force resource updates
through the replacement strategy of Helm, this is possible by setting
`.forceUpgrade`:

```yaml
spec:
  forceUpgrade: true
```

## Reusing values during upgrade

Due to the declarative behaviour of the Helm Operator it resets all values by
default, so that the only configuration being applied is what is defined in the
`HelmRelease` resource. It is possible to disable this behaviour, and make it
reuse values from the previous release by explicitly stating that values should
not be reset:

```yaml
spec:
  resetValues: false
```

## Configuring the timeout

To configure how many seconds Helm should wait for any individual Kubernetes operations
you can set `.timeout`, the default is `300`:

```yaml
spec:
  timeout: 300
```

{{% alert color="warning" title="Warning" %}}
When your chart requires a high non-default `timeout` value it is advised
to increase the `terminationGracePeriod` on the Helm Operator pod to not
end up with a release in a faulty state due to the operator receiving a
`SIGKILL` signal during an upgrade.
{{% /alert %}}

## Wait for resources to be ready

{{% alert color="info" title="Note" %}}
When you have many `HelmRelease` resources with the waiting enabled,
it is likely that you want to increase the amount of `--workers` to
ensure other releases are still processed, given the waiting process blocks
the worker from processing other `HelmRelease` resources.
{{% /alert %}}

By default, Helm will mark a release as successfully deployed as soon as the
changes have been applied to the cluster. To instruct Helm to wait until
all resources are in ready state before marking the release as successful
you can set `.wait`, or enable [tests](tests.md#enabling-tests) or [rollbacks](rollbacks.md#enabling-rollbacks) which has the same effect.
When set, it will wait for as long as [`.timeout`](#configuring-the-timeout):

```yaml
spec:
  wait: true
```

## Configuring the max number of revision saved

{{% alert color="info" title="Note" %}}
Setting this only has effect for a `HelmRelease` targeting Helm 3.
You can configure a global history limit for Helm 2 by [passing `--history-max
<int>` to `helm init`](https://v2.helm.sh/docs/using_helm/#initialize-helm-and-install-tiller).
{{% /alert %}}

To configure the maximum number of revision saved by Helm for a `HelmRelease`,
you can set `.maxHistory`. Use `0` for an unlimited number of revisions;
defaults to `10`:

```yaml
spec:
  maxHistory: 10
```
