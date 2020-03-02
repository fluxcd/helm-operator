#!/usr/bin/env bash
# Kindly borrowed from Kind
# https://github.com/kubernetes-sigs/kind/blob/c5298b2/hack/update/generated.sh
#
# Copyright 2018 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# 'go generate's kind, using tools from vendor (go-bindata)
set -o errexit -o nounset -o pipefail

# cd to the repo root
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
cd "${REPO_ROOT}"

# build the generators using the tools module
cd "hack/tools"
"${REPO_ROOT}/hack/go_container.sh" go build -o /out/controller-gen sigs.k8s.io/controller-tools/cmd/controller-gen
# go back to the root
cd "${REPO_ROOT}"

# turn off module mode before running the generators
# https://github.com/kubernetes/code-generator/issues/69
# we also need to populate vendor
hack/go_container.sh go mod tidy
hack/go_container.sh go mod vendor
export GO111MODULE="off"

# fake being in a gopath
FAKE_GOPATH="$(mktemp -d)"
trap 'rm -rf ${FAKE_GOPATH}' EXIT

FAKE_REPOPATH="${FAKE_GOPATH}/src/github.com/fluxcd/helm-operator"
mkdir -p "$(dirname "${FAKE_REPOPATH}")" && ln -s "${REPO_ROOT}" "${FAKE_REPOPATH}"

export GOPATH="${FAKE_GOPATH}"
cd "${FAKE_REPOPATH}"

# run the generators
CRD_DIR="./chart/helm-operator/crds"
echo "Generate OpenAPI v3 schemas for chart CRDs"
bin/controller-gen \
  schemapatch:manifests="${CRD_DIR}" \
  output:dir="${CRD_DIR}" \
  paths=./pkg/apis/...

echo "Forging CRD template for \`pkg/install\` from generated chart CRDs"
out="./pkg/install/templates/crds.yaml.tmpl"
rm "$out" || true
touch "$out"
for file in $(find "${CRD_DIR}" -type f | sort -V); do
 # concatenate all files while removing blank (^$) lines
  printf -- "---\n" >> "$out"
  < "$file" sed '/^$$/d' >> "$out"
done
chmod 644 "$out"

# set module mode back, return to repo root
export GO111MODULE="on"
cd "${REPO_ROOT}"
