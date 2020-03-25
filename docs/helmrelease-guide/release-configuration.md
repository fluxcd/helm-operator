# Release configuration

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

## Resetting values during upgrade

To ignore any previous used values for a release and make Helm use the values
as they are in the `values.yaml` file of the Helm chart used, you can set
`.resetValue`. You will likely want to use this when you have removed
overriding `.values` from your `HelmRelease`:

```yaml
spec:
  resetValues: true
```

## Configuring the timeout

To configure how many seconds Helm should wait for any individual Kubernetes operations
you can set `.timeout`, the default is `300`:

```yaml
spec:
  timeout: 300
```

!!! warning
    When your chart requires a high non-default `timeout` value it is advised
    to increase the `teriminationGracePeriod` on the Helm Operator pod to not
    end up with a release in a faulty state due to the operator receiving a
    `SIGKILL` signal during an upgrade.

## Wait for resources to be ready

!!! note
    When you have many `HelmRelease` resources with the `.wait` flag
    set, it is likely that you want to increase the amount of `--workers` to
    ensure other releases are still processed, given the waiting process blocks
    the worker from processing other `HelmRelease` resources.

Normally Helm will mark a release as successfully deployed as soon as the
changes have been applied to the cluster. To instruct Helm to wait until
all resources are in ready state before marking the release as successful
you can set `.wait`. When set, it will wait for as long as
[`.timeout`](#configuring-the-timeout):

```yaml
spec:
  wait: true
```

## Configuring the max number of revision saved

!!! note
    Setting this only has effect for a `HelmRelease` targeting Helm 3.
    You can configure a global history limit for Helm 2 by [passing `--history-max
    <int>` to `helm init`](https://v2.helm.sh/docs/using_helm/#initialize-helm-and-install-tiller).

To configure the maximum number of revision saved by Helm for a `HelmRelease`,
you can set `.maxHistory`. Use `0` for an unlimited number of revisions;
defaults to `10`:

```yaml
spec:
  maxHistory: 10
```
