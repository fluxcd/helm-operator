#!/usr/bin/env bats

function setup() {
  # Load libraries in setup() to access BATS_* variables
  load lib/env
  load lib/install
  load lib/poll

  kubectl create namespace "$E2E_NAMESPACE"
  install_gitsrv
  install_tiller
  install_helm_operator_with_helm
  kubectl create namespace "$DEMO_NAMESPACE"
}

@test "When max history is not set on release the most recent 10 revisions are kept" {
  if [ "$HELM_VERSION" != "v3" ]; then
    skip
  fi

  # Apply the HelmRelease
  kubectl apply -f "$FIXTURES_DIR/releases/helm-repository.yaml" >&3
  poll_until_equals 'release deploy' 'deployed' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o 'custom-columns=status:status.releaseStatus' --no-headers"

  for i in {1..15}
  do
    # Apply a patch to initiate an upgrade
    kubectl patch -f "$FIXTURES_DIR/releases/helm-repository.yaml" --type='json' -p="[{\"op\": \"replace\", \"path\": \"/spec/values/someField\", \"value\":\"$i\"}]" >&3
    poll_until_equals 'patch deploy' "$(($i+1))" "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o 'custom-columns=status:status.observedGeneration' --no-headers"
  done

  # Check count of revisions is <= 10
  count=$(helm history podinfo-helm-repository --namespace "$DEMO_NAMESPACE" --skip-headers | tail -n +2 | wc -l)
  [ "$count" -eq 10 ]

  poll_no_restarts
}

@test "When max history on release is set to 5 the most recent 5 revisions are kept" {
  if [ "$HELM_VERSION" != "v3" ]; then
    skip
  fi

  # Apply the HelmRelease
  kubectl apply -f "$FIXTURES_DIR/releases/helm-repository.yaml" >&3
  poll_until_equals 'release deploy' 'deployed' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o 'custom-columns=status:status.releaseStatus' --no-headers"

  # Patch HelmRelease to set maxHistory to 5
  kubectl patch -f "$FIXTURES_DIR/releases/helm-repository.yaml" --type='json' -p='[{"op": "replace", "path": "/spec/maxHistory", "value":5}]' >&3
  poll_until_equals 'max history deploy' '2' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o 'custom-columns=status:status.observedGeneration' --no-headers"

  for i in {1..8}
  do
    # Apply a patch to initiate an upgrade
    kubectl patch -f "$FIXTURES_DIR/releases/helm-repository.yaml" --type='json' -p="[{\"op\": \"replace\", \"path\": \"/spec/values/someField\", \"value\":\"$i\"}]" >&3
    poll_until_equals 'patch deploy' "$(($i+2))" "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-helm-repository -o 'custom-columns=status:status.observedGeneration' --no-headers"
  done

  # Check count of revisions is <= 5
  count=$(helm history podinfo-helm-repository --namespace "$DEMO_NAMESPACE" --skip-headers | tail -n +2 | wc -l)
  [ "$count" -eq 5 ]

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
