---
title: Get started using Kustomize
linkTitle: Using Kustomize
weight: 40
---

This guide walks you through setting up the Helm Operator using
[Kustomize](https://kustomize.io).

## Prerequisites

- Kubernetes cluster **>=1.13.0**
- `kustomize` **>=3.2.0**
- Some knowledge of Kustomize
- _(Optional)_
  Tiller [(secure setup)](https://v2.helm.sh/docs/securing_installation/)

## Prepare the manifests for installation

Create a directory called `helm-operator`:

```sh
mkdir helm-operator
```

Create a `kustomization.yaml` file and use the [Helm Operator
deployment YAMLs](https://github.com/fluxcd/helm-operator/tree/1.4.4/deploy)
as a base:

```sh
cat > helm-operator/kustomization.yaml <<EOF
bases:
- github.com/fluxcd/helm-operator//deploy
patchesJSON6902:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: helm-operator
    namespace: flux
  patch: |-
    - op: replace
      path: /spec/template/spec/containers/0/args
      value:
        - --enabled-helm-versions=v3
EOF
```

The `patchesJSON6902` target ensures only support for Helm 3 is
enabled, to also enable support for Helm 2 and connect to Tiller,
continue to read [Helm 2](#helm-2) below.

{{% alert color="info" title="Tip" %}}
If you want to install a specific Helm Operator release, add the
version number to the base URL:

```yaml
bases:
  - github.com/fluxcd/helm-operator//deploy?ref=1.4.4
```
{{% /alert %}}

### Helm 2

To also enable support for Helm 2 and configure the Tiller settings, we
need to make slight adjustment to the `patchesJSON6902` target.

First, make sure [your Tiller installation is
secure](https://v2.helm.sh/docs/securing_installation/), and add a
`secretGenerator` entry of type `kubernetes.io/tls` for the client
certificates:


```yaml
# helm-operator/kustomization.yaml
namespace: flux   # ensures secret is generated in the right namespace
bases:
- github.com/fluxcd/helm-operator//deploy
secretGenerator:
- name: tiller-tls-cert
  type: kubernetes.io/tls
  files:
  - tls.crt
  - tls.key
patchesJSON6902:
...
```

Create a patch file for the Helm Operator to mount the `tiller-tls-cert`
secret:

```sh
cat > helm-operator/patch-tiller-tls.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: helm-operator
  namespace: flux
spec:
  template:
    spec:
      volumes:
        - name: tiller-tls-cert
          secret:
            secretName: tiller-tls-cert
            defaultMode: 0400
      containers:
        - name: helm-operator
          volumeMounts:
          - name: tiller-tls-cert
            mountPath: /etc/fluxd/helm
            readOnly: true
EOF
```

Adapt your `kustomization.yaml` to include the patch:

```yaml
# helm-operator/kustomization.yaml
...
patchesStrategicMerge:
- patch-tiller-tls.yaml
```

Add (or replace) `v2` to `--enabled-helm-versions` and configure the
required [Tiller option flags](../references/operator.md#tiller-configuration)
for your setup:

```yaml
# helm-operator/kustomization.yaml
...
patchesJSON6902:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: helm-operator
    namespace: flux
  patch: |-
    - op: replace
      path: /spec/template/spec/containers/0/args
      value:
        - --enabled-helm-versions=v2,v3     # enables Helm 2
        - --tiller-namespace=kube-system    # defines the Tiller namespace
        - --tiller-tls-enable=true          # enables TLS communication with Tiller
...
```
    
## Install the Helm Operator using Kustomize

Deploy the Helm Operator to the cluster by applying the `helm-operator`
folder with the `kustomization.yaml` file in it onto the cluster:

```sh
kustomize build helm-operator | kubectl apply -f -
```

Confirm the Helm Operator deployed successfully to the default `flux`
namespace:

```sh
kubectl -n flux rollout status deployment/helm-operator
```

## Customize the Helm Operator deployment

### Configure a namespace

By default the Helm Operator is installed in the `flux` namespace when
making use of the published deployment YAMLs as a base. It is possible
to override this default namespace by creating a custom namespace
definition and configuring a `namespace` in your `kustomization.yaml`
file.

Create a custom namespace definition, this example uses `team-ns`:

```sh
cat > helm-operator/namespace.yaml <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: team-ns
EOF
```

Create a patch to remove the default namespace from the base:

```sh
cat > helm-operator/patch-default-ns.yaml <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: flux
\$patch: delete
EOF
```

Adapt your `kustomization.yaml` file to to include your own namespace
resource and the patch file, and define the `namespace`:

```yaml
# helm-operator/kustomization.yaml
namespace: team-ns
resources:
- namespace.yaml
bases:
- github.com/fluxcd/helm-operator//deploy
patchesJSON6902:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: helm-operator
    namespace: flux
  patch: |-
    - op: replace
      path: /spec/template/spec/containers/0/args
      value:
        - --enabled-helm-versions=v3
patchesStrategicMerge:
- patch-default-ns.yaml
```

Apply the `helm-operator` folder with the `kustomization.yaml` file
in it onto the cluster:

```sh
kubectl apply -k helm-operator
```

Confirm the Helm Operator deployed successfully to the `teamn-ns`
namespace:

```sh
kubectl -n team-ns rollout status deployment/helm-operator
```

### Provide Helm repository credentials

To mount a custom `repositories.yaml` file to e.g. provide credentials
to a Helm chart repository [as described in the `HelmRelease`
guide](../helmrelease-guide/chart-sources.md#authentication-and-certificates),
you can instruct Kustomize to create a secret from the
`repositories.yaml` file.

First, create the `repositores.yaml` file with the credentials:

```sh
cat > helm-operator/repositories.yaml <<EOF
apiVersion: ""
generated: "0001-01-01T00:00:00Z"
repositories:
- caFile: ""
  certFile: ""
  keyFile: ""
  name: private-repository
  url: https://charts.example.com
  password: john
  username: s3cr3t!
EOF
```

Create a patch file for the Helm Operator to mount the `flux-helm-repositories`
secret:

```sh
cat > helm-operator/patch-repositories-mount.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: helm-operator
  namespace: flux
spec:
  template:
    spec:
      volumes:
        - name: repositories-yaml
          secret:
            secretName: flux-helm-repositories
            defaultMode: 0400
      containers:
        - name: helm-operator
          volumeMounts:
            - name: repositories-yaml
              mountPath: /root/.helm/repository/repositories.yaml
              readOnly: true
EOF
```

Adapt the `kustomization.yaml` file to instruct it to generate the
secret and apply the patch file:

```yaml
# helm-operator/kustomization.yaml
namespace: flux         # ensures secret is generated in the right namespace                            
bases:
 - github.com/fluxcd/helm-operator//deploy
patchesJSON6902:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: helm-operator
    namespace: flux
  patch: |-
    - op: replace
      path: /spec/template/spec/containers/0/args
      value:
        - --enabled-helm-versions=v3
secretGenerator:
- name: flux-helm-repositories
  files:
  - repositories.yaml
patchesStrategicMerge:
- patch-repositories-mount.yaml
```

Apply the `helm-operator` folder with the `kustomization.yaml` file
in it onto the cluster:

```sh
kustomize build helm-operator | kubectl apply -f -
```

## Next

- Learn all about the available configuration options in the [operator
  reference](../references/operator.md).
- Continue learning about `HelmRelease` resources [in the
  guide](../helmrelease-guide/introduction.md).
