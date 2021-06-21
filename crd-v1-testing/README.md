# Verification in new cluster

Steps to verify CRD V1 functionality in a 1.20 cluster.

Create new kind cluster with crd/v1beta1 disabled.
```
kind create cluster --config ./kind-config.yaml
```

Apply local CRDs.
```
kubectl apply -f chart/helm-operator/crds/helmrelease.yaml
```

Install Helm Operator without the charts CRDs.
```
helm repo add fluxcd https://charts.fluxcd.io
kubectl create ns flux
helm upgrade --skip-crds -i helm-operator fluxcd/helm-operator --namespace flux --set helm.versions=v3
```

Apply podinfo helm chart, and verify it is working.
```
kubectl apply -f crd-v1-testing/podinfo-helmrelease.yaml
kubectl -n default get pods
kubectl -n default port-forward pod/<POD_NAME>
curl http://localhost:9898/
```

# Verification in upgraded cluster

Create a new cluster with v1beta1 CRD enabled.
```
kind create cluster`
```

Install Helm Operator.
```
helm repo add fluxcd https://charts.fluxcd.io
kubectl create ns flux
helm upgrade -i helm-operator fluxcd/helm-operator --namespace flux --set helm.versions=v3
```

Install chart museum.
```
helm repo add stable https://charts.helm.sh/stable
helm upgrade -i chartmuseum stable/chartmuseum -f crd-v1-testing/chart-museum-values.yaml
kubectl -n default get pods
```

Build and upload helm package.
```
kubectl -n default port-forward pod/<POD_NAME> 8080:8080
helm plugin install https://github.com/chartmuseum/helm-push.git
helm push crd-v1-testing/crd-test chartmuseum
```

Apply crd-test helm release
```
kubectl apply -f crd-v1-testing/crd-test-helmrelease.yaml
```

Exec into the docker container running the node.
```
docker ps
docker exec -i <CONTAINER_ID> bash
```

Get the kubeadm config.
```
kubeadm config view > kubeadm-config.yaml
```

Update the config with runtime config.
```
extraArgs:
  authorization-mode: Node,RBAC
  runtime-config: ""
```

```
extraArgs:
  authorization-mode: Node,RBAC
  runtime-config: "apiextensions.k8s.io/v1beta1=false"
```

Upgrade the cluster with the new settings.
```
kubeadm upgrade diff --config kubeadm-config.yaml
kubeadm upgrade apply --config kubeadm-config.yaml --ignore-preflight-errors all --force --v=5
```
