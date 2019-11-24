#!/usr/bin/env bash

# shellcheck disable=SC1090
source "${E2E_DIR}/lib/defer.bash"
# shellcheck disable=SC1090
source "${E2E_DIR}/lib/template.bash"

function install_tiller() {
  if ! helm2 version > /dev/null 2>&1; then # only if helm isn't already installed
    kubectl --namespace kube-system create sa tiller
    kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
    helm2 init --service-account tiller --upgrade --wait
  fi
}

function uninstall_tiller() {
  helm2 reset --force
  kubectl delete clusterrolebinding tiller-cluster-rule
  kubectl --namespace kube-system delete sa tiller
}

function install_helm_operator_with_helm() {
  local create_crds='true'
  if kubectl get crd helmreleases.helm.fluxcd.io > /dev/null 2>&1; then
    echo 'CRD existed, disabling CRD creation'
    create_crds='false'
  fi

  helm2 install --name helm-operator --wait \
    --namespace "${E2E_NAMESPACE}" \
    --set createCRD="${create_crds}" \
    --set chartsSyncInterval=10s \
    --set image.repository=docker.io/fluxcd/helm-operator \
    --set image.tag=latest \
    --set git.pollInterval=10s \
    --set git.config.secretName=gitconfig \
    --set git.config.enabled=true \
    --set-string git.config.data="${GITCONFIG}" \
    --set git.ssh.secretName=flux-git-deploy \
    --set-string git.ssh.known_hosts="${KNOWN_HOSTS}" \
    --set configureRepositories.enable=true \
    --set configureRepositories.repositories[0].name="podinfo" \
    --set configureRepositories.repositories[0].url="https://stefanprodan.github.io/podinfo" \
    --set extraEnvs[0].name="HELM_VERSION" \
    --set extraEnvs[0].value="${HELM_VERSION:-v2\,v3}" \
    "${ROOT_DIR}/chart/helm-operator"
}

function uninstall_helm_operator_with_helm() {
  helm2 delete --purge helm-operator > /dev/null 2>&1
  kubectl delete crd helmreleases.helm.fluxcd.io > /dev/null 2>&1
}

function install_git_srv() {
  local external_access_result_var=${1}
  local kustomization_dir=${2:-base/gitsrv}
  local gen_dir
  gen_dir=$(mktemp -d)

  ssh-keygen -t rsa -N "" -f "$gen_dir/id_rsa"
  defer rm -rf "'$gen_dir'"
  kubectl create secret generic flux-git-deploy \
    --namespace="${E2E_NAMESPACE}" \
    --from-file="${FIXTURES_DIR}/known_hosts" \
    --from-file="$gen_dir/id_rsa" \
    --from-file=identity="$gen_dir/id_rsa" \
    --from-file="$gen_dir/id_rsa.pub"

  kubectl apply -n "${E2E_NAMESPACE}" -k "${FIXTURES_DIR}/kustom/${kustomization_dir}" >&3

  # Wait for the git server to be ready
  kubectl -n "${E2E_NAMESPACE}" rollout status deployment/gitsrv

  if [ -n "$external_access_result_var" ]; then
    local git_srv_podname
    git_srv_podname=$(kubectl get pod -n "${E2E_NAMESPACE}" -l name=gitsrv -o jsonpath="{['items'][0].metadata.name}")
    coproc kubectl port-forward -n "${E2E_NAMESPACE}" "$git_srv_podname" :22
    local local_port
    read -r local_port <&"${COPROC[0]}"-
    # shellcheck disable=SC2001
    local_port=$(echo "$local_port" | sed 's%.*:\([0-9]*\).*%\1%')
    local ssh_cmd="ssh -o UserKnownHostsFile=/dev/null  -o StrictHostKeyChecking=no -i $gen_dir/id_rsa -p $local_port"
    # return the ssh command needed for git, and the PID of the port-forwarding PID into a variable of choice
    eval "${external_access_result_var}=('$ssh_cmd' '$COPROC_PID')"
  fi
}

function uninstall_git_srv() {
  local secret_name=${1:-flux-git-deploy}
  # Silence secret deletion errors since the secret can be missing (deleted by uninstalling Flux)
  kubectl delete -n "${E2E_NAMESPACE}" secret "$secret_name" &> /dev/null
  kubectl delete -n "${E2E_NAMESPACE}" -f "${FIXTURES_DIR}/gitsrv.yaml"
}
