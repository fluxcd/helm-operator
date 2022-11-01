---
weight: 70
title: Frequently asked questions
linkTitle: FAQ
---

### A `HelmRelease` is stuck in a `failed` release state. How do I force it through?

You need to manually perform a `helm rollback` or `helm upgrade` to get it in
a healthy `deployed` state again before the Helm Operator will continue.
Read more about this in [the `HelmRelease`
guide](helmrelease-guide/reconciliation-and-upgrades.md#upgrade-failures).

### A `HelmRelease` reports a `deployed` release status but also an error. What is going on?

The release status recorded in the `status` sub-resource of the `HelmRelease`
is the status as given by Helm for the release. It is polled from Helm on a
interval (see
[`--status-update-interval`](references/operator.md#reconciliation-configuration)).

The message is the error reported during the latest release failure. When this
failure was for example due to a validation error, the upgrade has not been
completed, and the status of the release will thus still be `deployed`.

### I want to automate chart releases from Helm chart repositories. Is this possible?

This is possible by wrapping the chart from the Helm chart repository in
an umbrella chart sourced from a Git source, [read more about this
here](helmrelease-guide/automation.md#helm-repository-chart-updates).

### It takes a long time before the operator processes an update to a resource. How can I speed up the processing of releases?

The operator watches for changes of Custom Resources of kind `HelmRelease`. It
receives Kubernetes Events and queues these to be processed. All resources will
also be re-queued every [`--charts-sync-interval`
(default `3m`)](references/operator.md#reconciliation-configuration) for a
dry-run to detect and revert manual changes made to a release.

Depending on how many resources the operator is watching and the complexity of
the charts (umbrella charts for example generally take a longer time to
process), the default number of workers may not be sufficient to instantly
process a release the moment it enters the queue as the queue works on a
[FIFO](https://en.wikipedia.org/wiki/FIFO_(computing_and_electronics)) basis,
and other releases may have to be processed first.

The solution is to increase the number of workers processing the releases using
the `--workers` flag (default `2`), i.e. by steps of `2` until the releases are
processed within a time frame that feels right to you.

If this does not give the desired effect, or if the number of workers required
is unacceptably high, there are two other tweaks possible:

1. increasing the `--charts-sync-interval`; this causes the queue to be less
   heavy occupied at the cost of detecting mutations less rapidly
1. using multiple Helm Operator instances, i.e. by having one operator per
   namespace; namespace scoping is possible by configuring the
   `--allow-namespace` flag
   
### The Helm Operator is taking up too many resources. How can I limit the resource usage?

When you are running the Helm Operator on an edge, home lab, or other resource
constrained cluster it may take up too many resources for your liking.

1. In case of a high number of `HelmRelease` resources, think about a namespace
   restricted multi-tenant setup as an option to spread the load and have finer
   grain control over configuration.
1. Increase the [various polling intervals](references/operator.md) based on
   what your environment really needs:

       - `--status-update-interval`: The interval at which Helm is polled for
         the current state of the release, generally safe to increase.
       - `--charts-sync-interval`: The interval at which `HelmRelease` resources
         are re-queued for reconciliation. Increasing this interval will result
         in mutations to a chart release being detected slower.
       - `--git-poll-interval`: The interval at which Git repositories are
         polled for changes for charts from Git sources.

### I have manually deleted a Helm release using `helm`. Why is the operator not able to restore it?

When you delete a Helm release with `helm delete <name>` using Helm 2, the
release name can not be re-used as the history of the old release is still
stored in the Helm storage under the same name. You need to use the
`helm delete --purge` option, only then the release history will be removed
from the Helm storage and the Helm Operator will then be able to reinstall the
release.

### I am using SSL between Helm and Tiller. How can I configure the operator to use the certificate?

When installing the Helm Operator, you can supply the CA and client-side
certificate using the `tls` options on the chart. More details about this
on:

- [Get started using Helm](get-started/using-helm.md#with-tiller-helm-2)
- [Get started using Kustomize](get-started/using-kustomize.md#prepare-the-manifests-for-installation)
- [Get started using YAMLs](get-started/using-yamls.md#helm-2)

### I am using Flux and have deleted a `HelmRelease` file from Git. Why is the Helm release still running on my cluster?

Flux does not delete resources by default; you can enable Flux garbage
collection feature by passing the command-line flag
`--sync-garbage-collection` to `fluxd`.

With Flux garbage collection enabled, the Helm Operator will receive the delete
event and will purge the Helm release.

### Are there prerelease builds I can run?

There are builds from CI for each merge to master branch. See
[fluxcd/helm-operator-prerelease](https://hub.docker.com/r/fluxcd/helm-operator-prerelease/tags).
