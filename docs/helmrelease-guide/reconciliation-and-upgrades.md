---
weight: 50
title: Reconciliation and upgrades
---

Now that you know the ins and outs of configuring a release, we are going to
have a better look at how the Helm Operator performs the actual Helm release.

## Reconciliation

On the creation and update of a `HelmRelease` the resource is queued for
reconciliation. Besides this all `HelmRelease` resources handled by the Helm
operator instance are also queued for reconciliation every
[`--charts-sync-interval`](../references/operator.md) (defaults to 3
minutes).

Once the queued resource has been picked up by a worker, the Helm Operator
attempts to receive the chart for the resource and performs several [safe guard
checks](#what-triggers-an-upgrade); if those do not result in an error or
instruct to return early, the Helm installation or upgrade is performed.

## What triggers an upgrade

To prevent spurious upgrades from happening the Helm Operator performs several
safe guard checks before performing the actual upgrade. Below you will find an
overview of the checks it performs, and what effect they have.

If any of the following equals to `false`, the sync process will exit with an
error and no upgrade will be performed:

1. The values composed of the merged `.valuesFrom` and `.values` are valid
   YAML.
1. The resources of the Helm release are [marked as being managed by the
   `HelmRelease`](#the-antecedent-annotation).
1. The current state of the Helm release is `deployed`.

The first of the following that equals to `true` will result in an upgrade
being performed, otherwise no action is taken:

1. No Helm release exists in the Helm storage for the `HelmRelease`.
1. This generation of the `HelmRelease` has not been processed before —
   the generation changes for example when the `.spec` is edited.
1. The result of a dry-run upgrade for the `HelmRelease` differs from the
   latest release in the Helm storage.

{{% alert color="info" title="Note" %}}
Mutations to live cluster-state are not detected and thus not
reverted. This will however be added in the foreseeable future.
{{% /alert %}}

## Upgrade failures

When an upgrade fails, the Helm Operator will stop performing upgrades for the
release as it cannot assume this is a safe procedure, nor does it
automatically perform [a rollback](rollbacks.md). Instead it will start logging
warnings about the `failed` status of the release.

Recovering from this is possible, after having inspected the state of the
release, by getting the Helm release manually in a `deployed` state, for
example by performing a rollback or upgrade for the release using `helm`:

```console
helm rollback <release name>
```

## The antecedent annotation

Right after the Helm Operator performs a Helm release for the
`HelmRelease` resource, all resources that were the result of this release
are annotated with the antecedent annotation `helm.fluxcd.io/antecedent`,
the value of the annotation equals to `<namespace>:helmrelease/<name>`.

The purpose of this annotation is to indicate that the cause of that resource
is a `HelmRelease`. It also functions as a safe guard during reconciliation
to ensure the release is only managed by a single `HelmRelease`, as it is
possible that due to a misconfiguration multiple `HelmRelease` resources exist
with the same `.releaseName` set.
