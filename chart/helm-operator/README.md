# Flux Helm Operator

This chart bootstraps a [Helm Operator](https://github.com/fluxcd/helm-operator) deployment on
a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

* Kubernetes >= v1.11
* Tiller >= v2.14

## Installation

Add the fluxcd repo:

```sh
helm repo add fluxcd https://charts.fluxcd.io
```

Install the HelmRelease CRD:

```sh
kubectl apply -f https://raw.githubusercontent.com/fluxcd/helm-operator/prepare-helm-op/deploy/flux-helm-release-crd.yaml
```

Install the chart with the release name `helm-operator`:

```sh
helm install --wait --name helm-operator \
--namespace flux \
fluxcd/helm-operator
```

#### To install with a private git host:

When using a private Git server to host your charts, setting the `git.ssh.known_hosts` variable
is required for enabling successful key matches because `StrictHostKeyChecking` is enabled during git pull operations.

By setting the `git.ssh.known_hosts` variable, a configmap will be created
called `helm-operator-ssh-config` which in turn will be mounted into a volume named
`sshdir` at `/root/.ssh/known_hosts`.

* Get the known hosts keys by running the following command:

```sh
ssh-keyscan <your_git_host_domain>
```

* Copy the known hosts keys into a temporary file `/tmp/flux_known_hosts` and install Helm Operator:

```sh
helm install --name helm-operator \
--set-file git.ssh.known_hosts=/tmp/flux_known_hosts \
--namespace flux \
chart/helm-operator
```

The [configuration](#configuration) section lists all the parameters that can be configured during installation.

## Uninstall

To uninstall/delete the `helm-operator` deployment:

```sh
helm delete --purge helm-operator
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following tables lists the configurable parameters of the Flux chart and their default values.

| Parameter                                         | Default                                              | Description
| -----------------------------------------------   | ---------------------------------------------------- | ---
| `image.repository`                                | `docker.io/fluxcd/helm-operator`                     | Image repository
| `image.tag`                                       | `<VERSION>`                                          | Image tag
| `replicaCount`                                    | `1`                                                  | Number of Helm Operator pods to deploy, more than one is not desirable.
| `image.pullPolicy`                                | `IfNotPresent`                                       | Image pull policy
| `image.pullSecret`                                | `None`                                               | Image pull secret
| `resources.requests.cpu`                          | `50m`                                                | CPU resource requests for the deployment
| `resources.requests.memory`                       | `64Mi`                                               | Memory resource requests for the deployment
| `resources.limits`                                | `None`                                               | CPU/memory resource limits for the deployment
| `nodeSelector`                                    | `{}`                                                 | Node Selector properties for the deployment
| `tolerations`                                     | `[]`                                                 | Tolerations properties for the deployment
| `affinity`                                        | `{}`                                                 | Affinity properties for the deployment
| `extraEnvs`                                       | `[]`                                                 | Extra environment variables for the Helm Operator pod(s)
| `rbac.create`                                     | `true`                                               | If `true`, create and use RBAC resources
| `rbac.pspEnabled`                                 | `false`                                              | If `true`, create and use a restricted pod security policy for Helm Operator pod(s)
| `serviceAccount.create`                           | `true`                                               | If `true`, create a new service account
| `serviceAccount.name`                             | `flux`                                               | Service account to be used
| `clusterRole.create`                              | `true`                                               | If `false`, Helm Operator will be restricted to the namespace where is deployed
| `createCRD`                                       | `true`                                               | Create the HelmRelease CRD
| `updateChartDeps`                                 | `true`                                               | Update dependencies for charts
| `git.pollInterval`                                | `git.pollInterval`                                   | Period on which to poll git chart sources for changes
| `git.timeout`                                     | `git.timeout`                                        | Duration after which git operations time out
| `git.secretName`                                  | `None`                                               | The name of the kubernetes secret with the SSH private key, supercedes `git.secretName`
| `git.ssh.known_hosts`                             | `None`                                               | The contents of an SSH `known_hosts` file, if you need to supply host key(s)
| `chartsSyncInterval`                              | `3m`                                                 | Interval at which to check for changed charts
| `workers`                                         | `None`                                               | (Experimental) amount of workers processing releases
| `logFormat`                                       | `fmt`                                                | Log format (fmt or json)
| `logReleaseDiffs`                                 | `false`                                              | Helm operator should log the diff when a chart release diverges (possibly insecure)
| `allowNamespace`                                  | `None`                                               | If set, this limits the scope to a single namespace. If not specified, all namespaces will be watched
| `tillerNamespace`                                 | `kube-system`                                        | Namespace in which the Tiller server can be found
| `tls.enable`                                      | `false`                                              | Enable TLS for communicating with Tiller
| `tls.verify`                                      | `false`                                              | Verify the Tiller certificate, also enables TLS when set to true
| `tls.secretName`                                  | `helm-client-certs`                                  | Name of the secret containing the TLS client certificates for communicating with Tiller
| `tls.keyFile`                                     | `tls.key`                                            | Name of the key file within the k8s secret
| `tls.certFile`                                    | `tls.crt`                                            | Name of the certificate file within the k8s secret
| `tls.caContent`                                   | `None`                                               | Certificate Authority content used to validate the Tiller server certificate
| `tls.hostname`                                    | `None`                                               | The server name used to verify the hostname on the returned certificates from the Tiller server
| `configureRepositories.enable`                    | `false`                                              | Enable volume mount for a `repositories.yaml` configuration file and repository cache
| `configureRepositories.volumeName`                | `repositories-yaml`                                  | Name of the volume for the `repositories.yaml` file
| `configureRepositories.secretName`                | `flux-helm-repositories`                             | Name of the secret containing the contents of the `repositories.yaml` file
| `configureRepositories.cacheName`                 | `repositories-cache`                                 | Name for the repository cache volume
| `configureRepositories.repositories`              | `None`                                               | List of custom Helm repositories to add. If non empty, the corresponding secret with a `repositories.yaml` will be created
| `kube.config`                                     | `None`                                               | Override for kubectl default config in the Helm Operator pod(s).
| `prometheus.enabled`                              | `false`                                              | If enabled, adds prometheus annotations to Helm Operator pod(s)

## Upgrade

Update Helm Operator version with:

```sh
helm upgrade --reuse-values helm-operator \
--set image.tag=0.10.1 \
fluxcd/helm-operator
```
