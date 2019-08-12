#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(realpath $(dirname ${BASH_SOURCE})/..)

# Grab code-generator version from go.sum.
CODEGEN_VERSION=$(grep 'k8s.io/code-generator' go.sum | awk '{print $2}' | head -1)
CODEGEN_PKG=$(echo `go env GOPATH`"/pkg/mod/k8s.io/code-generator@${CODEGEN_VERSION}")

echo ">> Using ${CODEGEN_PKG}"

# code-generator does work with go.mod but makes assumptions about
# the project living in `$GOPATH/src`. To work around this and support
# any location; create a temporary directory, use this as an output
# base, and copy everything back once generated.
TEMP_DIR=$(mktemp -d)
cleanup() {
    echo ">> Removing ${TEMP_DIR}"
    rm -rf ${TEMP_DIR}
}
trap "cleanup" EXIT SIGINT

echo ">> Temporary output directory ${TEMP_DIR}"

# Ensure we can execute.
chmod +x ${CODEGEN_PKG}/generate-groups.sh

# :magic:
${CODEGEN_PKG}/generate-groups.sh all \
    github.com/fluxcd/helm-operator/pkg/client github.com/fluxcd/helm-operator/pkg/apis \
    "flux.weave.works:v1beta1 helm.integrations.flux.weave.works:v1alpha2" \
    --output-base "${TEMP_DIR}" \
    --go-header-file "${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt"

# Copy everything back.
cp -r "${TEMP_DIR}/github.com/fluxcd/helm-operator/." "${SCRIPT_ROOT}/"
