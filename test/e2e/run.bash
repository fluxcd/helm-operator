#!/usr/bin/env bash

set -o errexit

# This script runs the bats tests, first ensuring there is a kubernetes
# cluster available, with a namespace and a git secret ready to use

# Directory paths we need to be aware of
ROOT_DIR="$(git rev-parse --show-toplevel)"
E2E_DIR="${ROOT_DIR}/test/e2e"
CACHE_DIR="${ROOT_DIR}/cache/$CURRENT_OS_ARCH"

KIND_VERSION=v0.7.0
KUBE_VERSION=v1.14.10
KIND_CACHE_PATH="${CACHE_DIR}/kind-$KIND_VERSION"
BATS_EXTRA_ARGS=""

# shellcheck disable=SC1090
source "${E2E_DIR}/lib/defer.bash"
trap run_deferred EXIT

if [[ -z "$HELM_VERSION" ]]; then
  echo "HELM_VERSION not set. Valid v2, v3. See $ROOT_DIR/docs/contributing/building.md"
  exit 1
fi

function install_kind() {
  if [ ! -f "${KIND_CACHE_PATH}" ]; then
    echo '>>> Downloading Kind'
    mkdir -p "${CACHE_DIR}"
    curl -sL "https://github.com/kubernetes-sigs/kind/releases/download/${KIND_VERSION}/kind-${CURRENT_OS_ARCH}" -o "${KIND_CACHE_PATH}"
  fi
  cp "${KIND_CACHE_PATH}" "${ROOT_DIR}/test/bin/kind"
  chmod +x "${ROOT_DIR}/test/bin/kind"
}

# Let users specify how many, e.g. with E2E_KIND_CLUSTER_NUM=3 make e2e
E2E_KIND_CLUSTER_NUM=${E2E_KIND_CLUSTER_NUM:-1}
KIND_CLUSTER_PREFIX=${KIND_CLUSTER_PREFIX:-helm-operator-e2e}

# Check if there is a kubernetes cluster running, otherwise use Kind
if ! kubectl version > /dev/null 2>&1; then
  install_kind

  # We require GNU Parallel, but some systems come with Tollef's parallel (moreutils)
  if ! parallel -h | grep -q "GNU Parallel"; then
    echo "GNU Parallel is not available on your system"; exit 1
  fi

  echo '>>> Creating Kind Kubernetes cluster(s)'
  KIND_CONFIG_PREFIX="${HOME}/.kube/kind-config-${KIND_CLUSTER_PREFIX}"
  for I in $(seq 1 "${E2E_KIND_CLUSTER_NUM}"); do
    defer kind --name "${KIND_CLUSTER_PREFIX}-${I}" delete cluster > /dev/null 2>&1 || true
    defer rm -rf "${KIND_CONFIG_PREFIX}-${I}"
    # Wire tests with the right cluster based on their BATS_JOB_SLOT env variable
    eval export "KUBECONFIG_SLOT_${I}=${KIND_CONFIG_PREFIX}-${I}"
  done
  # Due to https://github.com/kubernetes-sigs/kind/issues/1288
  # limit parallel creation of clusters to 1 job.
  seq 1 "${E2E_KIND_CLUSTER_NUM}" | time parallel -j 1 -- env KUBECONFIG="${KIND_CONFIG_PREFIX}-{}" kind create cluster --name "${KIND_CLUSTER_PREFIX}-{}" --wait 5m --image kindest/node:${KUBE_VERSION}

  echo '>>> Loading images into the Kind cluster(s)'
  seq 1 "${E2E_KIND_CLUSTER_NUM}" | time parallel -- kind --name "${KIND_CLUSTER_PREFIX}-{}" load docker-image 'docker.io/fluxcd/helm-operator:latest'
  if [ "${E2E_KIND_CLUSTER_NUM}" -gt 1 ]; then
    BATS_EXTRA_ARGS="--jobs ${E2E_KIND_CLUSTER_NUM}"
  fi
fi

echo '>>> Running the tests'
# Run all tests by default but let users specify which ones to run, e.g. with E2E_TESTS='11_*' make e2e
E2E_TESTS=${E2E_TESTS:-.}
HELM_VERSION=${HELM_VERSION:-}
(
  cd "${E2E_DIR}"
  export HELM_VERSION=${HELM_VERSION}
  # shellcheck disable=SC2086
  "${E2E_DIR}/bats/bin/bats" -t ${BATS_EXTRA_ARGS} ${E2E_TESTS}
)
