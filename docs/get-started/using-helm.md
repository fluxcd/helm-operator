# Get started using Helm

This guide walks you through setting up the Helm Operator using the
available [Helm chart](https://github.com/fluxcd/helm-operator/tree/master/chart/helm-operator).

## Prerequisites

- Kubernetes cluster **>=1.13.0**
- Up-to-date **Helm 2 or 3** [`helm` binary](https://github.com/helm/helm/releases)
- _(Optional)_
  Tiller [(secure setup)](https://v2.helm.sh/docs/securing_installation/)

## Install the Helm Operator chart

Install the `HelmRelease` Custom Resource Definition. By adding this
CRD it will be possible to define `HelmRelease` resources on the
cluster:

```sh
kubectl apply -f https://raw.githubusercontent.com/fluxcd/helm-operator/{{ version }}/deploy/crds.yaml
```

Using `helm`, add the Flux CD Helm repository:

```sh
helm repo add fluxcd https://charts.fluxcd.io
```

Install the Helm Operator using the available Helm chart:

```sh
helm upgrade -i helm-operator fluxcd/helm-operator \
    --set helm.versions=v3
```

Confirm the Helm Operator deployed successfully:

```console
$ helm status helm-operator
NAME: helm-operator
LAST DEPLOYED: Wed Jan 01 12:00:00 2020
NAMESPACE: default
STATUS: deployed
REVISION: 1
...
```

```console
$ kubectl get pods
NAME                             READY   STATUS    RESTARTS   AGE
helm-operator-6985656995-dpmdl   1/1     Running   0          31s
```

!!! note
    This installs the Helm Operator with only support for Helm 3
    enabled, to also enable support for Helm 2 and connect to Tiller, read
    [with Tiller](#with-tiller-helm-2) below. 

### With Tiller (Helm 2)

Make sure [your Tiller installation is
secure](https://v2.helm.sh/docs/securing_installation/), and create a
secret for the client certificates:

```sh
kubectl create secret tls helm-client-certs \
    --cert=cert.pem \
    --key=key.pem
```

Install (or upgrade) the Helm Operator with the [Tiller
configuration](../references/operator.md#tiller-configuration)
while also enabling Helm 2 support:

```sh
helm upgrade -i helm-operator fluxcd/helm-operator \
    --set tillerNamespace=kube-system \
    --set tls.enable=true \
    --set helm.versions="v2\,v3"
```

Confirm the Helm Operator deployed successfully and connected to
Tiller:

```console
$ kubectl get pods
NAME                             READY   STATUS    RESTARTS   AGE
helm-operator-7cc7c798cc-kn26w   1/1     Running   0          18s
```

```console
$ kubectl logs -f deploy/helm-operator
...
ts=2020-01-01T12:00:00.556712443Z caller=helm.go:71 component=helm version=v2 info="connected to Tiller" version="sem_ver:\"v2.16.3\" git_commit:\"1ee0254c86d4ed6887327dabed7aa7da29d7eb0d\" git_tree_state:\"clean\" " host=tiller-deploy.kube-system:44134 options="{Host: Port: Namespace:kube-system TLSVerify:false TLSEnable:true TLSKey:/etc/fluxd/helm/tls.key TLSCert:/etc/fluxd/helm/tls.crt TLSCACert: TLSHostname:}"
```

!!! tip
    Targeting a specific Helm version in a `HelmRelease` is possible by
    defining the `spec.helmVersion`, e.g: `helmVersion: v3` to target
    Helm 3. Read more about this in '[targeting a Helm
    version](../helmrelease-guide/release-configuration.md#targeting-a-helm-version)'
    in the `HelmRelease` guide.

## Next

- Learn all about the available configuration options in the [chart
  documentation](../references/chart.md#configuration)
  and [operator reference](../references/operator.md).
- Continue learning about `HelmRelease` resources [in the
  guide](../helmrelease-guide/introduction.md).