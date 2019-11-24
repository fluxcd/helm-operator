#!/usr/bin/env bash

export E2E_NAMESPACE=helm-operator-e2e
export DEMO_NAMESPACE=demo
ROOT_DIR=$(git rev-parse --show-toplevel)
export ROOT_DIR
export E2E_DIR="${ROOT_DIR}/test/e2e"
export FIXTURES_DIR="${E2E_DIR}/fixtures"
KNOWN_HOSTS=$(cat "${FIXTURES_DIR}/known_hosts")
export KNOWN_HOSTS
GITCONFIG=$(cat "${FIXTURES_DIR}/gitconfig")
export GITCONFIG
export HELM_VERSION=${HELM_VERSION}
