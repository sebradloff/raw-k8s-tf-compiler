#!/usr/bin/env bash

set -eu pipefail


DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

kind_cluster="integration-tests"


if [[ $(kind get clusters | grep "${kind_cluster}") == "" ]]; then
    echo "Creating kind cluster ${kind_cluster}"
    kind create cluster --image kindest/node:v1.18.0 --name "${kind_cluster}" --config kind.yaml
else
    echo "Using kind cluster ${kind_cluster} that already exists"
fi

kind export kubeconfig --name "${kind_cluster}"

# replace server block with master IP due to docker host networking
MASTER_IP=$(docker inspect --format='{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "${kind_cluster}"-control-plane)
sed -i "s/server:.*/server: https:\/\/$MASTER_IP:6443/" $HOME/.kube/config


set -x
k8s_file_dir="../k8s-files"
cd "${DIR}"

rawk8stfc inline -f "${k8s_file_dir}/one-obj.yaml" -o test-1.tf

terraform init
terraform apply -auto-approve
sleep 3
terraform destroy -auto-approve
rm test-1.tf


rawk8stfc inline -f "${k8s_file_dir}/multiple-objs.yaml" -o test-2.tf

terraform init
terraform apply -auto-approve
sleep 3
terraform destroy -auto-approve
rm test-2.tf

rawk8stfc file-reference -f "${k8s_file_dir}/one-obj.yaml" -o test-3.tf

terraform init
terraform apply -auto-approve
sleep 3
terraform destroy -auto-approve
rm test-3.tf


rawk8stfc file-reference -f "${k8s_file_dir}/multiple-objs.yaml" -o test-4.tf

terraform init
terraform apply -auto-approve
sleep 3
terraform destroy -auto-approve
rm test-4.tf