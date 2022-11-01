# Troubleshooting

> **🛑 Upgrade Advisory**
>
> This documentation is for Helm Operator (v1) which has [reached its end-of-life in November 2022](https://fluxcd.io/blog/2022/10/september-2022-update/#flux-legacy-v1-retirement-plan).
>
> We strongly recommend you familiarise yourself with the newest Flux and [migrate as soon as possible](https://fluxcd.io/flux/migration/).
>
> For documentation regarding the latest Flux, please refer to [this section](https://fluxcd.io/flux/).

Also see the [issues labeled with
`question`](https://github.com/fluxcd/helm-operator/labels/question), which often
explain workarounds.

## `failed to load chart from path [/tmp/flux-working012345678/chart/path] for release [x]: stat /tmp/flux-working012345678/chart/path: no such file or directory`

The `.chart.path` of your Git sourced chart in the `HelmRelease` is likely
incorrect. It is expected to be a path to the chart folder relative to the
root directory of the Git repository.

## `Error creating helm client: failed to append certificates from file: /etc/fluxd/helm-ca/ca.crt`

Your CA certificate content is not set correctly, check if your ConfigMap contains the correct values. Example:

```bash
$ kubectl get configmaps flux-helm-tls-ca-config -o yaml
apiVersion: v1
data:
  ca.crt: |
    -----BEGIN CERTIFICATE-----
    ....
    -----END CERTIFICATE-----
kind: ConfigMap
metadata:
  creationTimestamp: 2018-07-04T15:27:25Z
  name: flux-helm-tls-ca-config
  namespace: helm-system
  resourceVersion: "1267257"
  selfLink: /api/v1/namespaces/helm-system/configmaps/flux-helm-tls-ca-config
  uid: c106f866-7f9e-11e8-904a-025000000001
```
