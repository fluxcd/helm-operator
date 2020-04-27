# Automation

This section of the guide is mostly a clarification of some common
misconceptions about (non-existent) automation features of the Helm Operator.

## Image updates

Because the Helm Operator is a [Flux umbrella project](https://fluxcd.io),
occasionally people assume it is capable of updating image references in
the `HelmRelease` and/or associated charts. This feature is however baked
in to [Flux](https://github.com/fluxcd/flux), and not the Helm Operator
itself due to it having no knowledge of available images or the origin of the
`HelmRelease`.

For more details about this Flux feature, please refer to the [documentation
for the `HelmRelease` integration](https://docs.fluxcd.io/en/stable/references/helm-operator-integration).

## Helm repository chart updates

Another much requested feature is automated updates for charts from [Helm
repository chart sources](chart-sources.md#helm-repositories). The
development of this feature is currently blocked until the automation logic
has been untangled from Flux. To keep up-to-date about new developments of
this feature you can subscribe to
[fluxcd/helm-operator#12](https://github.com/fluxcd/helm-operator/issues/12).

It is possible to get a similar functionality by making use of an
[umbrella chart](https://helm.sh/docs/howto/charts_tips_and_tricks/#complex-charts-with-many-dependencies)
from a [Git repository chart source](chart-sources.md#git-repositories) with
a [version range set](https://helm.sh/docs/topics/chart_best_practices/dependencies/#versions),
as for charts from Git repository sources, a [dependency update is performed by
default](chart-sources.md#dependency-updates), and that will download the
latest available version within the defined range.

For example, to make the Helm Operator install the latest `1.2.x` patch release
for `foo-chart`, you would define the following in the `dependencies` of your
(dummy) umbrella chart in Git:

```yaml
dependencies:
- name: 
  version: ~1.2.0
  repository: https://charts.example.com
```
