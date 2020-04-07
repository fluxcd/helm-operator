#!/usr/bin/env bash

# Copyright 2017 The Kubernetes Authors.
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

set -o errexit -o nounset -o pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"

TMP_DIFFROOT="${REPO_ROOT}/_tmp"
DIFFROOT="${REPO_ROOT}"

cleanup() {
  rm -rf "${TMP_DIFFROOT}"
}
trap "cleanup" EXIT SIGINT

cleanup

mkdir -p "${TMP_DIFFROOT}"
cp -a "${DIFFROOT}"/{pkg,docs,deploy} "${TMP_DIFFROOT}"

"${REPO_ROOT}/hack/update/generate-all.sh"

echo "diffing ${DIFFROOT} against freshly generated files"
ret=0
for i in {pkg,docs,deploy}; do
  diff -Naupr --no-dereference "${TMP_DIFFROOT}/${i}" "${DIFFROOT}/${i}" || ret=$?
done
cp -a "${TMP_DIFFROOT}"/{pkg,docs,deploy} "${DIFFROOT}"
if [[ $ret -eq 0 ]]
then
  echo "${DIFFROOT} up to date."
else
  echo "${DIFFROOT} is out of date. Please run hack/update/generate-all.sh"
  exit 1
fi
