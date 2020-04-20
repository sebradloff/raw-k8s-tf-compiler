#!/usr/bin/env bash

set -eu pipefail

kind_cluster="integration-tests"


if [[ $(kind get clusters | grep "${kind_cluster}") == "" ]]; then
    echo "Creating kind cluster ${kind_cluster}"
    kind create cluster --image kindest/node:v1.17.0 --name "${kind_cluster}" --config kind.yaml
else
    echo "Using kind cluster ${kind_cluster} that already exists"
fi

kind export kubeconfig --name "${kind_cluster}"

# replace server block with master IP due to docker host networking
MASTER_IP=$(docker inspect --format='{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "${kind_cluster}"-control-plane)
sed -i "s/server:.*/server: https:\/\/$MASTER_IP:6443/" $HOME/.kube/config

terraform init
terraform plan
terraform apply -auto-approve