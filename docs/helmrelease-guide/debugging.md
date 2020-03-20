# Debugging

Even after having read everything this guide has to offer it is possible that a
`HelmRelease` fails and you want to debug it to get to the cause. This may be
a bit harder than you were used to while working with just `helm` because you
are no longer in direct controlm but the Helm Operator is doing the work for
you.

This last section of the guide will give you some pointers on how to debug a
failing `HelmRelease` resource.

## Getting the reason of failure

If a release fails the reason of failure will be logged in the Helm Operator's
logs _and_ recorded as a condition on the `HelmRelease` resource. You can view
this condition by describing the `HelmRelease` resource using `kubectl`:

```console
kubectl describe -n <namespace> helmrelease/<name>
```

In case of a release upgrade failure, the error as returned by Helm will be
recorded in the message of `HelmUpgradeFailed`.

## Manually performing a release to debug

When describing the `HelmRelease` and the logs did not give any clues, it may
help to perform the release manually using the same values as specified in the
`HelmRelease` resource. When no `.valuesFrom` are defined, this can be done
by making use of [`jq`](https://stedolan.github.io/jq/) and `kubectl`:

```console
kubectl get -n <namespace> helmrelease/<name> -o json | jq .spec.values |  helm upgrade -i <release name> -f - <chart>
```

## Getting help

Invite yourself to the [CNCF community](https://slack.cncf.io) Slack and ask
a question on the [`#flux` channel](https://cloud-native.slack.com/archives/flux).
