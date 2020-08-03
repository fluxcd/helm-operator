# Helm Operator chart

The [Helm operator chart](https://github.com/fluxcd/helm-operator/tree/master/chart/helm-operator)
bootstraps the Helm Operator on a [Kubernetes](http://kubernetes.io) cluster
using the [Helm](https://helm.sh) package manager.

## Prerequisites

* Kubernetes **>=v1.13**

## Installation

Add the Flux CD Helm repository:

```sh
helm repo add fluxcd https://charts.fluxcd.io
```

Install the `HelmRelease` Custom Resource Definition. By adding this
CRD it will be possible to define `HelmRelease` resources on the
cluster:

```sh
kubectl apply -f https://raw.githubusercontent.com/fluxcd/helm-operator/{{ version }}/deploy/crds.yaml
```

Install the Helm operator using the chart:

Chart defaults (Helm 2 and 3):

```sh
# Default with support for Helm 2 and 3 enabled
# NB: the presence of Tiller is a requirement when
# Helm 2 is enabled.
helm upgrade -i helm-operator fluxcd/helm-operator \
    --namespace flux
```

Helm 3:

```sh
# Only Helm 3 support enabled using helm.versions
helm upgrade -i helm-operator fluxcd/helm-operator \
    --namespace flux \
    --set helm.versions=v3
```

Helm 2:

```sh
# Only Helm 2 support enabled using helm.versions
# NB: the presence of Tiller is a requirement when
# Helm 2 is enabled.
helm upgrade -i helm-operator fluxcd/helm-operator \
    --namespace flux \
    --set helm.versions=v2
```

## Configuration

The following tables lists the configurable parameters of the Helm Operator
chart and their default values.

| Parameter                                         | Default                                              | Description
| -----------------------------------------------   | ---------------------------------------------------- | ---
| `image.repository`                                | `docker.io/fluxcd/helm-operator`                     | Image repository
| `image.tag`                                       | `{{ version }}`                                      | Image tag
| `image.pullPolicy`                                | `IfNotPresent`                                       | Image pull policy
| `image.pullSecret`                                | `None`                                               | Image pull secret
| `resources.requests.cpu`                          | `50m`                                                | CPU resource requests for the deployment
| `resources.requests.memory`                       | `64Mi`                                               | Memory resource requests for the deployment
| `resources.limits.cpu`                            | `None`                                               | CPU resource limits for the deployment
| `resources.limits.memory`                         | `None`                                               | Memory resource limits for the deployment
| `nodeSelector`                                    | `{}`                                                 | Node Selector properties for the deployment
| `tolerations`                                     | `[]`                                                 | Tolerations properties for the deployment
| `affinity`                                        | `{}`                                                 | Affinity properties for the deployment
| `extraVolumeMounts`                               | `[]`                                                 | Extra volume mounts to be added to the Helm Operator pod(s)
| `extraVolumes`                                    | `[]`                                                 | Extra volume to be added to the Helm Operator pod(s)
| `priorityClassName`                               | `""`                                                 | Set priority class for Helm Operator
| `extraEnvs`                                       | `[]`                                                 | Extra environment variables for the Helm Operator pod(s)
| `podAnnotations`                                  | `{}`                                                 | Additional pod annotations
| `podLabels`                                       | `{}`                                                 | Additional pod labels
| `rbac.create`                                     | `true`                                               | If `true`, create and use RBAC resources
| `rbac.pspEnabled`                                 | `false`                                              | If `true`, create and use a restricted pod security policy for Helm Operator pod(s)
| `serviceAccount.create`                           | `true`                                               | If `true`, create a new service account
| `serviceAccount.annotations`                      | `{}`                                                 | Additional Service Account annotations
| `serviceAccount.name`                             | `flux`                                               | Service account to be used
| `clusterRole.create`                              | `true`                                               | If `false`, Helm Operator will be restricted to the namespace where is deployed
| `clusterRole.name`                                | `None`                                               | The name of a cluster role to bind to
| `createCRD`                                       | `false`                                              | Install the `HelmRelease` CRD. Setting this value only has effect for Helm 2, as Helm 3 uses `--skip-crds` and will skip installation if the CRD is already present. Managing CRDs outside of Helm is recommended, also see the [Helm best practices](https://helm.sh/docs/chart_best_practices/custom_resource_definitions/)
| `service.type`                                    | `ClusterIP`                                          | Service type to be used (exposing the Helm Operator API outside of the cluster is not advised)
| `service.port`                                    | `3030`                                               | Service port to be used
| `updateChartDeps`                                 | `true`                                               | Update dependencies for charts
| `git.pollInterval`                                | `5m`                                                 | Period on which to poll git chart sources for changes
| `git.timeout`                                     | `20s`                                                | Duration after which git operations time out
| `git.defaultRef`                                  | `master`                                             | Ref to clone chart from if ref is unspecified in a HelmRelease
| `git.ssh.secretName`                              | `None`                                               | The name of the kubernetes secret with the SSH private key, supercedes `git.secretName`
| `git.ssh.known_hosts`                             | `None`                                               | The contents of an SSH `known_hosts` file, if you need to supply host key(s)
| `git.ssh.configMapName`                           | `None`                                               | The name of a kubernetes config map containing the ssh config
| `git.ssh.configMapKey`                            | `config`                                             | The name of the key in the kubernetes config map specified above
| `git.config.enabled`                              | `false`                                              | If `true`, mount the .gitconfig into the Helm Operator pod created from the `git.config.data`
| `git.config.secretName`                           | `None`                                               | The name of the kubernetes secret to store .gitconfig data created from the `git.config.data`
| `git.config.data`                                 | `None`                                               | The .gitconfig to be mounted into the home directory of the Helm Operator pod
| `chartsSyncInterval`                              | `3m`                                                 | Period on which to reconcile the Helm releases with `HelmRelease` resources
| `statusUpdateInterval`                            | `30s`                                                | Period on which to update the Helm release status in `HelmRelease` resources
| `workers`                                         | `4`                                                  | Number of workers processing releases
| `logFormat`                                       | `fmt`                                                | Log format (fmt or json)
| `logReleaseDiffs`                                 | `false`                                              | Helm Operator should log the diff when a chart release diverges (possibly insecure)
| `allowNamespace`                                  | `None`                                               | If set, this limits the scope to a single namespace. If not specified, all namespaces will be watched
| `helm.versions`                                   | `v2,v3`                                              | Helm versions supported by this operator instance, if v2 is specified then Tiller is required
| `tillerNamespace`                                 | `kube-system`                                        | Namespace in which the Tiller server can be found
| `tillerSidecar.enabled`                           | `false`                                              | Whether to deploy Tiller as a sidecar (and listening on `localhost` only).
| `tillerSidecar.image.repository`                  | `gcr.io/kubernetes-helm/tiller`                      | Image repository to use for the Tiller sidecar.
| `tillerSidecar.image.tag`                         | `v2.16.1`                                            | Image tag to use for the Tiller sidecar.
| `tillerSidecar.storage`                           | `secret`                                             | Storage engine to use for the Tiller sidecar.
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
| `configureRepositories.cacheVolumeName`           | `repositories-cache`                                 | Name for the repository cache volume
| `configureRepositories.repositories`              | `None`                                               | List of custom Helm repositories to add. If non empty, the corresponding secret with a `repositories.yaml` will be created
| `initPlugins.enable`                              | `false`                                              | Enable the initialization of Helm plugins using init containers
| `initPlugins.cacheVolumeName`                     | `plugins-cache`                                      | Name for the plugins cache volume
| `initPlugins.plugins`                             | `None`                                               | List of Helm plugins to initialize before starting the operator. If non empty, an init container will be added for every entry
| `kube.config`                                     | `None`                                               | Override for kubectl default config in the Helm Operator pod(s)
| `prometheus.enabled`                              | `false`                                              | If enabled, adds prometheus annotations to Helm Operator pod(s)
| `prometheus.serviceMonitor.create`                | `false`                                              | Set to true if using the Prometheus Operator
| `prometheus.serviceMonitor.interval`              | `None`                                               | Interval at which metrics should be scraped
| `prometheus.serviceMonitor.scrapeTimeout`         | `None`                                               | The timeout to configure the service monitor scrape task e.g `5s`
| `prometheus.serviceMonitor.namespace`             | `None`                                               | The namespace where the ServiceMonitor is deployed
| `prometheus.serviceMonitor.additionalLabels`      | `{}`                                                 | Additional labels to add to the ServiceMonitor
| `livenessProbe.initialDelaySeconds`               | `1`                                                  | The initial delay in seconds before the first liveness probe is initiated
| `livenessProbe.periodSeconds`                     | `10`                                                 | The number of seconds between the liveness probe is checked
| `livenessProbe.timeoutSeconds`                    | `5`                                                  | The number of seconds after which the liveness probe times out
| `livenessProbe.successThreshold`                  | `1`                                                  | The minimum number of consecutive successful probe results for the liveness probe to be considered successful
| `livenessProbe.failureThreshold`                  | `3`                                                  | The number of times the liveness probe can failed before restarting the container
| `readinessProbe.initialDelaySeconds`              | `1`                                                  | The initial delay in seconds before the first readiness probe is initiated
| `readinessProbe.periodSeconds`                    | `10`                                                 | The number of seconds between the readiness probe is checked
| `readinessProbe.timeoutSeconds`                   | `5`                                                  | The number of seconds after which the readiness probe times out
| `readinessProbe.successThreshold`                 | `1`                                                  | The minimum number of consecutive successful probe results for the readiness probe to be considered successful
| `readinessProbe.failureThreshold`                 | `3`                                                  | The number of times the readiness probe can failed before the container is marked as unready
| `initContainers`                                  | `[]`                                                 | Init containers and their specs
| `hostAliases`                                     | `{}`                                                 | Host aliases allow the modification of the hosts file (`/etc/hosts`) inside Helm Operator container. See <https://kubernetes.io/docs/concepts/services-networking/add-entries-to-pod-etc-hosts-with-host-aliases/>
| `dashboards.enabled`                              | `false`                                              | If enabled, helm-operator will create a configmap with a dashboard in json that's going to be picked up by grafana (see [sidecar.dashboards.enabled](https://github.com/helm/charts/tree/master/stable/grafana#configuration))
| `securityContext`                                 | `{}`                                                 | Adding `securityContext` options to the pod
| `containerSecurityContext.helmOperator`           | `{}`                                                 | Adding `securityContext` options to the helm operator container
| `containerSecurityContext.tiller`                 | `{}`                                                 | Adding `securityContext` options to the tiller container
| `sidecarContainers`                               | `{}`                                                 | Sidecar containers along with the specifications.

## How-to

### Use a custom Helm repository

Public Helm chart repositories that do not require any authentication do
not have to be configured and can just be referenced by their URL in a
`HelmRelease` resource. However, for Helm chart repositories that do
require authentication repository entries with the credentials need to be
added so the Helm Operator is able to authenticate against the repository.

Helm chart repository entries can be added with the chart using the
`configureRepositories.repositories` value, which accepts an array
of objects with the following keys:

| Key | Description
|-----|------------
| `name` | The name (alias) for the Helm chart repository
| `url`  | The URL of the Helm chart repository
| `username` | Helm chart repository username
| `password` | Helm chart repository password
| `certFile` | The path to a SSL certificate file used to identify the HTTPS client
| `keyFile` | The path to a SSL key file used to identify the HTTPS client
| `caFile` | The path to a CA bundle used to verify HTTPS enabled servers

---

For example, to add an Helm chart repository with username and password
protection:

```sh
helm upgrade -i helm-operator fluxcd/helm-operator \
    --namespace flux \
    --set configureRepositories.enable=true \
    --set 'configureRepositories.repositories[0].name=example' \
    --set 'configureRepositories.repositories[0].url=https://charts.example.com' \
    --set 'configureRepositories.repositories[0].username=john' \
    --set 'configureRepositories.repositories[0].password=s3cr3t!'
```

After adding the entry, the Helm chart in the repository can then be referred
to by the URL of the repository as usual:

```yaml
apiVersion: helm.fluxcd.io/v1
kind: HelmRelease
metadata:
  name: awesome-example
spec:
  chart:
    repository: https://charts.example.com
    version: 1.0.0
    name: awesome
```

### Use a private Git server

When using a private Git server to host your charts, setting the
`git.ssh.known_hosts` variable is required for enabling successful key matches
because `StrictHostKeyChecking` is enabled during `git pull` operations.

By setting the `git.ssh.known_hosts` variable, a configmap will be created
called `helm-operator-ssh-config` which in turn will be mounted into a volume
named `sshdir` at `/root/.ssh/known_hosts`.

Get the known hosts keys by running the following command:

```sh
ssh-keyscan <your_git_host_domain> > /tmp/flux_known_hosts
```

Generate a SSH key named `identity` and create a secret with it:

```sh
ssh-keygen -q -N "" -f /tmp/identity
kubectl create secret generic helm-operator-ssh \
    --from-file=/tmp/identity
    --namespace flux
```

Add `identity.pub` as a read-only deployment key in your Git repo and install
the Helm Operator:

```sh
helm upgrade -i helm-operator fluxcd/helm-operator \
    --namespace flux \
    --set git.ssh.secretName=helm-operator-ssh \
    --set-file git.ssh.known_hosts=/tmp/flux_known_hosts
```

You can refer to a chart from your private Git with:

```yaml
apiVersion: helm.fluxcd.io/v1
kind: HelmRelease
metadata:
  name: some-app
  namespace: default
spec:
  releaseName: some-app
  chart:
    git: git@your_git_host_domain:org/repo
    ref: master
    path: charts/some-app
  values:
    replicaCount: 1
```

### Use Flux's Git deploy key

You can configure the Helm Operator to use the Git SSH key generated by
[Flux](https://github.com/fluxcd/flux).

Assuming you have installed Flux with:

```sh
helm upgrade -i flux fluxcd/flux \
    --namespace flux \
    --set git.url=git@github.com:org/repo
```

When installing Helm Operator, you can refer to the Flux deploy key by the name
of the Kubernetes Secret:

```sh
helm upgrade -i helm-operator fluxcd/helm-operator \
    --namespace flux \
    --set git.ssh.secretName=flux-git-deploy
```

The deploy key naming convention is `<Flux Release Name>-git-deploy`.

### Use Helm downloader plugins

Helm downloader plugins like [`hypnoglow/helm-s3`](https://github.com/hypnoglow/helm-s3)
and [`hayorov/helm-gcs`](https://github.com/hayorov/helm-gcs) make it possible
to extend the protocols Helm recognizes to e.g. pull charts from a S3 bucket.

The chart offers an utility to install plugins before starting the operator
using init containers:

```sh
helm upgrade -i helm-operator fluxcd/helm-operator \
    --namespace flux \
    --set initPlugins.enable=true \
    --set 'initPlugins.plugins[0].plugin=https://github.com/hypnoglow/helm-s3.git' \
    --set 'initPlugins.plugins[0].version=0.9.2' \
    --set 'initPlugins.plugins[0].helmVersion=v3'
```

> **Note:** Most plugins assume credentials are available on the system they
> run on, make sure those are available at the expected paths using e.g.
> `extraVolumes` and `extraVolumeMounts`.

You should now be able to make use of the protocol added by the plugin:

```sh
cat <<EOF | kubectl apply -f -
apiVersion: helm.fluxcd.io/v1
kind: HelmRelease
metadata:
  name: chart-from-s3
  namespace: default
spec:
  chart:
    repository: s3://bucket-name/charts
    name: chart
    version: 0.1.0
  values:
    replicaCount: 1
EOF
```

### Uninstall

To uninstall/delete the `helm-operator` Helm release:

```sh
helm delete --purge helm-operator
```

The command removes all the Kubernetes components associated with the chart and
deletes the release.

> **Note:** `helm delete` will not remove the `HelmRelease` CRD.
> Deleting the CRD will trigger a cascade delete of all Helm release objects.
