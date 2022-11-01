---
title: Values
weight: 30
---

Now that we have a good understanding of where we can get our charts from and
what they have to offer, it is time to examine how you can supply
[values](https://helm.sh/docs/glossary/#values-values-files-values-yaml) to
be used with the chart when the Helm Operator makes a release.

### Inlined values

The most simple option to define the values for your Helm release. This is a
YAML map as you would put in a file and supply to Helm with `-f values.yaml`,
but inlined into the `HelmRelease` manifest. For example:

```yaml
spec:
  values:
    foo: value1
    bar:
    baz: value2
    oof:
    - item1
    - item2
```

## Values from sources

It is possible to define a list of config maps, secrets (in the same namespace
as the `HelmRelease` by default, or in a configured namespace) or external
sources (URLs) from which to take values. For charts from a Git
repository, there is an additional option available to refer to a file in
the chart folder.

The values are merged in the order given, with later values overwriting
earlier. These values always have a lower priority than the values
inlined in the `HelmRelease` via the `spec.values` parameter.

This is useful if you want to have defaults such as the `region`,
`clustername`, `environment`, a local docker registry URL, etc., or if you
simply do not want the values to be visible as plaintext in the `HelmRelease`.

{{% alert color="info" title="Note" %}}
The Helm Operator does not watch the sources for updates. Changes to
sources are detected during the [reconciliation
loop](reconciliation-and-upgrades.md#reconciliation).
{{% /alert %}}

### Config maps

The reference to a config map is defined by adding a `configMapKeyRef` to the
`spec.valuesFrom` list.

```yaml
spec:
  valuesFrom:
  - configMapKeyRef:
      name: default-values
      namespace: my-ns
      key: values.yaml
      optional: false
```

The definition of the listed keys is as follows:

- `name`: The name of the config map.
- `namespace` _(Optional)_: The namespace the config map is in. Defaults to the
  namespace of the `HelmRelease` when omitted.
- `key` _(Optional)_: The key in the config map to get the values from.
  Defaults to `values.yaml` when omitted.
- `optional` _(Optional)_: When set to `true`, successful retrieval of the
  config map is optional and a release will still be made if it is missing.
  Defaults to `false` when omitted.

### Secrets

The reference to a secret is defined by adding a `secretKeyRef` to the
`spec.valuesFrom` list.

```yaml
spec:
  valuesFrom:
  - secretKeyRef:
      name: default-values
      namespace: my-ns
      key: values.yaml
      optional: true
```

The definition of the listed keys is as follows:

- `name`: The name of the secret.
- `namespace` _(Optional)_: The namespace the secret is in. Defaults to the
  namespace of the `HelmRelease` when omitted.
- `key` _(Optional)_: The key in the secret to get the values from.
  Defaults to `values.yaml` when omitted.
- `optional` _(Optional)_: When set to `true`, successful retrieval of the
  secret is optional and a release will still be made if it is missing.
  Defaults to `false` when omitted.

### External sources

The reference to an external source (URL) is defined by adding a
`externalSourceRef` to the `spec.valuesFrom` list. The external
source is expected to be a plain YAML file.

```yaml
spec:
  valuesFrom:
  - externalSourceRef:
      url: https://example.com/static/raw/values.yaml
      optional: true
```

The definition of the listed keys is as follows:

- `url`: The URL of the plain YAML file.
- `optional` _(Optional)_: When set to `true`, successful retrieval of the
  YAML file is optional and a release will still be made if it could not be
  found. Defaults to `false` when omitted.

### Chart files

{{% alert color="info" title="Note" %}}
Values from chart files are only supported for charts from a [Git
repository](chart-sources.md#git-repositories).
{{% /alert %}}

When making use of a chart sourced from a Git repository, it is possible to
refer to a values file relative to the path of the chart. This can for example
be utilized to select values optimized for production environments, that you ship
with your chart.

```yaml
spec:
  valuesFrom:
  - chartFileRef:
      path: overrides/environment-prod.yaml
      optional: true
```

The definition of the listed keys is as follows:

- `path`: The path relative to the `.chart.path` where the values file can be
  found.
- `optional` _(Optional)_: When set to `true`, successful retrieval of the
  YAML file is optional and a release will still be made if it could not be
  found. Defaults to `false` when omitted.
