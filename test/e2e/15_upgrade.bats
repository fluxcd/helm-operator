#!/usr/bin/env bats

function setup() {
  # Load libraries in setup() to access BATS_* variables
  load lib/env
  load lib/defer
  load lib/install
  load lib/poll

  kubectl create namespace "$E2E_NAMESPACE"

  # Install the git server, allowing external access
  install_git_srv git_srv_result
  # shellcheck disable=SC2154
  export GIT_SSH_COMMAND="${git_srv_result[0]}"
  # Teardown the created port-forward to gitsrv.
  defer kill "${git_srv_result[1]}"

  install_tiller
  install_helm_operator_with_helm

  kubectl create namespace "$DEMO_NAMESPACE"
}

@test "Git mituation causes upgrade" {
  # Apply the HelmRelease fixtures
  kubectl apply -f "$FIXTURES_DIR/releases/git.yaml" >&3

  # Wait for it to be deployed
  poll_until_equals 'podinfo-git HelmRelease' 'deployed' "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-git -o 'custom-columns=status:status.releaseStatus' --no-headers"

  # Clone the charts repository
  local clone_dir
  clone_dir="$(mktemp -d)"
  defer rm -rf "'$clone_dir'"
  git clone -b master ssh://git@localhost/git-server/repos/cluster.git "$clone_dir"
  cd "$clone_dir"

  # Make a chart template mutation in Git without bumping the version number
  sed -i 's%these commands:%these commands;%' charts/podinfo/templates/NOTES.txt
  git add charts/podinfo/templates/NOTES.txt
  git -c 'user.email=foo@bar.com' -c 'user.name=Foo' commit -m "Modify NOTES.txt"

  # Record new HEAD and push change
  head_hash=$(git rev-list -n 1 HEAD)
  git push >&3

  # Assert change is rolled out
  poll_until_equals 'podinfo-git HelmRelease chart update' "successfully cloned chart revision: $head_hash" "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-git -o jsonpath='{.status.conditions[?(@.type==\"ChartFetched\")].message}'"
  poll_until_equals 'podinfo-git HelmRelease revision matches' "$head_hash" "kubectl -n $DEMO_NAMESPACE get helmrelease/podinfo-git -o jsonpath='{.status.revision}'"
}

function teardown() {
  run_deffered

  # Removing the operator also takes care of the global resources it installs.
  uninstall_helm_operator_with_helm
  uninstall_tiller
  # Removing the namespace also takes care of removing gitsrv.
  kubectl delete namespace "$E2E_NAMESPACE"
  # Only remove the demo workloads after the operator, so that they cannot be recreated.
  kubectl delete namespace "$DEMO_NAMESPACE"
}
