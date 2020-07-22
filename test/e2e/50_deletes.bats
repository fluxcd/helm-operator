#!/usr/bin/env bats

function setup() {
  # Load libraries in setup() to access BATS_* variables
  load lib/env
  load lib/helm
  load lib/install
  load lib/poll

  kubectl create namespace "$E2E_NAMESPACE"
  install_gitsrv
  install_tiller
  install_helm_operator_with_helm
  kubectl create namespace "$DEMO_NAMESPACE"
}

@test "Deletes cleanup resources successfully" {
  # Apply the HelmRelease
  kubectl apply -f "$FIXTURES_DIR/releases/helm-repository.yaml" >&3

  # Wait for it to be released
  poll_until_equals 'release deploy' 'True' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o jsonpath='{.status.conditions[?(@.type==\"Released\")].status}'"

  kubectl delete -f "$FIXTURES_DIR/releases/helm-repository.yaml" >&3

  poll_until_true 'deployment is deleted' "kubectl get deploy -n $DEMO_NAMESPACE podinfo-helm-repository 2>&1 | grep 'NotFound'"

  poll_no_restarts
}

function teardown() {
  # Teardown is verbose when a test fails, and this will help most of the time
  # to determine _why_ it failed.
  echo ""
  echo "### Previous container:"
  kubectl logs -n "$E2E_NAMESPACE" deploy/helm-operator -p
  echo ""
  echo "### Current container:"
  kubectl logs -n "$E2E_NAMESPACE" deploy/helm-operator

  # Removing the operator also takes care of the global resources it installs.
  uninstall_helm_operator_with_helm
  uninstall_tiller
  # Removing the namespace also takes care of removing gitsrv.
  kubectl delete namespace "$E2E_NAMESPACE"
  # Only remove the demo workloads after the operator, so that they cannot be recreated.
  kubectl delete namespace "$DEMO_NAMESPACE"
}
