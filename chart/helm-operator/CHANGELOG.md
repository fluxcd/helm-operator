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
