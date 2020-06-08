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

@test "When test.enable is set, tests are run" {
  # Apply the HelmRelease
  kubectl apply -f "$FIXTURES_DIR/releases/test/success.yaml" >&3

  # Wait for it to be released
  poll_until_equals 'release deploy' 'True' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o jsonpath='{.status.conditions[?(@.type==\"Released\")].status}'"

  if [ $HELM_VERSION = "v2" ]; then
    test_results_exist_jq_filter='.info.status | has("last_test_suite_run")'
    extra_args=""
  else
    test_results_exist_jq_filter='.hooks | map(.events) | map(select(contains(["test"]))) | length > 0'
    extra_args="--namespace $DEMO_NAMESPACE"
  fi

  # Wait for test results to exist
  poll_until_equals 'test results exist' 'true' "helm status $extra_args podinfo-helm-repository -o json | jq -r '$test_results_exist_jq_filter'"

  # Wait for `Tested` condition to be `True`
  poll_until_equals 'release deploy' 'True' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o jsonpath='{.status.conditions[?(@.type==\"Tested\")].status}'"
}

@test "When test.enable is set, releases with failed tests are uninstalled" {
  # Apply the HelmRelease
  kubectl apply -f "$FIXTURES_DIR/releases/test/fail.yaml" >&3

  # Wait for test failure
  poll_until_true 'test failure' "kubectl -n $E2E_NAMESPACE logs deploy/helm-operator | grep -E \"test failed\""

  # Assert release uninstalled
  # TODO: Poll `helm ls` results directly for release removal once install retries can be disabled.
  poll_until_true 'release uninstalled' "kubectl -n $E2E_NAMESPACE logs deploy/helm-operator | grep -E \"running uninstall\""
}

# TODO: Fail tests on install instead of upgrade once install retries can be disabled.
@test "When tests fail, Tested and Released conditions are False" {
  # Apply the HelmRelease
  kubectl apply -f "$FIXTURES_DIR/releases/test/success.yaml" >&3

  # Wait for it to be deployed
  poll_until_equals 'release deploy' 'True' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o jsonpath='{.status.conditions[?(@.type==\"Released\")].status}'"

  # Apply a patch which causes helm tests to fail
  kubectl apply -f "$FIXTURES_DIR/releases/test/fail.yaml" >&3

  # Wait for test failure
  poll_until_true 'test failure' "kubectl -n $E2E_NAMESPACE logs deploy/helm-operator | grep -E \"test failed\""

  # Assert `Released` condition becomes `False`
  poll_until_equals 'released condition false' 'False' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o jsonpath='{.status.conditions[?(@.type==\"Released\")].status}'"

  # Assert `Tested` condition is `False`
  run kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o jsonpath='{.status.conditions[?(@.type=="Tested")].status}'
  [ "$output" = 'False' ]
}

@test "When test.enable and rollback.enable are set, releases with failed tests are rolled back" {
  # Apply the HelmRelease
  kubectl apply -f "$FIXTURES_DIR/releases/test/success.yaml" >&3

  # Wait for it to be deployed
  poll_until_equals 'release deploy' 'True' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o jsonpath='{.status.conditions[?(@.type==\"Released\")].status}'"

  # Apply a patch which causes helm tests to fail
  kubectl apply -f "$FIXTURES_DIR/releases/test/fail.yaml" >&3

  # Wait for test failure
  poll_until_true 'test failure' "kubectl -n $E2E_NAMESPACE logs deploy/helm-operator | grep -E \"test failed\""

  # Wait for rollback
  poll_until_equals 'rollback' 'True' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o jsonpath='{.status.conditions[?(@.type==\"RolledBack\")].status}'"

  # Apply fix patch
  kubectl apply -f "$FIXTURES_DIR/releases/test/success.yaml" >&3

  # Assert recovery
  poll_until_equals 'recovery' 'True' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o jsonpath='{.status.conditions[?(@.type==\"Released\")].status}'"
}

@test "When test.enable and rollback.enable are set, releases with timed out tests are rolled back" {
  # Apply the HelmRelease
  kubectl apply -f "$FIXTURES_DIR/releases/test/success.yaml" >&3

  # Wait for it to be deployed
  poll_until_equals 'release deploy' 'True' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o jsonpath='{.status.conditions[?(@.type==\"Released\")].status}'"

  # Apply a patch which causes helm tests to fail
  kubectl apply -f "$FIXTURES_DIR/releases/test/timeout.yaml" >&3

  # Wait for test failure
  poll_until_true 'test failure' "kubectl -n $E2E_NAMESPACE logs deploy/helm-operator | grep -E \"test failed\""

  # Wait for rollback
  poll_until_equals 'rollback' 'True' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o jsonpath='{.status.conditions[?(@.type==\"RolledBack\")].status}'"

  # Apply fix patch
  kubectl apply -f "$FIXTURES_DIR/releases/test/success.yaml" >&3

  # Assert recovery
  poll_until_equals 'recovery' 'True' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o jsonpath='{.status.conditions[?(@.type==\"Released\")].status}'"
}

function teardown() {
  # Teardown is verbose when a test fails, and this will help most of the time
  # to determine _why_ it failed.
  kubectl logs -n "$E2E_NAMESPACE" deploy/helm-operator

  # Removing the operator also takes care of the global resources it installs.
  uninstall_helm_operator_with_helm
  uninstall_tiller
  # Removing the namespace also takes care of removing gitsrv.
  kubectl delete namespace "$E2E_NAMESPACE"
  # Only remove the demo workloads after the operator, so that they cannot be recreated.
  kubectl delete namespace "$DEMO_NAMESPACE"
}
