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
  # for this test, we need the operator to be able to handle both v2 and v3 manifests
  HELM_ENABLED_VERSIONS="v2\,v3" install_helm_operator_with_helm
  kubectl create namespace "$DEMO_NAMESPACE"
}

@test "Migration succeeds from v2 to v3" {
  # Apply the HelmRelease
  kubectl apply -f "$FIXTURES_DIR/releases/convert-2to3-v2.yaml" >&3

  # Wait for it to be released
  poll_until_equals 'release deploy' 'True' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o jsonpath='{.status.conditions[?(@.type==\"Released\")].status}'"

  poll_until_equals 'helm2 shows helm release' 'podinfo-helm-repository' "HELM_VERSION=v2 helm ls | grep podinfo-helm-repository | awk '{print \$1}'"

  kubectl apply -f "$FIXTURES_DIR/releases/convert-2to3-v3.yaml" >&3
  poll_until_equals 'helm2 no longer shows helm release' '0' "HELM_VERSION=v2 helm ls | grep podinfo-helm-repository | wc -l | awk '{\$1=\$1};1'"
  poll_until_equals 'helm3 shows helm release' 'podinfo-helm-repository' "HELM_VERSION=v3 helm ls -n $DEMO_NAMESPACE | grep podinfo-helm-repository | awk '{print \$1}'"
  poll_until_equals 'release migrated' 'True' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o jsonpath='{.status.conditions[?(@.type==\"Released\")].status}'"

  kubectl apply -f "$FIXTURES_DIR/releases/convert-2to3-v3-upgrade.yaml" >&3
  poll_until_equals 'upgrades work after migration' '1' "kubectl get deploy/podinfo-helm-repository -n "$DEMO_NAMESPACE" -o jsonpath='{.spec.replicas}'"
  poll_no_restarts
}

@test "Migration is skipped and install works when no v2 release exists" {
  kubectl apply -f "$FIXTURES_DIR/releases/convert-2to3-v3.yaml" >&3
  poll_until_equals 'helm3 shows helm release' 'podinfo-helm-repository' "HELM_VERSION=v3 helm ls -n $DEMO_NAMESPACE | grep podinfo-helm-repository | awk '{print \$1}'"
  poll_until_equals 'install successful' 'True' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o jsonpath='{.status.conditions[?(@.type==\"Released\")].status}'"
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
