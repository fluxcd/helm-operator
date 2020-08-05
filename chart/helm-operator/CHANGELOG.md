## 1.2.0 (2020-08-05)

### Improvements

 - Update Helm Operator to `1.2.0`
   [fluxcd/helm-operator#505](https://github.com/fluxcd/helm-operator/pull/505)
 - Adding securityContext options at the pod and container level
   [fluxcd/helm-operator#494](https://github.com/fluxcd/helm-operator/pull/494)
 - Enable additional sidecar containers
   [fluxcd/helm-operator#484](https://github.com/fluxcd/helm-operator/pull/484)
 - Add grafana helm-operator dashboard
   [fluxcd/helm-operator#482](https://github.com/fluxcd/helm-operator/pull/482)
 - Roll deployment on update known_hosts
   [fluxcd/helm-operator#441](https://github.com/fluxcd/helm-operator/pull/441)

### Bug fixes

 - Fix kubeconfig override in chart/deployment
   [fluxcd/helm-operator#507](https://github.com/fluxcd/helm-operator/pull/507)
 - Documentation fixes
   [fluxcd/helm-operator#502](https://github.com/fluxcd/helm-operator/pull/502)
   [fluxcd/helm-operator#451](https://github.com/fluxcd/helm-operator/pull/451)
   [fluxcd/helm-operator#438](https://github.com/fluxcd/helm-operator/pull/438)
   
## 1.1.0 (2020-05-28)

### Improvements

 - Update Helm Operator to `1.1.0`
   [fluxcd/helm-operator#424](https://github.com/fluxcd/helm-operator/pull/424)
 - Make probes configurable on the deployment manifest
   [fluxcd/helm-operator#397](https://github.com/fluxcd/helm-operator/pull/397)
 - Update version of example Redis chart in `NOTES.txt` to `10.5.7`
   [fluxcd/helm-operator#430](https://github.com/fluxcd/helm-operator/pull/430)

## 1.0.2 (2020-05-06)

### Improvements

 - Add `priorityClassName` setting
   [fluxcd/helm-operator#401](https://github.com/fluxcd/helm-operator/pull/401)
 - Surpress creation of ClusterRole(Binding)s when `clusterRole.create`
   is set to `false`.
   [fluxcd/helm-operator#402](https://github.com/fluxcd/helm-operator/pull/402)

## 1.0.1 (2020-04-16)

### Improvements

 - Update Helm Operator to `1.0.1`
   [fluxcd/helm-operator#375](https://github.com/fluxcd/helm-operator/pull/375)
 - Revert default memory limit
   [fluxcd/helm-operator#373](https://github.com/fluxcd/helm-operator/pull/373)
 - Use `Recreate` strategy
   [fluxcd/helm-operator#374](https://github.com/fluxcd/helm-operator/pull/374)
   
## 1.0.0 (2020-04-08)

### Improvements

 - Update Helm Operator to `1.0.0`
   [fluxcd/helm-operator#354](https://github.com/fluxcd/helm-operator/pull/354)
 - Tune chart defaults for production use
   [fluxcd/helm-operator#353](https://github.com/fluxcd/helm-operator/pull/353)
 - Add hostAliases to chart options
   [fluxcd/helm-operator#328](https://github.com/fluxcd/helm-operator/pull/328)

## 0.7.0 (2020-02-14)

### Improvements

 - Update Helm Operator to `1.0.0-rc9`
   [fluxcd/helm-operator#298](https://github.com/fluxcd/helm-operator/pull/298)
 - Allow init containers to be set
   [fluxcd/helm-operator#276](https://github.com/fluxcd/helm-operator/pull/276)
 - Support installation of Helm plugins
   [fluxcd/helm-operator#276](https://github.com/fluxcd/helm-operator/pull/276)
 - Support Helm 3 CRD installation
   [fluxcd/helm-operator#287](https://github.com/fluxcd/helm-operator/pull/287)

## 0.6.0 (2020-01-26)

### Improvements

 - Update Helm Operator to `1.0.0-rc8`
   [fluxcd/helm-operator#244](https://github.com/fluxcd/helm-operator/pull/244)
 - Allow pod annotations, labels and account annotations to be set
   [fluxcd/helm-operator#229](https://github.com/fluxcd/helm-operator/pull/229)

## 0.5.0 (2020-01-10)

### Improvements

 - Update Helm Operator to `1.0.0-rc7`
   [fluxcd/helm-operator#197](https://github.com/fluxcd/helm-operator/pull/197)
 - Add support for configuring cert files for repositories
   [fluxcd/helm-operator#183](https://github.com/fluxcd/helm-operator/pull/183)
 - Add support for configuring Helm v3 repositories
   [fluxcd/helm-operator#173](https://github.com/fluxcd/helm-operator/pull/173)
 - Add Prometheus Operator ServiceMonitor templates
   [fluxcd/helm-operator#139](https://github.com/fluxcd/helm-operator/pull/139)

## 0.4.0 (2019-12-23)

### Improvements

 - Add `helm.versions` option to chart values
   [fluxcd/helm-operator#159](https://github.com/fluxcd/helm-operator/pull/159)
 - Update Helm Operator to `1.0.0-rc5`
   [fluxcd/helm-operator#157](https://github.com/fluxcd/helm-operator/pull/157)
 - Add Service and ServiceMonitor templates
   [fluxcd/helm-operator#139](https://github.com/fluxcd/helm-operator/pull/139)
 - Add extraVolumes and extraVolumeMounts
   [fluxcd/helm-operator#125](https://github.com/fluxcd/helm-operator/pull/125)

## 0.3.0 (2019-11-22)

### Improvements

 - Update Helm Operator to `1.0.0-rc4`
   [fluxcd/helm-operator#114](https://github.com/fluxcd/helm-operator/pull/114)
 - Fix upgrade command install instructions in `README.md`
   [fluxcd/helm-operator#92](https://github.com/fluxcd/helm-operator/pull/92)
 - Add `git.defaultRef` option for configuring an alternative Git default ref
   [fluxcd/helm-operator#83](https://github.com/fluxcd/helm-operator/pull/83)
 - Allow for deploying Tiller as a sidecar by setting `tillerSidecar.enabled`
   [fluxcd/helm-operator#79](https://github.com/fluxcd/helm-operator/pull/79)

## 0.2.1 (2019-10-18)

### Improvements

 - Update Helm Operator to `1.0.0-rc3`
   [fluxcd/helm-operator#74](https://github.com/fluxcd/helm-operator/pull/74)

## 0.2.0 (2019-10-07)

### Improvements

 - Update Helm Operator to `1.0.0-rc2`
   [fluxcd/helm-operator#59](https://github.com/fluxcd/helm-operator/pull/59)
 - Expand the list of public Helm repositories in the default config
   [fluxcd/helm-operator#53](https://github.com/fluxcd/helm-operator/pull/53)
 - Add `statusUpdateInterval` option for configuring the interval at which the operator consults Tiller for the status of a release
   [fluxcd/helm-operator#44](https://github.com/fluxcd/helm-operator/pull/44)

## 0.1.1 (2019-09-15)

### Improvements

 - Restart operator on helm repositories changes
   [fluxcd/helm-operator#30](https://github.com/fluxcd/helm-operator/pull/30)
 - Add liveness and readiness probes
   [fluxcd/helm-operator#30](https://github.com/fluxcd/helm-operator/pull/30)
 - Add `HelmRelease` example to chart notes
   [fluxcd/helm-operator#30](https://github.com/fluxcd/helm-operator/pull/30)

### Bug fixes

 - Fix SSH key mapping
   [fluxcd/helm-operator#30](https://github.com/fluxcd/helm-operator/pull/30)

## 0.1.0 (2019-09-14)

Initial chart release with Helm Operator [1.0.0-rc1](https://github.com/fluxcd/helm-operator/blob/master/CHANGELOG.md#100-rc1-2019-08-14)
