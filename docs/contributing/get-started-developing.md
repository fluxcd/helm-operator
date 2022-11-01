---
weight: 30
title: Get started developing
---

This guide shows you how to make a small change to the Helm Operator and then build and test that change locally using a Kind cluster.

## Prepare your environment

To get started you will need to prepare your development environment, the following will need to be installed:

- [docker](https://docs.docker.com/install/)
- [go](https://golang.org/doc/install)
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [helm](https://helm.sh/docs/intro/quickstart/)

### Configure Kind

You will want to create a Kind cluster with a [local registry](https://kind.sigs.k8s.io/docs/user/local-registry/) so that you have somewhere to easily push your helm-operator image without relying on an external registry, you can do this by running the following:

```bash
KIND_CLUSTER_NAME=helm-operator \
  sh -c "$(curl -sSL https://kind.sigs.k8s.io/examples/kind-with-registry.sh)"
```

You should now have your local development environment ready to make a change to the Helm Operator, continue on to see how to make your first change and deploy it to the Kind cluster.

## Make the change

Your first change will involve adding a new field to the HelmRelease CRD that will simply output a custom log message when applying the release, first you will need to modify the CRD definition to add the new field:

```diff
--- a/pkg/apis/helm.fluxcd.io/v1/types_helmrelease.go
+++ b/pkg/apis/helm.fluxcd.io/v1/types_helmrelease.go
@@ -420,6 +420,9 @@ type HelmReleaseSpec struct {
        // Values holds the values for this Helm release.
        // +optional
        Values HelmValues `json:"values,omitempty"`
+       // A custom message to emit when applying the release.
+       // +optional
+       EchoMessage *string `json:"echoMessage,omitempty"`
 }
```

Now modify the sync process to output your message if it has been set:

```diff
--- a/pkg/release/release.go
+++ b/pkg/release/release.go
@@ -74,6 +74,10 @@ func (r *Release) Sync(hr *v1.HelmRelease) (err error) {
        logger := releaseLogger(r.logger, client, hr)
        logger.Log("info", "starting sync run")
 
+       if hr.Spec.EchoMessage != nil {
+               logger.Log("info", *hr.Spec.EchoMessage)
+       }
+
        chart, cleanup, err := r.prepareChart(client, hr)
```

That should be all that is needed for your first change, you can now move on to building and pushing the image so that you can see your changes in action.

## Building

The next step is to build everything, as you have made changes to the CRD you will first want to run the code-generation tasks:

```bash
make generate
```

Now you can build everything else including the Docker image:

```bash
make all
```

## Pushing the image

Once everything is successfully built you should be ready to push the image to your local Kind registry, you will have to re-tag the image and then push it to the local registry:

```bash
# Tag the image
docker tag \
  "fluxcd/helm-operator:$(./docker/image-tag)" \
  "localhost:5000/helm-operator:$(./docker/image-tag)"

# Push the image
docker push \
  "localhost:5000/helm-operator:$(./docker/image-tag)"
```

## Deploying your changes

You can now deploy your changes to the Kind cluster using your newly built and pushed image:

```bash
# Create fluxcd namespace
kubectl create namespace fluxcd

# Apply HelmRelease CRD
kubectl apply -f deploy/crds.yaml

# Install helm-operator using pushed image
helm upgrade -i helm-operator fluxcd/helm-operator \
  --namespace fluxcd \
  --set helm.versions=v3 \
  --set image.repository=localhost:5000/helm-operator \
  --set image.tag="$(./docker/image-tag)"
```

You should now have an instance of the Helm Operator running with your changes!

## See your changes in action

To see your changes in action, first deploy a HelmRelease setting the new `echoMessage` field:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: helm.fluxcd.io/v1
kind: HelmRelease
metadata:
  name: podinfo
  namespace: default
spec:
  echoMessage: This is a test
  releaseName: podinfo
  chart:
    repository: https://stefanprodan.github.io/podinfo
    version: 2.1.0
    name: podinfo
  values:
    replicaCount: 1
EOF
```

Once the release is deployed you can view the helm-operator pods logs to see your emitted message:

```bash
> kubectl logs -n fluxcd --tail=25 --selector=app=helm-operator

...
ts=2020-04-15T20:35:16.249414781Z caller=release.go:78 component=release release=podinfo targetNamespace=default resource=default:helmrelease/podinfo helmVersion=v3 info="This is a test"
ts=2020-04-15T20:35:16.540354042Z caller=release.go:272 component=release release=podinfo targetNamespace=default resource=default:helmrelease/podinfo helmVersion=v3 info="running installation" phase=install
ts=2020-04-15T20:35:16.89689355Z caller=helm.go:69 component=helm version=v3 info="creating 3 resource(s)" targetNamespace=default release=podinfo
ts=2020-04-15T20:35:16.938861204Z caller=release.go:281 component=release release=podinfo targetNamespace=default resource=default:helmrelease/podinfo helmVersion=v3 info="installation succeeded" revision=2.1.0 phase=install
...
```

## Congratulations!

You did it! you made your first change to the HelmRelease CRD, built, deployed and ran an instance of the Helm Operator to see it in action.

You should now hopefully be more comfortable with making changes and running the Helm Operator locally, and be ready to tackle your [first issue](https://github.com/fluxcd/helm-operator/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22).

To find out more about the Helm Operator community and our contribution workflow have a look at the [contributing guide](introduction.md).
