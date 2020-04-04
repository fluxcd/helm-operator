#!/usr/bin/env bash

# shellcheck disable=SC1090
source "${E2E_DIR}/lib/defer.bash"

function helm_binary() {
    case ${HELM_VERSION} in
        v2)
            helm2 --tiller-namespace "$E2E_NAMESPACE" "$@"
            ;;
        v3)
            helm3 "$@"
            ;;
        *)
            echo "No Helm binary found for version $HELM_VERSION" >&2
            return 1
    esac
}

function package_and_upload_chart() {
  local chart=${1}
  local chart_repository=${2}

  gen_dir=$(mktemp -d)
  defer rm -rf "'$gen_dir'"

  helm_binary package --destination "$gen_dir" "$chart"

  # Upload
  chart_tarbal=$(find "$gen_dir" -type f -name "*.tgz" | head -n1)
  curl --data-binary "@$chart_tarbal" "$chart_repository/api/charts"
}
