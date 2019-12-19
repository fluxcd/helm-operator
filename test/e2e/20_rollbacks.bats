#!/usr/bin/env bats

function setup() {
  # Load libraries in setup() to access BATS_* variables
  load lib/env
  load lib/install
  load lib/poll

  kubectl create namespace "$E2E_NAMESPACE"
  install_git_srv
  install_tiller
  install_helm_operator_with_helm
  kubectl create namespace "$DEMO_NAMESPACE"
}

@test "Rollback Helm repository release" {
  # Apply the HelmRelease
  kubectl apply -f "$FIXTURES_DIR/releases/helm-repository.yaml" >&3

  # Wait for it to be deployed
  poll_until_equals 'podinfo-helm-repository HelmRelease' 'deployed' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o 'custom-columns=status:status.releaseStatus' --no-headers"

  # Apply a faulty patch
  kubectl patch -f "$FIXTURES_DIR/releases/helm-repository.yaml" --type='json' -p='[{"op": "replace", "path": "/spec/values/service/type", "value":"LoadBalancer"}]' >&3

  # Wait for release failure
  poll_until_equals 'podinfo-helm-repository HelmRelease upgrade failure' 'HelmUpgradeFailed' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o jsonpath='{.status.conditions[?(@.type==\"Released\")].reason}'"

  # Wait for rollback
  poll_until_equals 'podinfo-helm-repository HelmRelease rollback' 'True' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o jsonpath='{.status.conditions[?(@.type==\"RolledBack\")].status}'"

  # Apply fix patch
  kubectl apply -f "$FIXTURES_DIR/releases/helm-repository.yaml" >&3

  # Assert recovery
  poll_until_equals 'podinfo-helm-repository HelmRelease recovery' 'HelmSuccess' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o jsonpath='{.status.conditions[?(@.type==\"Released\")].reason}'"
}

@test "Rollback git release" {
  # Apply the HelmRelease
  kubectl apply -f "$FIXTURES_DIR/releases/git.yaml" >&3

  # Wait for it to be deployed
  poll_until_equals 'podinfo-git HelmRelease' 'deployed' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-git -o 'custom-columns=status:status.releaseStatus' --no-headers"

  # Apply a faulty patch
  kubectl patch -f "$FIXTURES_DIR/releases/git.yaml" --type='json' -p='[{"op": "replace", "path": "/spec/values/service/type", "value":"LoadBalancer"}]' >&3

  # Wait for release failure
  poll_until_equals 'podinfo-git HelmRelease upgrade failure' 'HelmUpgradeFailed' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-git -o jsonpath='{.status.conditions[?(@.type==\"Released\")].reason}'"

  # Wait for rollback
  poll_until_equals 'podinfo-git HelmRelease rollback' 'True' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-git -o jsonpath='{.status.conditions[?(@.type==\"RolledBack\")].status}'"

  # Apply fix patch
  kubectl apply -f "$FIXTURES_DIR/releases/git.yaml" >&3

  # Assert recovery
  poll_until_equals 'podinfo-git HelmRelease recovery' 'HelmSuccess' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-git -o jsonpath='{.status.conditions[?(@.type==\"Released\")].reason}'"
}

@test "Validation error does not trigger rollback" {
  if [ "$HELM_VERSION" != "v3" ]; then
    skip
  fi

  # Apply the HelmRelease
  kubectl apply -f "$FIXTURES_DIR/releases/git.yaml" >&3

  # Wait for it to be deployed
  poll_until_equals 'podinfo-git HelmRelease' 'deployed' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-git -o 'custom-columns=status:status.releaseStatus' --no-headers"

  # Apply a faulty patch
  kubectl patch -f "$FIXTURES_DIR/releases/git.yaml" --type='json' -p='[{"op": "replace", "path": "/spec/values/replicaCount", "value":"faulty"}]' >&3

  # Wait for release failure
  poll_until_equals 'podinfo-git HelmRelease upgrade failure' 'HelmUpgradeFailed' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-git -o jsonpath='{.status.conditions[?(@.type==\"Released\")].reason}'"

  # Assert release version
  version=$(kubectl exec -n "$E2E_NAMESPACE" deploy/helm-operator -- helm3 status podinfo-git --namespace "$DEMO_NAMESPACE" -o json | jq .version)
  [ "$version" -eq 1 ]

  # Apply fix patch
  kubectl apply -f "$FIXTURES_DIR/releases/git.yaml" >&3

  # Assert recovery
  poll_until_equals 'podinfo-git HelmRelease recovery' 'HelmSuccess' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-git -o jsonpath='{.status.conditions[?(@.type==\"Released\")].reason}'"

  # Assert release version
  version=$(kubectl exec -n "$E2E_NAMESPACE" deploy/helm-operator -- helm3 status podinfo-git --namespace "$DEMO_NAMESPACE" -o json | jq .version)
  [ "$version" -eq 2 ]
}

function teardown() {
  # Removing the operator also takes care of the global resources it installs.
  uninstall_helm_operator_with_helm
  uninstall_tiller
  # Removing the namespace also takes care of removing gitsrv.
  kubectl delete namespace "$E2E_NAMESPACE"
  # Only remove the demo workloads after the operator, so that they cannot be recreated.
  kubectl delete namespace "$DEMO_NAMESPACE"
}
