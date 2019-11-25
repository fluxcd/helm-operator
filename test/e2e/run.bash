#!/usr/bin/env bash

set -o errexit

# This script runs the bats tests, first ensuring there is a kubernetes
# cluster available, with a namespace and a git secret ready to use

# Directory paths we need to be aware of
ROOT_DIR="$(git rev-parse --show-toplevel)"
E2E_DIR="${ROOT_DIR}/test/e2e"
CACHE_DIR="${ROOT_DIR}/cache/$CURRENT_OS_ARCH"

KIND_VERSION="v0.5.1"
KIND_CACHE_PATH="${CACHE_DIR}/kind-$KIND_VERSION"
KIND_CLUSTER=helm-operator-e2e
USING_KIND=false

# shellcheck disable=SC1090
source "${E2E_DIR}/lib/defer.bash"
trap run_deferred EXIT

# Check if there is a kubernetes cluster running, otherwise use Kind
if ! kubectl version > /dev/null 2>&1; then
  if [ ! -f "${KIND_CACHE_PATH}" ]; then
    echo '>>> Downloading Kind'
    mkdir -p "${CACHE_DIR}"
    curl -sL "https://github.com/kubernetes-sigs/kind/releases/download/${KIND_VERSION}/kind-${CURRENT_OS_ARCH}" -o "${KIND_CACHE_PATH}"
  fi
  echo '>>> Creating Kind Kubernetes cluster'
  cp "${KIND_CACHE_PATH}" "${ROOT_DIR}/test/bin/kind"
  chmod +x "${ROOT_DIR}/test/bin/kind"
  kind create cluster --name "${KIND_CLUSTER}" --wait 5m
  defer kind --name "${KIND_CLUSTER}" delete cluster > /dev/null 2>&1 || true
  KUBECONFIG="$(kind --name="${KIND_CLUSTER}" get kubeconfig-path)"
  export KUBECONFIG
  USING_KIND=true
  kubectl get pods --all-namespaces
fi

if [ "${USING_KIND}" = 'true' ]; then
  echo '>>> Loading images into the Kind cluster'
  kind --name "${KIND_CLUSTER}" load docker-image 'docker.io/fluxcd/helm-operator:latest'
fi

echo '>>> Running the tests'
# Run all tests by default but let users specify which ones to run, e.g. with E2E_TESTS='11_*' make e2e
E2E_TESTS=${E2E_TESTS:-.}
HELM_VERSION=${HELM_VERSION:-}
(
  cd "${E2E_DIR}"
  export HELM_VERSION=${HELM_VERSION}
  # shellcheck disable=SC2086
  "${E2E_DIR}/bats/bin/bats" -t ${E2E_TESTS}
)
