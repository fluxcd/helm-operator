# Chart sources

In the introduction we created a simple `HelmRelease` that made use of a chart
from a Helm repository, but the Helm Operator does support multiple chart
sources, and virtually any protocol and/or source that can be shelled-in
through a [Helm downloader plugin](#extending-the-supported-helm-repository-protocols).

In this section of the guide we will take a deeper dive into the available
chart sources, and the unique features they have.

## Comparison

The following are short lists of the major characteristics of the different
chart sources. Keep these in mind when you have to make a decision about what
type of chart source to use for your `HelmRelease`, as they result in quite
different behaviour.

### Charts from [Helm repositories](#helm-repositories)

- Are immutable and non-moving (i.e. no updates for the chart itself are
  received unless the `.chart.version` is changed).
- Are cached for the lifetime duration of the Helm Operator pod.
- Are shared between `HelmRelease` resources making use of the same chart
  and version.
- Require a custom `repositories.yaml` to be [mounted and imported](#authentication-and-certificates)
  when authentication is required.
- [Are not just limited to HTTP/S](#extending-the-supported-helm-repository-protocols).
- _Do not support_ chart dependency updates (but instead use the dependencies
  bundled with the chart).
- _Do not support_ `valuesFrom.chartFileRef`.

### Charts from [Git repositories](#git-repositories)

- Move by mirroring the Git repository and fetching the latest `HEAD` for the
  configured `.chart.ref` on an interval (i.e. when a change is detected in
  Git under the `.chart.path`, a release will be scheduled for an upgrade).
- Share their Git repository mirror with `HelmRelease` resources making use
  of the same `.chart.git`, `.chart.ref` and `.chart.secretKeyRef`.
- Require a [private key to be mounted for Git over SSH](#ssh).
- Support credentials from a secret or a global `.netrc` file for [Git over
  HTTPS](#https).
- _Do support_ [chart dependency updates](#dependency-updates)
- _Do support_ [`valuesFrom.chartFileRef`](values.md#chart-files) to make use
  of alternative value files present in the `.chart.path`.

## Helm repositories

The Helm repository chart source is defined as follows in the `.spec` of a
`HelmRelease`. All listed fields are mandatory:

```yaml
spec:
  chart:
    repository: https://stefanprodan.github.io/podinfo
    name: podinfo
    version: 3.2.0
```

The definition of the listed keys is as follows:

* `repository`: The URL of the Helm repository, e.g. `https://kubernetes-charts.storage.googleapis.com`
  or `https://charts.example.com`.
  
  Putting `/index.yaml` behind the URL it should return an index file with all
  available charts for the Helm repository.

* `name`: The name of the Helm chart _without_ an alias, e.g. `podinfo`
  (instead of `<alias>/podinfo`).
  
  Having doubts about what to put here? Use the `name` as listed in the
  `Chart.yaml` of the Helm chart you want to use.

* `version`: The targeted Helm chart version, e.g. `3.2.0`.

In the [introduction](introduction.md) you already had a brief exposure to this
chart source, and in essence Helm repositories are the simplest way to make use
of a Helm chart in a `HelmRelease`.

To be able to perform releases with them the Helm Operator only makes use of
native Helm features and a tiny bit of glue to wire things together:

It will first attempt a reverse lookup for a repository alias in the local
`repositories.yaml` for the defined `repository` URL, if an alias is found it
will use this alias with the given `name` and `version` to instruct Helm to
fetch the chart to a cache path defined by the Helm Operator.

If the reverse lookup failed and no alias was found for the given URL it will
fallback to attempting to retrieve the absolute URL of the chart from the index
of the given `repository` URL, this URL is then used to instruct Helm to fetch
the chart to a cache path defined by the Helm Operator.

When this does not succeed either a status condition of type `ChartFetched`
will be recorded on the `HelmRelease` resource with the returned error.

### Authentication and certificates

Some Helm repositories require authentication or certificates before you are
able to make use of any charts they hold. At present, per-resource
authentication is not implemented for Helm repositories. The `HelmRelease`
Custom Resource does include a field `chartPullSecret` for attaching a
`repositories.yaml` file, but this has not been actually implemented.

As a workaround, you can mount a `repositories.yaml` file with authentication
already configured (and any required certificates) into the Helm Operator
container, and import it using the `--helm-repository-import` flag.

First, create a new empty `repositories.yaml` file _locally_:

```sh
touch repositories.yaml
```

You can now use `helm` to write the repository entry to this new file. Using
Helm 3 for this is the best option as it offers a `--repository-config` flag
and the generated output works for both versions:

```sh
helm repo add \
    --repository-config $PWD/repositories.yaml \
    --username <username> \
    --password <password> \
    <alias> <URL>
```

!!! note
    For Azure ACR repositories, you will need to [create a service
    principal](https://docs.microsoft.com/en-us/azure/container-registry/container-registry-auth-service-principal#create-a-service-principal)
    and use the plain text ID and password this gives you.

If you need to define any certificates, edit the respective `caFile`, `certFile`
and `keyFile` values of the entry you just added to the mount paths you will
later add to the Helm Operator _container_  (example path used here is
`/var/certs/*`):

```yaml
- caFile: /var/certs/ca.crt
  certFile: /var/certs/cert.crt
  keyFile: /var/certs/auth.key
  name: <alias>
```

Now you can create a secret in the same namespace as you are running the Helm
operator, from the repositories file:

```sh
kubectl create secret generic flux-helm-repositories \
    --from-file=$PWD/repositories.yaml
```

In case you defined any certificate entries, also create a secret for those
files in the same namespace as you are running the Helm Operator:

```sh
kubectl create secret generic flux-helm-repository-certs \
    --from-file=$PWD/ca.crt \
    --from-file=$PWD/cert.crt \
    --from-file=$PWD/auth.key
```

Mount the created secret(s) by adding to `volumes` in the pod spec of the Helm
operator deployment, and `volumeMounts` of the Helm Operator container. A good
mount path for the `repositories.yaml` file that does not give conflicts with
any Helm paths is `/root/.helm/repository/repositories.yaml`. Examples of this
can be found in the commented-out sections of the [example
deployment](https://github.com/fluxcd/helm-operator/blob/{{ version }}/deploy/deployment.yaml).

Lastly, configure the `--helm-repository-import` argument on the Helm Operator
container for your enabled Helm versions:

```yaml
        args:
        - --helm-repository-import=v2:/root/.helm/repository/repositories.yaml,v3:/root/.helm/repository/repositories.yaml
```

!!! note
    There is no limit to the amount of repository files that can be imported
    as the `--helm-repository-import` flag accepts a comma separated string
    slice of `<Helm version>:<filepath>`, e.g. `v3:/my/path.yaml`.
    Adding all entries to the same file and/or secret is thus not a requirement.

!!! hint
    For the [Helm chart](../references/chart.md), this all can be done by setting
    `configureRepositories.enable` to `true`, it will automatically pick up the 
    `flux-helm-reposities` secret created earlier in this guide and configure the
    `--helm-repository-import` flag for the enabled Helm versions. The certificate
    secret can be mounted by configuring the `extraVolumes` and `extraVolumeMounts`
    values.

### Extending the supported Helm repository protocols

By default, the Helm Operator is able to pull charts from repositories using
HTTP/S. It is however possible to extend the supported protocols by making use
of a [Helm downloader plugin](https://helm.sh/docs/topics/plugins/#downloader-plugins),
this allows you for example to use charts hosted on [Amazon S3](https://github.com/hypnoglow/helm-s3)
or [Google Cloud Storage](https://github.com/hayorov/helm-gcs).

#### Installing a Helm downloader plugin

The easiest way to install a plugin so that it becomes accessible to the Helm
operator to use an [init container](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/)
and one of the available `helm` binaries in the Helm Operator's image and a
volume mount. For the Helm chart,
[refer to the chart the documentation](../references/chart.md#use-helm-downloader-plugins).

**Plugin folder paths per Helm version:**
 

| Version | Plugins                         | Config                            |
|---------|---------------------------------|-----------------------------------|
| Helm 2  | `/var/fluxd/helm/cache/plugins` | `/var/fluxd/helm/plugins`         |
| Helm 3  | `/root/.cache/helm/plugins`     | `/root/.local/share/helm/plugins` |

Add a volume entry of [type `emptyDir`](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir)
to the deployment of your Helm Operator, this is where the plugins will be
stored for the lifetime duration of the pod:
   
```yaml
spec:
 volumes:
 - name: helm-plugins-cache
   emptyDir: {}
```

Next, add a new init container that uses the same image as the Helm
operator's container, and makes use of the earlier mentioned volume, with
correct volume mounts for the Helm version you are making use of. The
available `helm2` and `helm3` binaries can then be used to install the plugin:
   
```yaml
spec:
 initContainers:
 - name: helm-3-downloader-plugin
   image: docker.io/fluxcd/helm-operator:<tag>
   imagePullPolicy: IfNotPresent
   command:
     - 'sh'
     - '-c'
     # Replace '<plugin>' and '<version>' with the respective
     # values of the plugin you want to install
     - 'helm3 plugin install <plugin> --version <version>'
   volumeMounts:
   - name: helm-plugins-cache
     # See: 'plugin folder paths per Helm version'
     mountPath: /root/.cache/helm/plugins
     subPath: v3
   - name: helm-plugins-cache
     # See: 'plugin folder paths per Helm version'
     mountPath: /root/.local/share/helm/plugins
     subPath: v3-config
```

Last, add the same volume mounts to the Helm Operator container so that the
downloaded plugin becomes available:

```yaml
spec:
 containers:
 - name: flux-helm-operator
   image: docker.io/fluxcd/helm-operator:<tag>
   ...
   volumeMounts:
   - name: helm-plugins-cache
     # See: 'plugin folder paths per Helm version'
     mountPath: /root/.cache/helm/plugins
     subPath: v3
   - name: helm-plugins-cache
     # See: 'plugin folder paths per Helm version'
     mountPath: /root/.local/share/helm/plugins
     subPath: v3-config
```

#### Using an installed protocol in your HelmRelease

Once a Helm downloader plugin has been successfully installed, the newly added
protocol can be used in the `.chart.repository` value of a `HelmRelease`:

```yaml
spec:
  chart:
    repository: s3://bucket-name/charts
    name: chart-name
    version: 1.0.0
```

!!! caution
    Most downloader plugins expect some form of credentials to be present to be
    able to download a chart, make sure those are available in the Helm
    operator's container before attempting to make use of the newly added
    protocol.

## Git repositories

Besides the extensive support for Helm repositories, the Helm Operator also
offers support for charts from Git repository sources. You can refer to a chart
from a _Git_ repository, rather than a Helm repository, with a `.chart` section
like this:

```yaml
spec:
  chart:
    git: git@github.com:org/repo
    ref: master
    path: charts/podinfo
```

The definition of the listed keys is as follows:

* `git`: The URL of the Git repository, e.g. `git@github.com:org/repo` or
   `https://github.com/org/repo.git`. **Note:** specifying a custom port only
   works when the protocol is specified,
   e.g. `ssh://git@github.com:2222/org/repo.git` and not `git@github.com:2222/org/repo`.
* `ref` _(Optional)_: The Git reference, e.g. a branch, tag, or (short) commit
   hash. When omitted, defaults to `master` or the configured `--git-default-ref`.
* `path`: The path of the chart relative to the root of the Git repository.

In this case, the Helm Operator will start a mirror for the Git repository, and
a temporary working clone at the current `HEAD` of the defined `ref` of the
mirror will be created, before performing a release with the `path` given.

Mirrored Git repositories are polled for changes by fetching from the upstream
on the configured `--git-poll-interval` (defaults to 5 minutes). When a change
is detected the Helm Operator will collect all `HelmRelease` resources making
use of the mirror, and inspect if the change updates the chart at the `path`
given. When this is true, it will schedule a new release and an upgrade will
follow.

When a temporary working clone cannot be created due to e.g. the mirror not
being available yet or a cloning failure because of missing [credentials](#authentication),
a status condition of type `ChartFetched` will be recorded on the `HelmRelease` resource with the
returned error.

!!! note
    You can pin a chart to a specific version by changing the `.ref` to a tag
    or commit hash.

### Authentication

Unauthenticated cloning from Git repositories is possible for public Git
repositories by making the Helm Operator fetch them over HTTP/S. It is
however likely that most of the time you will be using a Git repository
chart source, some form of authentication is required before the repository
can be accessed by the Helm Operator.

!!! tip
    Because the Helm Operator does not perform any write operations on the Git
    repository, credentials with read permissions are always sufficient.

#### SSH

For Git over SSH the Helm Operator makes use of private keys available in the
container. Because of this, any `HelmRelease` under the management of a
Helm Operator instance has access to the same repositories once a private key
has been provided and no additional configuration is required for the resource
itself other than defining the Git repository in the `.chart.repo`.

To provide a private key to be used for Git operations over SSH, put the key in
a secret under the entry `identity`:

```sh
kubectl create secret generic flux-git-deploy \
    --from-file=identity=<path to key file>
```

Next, mount it into the Helm Operator container as shown in the
[example deployment](https://github.com/fluxcd/helm-operator/blob/{{ version }}/deploy/deployment.yaml).

The default `ssh_config` that ships with the Helm Operator's Docker image
expects an identity file at `/etc/fluxd/ssh/identity`, which is where it will
be if you just uncomment the blocks from the example.

##### Providing multiple private keys

If you are using more than one repository, you may need to provide more than
one private key. In that case, you can create a secret with an entry for each
key and mount that _as well as_ an `ssh_config` file mentioning each key as an
`IdentityFile` in `/root/.ssh`.

For example, to provide different credentials for `github.com` and
`bitbucket.org` you would create a `ssh_config` file looking like this:

```text
Host github.com
    HostName github.com
    User git
    IdentityFile <github private key path>
    IdentitiesOnly yes

Host bitbucket.org
    HostName bitbucket.org
    User git
    IdentityFile <bitbucket private key path>
    IdentitiesOnly yes
```

!!! note
    The `IdentitiesOnly` ensures that only the `IdentityFile` for the
    `Host` is used and any other identity files known are ignored.

###### Multiple private keys for Git repositories on the same host

There is one caveat to the example illustrated above; due to the fact that
permissions are being handled by the Git server and not SSH itself, any public
key known to the Git server will result in a successful login while the private
counterpart it belongs to may not actually have access to the Git repository
that is targeted. This poses a problem when you have multiple repositories on
the same Git server with a dedicated private key per repository.

The workaround is to use an alias for the `Host` value, and then use this as a
replacement for the hostname in the defined Git repository URL of the
`HelmRelease`:

```text
Host github-repository1
    HostName github.com
    User git
    IdentityFile <repository specific private key path>
    IdentitiesOnly yes
```

```yaml
spec:
  chart:
    git: git@github-repository1:org/repo
    ref: master
    path: charts/podinfo
```

#### HTTPS

For Git over HTTPS the Helm Operator offers two ways of providing credentials.

##### Per-resource credentials using .chart.secretRef

To provide HTTPS credentials per `HelmRelease` resource you can make use of
a `secretRef` in the `.chart` and a secret with a username and password.
The defined secret is retrieved from Kubernetes and appended to the
`.chart.git` URL before starting the Git mirror.

First, create a secret with the `username` and `password` that give access
to the Git repository:

```sh
kubectl create secret generic git-https-credentials \
    --from-literal=username=<username> \
    --from-literal=password=<password>
```

Now, add the reference to the secret to the `.chart`:

```yaml
spec:
  chart:
    git: https://github.com/org/repo
    ref: master
    path: charts/podinfo
    secretRef:
      name: git-https-credentials
```

##### Global credentials using .netrc

It is also possible to provide `HelmRelease` resources access to global
credentials via a
[`.netrc` file](https://ec.haxx.se/usingcurl/usingcurl-netrc) mounted in the
`/root/` directory of the Helm Operator container.

!!! caution
     This approach suffers essentially from [the same caveat as mentioned for
     Git over SSH](#multiple-private-keys-for-git-repositories-on-the-same-host).

To provide credentials for `github.com`, you would create a `.netrc` file like
this:

```text
machine github.com
login <username>
password <password>
```

After mounting the file from a secret, you can then define the plain HTTPS URL
of the Git repository in your `HelmRelease`:

```yaml
spec:
  chart:
    git: https://github.com/org/repo.git
    ref: master
    path: charts/podinfo
```

### Dependency updates

For a chart from a Git repository the Helm Operator runs a dependency update
by default, this is done through an action that equals to `helm dep update`.
You may want to disable this behaviour, for example because your dependencies
are already in git, to do so set `skipDepUpdate` to `true` in `.chart`:

```yaml
spec:
  chart:
    git: git@github.com:org/repo
    ref: master
    path: charts/podinfo
    skipDepUpdate: true
```

### Notifying the Helm Operator about Git changes

As earlier laid out in this guide the Helm Operator fetches the upstream of
mirrored Git repositories on the configured `--git-poll-interval` (defaults
to 5 minutes). In some scenarios (think CI/CD), you may not want to wait for
this interval to pass.

To help you with this the Helm Operator serves a HTTP API endpoint to
instruct it to immediately refresh all Git mirrors:

```console
$ kubectl port-forward deployment/flux-helm-operator 3030:3030 &
$ curl -XPOST http://localhost:3030/api/v1/sync-git
OK
```

!!! warning
    The HTTP API has no built-in authentication, this means you either need to
    port forward before making the request or put something in front of it to
    serve as a gatekeeper.
