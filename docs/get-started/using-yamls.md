# Get started using YAMLs

> **🛑 Upgrade Advisory**
>
> This documentation is for Helm Operator (v1) which has [reached its end-of-life in November 2022](https://fluxcd.io/blog/2022/10/september-2022-update/#flux-legacy-v1-retirement-plan).
>
> We strongly recommend you familiarise yourself with the newest Flux and [migrate as soon as possible](https://fluxcd.io/flux/migration/).
>
> For documentation regarding the latest Flux, please refer to [this section](https://fluxcd.io/flux/).

This guide walks you through setting up the Helm Operator using
[deployment YAMLs](https://github.com/fluxcd/helm-operator/tree/1.4.4/deploy).

## Prerequisites

- Kubernetes cluster **>=1.1.3.0**
- `kubectl`
- _(Optional)_
  Tiller [(secure setup)](https://v2.helm.sh/docs/securing_installation/)

## Install the Helm Operator

First, install the `HelmRelease` Custom Resource Definition. By adding this CRD
it will be possible to define `HelmRelease` resources on the cluster:

```sh
kubectl apply -f https://raw.githubusercontent.com/fluxcd/helm-operator/1.4.4/deploy/crds.yaml
```

Proceed to create the `flux` namespace, this is the namespace the Helm Operator
will be deployed to:

```sh
kubectl create namespace flux
```

Apply the `ServiceAccount`, `ClusterRole` and `ClusterRoleBinding` so that the
Helm Operator can access cluster resources:

```sh
kubectl apply -f https://raw.githubusercontent.com/fluxcd/helm-operator/1.4.4/deploy/rbac.yaml
```

Apply the Helm Operator deployment itself:

```sh
kubectl deploy -f https://raw.githubusercontent.com/fluxcd/helm-operator/1.4.4/deploy/deployment.yaml
```

### Helm 3

The default deployment of the Helm Operator comes with support for Helm 2 and 3
enabled. To disable support for Helm 2 (and recover from the Tiller connection
failure), patch the resource to set `--enabled-helm-versions=v3`:

```
kubectl deploy -f https://raw.githubusercontent.com/fluxcd/helm-operator/1.4.4/deploy/deployment.yaml \
    --type='json' \
    -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value":"--enabled-helm-versions=v3"}]'
```

### Helm 2

The default deployment of the Helm Operator does enable support for Helm 2 but
does not take any custom configurations like Tiller TLS settings into account.
If your Tiller is e.g. in a different namespace than `kube-system` or
[securely setup](https://v2.helm.sh/docs/securing_installation/), take a look
at the available [Tiller configuration flags](../references/operator.md#tiller-configuration)
and [commented out sections](https://github.com/fluxcd/helm-operator/blob/1.4.4/deploy/deployment.yaml)
in the example deployment to further tweak your Helm Operator installation.

## Next

- Learn all about the available configuration options in the [operator
  reference](../references/operator.md).
- Continue learning about `HelmRelease` resources [in the
  guide](../helmrelease-guide/introduction.md).
