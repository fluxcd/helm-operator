#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/../
# This corresponds to the tag kubernetes-1.14.4, and to the pinned version in go.mod
CODEGEN_PKG=${CODEGEN_PKG:-$(echo `go env GOPATH`'/pkg/mod/k8s.io/code-generator@v0.0.0-20190311093542-50b561225d70')}

go mod download # make sure the code-generator is downloaded
chmod u+w ${CODEGEN_PKG}
env GOPATH=`go env GOPATH` bash ${CODEGEN_PKG}/generate-groups.sh all github.com/fluxcd/helm-operator/pkg/client \
              github.com/fluxcd/helm-operator/pkg/apis \
              "flux.weave.works:v1beta1 helm.integrations.flux.weave.works:v1alpha2" \
  --go-header-file "${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt"
