## 1.4.1 (2021-11-26)

This release of Helm Operator has no code changes, only base image and dependency updates.

CHANGELOG - TODO

## 1.4.0 (2021-07-08)

> **Note: Breaking Changes**

This release of Helm Operator will break compatibility with older releases of Kubernetes, in order to ensure forward compatibility with long-awaited breaking changes in Kubernetes 1.22.0.

The `v1beta1` release of `CustomResourceDefinition` as well as `Role`, `ClusterRole`, and role bindings, have been replaced by their `v1` counterparts, and will finally be removed.

> **Helm Operator and Flux are in maintenance:**
> Efforts have been focused on the next generation of Flux, also called the [GitOps Toolkit](https://toolkit.fluxcd.io), which has crossed the feature-parity milestone, and is already recommended for production use in many cases. The [helm-controller](https://toolkit.fluxcd.io/components/helm/controller/) is the replacement for Helm Operator. The roadmap for Flux v2 development including Helm Controller can be found [here](https://fluxcd.io/docs/roadmap/).
>
> We are eager to hear [feedback, suggestions, and/or feature requests](https://github.com/fluxcd/toolkit/discussions) for the helm-controller and other Toolkit components. The [migration timetable](https://fluxcd.io/docs/migration/timetable/) will be kept updated with developments regarding the ongoing support of Helm Operator.
>
> Users of Helm Operator should be planning or executing their migrations, and report any blocking issues so that they can be addressed as early as possible.

NOTE: Make sure to update the CRD when upgrading from a previous version as they have been changed since the prior release.

Please note, while we continue the Helm Operator support, there are some known issues in Helm Operator that cannot be addressed. Users are advised strongly to plan their infrastructure upgrades and [migrate to Flux v2 and Helm Controller](https://fluxcd.io/docs/migration/helm-operator-migration/) as soon as possible, as Helm Operator will soon [no longer be maintained](https://fluxcd.io/docs/migration/timetable/).

### Maintenance

- Update CRD version for CustomResourceDefinition [fluxcd/helm-operator#599][#599]
- Update CRD version for rbac.authorization.k8s.io [fluxcd/helm-operator#618][#618]

[#599]: https://github.com/fluxcd/helm-operator/pull/599
[#618]: https://github.com/fluxcd/helm-operator/pull/618

## 1.3.0 (2021-07-07)

> **Helm Operator and Flux are in maintenance:**
> Efforts have been focused on the next generation of Flux, also called the [GitOps Toolkit](https://toolkit.fluxcd.io), which has crossed the feature-parity milestone, and is already recommended for production use in many cases. The [helm-controller](https://toolkit.fluxcd.io/components/helm/controller/) is the replacement for Helm Operator. The roadmap for Flux v2 development including Helm Controller can be found [here](https://fluxcd.io/docs/roadmap/).
>
> We are eager to hear [feedback, suggestions, and/or feature requests](https://github.com/fluxcd/toolkit/discussions) for the helm-controller and other Toolkit components. The [migration timetable](https://fluxcd.io/docs/migration/timetable/) will be kept updated with developments regarding the ongoing support of Helm Operator.
>
> Users of Helm Operator should be planning their migrations, and report any blocking issues so that they can be addressed as early as possible.

NOTE: Make sure to update the CRD when upgrading from a previous version as they may have been changed since the prior release.

Documentation for Helm Operator has moved under [fluxcd.io/legacy/flux](https://fluxcd.io/legacy/flux/). This release includes any merged fixes that were unreleased for the past year, including an upgrade from Helm 3.1.x to Helm 3.5.4 which also covered some [breaking changes](https://github.com/helm/helm/releases/tag/v3.5.2) in Helm.

The next MINOR (1.4.0) release of Helm Operator will break compatibility with older releases of Kubernetes, in order to ensure forward compatibility with long-awaited breaking changes in Kubernetes 1.22.0.

([fluxcd/helm-operator#599][#599] and [fluxcd/helm-operator#618][#618] describe the breaking changes that are upcoming in Helm Operator 1.4.0.)

Please note, this is a security update and while we continue the Helm Operator support, there are some known issues in Helm Operator that cannot be addressed. Users are advised strongly to plan their infrastructure upgrades and [migrate to Flux v2 and Helm Controller](https://fluxcd.io/docs/migration/helm-operator-migration/) as soon as possible, as Helm Operator will soon [no longer be maintained](https://fluxcd.io/docs/migration/timetable/).

### Maintenance

- Update helm v2 and v3 to latest [fluxcd/helm-operator#604][]
- Update Alpine to 3.13 [fluxcd/helm-operator#589][]
- Update Helm stable repository url [fluxcd/helm-operator#577][]
- Move Fons to maintainer emeritus [fluxcd/helm-operator#547][]
- Disable link-checking [fluxcd/helm-operator#594][]
- Add maintenance note to GitHub templates and README [fluxcd/helm-operator#548][]

### Enhancements

- Add `terminationGracePeriodSeconds` value to Helm Operator Chart [fluxcd/helm-operator#564][]
- Default kubeconfig to None in chart [fluxcd/helm-operator#521][]
- Remove the query string from the CleanURL [fluxcd/helm-operator#571][]
- Sync chart mirror on chart spec change to prevent incorrect reconciliation [fluxcd/helm-operator#573][]
- Add explicit namespace field to namespaced resources [fluxcd/helm-operator#517][]
- Add more Helm chart dashboard options [fluxcd/helm-operator#522][]
- Unnest convert settings from tls [fluxcd/helm-operator#524][]
- Enable additional sidecar containers [fluxcd/helm-operator#484][]
- Adding securityContext options at the pod and container level. [fluxcd/helm-operator#494][]

### Fixes

- fix: kubeconfig override in chart/deployment [fluxcd/helm-operator#507][]
- fix: make passing gitconfig as a value optional [fluxcd/helm-operator#551][]
- fix: Address constant release creation [fluxcd/helm-operator#533][]
- fix: take ignored OptionalSecretKeySelector into account [fluxcd/helm-operator#616][]

### Documentation

- docs: Fix Readme (configureRepositories.cacheVolumeName) [fluxcd/helm-operator#502][]
- docs: Remove docs, point to new location where relevant [fluxcd/helm-operator#611][]
- docs: Update issue/PR templates with new info on v2 [fluxcd/helm-operator#583][]
- docs: Add website, twitter, linkedin to docs.f.i [fluxcd/helm-operator#592][]
- docs: how to update the helm-op docs post-release [fluxcd/helm-operator#610][]
- docs: Make docs v1 warning global [fluxcd/helm-operator#590][]
- docs: Point to flux2 from our docs [fluxcd/helm-operator#588][]
- docs: Malformed list in Chart sources -> Git repositories [fluxcd/helm-operator#549][]
- docs: Remove redundant text [fluxcd/helm-operator#542][]
- docs: Fix bullet lists in chart-sources.md [fluxcd/helm-operator#529][]

### Thanks

Thanks to @Carles-Figuerola, @amit-handda, @bmalynovytch, @coultenholt, @dholbach, @em-schmidt, @flimzy, @fredr, @hiddeco, @jbuettnerbild, @kingdonb, @krichardson-apexclearing, @mattjw, @mbrancato, @mnaser, @neil-greenwood, @nilesh892003, @smarthall, @stefanprodan, @swade1987, @t1bb4r, @tux-00 and @wujiangfa-xlauncher for their contributions to this release.

[#618]: https://github.com/fluxcd/helm-operator/pull/618
[fluxcd/helm-operator#616]: https://github.com/fluxcd/helm-operator/pull/616
[fluxcd/helm-operator#611]: https://github.com/fluxcd/helm-operator/pull/611
[fluxcd/helm-operator#610]: https://github.com/fluxcd/helm-operator/pull/610
[fluxcd/helm-operator#605]: https://github.com/fluxcd/helm-operator/pull/605
[fluxcd/helm-operator#604]: https://github.com/fluxcd/helm-operator/pull/604
[#599]: https://github.com/fluxcd/helm-operator/pull/599
[fluxcd/helm-operator#594]: https://github.com/fluxcd/helm-operator/pull/594
[fluxcd/helm-operator#592]: https://github.com/fluxcd/helm-operator/pull/592
[fluxcd/helm-operator#590]: https://github.com/fluxcd/helm-operator/pull/590
[fluxcd/helm-operator#589]: https://github.com/fluxcd/helm-operator/pull/589
[fluxcd/helm-operator#588]: https://github.com/fluxcd/helm-operator/pull/588
[fluxcd/helm-operator#583]: https://github.com/fluxcd/helm-operator/pull/583
[fluxcd/helm-operator#577]: https://github.com/fluxcd/helm-operator/pull/577
[fluxcd/helm-operator#573]: https://github.com/fluxcd/helm-operator/pull/573
[fluxcd/helm-operator#571]: https://github.com/fluxcd/helm-operator/pull/571
[fluxcd/helm-operator#564]: https://github.com/fluxcd/helm-operator/pull/564
[fluxcd/helm-operator#551]: https://github.com/fluxcd/helm-operator/pull/551
[fluxcd/helm-operator#549]: https://github.com/fluxcd/helm-operator/pull/549
[fluxcd/helm-operator#548]: https://github.com/fluxcd/helm-operator/pull/548
[fluxcd/helm-operator#547]: https://github.com/fluxcd/helm-operator/pull/547
[fluxcd/helm-operator#542]: https://github.com/fluxcd/helm-operator/pull/542
[fluxcd/helm-operator#533]: https://github.com/fluxcd/helm-operator/pull/533
[fluxcd/helm-operator#529]: https://github.com/fluxcd/helm-operator/pull/529
[fluxcd/helm-operator#524]: https://github.com/fluxcd/helm-operator/pull/524
[fluxcd/helm-operator#522]: https://github.com/fluxcd/helm-operator/pull/522
[fluxcd/helm-operator#521]: https://github.com/fluxcd/helm-operator/pull/521
[fluxcd/helm-operator#517]: https://github.com/fluxcd/helm-operator/pull/517
[fluxcd/helm-operator#507]: https://github.com/fluxcd/helm-operator/pull/507
[fluxcd/helm-operator#506]: https://github.com/fluxcd/helm-operator/pull/506
[fluxcd/helm-operator#502]: https://github.com/fluxcd/helm-operator/pull/502
[fluxcd/helm-operator#494]: https://github.com/fluxcd/helm-operator/pull/494
[fluxcd/helm-operator#484]: https://github.com/fluxcd/helm-operator/pull/484

## 1.2.0 (2020-07-29)

> **Note on the future of the Helm Operator and Flux:**
> We are working on a next generation Flux assembled from components
> as part of a bigger [GitOps Toolkit](https://toolkit.fluxcd.io) project.
> One of the components is the [helm-controller](https://toolkit.fluxcd.io/components/helm/controller/)
> which eventually will replace the Helm Operator. The roadmap for this
> can be found [here](https://toolkit.fluxcd.io/roadmap/#the-road-to-helm-operator-v2).
>
> We are eager to hear [feedback, suggestions, and/or feature requests](https://github.com/fluxcd/toolkit/discussions)
> for the helm-controller and other Toolkit components.

This is the second minor release, it adds support for Helm tests and
v2 to v3 release conversions, and includes a variety of bug fixes.

NOTE: Make sure to update the CRD when upgrading from a previous version as they have been changed in this release.

### Bug fixes

 - metrics: use release name and namespace in `release_condition_info`
   labels
   [fluxcd/helm-operator#431][#431]
 - operator: obtain lock before obtaining release data
   fluxcd/helm-operator{[#437][], [#445][]}
 - misc: use sigs.k8s.io/yaml everywhere
   [fluxcd/helm-operator#455][#455]
 - release: increase timeout for annotator to support large umbrella
   charts
   [fluxcd/helm-operator#478][#478]
 - metrics: remove `release_condition_info` data on delete
   fluxcd/helm-operator{[#485][], [#492][], [#495][]}
 - release: prevent spurious upgrades for semver ranges
   [fluxcd/helm-operator#490][#490]
 - helm/v3: slightly increase GC offset anonymous files
   [fluxcd/helm-operator#491][#491]
   
### Enhancement

 - helm: add `helm test` integration
   fluxcd/helm-operator{[#415][], [#472][]}
 - helm/v3: add v2->v3 release converter
   fluxcd/helm-operator{[#415][], [#471][], [#486][]}
 - helm/v3: add flag for disabling OpenAPI Validation
   [fluxcd/helm-operator#480][#480]
 
### Maintenance and documentation

 - docs: fix `HelmReleaseError` alert rule expression in monitoring
   reference
   [fluxcd/helm-operator#429][#429]
 - docs: fix comment regarding repositories.yaml setup
   [fluxcd/helm-operator#433][#433]
 - docs: replace v1.13 API reference URLs with v1.18
   [fluxcd/helm-operator#442][#442]
 - e2e: increase liveness and readiness probe failureThreshold during
   tests
   [fluxcd/helm-operator#500][#500] 

### Thanks

Thanks to @stefanprodan, @sa-spag, @MMartyn, @seaneagan, @hezhizhen,
@saada, @chrisjholly, @avielb, @luxas, @waseem-h, and others for their
contributions to this release, feedback, and reporting issues.

[#415]: https://github.com/fluxcd/helm-operator/pull/415
[#429]: https://github.com/fluxcd/helm-operator/pull/429
[#431]: https://github.com/fluxcd/helm-operator/pull/431
[#433]: https://github.com/fluxcd/helm-operator/pull/433
[#437]: https://github.com/fluxcd/helm-operator/pull/437
[#442]: https://github.com/fluxcd/helm-operator/pull/442
[#445]: https://github.com/fluxcd/helm-operator/pull/445
[#455]: https://github.com/fluxcd/helm-operator/pull/455
[#471]: https://github.com/fluxcd/helm-operator/pull/471
[#472]: https://github.com/fluxcd/helm-operator/pull/472
[#478]: https://github.com/fluxcd/helm-operator/pull/478
[#480]: https://github.com/fluxcd/helm-operator/pull/480
[#485]: https://github.com/fluxcd/helm-operator/pull/485
[#486]: https://github.com/fluxcd/helm-operator/pull/486
[#490]: https://github.com/fluxcd/helm-operator/pull/490
[#491]: https://github.com/fluxcd/helm-operator/pull/491
[#492]: https://github.com/fluxcd/helm-operator/pull/492
[#495]: https://github.com/fluxcd/helm-operator/pull/495
[#500]: https://github.com/fluxcd/helm-operator/pull/500

## 1.1.0 (2020-05-21)

This is the first minor release, it focuses on metrics improvements,
state logic improvements, and the fixing of spurious disk usage bug
due to anonymous index files not getting cleaned up in Helm 3.

### Bug fixes

 - release: reset recorded failure cond on dry-run
   fluxcd/helm-operator{[#385][], [#425][]}
 - release: imply release failure upon chart fetch
   failure
   [fluxcd/helm-operator#399][#399]
 - helm/v3: garbage collect anonymous index files
   [fluxcd/helm-operator#422][#422]

### Enhancement

 - release: keep track of last attempted revision
   [fluxcd/helm-operator#382][#382]
 - metrics: add `release_count` to expose number of Helm Operator
   managed releases
   [fluxcd/helm-operator#387][#387]
 - metrics: add `release_condition_info` condition gauge
   [fluxcd/helm-operator#403][#403]
 - metrics: introduce `release_action_duration_seconds` as a
   replacement for `release_phase_duration_seconds`
   [fluxcd/helm-operator#407][#407]
 - metrics: add `600` and `1800` duration buckets
   [fluxcd/helm-operator#407][#407]
 - operator: fix typo in synchronization log
   [fluxcd/helm-operator#411][#411]
 - chartsync/git: allow selecting cross-namespace secrets for
   HTTPS credentials
   [fluxcd/helm-operator#421][#421]
   
### Maintenance and documentation

 - docs: improvements to contributing documentation
   [fluxcd/helm-operator#297][#297]
 - docs: fix link-checking and a number of links
   [fluxcd/helm-operator#347][#347]
 - docs: improve monitoring documentation
   [fluxcd/helm-operator#384][#384]
 - build: update Go to `1.14.x`
   [fluxcd/helm-operator#386][#386]
 - e2e: fix tests on MacOS
   [fluxcd/helm-operator#400][#400]
   
### Thanks

Thanks to @stefansedich, @sa-spag, @dholbach, @seaneagan,@stefanprodan,
@vladlosev, @fllaca, @hiddeco, @Sayrus, @squaremo, and others for their
contributions to this release, feedback, and reporting issues.

[#297]: https://github.com/fluxcd/helm-operator/pull/297
[#347]: https://github.com/fluxcd/helm-operator/pull/347
[#382]: https://github.com/fluxcd/helm-operator/pull/382
[#384]: https://github.com/fluxcd/helm-operator/pull/384
[#385]: https://github.com/fluxcd/helm-operator/pull/385
[#387]: https://github.com/fluxcd/helm-operator/pull/387
[#386]: https://github.com/fluxcd/helm-operator/pull/386
[#399]: https://github.com/fluxcd/helm-operator/pull/399
[#400]: https://github.com/fluxcd/helm-operator/pull/400
[#403]: https://github.com/fluxcd/helm-operator/pull/403
[#407]: https://github.com/fluxcd/helm-operator/pull/407
[#411]: https://github.com/fluxcd/helm-operator/pull/411
[#421]: https://github.com/fluxcd/helm-operator/pull/421
[#422]: https://github.com/fluxcd/helm-operator/pull/422
[#425]: https://github.com/fluxcd/helm-operator/pull/425

## 1.0.1 (2020-04-15)

This is a patch release.

### Bug fixes

 - release: do not swallow dependency update errors
   [fluxcd/helm-operator#372][#372]

### Thanks

Thanks to @brew, @qvmedvedev, @stefansedich, @hiddeco,
@stefanprodan, and others for their contributions to this release,
feedback, and reporting issues.

[#372]: https://github.com/fluxcd/helm-operator/pull/372

## 1.0.0 (2020-04-07)

> **Notice:** upgrading to this version from `<=0.10.x` by just
> updating your Helm Operator image tag is not possible as the
> CRD domain and version have changed. An upgrade guide can be
> found [here](./docs/how-to/upgrade-to-ga.md).

> **Notice:**  due to the multiple added fields, you need to
> re-apply the `HelmRelease` CRD.

This release marks the first GA release of the Helm Operator,
and the end of the release candidate stretch releases.

### Bug fixes

 - status: unset `RolledBack` condition on `Released` == `True`
   [fluxcd/helm-operator#326][#326]

### Enhancements

 - apis: generate CRDs using `controller-gen`
   [fluxcd/helm-operator#270][#270]
 - release: configure go logger to log using our go-kit logger
   [fluxcd/helm-operator#306][#306]
 - install: rename files and resource names
   [fluxcd/helm-operator#322][#322]
 - helm: update Helm 3 to `3.1.2` and Helm 2 to `2.16.3`
   [fluxcd/helm-operator#333][#333]
 - release: record release phases on the `HelmRelease` resource
   [fluxcd/helm-operator#334][#334]
 - release: introduce `--reuse-values` functionality
   [fluxcd/helm-operator#359][#359]
   
### Maintenance and documentation

 - docs: update `helm delete` documentation for Helm 3
   [fluxcd/helm-operator#330][#330]
 - docs: switch to `mkdocs`
   [fluxcd/helm-operator#332][#332]
 - docs: new quickstart guide, `HelmRelease` guide, FAQ, etc.
   [fluxcd/helm-operator#321][#321]
 - e2e: update `gitsrv` to `1.0.0`
   [fluxcd/helm-operator#341][#341]
 - docs: fix the name description in the secret section
   [fluxcd/helm-operator#342][#342]
 - docs: fix 'add Helm chart' example
   [fluxcd/helm-operator#343][#343]
 - docs: update `MAINTAINERS`
   [fluxcd/helm-operator#351][#351]
 - e2e: use binary of targeted Helm version in tests
   [fluxcd/helm-operator#360][#360]
   
### Thanks

Thanks to @stefansedich, @mmorejon, @sa-spag, @stefanprodan, @dholbach,
@hiddeco, and others for their contributions to this release, feedback,
and reporting issues.

[#270]: https://github.com/fluxcd/helm-operator/pull/270
[#306]: https://github.com/fluxcd/helm-operator/pull/306
[#321]: https://github.com/fluxcd/helm-operator/pull/321
[#322]: https://github.com/fluxcd/helm-operator/pull/322
[#326]: https://github.com/fluxcd/helm-operator/pull/326
[#330]: https://github.com/fluxcd/helm-operator/pull/330
[#332]: https://github.com/fluxcd/helm-operator/pull/332
[#333]: https://github.com/fluxcd/helm-operator/pull/333
[#334]: https://github.com/fluxcd/helm-operator/pull/334
[#341]: https://github.com/fluxcd/helm-operator/pull/341
[#342]: https://github.com/fluxcd/helm-operator/pull/342
[#343]: https://github.com/fluxcd/helm-operator/pull/343
[#351]: https://github.com/fluxcd/helm-operator/pull/351
[#359]: https://github.com/fluxcd/helm-operator/pull/359
[#360]: https://github.com/fluxcd/helm-operator/pull/360

## 1.0.0-rc9 (2020-02-13)

> **Notice:** upgrading to this version from `<=0.10.x` by just
> updating your Helm Operator image tag is not possible as the
> CRD domain and version have changed. An upgrade guide can be
> found [here](docs/how-to/upgrade-to-ga.md).

> **Notice:**  due to the multiple added fields, you need to
> re-apply the `HelmRelease` CRD.

### Bug fixes

 - release: propagate all configured release flags to dry-run upgrade
   [fluxcd/helm-operator#250][#250]
 - chartsync: honour the configured default Git ref when reconciling
   charts source
   [fluxcd/helm-operator#253][#253]
 - release: disable atomic flag for Helm chart installation
   [fluxcd/helm-operator#256][#256]
 - apis: correct JSON namespace tag for key selectors
   [fluxcd/helm-operator#262][#262]
 - helm/v3: support upgrades of releases with nested `HelmRelease`
   resources (using a patched Helm `3.0.3` release)
   [fluxcd/helm-operator#292][#292]

### Enhancements

 - release: support retrying rollbacks
   [fluxcd/helm-operator#252][#252]
 - helm: support downloader plugins
   [fluxcd/helm-operator#263][#263]
 - helm/v3: support skipping CRD installation using `.spec.skipCRDs`
   [fluxcd/helm-operator#282][#282]
 - helm/v3: enrich Helm logger with release name and namespace
   metadata
   [fluxcd/helm-operator#291][#291]

### Maintenance and documentation

 - e2e: use podinfo's `--unready` to make a release fail
   [fluxcd/helm-operator#258][#258]
 - Pkg: update Helm 3 to `3.0.3`
   fluxcd/helm-operator{[#260][], [#292][]}
 - build: include `bash` and `curl` in image
   [fluxcd/helm-operator#276][#267]
 - build: make sure we test all the local modules
   [fluxcd/helm-operator#269][#269]
 - build: add `generate-codegen` target to Makefile
   [fluxcd/helm-operator#289][#289]
 - e2e: install Tiller in operator namespace for more reliable cleanup
   [fluxcd/helm-operator#290][#290]
 - e2e: do not create kind clusters in parallel
   [fluxcd/helm-operator#290][#290]
 - docs: document usage of Helm downloader plugins
   [fluxcd/helm-operator#295][#295]
 - docs: highlight standalone usage in `README.md`
   [fluxcd/helm-operator#296][#296]

### Thanks

Thanks to @sa-spag, @stefanprodan, @mcharriere, @GODBS, @derrickburns,
@autarchprinceps, @stefanseditch, @infinitydon, @cbenjemaa, @sayboras,
@2opremio, @hiddeco, and others for their contributions to this
release, feedback, and reporting issues.

[#250]: https://github.com/fluxcd/helm-operator/pull/250
[#252]: https://github.com/fluxcd/helm-operator/pull/252
[#253]: https://github.com/fluxcd/helm-operator/pull/253
[#256]: https://github.com/fluxcd/helm-operator/pull/256
[#258]: https://github.com/fluxcd/helm-operator/pull/258
[#260]: https://github.com/fluxcd/helm-operator/pull/260
[#262]: https://github.com/fluxcd/helm-operator/pull/262
[#263]: https://github.com/fluxcd/helm-operator/pull/263
[#267]: https://github.com/fluxcd/helm-operator/pull/267
[#269]: https://github.com/fluxcd/helm-operator/pull/269
[#282]: https://github.com/fluxcd/helm-operator/pull/282
[#289]: https://github.com/fluxcd/helm-operator/pull/289
[#290]: https://github.com/fluxcd/helm-operator/pull/290
[#291]: https://github.com/fluxcd/helm-operator/pull/291
[#292]: https://github.com/fluxcd/helm-operator/pull/292
[#295]: https://github.com/fluxcd/helm-operator/pull/295
[#296]: https://github.com/fluxcd/helm-operator/pull/296

## 1.0.0-rc8 (2020-01-25)

> **Notice:** upgrading to this version from `<=0.10.x` by just
> updating your Helm Operator image tag is not possible as the
> CRD domain and version have changed. An upgrade guide can be
> found [here](docs/how-to/upgrade-to-ga.md).

> **Notice:**  due to the multiple added fields, you need to
> re-apply the `HelmRelease` CRD.

### Bug fixes

 - release: push returned error as condition on sync check
   failure
   [fluxcd/helm-operator#209][#209]
 - release: reject git source if URL and path are missing
   [fluxcd/helm-operator#223][#223]
 - helm: only hold repository config lock for duration of
   read so dry-runs are run in parallel again
   [fluxcd/helm-operator#225][#225]
 - release: use all set `rollback` values when performing
   a rollback operation
   [fluxcd/helm-operator#239][#239]
 - helm: do not include non-template files in chart data
   so that the generation of a `requirement.lock` due to
   a dependency update does not cause spurious upgrades
   [fluxcd/helm-operator#242][#242]

### Enhancements

 - release: allow `.spec.wait` to be set for upgrades
   [fluxcd/helm-operator#95][#95]
 - chartsync: support supplying Git HTTPS credentials
   using `secretRef`
   [fluxcd/helm-operator#172][#172]
 - status: retry status and condition updates on conflicts
   [fluxcd/helm-operator#210][#210]
 - release: allow `secretKeyRef` and `configMapKeyRef` to be
   selected from other namespaces using the `namespace` key
   [fluxcd/helm-operator#219][#219]
 - helm: only index missing repositories when fetching a
   chart from an URL
   [fluxcd/helm-operator#225][#225]
 - helm/v3: propagate main application logger to client
   [fluxcd/helm-operator#232][#232]
 - release: allow max history to be overridden using
   `.spec.maxHistory`
   [fluxcd/helm-operator#235][#235]
 - release: rely on Helm storage for determining when to
   upgrade after rolling back
   [fluxcd/helm-operator#239][#239]

### Maintenance and documentation

 - build: update Kubernetes Kind to `v0.7.0` and set
   Kubernetes `v1.14.10` for end-to-end tests
   [fluxcd/helm-operator#207][#207]
 - build: upgrade code-generator to Kubernetes 1.16.2
   [fluxcd/helm-operator#214][#214]
 - docs: update FAQ on Flux garbage collection
   [fluxcd/helm-operator#221][#221]
 - Pkg: update Flux to `v1.17.2-0.20200121140732-3903cf8e71c3`
   [fluxcd/helm-operator#230][#230]
 - Pkg: make `pkg/install` a Go module to reduce its
   dependencies
   [fluxcd/helm-operator#234][#234]


### Thanks

Thanks to @sa-spag, @carlpett, @sureshamk, @ingeknudsen, @cep21,
@HaveFun83, @stefanprodan, @runningman84, @nabadger, @Helcaraxan,
@stefansedich, @hiddeco, @grrywlsn, @niall-weedon, @richardcase,
@REBELinBLUE, @derrickburns, and others for their contributions to
this release, feedback, and reporting issues.

[#95]: https://github.com/fluxcd/helm-operator/pull/95
[#172]: https://github.com/fluxcd/helm-operator/pull/172
[#207]: https://github.com/fluxcd/helm-operator/pull/207
[#209]: https://github.com/fluxcd/helm-operator/pull/209
[#210]: https://github.com/fluxcd/helm-operator/pull/210
[#214]: https://github.com/fluxcd/helm-operator/pull/214
[#219]: https://github.com/fluxcd/helm-operator/pull/219
[#221]: https://github.com/fluxcd/helm-operator/pull/221
[#223]: https://github.com/fluxcd/helm-operator/pull/223
[#225]: https://github.com/fluxcd/helm-operator/pull/225
[#230]: https://github.com/fluxcd/helm-operator/pull/230
[#232]: https://github.com/fluxcd/helm-operator/pull/232
[#234]: https://github.com/fluxcd/helm-operator/pull/234
[#235]: https://github.com/fluxcd/helm-operator/pull/235
[#239]: https://github.com/fluxcd/helm-operator/pull/239
[#242]: https://github.com/fluxcd/helm-operator/pull/242

## 1.0.0-rc7 (2020-01-10)

> **Notice:** upgrading to this version from `<=0.10.x` by just
> updating your Helm Operator image tag is not possible as the
> CRD domain and version have changed. An upgrade guide can be
> found [here](docs/how-to/upgrade-to-ga.md).

### Bug fixes

 - release: ignore manifest in annotator if we fail to unmarshal
   [fluxcd/helm-operator#190][#190]
 - helm/v2: if present, use alias to pull a chart so that the
   configured credentials are used
   [fluxcd/helm-operator#193][#193]
 - helm/v3: if present, use alias to pull a chart so that the
   configured credentials are used
   [fluxcd/helm-operator#193][#193]
 - helm/v3: prevent spurious upgrades due to missing dependencies
   in chart metadata
   [fluxcd/helm-operator#196][#196]
  
### Enhancements

 - helm: only log errors on repository index
   [fluxcd/helm-operator#193][#193]
   
### Maintenance and documentation

 - Pkg: update Flux package to `v1.17.0` (`kube-1.16` branch)
   [fluxcd/helm-operator#177][#177]

### Thanks

Thanks to @stefansedich, @domg123, @gaieges, @PaulFarver, @rowecharles,
@apenney, @stefanprodan, @hiddeco, and others for their contributions
to this release, feedback, and reporting issues.

[#177]: https://github.com/fluxcd/helm-operator/pull/177
[#190]: https://github.com/fluxcd/helm-operator/pull/190
[#193]: https://github.com/fluxcd/helm-operator/pull/193
[#196]: https://github.com/fluxcd/helm-operator/pull/196

## 1.0.0-rc6 (2020-01-08)

> **Notice:** upgrading to this version from `<=0.10.x` by just
> updating your Helm Operator image tag is not possible as the
> CRD domain and version have changed. An upgrade guide can be
> found [here](docs/how-to/upgrade-to-ga.md).

This release fixes some (but not all) of the more critical bugs
reported since Helm v3 was introduced in the last release.

### Bug fixes

 - helm/v3: include available credentials when pulling chart
   [fluxcd/helm-operator#171][#171]
 - helm/v2: index repositories before pulling chart
   [fluxcd/helm-operator#181][#181]
 - release: detect change of chart base values
   [fluxcd/helm-operator#182][#182]
 - release: filter out nil release resources during parsing
   [fluxcd/helm-operator#182][#185]
   
### Maintenance and documentation

 - Pkg: update Helm packages to `3.0.2` and `2.16.1`
   [fluxcd/helm-operator#177][#177]
 - Build: fix flakiness of various end-to-end tests
   [fluxcd/helm-operator#178][#178]
 - Build: fix authentication issue on `git clone` in end-to-end tests
   [fluxcd/helm-operator#180][#180]
   
### Thanks

Thanks to @PaulFarver, @stefansedich, @richardcase, @stefanprodan, @hiddeco,
and others for their contributions to this release, feedback, and reporting
issues.

[#171]: https://github.com/fluxcd/helm-operator/pull/171
[#177]: https://github.com/fluxcd/helm-operator/pull/177
[#178]: https://github.com/fluxcd/helm-operator/pull/178
[#180]: https://github.com/fluxcd/helm-operator/pull/180
[#181]: https://github.com/fluxcd/helm-operator/pull/181
[#182]: https://github.com/fluxcd/helm-operator/pull/182
[#185]: https://github.com/fluxcd/helm-operator/pull/185

## 1.0.0-rc5 (2019-12-23)

> **Notice:** upgrading to this version from `<=0.10.x` by just
> updating your Helm Operator image tag is not possible as the
> CRD domain and version have changed. An upgrade guide can be
> found [here](docs/how-to/upgrade-to-ga.md).

> **Notice:**  due to the added `helmVersion` field, you need
> to re-apply the `HelmRelease` CRD.

This release brings Helm v3 support to the release candidate,
Helm v3 functionalities should be considered _beta_. Support for
Helm v2 and v3 is enabled by default. To target Helm v3, set the
`.spec.helmVersion` in a `HelmRelease` to `v3`.

Enabling _just_ Helm v3 is possible by configuring
`--enabled-helm-versions=v3`, this will also make the
`.spec.helmVersion` default to `v3`.

To be able to support multiple Helm versions large parts of the
operator had to be rewritten, which lead to several improvements
around release deciscion making and keeping track of charts from
Git sources. We also no longer shell out to the `helm` binary
to achieve certain functionalities but instead make directly use
of the available Helm packages, this will also ease the support
of charts from OCI sources in upcoming releases.

Extensive documentation will be added in the next release candidate,
which will likely also be the last RC before moving to GA.

### Improvements

 - Helm v3 support (`v3.0.1`)
   [fluxcd/helm-operator#156][#156]

   With a subset of notable PRs:
   - Include of `helm2` and `helm3` binaries in Docker image
     [fluxcd/helm-operator#118][#118]
   - Support for importing Helm v2 and v3 repositories using
     the `--helm-repository-import` flag
     [fluxcd/helm-operator#141][#141]
   - Refactor of downloads from Helm chart repositories; it
     now uses the download manager from Helm instead of our
     own logic
     [fluxcd/helm-operator#145][#145]
   - Refactor of dependency updates; it now uses the download
     manager from Helm instead of shelling out to the `helm`
     binary
     [fluxcd/helm-operator#145][#145]
 - Decoupling of release reconciliation from chart source sync
   [fluxcd/helm-operator#99][#99]

### Maintenance and documentation

 - Build: upgrade Go to `1.13.3`
   [fluxcd/helm-operator#104][#90]
 - Pkg: only use `fluxcd/flux`
   [fluxcd/helm-operator#104][#104]
 - Build: end-to-end tests
   [fluxcd/helm-operator#]{[#110][], [#118][], [#148][], [#150][]}

### Thanks

Thanks @carnott-snap, @karuppiah7890, @hiddeco, @stefanprodan,
@2opremio and @stefansedich for contributions to this release.

Plus a special thanks to users testing the alpha version with Helm
v3 support, notably @gsf, @dminca, @rowecharles, @eschereisin,
@stromvirvel, @timja, @dragonsmith, @maxstepanov, @jan-schumacher,
@StupidScience, @brew, and all others that may have gone unnoticed.

[#90]: https://github.com/fluxcd/helm-operator/pull/90
[#99]: https://github.com/fluxcd/helm-operator/pull/99
[#104]: https://github.com/fluxcd/helm-operator/pull/104
[#110]: https://github.com/fluxcd/helm-operator/pull/110
[#118]: https://github.com/fluxcd/helm-operator/pull/118
[#141]: https://github.com/fluxcd/helm-operator/pull/141
[#145]: https://github.com/fluxcd/helm-operator/pull/145
[#148]: https://github.com/fluxcd/helm-operator/pull/148
[#150]: https://github.com/fluxcd/helm-operator/pull/150
[#156]: https://github.com/fluxcd/helm-operator/pull/156

## 1.0.0-rc4 (2019-11-22)

> **Notice:** upgrading to this version from `<=0.10.x` by just
> updating your Helm Operator image tag is not possible as the
> CRD domain and version have changed. An upgrade guide can be
> found [here](docs/how-to/upgrade-to-ga.md).

### Improvements

 - Update Helm to `2.16.1`
   [fluxcd/helm-operator#107][#107]
 - Add flag to specify an alternative default Git ref
   [fluxcd/helm-operator#83][#83]

### Maintenance and documentation

 - Build: use `fluxcd/flux` import in `chartsync` package
   [fluxcd/helm-operator#101][#101]
 - Documentation: document `timeout`, `resetValues` and `forceUpgrade`
   `HelmRelease` fields
   [fluxcd/helm-operator#82][#82]

### Thanks 

Many thanks to @carnott-snap, @frabar-lancom, @hiddeco, @niall-weedon,
and @stefanprodan for contributions to this release.

[#82]: https://github.com/fluxcd/helm-operator/pull/82
[#83]: https://github.com/fluxcd/helm-operator/pull/83
[#101]: https://github.com/fluxcd/helm-operator/pull/101
[#107]: https://github.com/fluxcd/helm-operator/pull/107

## 1.0.0-rc3 (2019-10-18)

> **Notice:** upgrading to this version from `<=0.10.x` by just
> updating your Helm Operator image tag is not possible as the
> CRD domain and version have changed. An upgrade guide can be
> found [here](docs/how-to/upgrade-to-ga.md).

This is the third release candidate.

### Improvements

 - Update Helm to `v2.14.3` and kubectl to `v1.14.7`
   [fluxcd/helm-operator#73][#73]
 - Add EKS repo to the list of public Helm repositories in the default config
   [fluxcd/helm-operator#64][#64]
 - De-experimental-ise `--workers` flag and set the default value to two workers
   [fluxcd/helm-operator#53][#53]

### Maintenance and documentation

 - Build: Make release build cope with `v` prefixed tags 
   [fluxcd/helm-operator#61][#61]

### Thanks

Many thanks to @hiddeco and @stefanprodan
for contributions to this release.

[#73]: https://github.com/fluxcd/helm-operator/pull/73
[#64]: https://github.com/fluxcd/helm-operator/pull/64
[#53]: https://github.com/fluxcd/helm-operator/pull/53
[#61]: https://github.com/fluxcd/helm-operator/pull/61

## 1.0.0-rc2 (2019-10-02)

> **Notice:** upgrading to this version from `<=0.10.x` by just
> updating your Helm Operator image tag is not possible as the
> CRD domain and version have changed. An upgrade guide can be
> found [here](docs/how-to/upgrade-to-ga.md).

This is the second release candidate.

### Bug fixes

 - Fix permissions on chart directory creation for non-root users
   [fluxcd/helm-operator#31][#31]
 - Filter out `nil` resources during parsing of released resources,
   as it caused confusion due to a harmless `Object 'Kind' is missing
   in 'null'` error being logged
   [fluxcd/helm-operator#47][#47]
 - Make `OwnedByHelmRelease` default to `true`, to work around some
   edge case scenarios where no resources are present for the release,
   or they are all skipped
   [fluxcd/helm-operator#56][#56]

### Improvements

 - Add `--status-update-interval` flag, for configuring the interval
   at which the operator consults Tiller for the status of a release
   [fluxcd/helm-operator#44][#44]
 - Expand the list of public Helm repositories in the default config
   [fluxcd/helm-operator#53][#53]

### Maintenance and documentation

 - Build: avoid spurious diffs in generated files by fixing their
   modtimes to Unix epoch [fluxcd/helm-operator#50][#50]
 - Build: update Flux dependency to `v1.15.0`
   [fluxcd/helm-operator#58][#58]
 - Documentation: Kustomize installation tutorial and various fixes
   [fluxcd/helm-operator#32][#32]
 - Documentation: add Helm v3 (alpha) workshop to `README.md`
   [fluxcd/helm-operator#52][#52]
   
### Thanks

Many thanks to @knackaron, @stefanprodan, @hiddeco, @swade1987
for contributions to this release.

[#31]: https://github.com/fluxcd/helm-operator/pull/31
[#32]: https://github.com/fluxcd/helm-operator/pull/32
[#44]: https://github.com/fluxcd/helm-operator/pull/44
[#47]: https://github.com/fluxcd/helm-operator/pull/47
[#50]: https://github.com/fluxcd/helm-operator/pull/50
[#52]: https://github.com/fluxcd/helm-operator/pull/52
[#53]: https://github.com/fluxcd/helm-operator/pull/53
[#56]: https://github.com/fluxcd/helm-operator/pull/56
[#58]: https://github.com/fluxcd/helm-operator/pull/58

## 1.0.0-rc1 (2019-08-14)

> **Notice:** upgrading to this version by just updating your Helm
> Operator image tag is not possible as the CRD domain and version
> have changed. An upgrade guide can be found [here](docs/how-to/upgrade-to-ga.md).

This is the first `1.0.0` release candidate, there is no rule about
the amount of published release candidates before it is officially
promoted to GA `1.0.0`, except for 'when it is considered stable'.

### Improvements

 - Support releasing to a different namespace than the CR is in, by
   configuring a `targetNamespace` in the `HelmRelease`
   [fluxcd/flux#2334][#2334]
 - Deployment in `deploy/` examples are now generated from templates
   fluxcd/helm-operator{[#2][#2], [#4][#4], [#19][#19]}
 - Display the release name, status, and message on `kubectl get hr`
   [fluxcd/helm-operator#23][#23]
 - Added descriptions for the CRD fields
   [fluxcd/helm-operator#24][#24]

### Maintenance and documentation

 - Project has been moved to a dedicated repository
   [fluxcd/helm-operator][helm-op-repo]
 - Support code generation from a non `$GOPATH/src` location
   [fluxcd/helm-operator#6][#6]
 - `HelmRelease` API version has been bumped to `v1` and group domain has changed to
   `helm.fluxcd.io`
   [fluxcd/helm-operator#7][#7]
 - Deprecated `FluxHelmRelease` resource and `helm.integratations.flux.weave.works`
   group domain have been removed
   [fluxcd/helm-operator#7][#7]   
 - Various updates to the documentation and a dedicated
   directory for Helm Operator related docs
   fluxcd/helm-operator{[#20][#20], [#22][#22], [#24][#24]}
 - Update Docker labels to reflect move to dedicated
   repository
   [fluxcd/helm-operator#21][#21]
 - Use Helm chart in e2e tests
   [fluxcd/helm-operator#26][#26]

[#2]: https://github.com/fluxcd/helm-operator/pull/2
[#4]: https://github.com/fluxcd/helm-operator/pull/4
[#6]: https://github.com/fluxcd/helm-operator/pull/6
[#7]: https://github.com/fluxcd/helm-operator/pull/7
[#19]: https://github.com/fluxcd/helm-operator/pull/19
[#20]: https://github.com/fluxcd/helm-operator/pull/20
[#21]: https://github.com/fluxcd/helm-operator/pull/21
[#22]: https://github.com/fluxcd/helm-operator/pull/22
[#23]: https://github.com/fluxcd/helm-operator/pull/23
[#24]: https://github.com/fluxcd/helm-operator/pull/24
[#26]: https://github.com/fluxcd/helm-operator/pull/26
[#2334]: https://github.com/fluxcd/flux/pull/2334
[helm-op-repo]: https://github.com/fluxcd/helm-operator

## 0.10.1 (2019-08-07)

> **Notice:** this release contains a `HelmRelease`
> [Custom Resource Definition][helm 0.10.1 crd] fix. Please make sure
> you patch the CRD in your cluster.

### Bug fixes

 - Fixed `rollback.timeout` definition in the `CustomResourceDefinition`
   [weaveworks/flux#2251][#2251]
 - Fixed the merge of values
   [weaveworks/flux#2292][#2292]
 - Correct spelling of integrations, and fix `make check-generated`
   [weaveworks/flux#2312][#2312]
 - Moved successful chart fetch signal to reconcile action (to prevent
   an infinite loop due to the `LastUpdateTime` on the condition getting
   accidentally updated during rollback checks).
   [weaveworks/flux#2316][#2316]
 - Fixed typo in `ReasonUpgradeFailed` condition change reason
   [weaveworks/flux#2317][#2317]

### Thanks

This release was made possible by contributions from @jfrndz, @adrian,
@stefanprodan, @obiesmans, @chriscorn-takt, @sureshamk, @dholbach,
@squaremo, and @hiddeco.

[#2251]: https://github.com/fluxcd/flux/pull/2251
[#2292]: https://github.com/fluxcd/flux/pull/2292
[#2312]: https://github.com/fluxcd/flux/pull/2312
[#2316]: https://github.com/fluxcd/flux/pull/2316
[#2317]: https://github.com/fluxcd/flux/pull/2317
[helm 0.10.1 crd]: https://github.com/fluxcd/flux/blob/helm-0.10.1/deploy-helm/flux-helm-release-crd.yaml

## 0.10.0 (2019-07-10)

This release brings you [opt-in automated rollback support][rollback docs],
new Prometheus metrics, and _experimental_ support of spawning
multiple workers with the `--workers=<num>` flag to speed up the
processing of releases.

This will likely also be the last _minor_ beta release before we
promote the Helm operator to its first GA `1.0.0` release.

> **Notice:** the Helm operator relies on changes in the `HelmRelease`
> [Custom Resource Definition][helm 0.10.0 crd]. Please make sure you patch the
> CRD in your cluster _before_ upgrading the Helm operator.

### Bug fixes

 - Prevent an infinite release loop when multiple `HelmRelease`
   resources with the same release name configuration coexist,
   by looking at the antecedent annotation set on release resources
   and confirming ownership
   [weaveworks/flux#2123][#2123]

### Improvements

 - Opt-in automated rollback support; when enabled, a failed release
   will be rolled back automatically and the operator will not attempt
   a new release until it detects a change in the chart and/or the
   configured values
   [weaveworks/flux#2006][#2006]
 - Increase timeout for annotating resources from a Helm release, to
   cope with large umbrella charts
   [weaveworks/flux#2123][#2123]
 - New Prometheus metrics

   + `release_queue_length_count`
   + `release_duration_seconds{action=['INSTALL','UPGRADE'], dry-run=['true', 'false'], success=['true','false'], namespace, releasename}`
   
   [weaveworks/flux#2191][#2191]
 - Experimental support of spawning multiple queue workers processing
   releases by configuring the `--workers=<num>` flag
   [weaveworks/flux#2194][#2194]

### Maintenance and documentation

 - Publish images to [fluxcd DockerHub][] organization
   [weaveworks/flux#2213][#2213]
 - Document opt-in rollback feature
   [weaveworks/flux#2220][#2220]

### Thanks

Many thanks to @adrian, @2opremio, @semyonslepov, @gtseres, @squaremo, @stefanprodan, @kingdonb, @ncabatoff,
@dholbach, @cristian-radu, @simonmacklin, @hiddeco for contributing to this release.

[#2006]: https://github.com/weaveworks/flux/pull/2006
[#2123]: https://github.com/weaveworks/flux/pull/2123
[#2191]: https://github.com/weaveworks/flux/pull/2191
[#2194]: https://github.com/weaveworks/flux/pull/2194
[#2213]: https://github.com/weaveworks/flux/pull/2213
[#2220]: https://github.com/weaveworks/flux/pull/2220
[helm 0.10.0 crd]: https://github.com/weaveworks/flux/blob/release/helm-0.10.x/deploy-helm/flux-helm-release-crd.yaml
[rollback docs]: docs/helmrelease-guide/rollbacks.md
[fluxcd DockerHub]: https://hub.docker.com/r/weaveworks/helm-operator/

## 0.9.2 (2019-06-13)

### Bug fixes

 - Ensure releases are enqueued on clone change only
   [weaveworks/flux#2081][#2081]
 - Reorder start of processes on boot and verify informer cache sync
   early, to prevent the operator from hanging on boot
   [weaveworks/flux#2103][#2103]
 - Use openssh-client rather than openssh in container image
   [weaveworks/flux#2142][#2142]

### Improvements

 - Enable pprof to ease profiling
   [weaveworks/flux#2095][#2095]

### Maintenance and documentation

 - Add notes about production setup Tiller
   [weaveworks/flux#2146][#2146]
   
### Thanks

Thanks @2opremio, @willholley ,@runningman84, @stefanprodan, @squaremo,
@rossf7, @hiddeco for contributing.

[#2081]: https://github.com/weaveworks/flux/pull/2081
[#2095]: https://github.com/weaveworks/flux/pull/2095
[#2103]: https://github.com/weaveworks/flux/pull/2103
[#2142]: https://github.com/weaveworks/flux/pull/2142
[#2146]: https://github.com/weaveworks/flux/pull/2146

## 0.9.1 (2019-05-09)

### Bug fixes

 - During the lookup of `HelmRelease`s for a mirror, ensure the
   resource has a git chart source before comparing the mirror name
   [weaveworks/flux#2027][#2027]

### Thanks

Thanks to @puzza007, @squaremo, @2opremio, @stefanprodan, @hiddeco
for reporting the issue, patching and reviewing it.

[#2027]: https://github.com/weaveworks/flux/pull/2027

## 0.9.0 (2019-05-08)

### Bug fixes

 - Make sure client-go logs to stderr
   [weaveworks/flux#1945][#1945]
 - Prevent garbage collected `HelmRelease`s from getting upgraded
   [weaveworks/flux#1906][#1906]

### Improvements

 - Enqueue release update on git chart source changes and improve
   mirror change calculations
   [weaveworks/flux#1906][#1906], [weaveworks/flux#2005][#2005]
 - The operator now checks if the `HelmRelease` spec has changed after
   it performed a dry-run, this prevents scenarios where it could
   enroll an older revision of a `HelmRelease` while a newer version
   was already known
   [weaveworks/flux#1906][#1906]
 - Stop logging broadcasted Kubernetes events
   [weaveworks/flux#1906][#1906]
 - Log and return early if release is not upgradable
   [weaveworks/flux#2008][#2008]

### Maintenance and documentation

 - Update client-go to `v1.11`
   [weaveworks/flux#1929][#1929]
 - Move images to DockerHub and have a separate pre-releases image repo
   [weaveworks/flux#1949][#1949], [weaveworks/flux#1956][#1956]
 - Support `arm` and `arm64` builds
   [weaveworks/flux#1950][#1950]
 - Retry keyscan when building images, to mitigate for occasional
   timeouts
   [weaveworks/flux#1971][#1971]

### Thanks

Thanks @brezerk, @jpds, @stefanprodan, @2opremio, @hiddeco, @squaremo,
@dholbach, @bboreham, @bricef and @stevenpall for their contributions
to this release, and anyone who I have missed during this manual
labour.

[#1906]: https://github.com/weaveworks/flux/pull/1906
[#1929]: https://github.com/weaveworks/flux/pull/1929
[#1945]: https://github.com/weaveworks/flux/pull/1945
[#1949]: https://github.com/weaveworks/flux/pull/1949
[#1950]: https://github.com/weaveworks/flux/pull/1950
[#1956]: https://github.com/weaveworks/flux/pull/1956
[#1971]: https://github.com/weaveworks/flux/pull/1971
[#2005]: https://github.com/weaveworks/flux/pull/2005
[#2008]: https://github.com/weaveworks/flux/pull/2008

## 0.8.0 (2019-04-11)

This release bumps the Helm API package and binary to `v2.13.0`;
although we have tested and found it to be backwards compatible, we
recommend running Tiller `>=2.13.0` from now on.

### Improvements

 - Detect changes made to git chart source in `HelmRelease`
   [weaveworks/flux#1865][#1865]
 - Cleanup git chart source clone on `HelmRelease` removal
   [weaveworks/flux#1865][#1865]
 - Add `chartFileRef` option to `valuesFrom` to support using a
   non-default values yamel from a git-sourced Helm chart
   [weaveworks#1909][#1909]
 - Reimplement `--git-poll-interval` to control polling interval of
   git mirrors for chart sources
   [weaveworks/flux#1910][#1910]

### Maintenance and documentation

 - Bump Helm API package and binary to `v2.13.0`
   [weaveworks/flux#1828][#1828]
 - Verify scanned keys in same build step as scan
   [weaveworks/flux#1908][#1908]
 - Use Helm operator image from build in e2e tests
   [weaveworks/flux#1910][#1910]

[#1828]: https://github.com/weaveworks/flux/pull/1828
[#1865]: https://github.com/weaveworks/flux/pull/1865
[#1908]: https://github.com/weaveworks/flux/pull/1908
[#1909]: https://github.com/weaveworks/flux/pull/1909
[#1910]: https://github.com/weaveworks/flux/pull/1910

### Thanks

Thanks to @hpurmann, @2opremio, @arturo-c, @squaremo, @stefanprodan,
@hiddeco, and others for their contributions to this release, feedback,
and bringing us one step closer to a GA-release.

## 0.7.1 (2019-03-27)

### Bug fixes

 - Prevent panic on `.spec.values` in `HelmRelease` due to merge
   attempt on uninitialized value
   [weaveworks/flux#1867](https://github.com/weaveworks/flux/pull/1867)

## 0.7.0 (2019-03-25)

### Bug fixes

 - Run signal listener in a goroutine instead of deferring
   [weaveworks/flux#1680](https://github.com/weaveworks/flux/pull/1680)
 - Make chart operations insensitive to (missing) slashes in Helm
   repository URLs
   [weaveworks/flux#1735](https://github.com/weaveworks/flux/pull/1735)
 - Annotating resources outside of the `HelmRelease` namespace
   [weaveworks/flux#1757](https://github.com/weaveworks/flux/pull/1757)

### Improvements

 - The `HelmRelease` CRD now supports a `skipDepUpdate` to instruct the
   operator to not update dependencies for charts from a git source
   [weaveworks/flux#1712](https://github.com/weaveworks/flux/pull/1712)
   [weaveworks/flux#1823](https://github.com/weaveworks/flux/pull/1823)
 - Azure DevOps Git host support
   [weaveworks/flux#1729](https://github.com/weaveworks/flux/pull/1729)
 - The UID of the `HelmRelease` is now used as dry run release name
   [weaveworks/flux#1745](https://github.com/weaveworks/flux/pull/1745)
 - Removed deprecated `--git-poll-interval` flag
   [weaveworks/flux#1757](https://github.com/weaveworks/flux/pull/1757)
 - Sync hook to instruct the operator to refresh Git mirrors
   [weaveworks/flux#1776](https://github.com/weaveworks/flux/pull/1776)
 - Docker image is now based on Alpine `3.9`
   [weaveworks/flux#1801](https://github.com/weaveworks/flux/pull/1801)
 - `.spec.values` in the `HelmRelease` CRD is no longer mandatory
   [weaveworks/flux#1824](https://github.com/weaveworks/flux/pull/1824)
 - With `valuesFrom` it is now possible to load values from secrets,
   config maps and URLs
   [weaveworks/flux#1836](https://github.com/weaveworks/flux/pull/1836)

### Thanks

Thanks to @captncraig, @2opremio, @squaremo, @hiddeco, @endrec, @ahmadiq,
@nmaupu, @samisq, @yinzara, @stefanprodan, and @sarath-p for their
contributions.

## 0.6.0 (2019-02-07)

### Improvements

 - Add option to limit the Helm operator to a single namespace
   [weaveworks/flux#1664](https://github.com/weaveworks/flux/pull/1664)

### Thanks

Without the contributions of @brandon-bethke-neudesic, @errordeveloper,
@ncabatoff, @stefanprodan, @squaremo, and feedback of our
[#flux](https://slack.weave.works/) inhabitants this release would not
have been possible -- thanks to all of you!

## 0.5.3 (2019-01-14)

### Improvements

  - `HelmRelease` now has a `resetValues` field which when set to `true`
    resets the values to the ones built into the chart
    [weaveworks/flux#1628](https://github.com/weaveworks/flux/pull/1628)
  - The operator now exposes a HTTP webserver (by default on port
    `:3030`) with Prometheus metrics on `/metrics` and a health check
    endpoint on `/healthz`
    [weaveworks/flux#1653](https://github.com/weaveworks/flux/pull/1653)

### Thanks

A thousand thanks to @davidkarlsen, @hiddeco, @ncabatoff, @stefanprodan,
@squaremo and others for their contributions leading to this release.

## 0.5.2 (2018-12-20)

### Bug fixes

  - Respect proxy env entries for git operations
    [weaveworks/flux#1556](https://github.com/weaveworks/flux/pull/1556)
  - Reimplement git timeout after accidentally removing it in `0.5.0`
    [weaveworks/flux#1565](https://github.com/weaveworks/flux/pull/1565)
  - Mark `--git-poll-interval` flag as deprecated
    [weaveworks/flux#1565](https://github.com/weaveworks/flux/pull/1565)
  - Only update chart dependencies if a `requirements.yaml` exists
    weaveworks/flux{[#1561](https://github.com/weaveworks/flux/pull/1561), [#1606](https://github.com/weaveworks/flux/pull/1606)}
    
### Improvements

  - `HelmRelease` now has a `timeout` field (defaults to `300s`),
    giving you control over the amount of time it may take for Helm to
    install or upgrade your chart
    [weaveworks/flux#1566](https://github.com/weaveworks/flux/pull/1566)
  - The Helm operator [flag docs](docs/references/operator.md#setup-and-configuration)
    have been updated
    [weaveworks/flux#1594](https://github.com/weaveworks/flux/pull/1594)
  - Added tests to ensure Helm dependencies update behaviour is always as
    expected
    [weaveworks/flux#1562](https://github.com/weaveworks/flux/pull/1562)

### Thanks

Thanks to @stephenmoloney, @sfrique, @mgazza, @stefanprodan, @squaremo,
@rade and @hiddeco for their contributions.

## 0.5.1 (2018-11-21)

### Bug fixes

  - Helm releases will now stay put when an upgrade fails or the
    Kubernetes API connectivity is flaky, instead of getting purged
    [weaveworks/flux#1530](https://github.com/weaveworks/flux/pull/1530)

### Thanks

Thanks to @sfrique, @brantb and @squaremo for helping document the
issues leading to this bug fix, @stefanprodan for actually squashing
the bug and all others that may have gone unnoticed while writing this
release note.

## 0.5.0 (2018-11-14)

WARNING: this release of the Helm operator is not backward-compatible:

 - It uses a new custom resource `HelmRelease`, and will ignore
   `FluxHelmRelease` resources
 - Some command-line arguments have changed, so the [deployment
   manifests](deploy/) must also be updated

To use it, you will need to migrate custom resources to the new format
supported by this version. See the [upgrade
guide](docs/how-to/upgrade-to-beta.md).

This version of the Helm operator supports HelmRelease custom
resources, which each specify a chart and values to use in a Helm
release, as in previous versions. The main improvement is that you are
now able to specify charts from Helm repos, as well as from git repo,
per resource (rather than a single git repo, which is supplied to the
operator).

### Improvements

All of these were added in
[weaveworks/flux#1382](https://github.com/weaveworks/flux/pull/1382).

See the [Helm operator
guide](https://docs.fluxcd.io/en/latest/references/helm-operator-integration/)
for details.

 - You can now release charts from arbitrary Helm repos
 - You can now release charts from arbitrary git repos

### Thanks

Thanks to @demikl, @dholbach, @hiddeco, @mellana1, @squaremo,
@stefanprodan, @stephenmoloney, @whereismyjetpack and others who made
suggestions, logged problems, and tried out nightly builds.

## 0.4.0 (2018-11-01)

This release improves support for TLS connections to Tiller; in
particular it makes it much easier to get server certificate
verification (`--tiller-tls-verify`) to work.

It also adds the ability to supply additional values to
`FluxHelmRelease` resources by attaching Kubernetes secrets. This
helps with a few use cases:

 - supplying the same default values to several releases
 - providing secrets (e.g., a password) to a chart that expects them as values
 - using values files without inlining them into FluxHelmReleases

**NB** It is advised that you deploy the operator alongside Tiller
v2.10 or more recent. To properly support TLS, the operator now
includes code from Helm v2.10, and this may have difficulty connecting
to older versions of Tiller.

### Bug fixes

 - Make `--tiller-tls-verify` work as intended, by giving better
   instructions, and adding the argument `--tiller-tls-hostname` which
   lets you specify the hostname that TLS should expect in the
   certificate
   [weaveworks/flux#1484](https://github.com/weaveworks/flux/pull/1484)

### Improvements

 - You can now create secrets containing a `values.yaml` file, and
   attach them to a `FluxHelmRelease` as additional values to use
   [weaveworks/flux#1468](https://github.com/weaveworks/flux/pull/1468)

### Thanks

Thanks to @hiddeco, @Smirl, @stefanprodan, @arthurk, @the-fine,
@wstrange, @sfitts, @squaremo, @mpareja, @stephenmoloney,
@justinbarrick, @pcfens for contributions to the PRs and issues
leading to this release, as well as the inhabitants of
[#flux](https://slack.weave.works/) for high-quality, helpful
discussion.

## 0.3.0 (2018-10-24)

This release adds dependency handling to the Helm operator.

**NB** The helm operator will now update dependencies for charts _by
default_, which means you no longer need to vendor them. You can
switch this behaviour off with the flag `--update-chart-deps=false`.

### Bug fixes

 - Improve chance of graceful shutdown
   [weaveworks/flux#1439](https://github.com/weaveworks/flux/pull/1439)
   and
   [weaveworks/flux#1438](https://github.com/weaveworks/flux/pull/1438)
 
### Improvements

 - The operator now runs `helm dep build` for charts before installing
   or upgrading releases. This will use a lockfile if present, and
   update the dependencies according to `requirements.yaml` otherwise
   [weaveworks/flux#1450](https://github.com/weaveworks/flux/pull/1450)
 - A new flag `--git-timeout` controls how long the Helm operator will
   allow for git operations
   [weaveworks/flux#1416](https://github.com/weaveworks/flux/pull/1416)
 - The Helm operator image now includes the Helm command-line client,
   which makes it easier to troubleshoot problems using `kubectl exec`
   (as part of
   [weaveworks/flux#1450](https://github.com/weaveworks/flux/pull/1450))

## 0.2.1 (2018-09-17)

This is a patch release that allows helm-op to recover from a failed release install.
If a chart is broken, Tiller will reserve the name and mark the release as failed. 
If at a later time the chart is fixed, helm-op can't install it anymore because the release name is in use. 
Purging the release after each failed install allows helm-op to keep retrying the install.

- Purge release if install fails
  [weaveworks/flux#1344](https://github.com/weaveworks/flux/pull/1344)

## 0.2.0 (2018-08-23)

In large part this release simplifies and improves the Helm operator
machinery, without changing its effect.

This release drops the `-alpha` suffix, but remains <1.0 and should
(still) be considered unready for production use.

- Use the same git implementation as fluxd, fixing a number of
  problems with SSH known_hosts and git URLs and so on
  [weaveworks/flux#1240](https://github.com/weaveworks/flux/pull/1240)
- Always check that a chart release will be a change, before releasing
  [weaveworks/flux#1254](https://github.com/weaveworks/flux/pull/1254)
- Add validation to the FluxHelmRelease custom resource definition,
  giving the kind the short name `fhr`
  [weaveworks/flux#1253](https://github.com/weaveworks/flux/pull/1253)
- Detect chart release differences more reliably
  [weaveworks/flux#1272](https://github.com/weaveworks/flux/pull/1272)
- Check for more recent versions and report in logs when out of date
  [weaveworks/flux#1276](https://github.com/weaveworks/flux/pull/1276)

See [getting started with
Helm](docs/get-started/quickstart.md)
and the [Helm chart
instructions](https://github.com/weaveworks/flux/blob/master/chart/flux/README.md)
for information on installing the Flux with the Helm operator.

## 0.1.1-alpha (2018-07-16)

- Support using TLS connections to Tiller
  [weaveworks/flux#1200](https://github.com/weaveworks/flux/pull/1200)
- Avoid continual, spurious installs in newer Kubernetes
  [weaveworks/flux#1193](https://github.com/weaveworks/flux/pull/1193)
- Make it easier to override SSH config (and `known_hosts`)
  [weaveworks/flux#1188](https://github.com/weaveworks/flux/pull/1188)
- Annotate resources created by a Helm release with the name of the
  FluxHelmRelease custom resource, so they can be linked
  [weaveworks/flux#1134](https://github.com/weaveworks/flux/pull/1134)
- Purge release when FluxHelmRelease is deleted, so restoring the
  resource can succeed
  [weaveworks/flux#1106](https://github.com/weaveworks/flux/pull/1106)
- Correct permissions on baked-in SSH config
  [weaveworks/flux#1098](https://github.com/weaveworks/flux/pull/1098)
- Test coverage for releasesync package
  [weaveworks/flux#1089](https://github.com/weaveworks/flux/pull/1089)).

It is now possible to install Flux and the Helm operator using the
[helm chart in this
repository](https://github.com/weaveworks/flux/tree/master/chart/flux).

## 0.1.0-alpha (2018-05-01)

First versioned release of the Flux Helm operator. The target features are:

- release Helm charts as specified in FluxHelmRelease resources
  - these refer to charts in a single git repo, readable by the operator
  - update releases when either the FluxHelmRelease resource or the
    chart (in git) changes

See
https://github.com/weaveworks/flux/blob/helm-0.1.0-alpha/site/helm/
for more detailed explanations.
