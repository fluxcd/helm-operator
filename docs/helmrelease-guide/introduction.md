---
title: Introduction
weight: 10
---

The intention of this guide is to give you more detailed information on various
elements of the `HelmRelease` Custom Resource. You can follow it in one take,
use it as a more explanatory reference, or a combination of both.

It assumes you have the Helm Operator already installed in your cluster. If
you have not done this yet, [follow the installation instructions from the
quickstart](../get-started/quickstart.md#install-the-helm-operator)

The guide tries not to presume a specific enabled Helm version but for the sake
of simplicity it was written while making use of Helm 3.

## A minimal HelmRelease

To get started, we will first create the most minimal `HelmRelease` possible.
We will use (parts) of this minimal `HelmRelease` as an example throughout the
rest of this guide.

```yaml
# podinfo.yaml
---
apiVersion: helm.fluxcd.io/v1
kind: HelmRelease
metadata:
  name: podinfo
  namespace: default
spec:
  chart:
    repository: https://stefanprodan.github.io/podinfo
    name: podinfo
    version: 3.2.0
```

The `spec.chart` object is the only mandatory property of the `HelmRelease`
and defines the Helm chart that should be installed by the Helm Operator. This
`HelmRelease` will manage [`stefanprodan/podinfo`](https://github.com/stefanprodan/podinfo)

> a tiny web application made with Go that showcases best practices of running
> microservices in Kubernetes

from a [Helm repository chart source](chart-sources.md#helm-repositories).

Applying this to the cluster and making the Helm Operator do the release is
equivalent to a human running the following `helm` commands:

```console
$ helm repo add podinfo https://stefanprodan.github.io/podinfo
$ helm upgrade -i default-podinfo podinfo/podinfo --version '3.2.0'
```

A couple of differences between the `HelmRelease` resource and listed `helm`
commands should stand out:

1. The `HelmRelease` does not make use of a repository alias, instead the
   absolute URL of the Helm repository is used.<br><br>
   This is to ensure the `HelmRelease` is able to stand on its own. If we used
   names in the spec, which were resolved to URLs elsewhere (e.g., in a
   `repositories.yaml` supplied to the operator), it would be possible to
   change the meaning of a `HelmRelease` without altering it. This is
   undesirable because it makes it hard to specify exactly what you want, in
   the one place; or to read exactly what is being specified, in the one place.
1. The `spec.chart.version` is mandatory.<br><br>
   The reasoning behind this is the same as the explanation given above.
1. The default release name used by the Helm Operator is composed from the
  `metadata.namespace` and the `metadata.name` of the `HelmRelease` resource
  (`<namespace>-<name>`).<br><br>
   This to ensure it does not collide with other `HelmRelease` resources that
   may have the same `.metadata.name` but different namespace definitions.

Having absorbed this information, you can now go ahead and apply the resource
using `kubectl`:

```console
$ kubectl apply -f podinfo.yaml
```

Take a look at the logs of the Helm Operator deployment and you will see the
Helm installation was performed almost instantly after applying the resource:

```console
$ kubectl logs deploy/flux-helm-operator
...
ts=2020-01-01T12:00:00.000000000Z caller=release.go:335 component=release release=default-podinfo targetNamespace=default resource=default:helmrelease/podinfo helmVersion=v3 info="no existing release" action=install
ts=2020-01-01T12:00:00.000000000Z caller=helm.go:69 component=helm version=v3 info="creating 2
resource(s)" targetNamespace=default release=default-podinfo
ts=2020-01-01T12:00:00.000000000Z caller=release.go:266 component=release release=default-podinfo targetNamespace=default resource=default:helmrelease/podinfo helmVersion=v3 info="Helm release sync succeeded" revision=3.2.0
$ kubectl get helmrelease
NAME      RELEASE           STATUS     MESSAGE                       AGE
podinfo   default-podinfo   deployed   Helm release sync succeeded   10s
```

As expected, the release will now also show up for e.g. `helm list`:

```console
$ helm list
NAME            NAMESPACE       REVISION        UPDATED                                 STATUS          CHART           APP VERSION
default-podinfo default         1               2020-01-01 12:00:00.000000000 +0000 UTC deployed        podinfo-3.2.0   3.2.0
```

Congratulations! You made your first Helm release using a `HelmRelease`
resource.
