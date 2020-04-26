# raw-k8s-tf-compiler

This is a tool to generate terraform resource blocks for kubernetes manifests to be used with the [terraform-provider-k8s plugin](https://github.com/banzaicloud/terraform-provider-k8s).

## Installation

### Curl binary from latest github release

#### On MacOS

```
latest_version=$(curl -s "https://api.github.com/repos/sebradloff/raw-k8s-tf-compiler/releases/latest" | grep '"tag_name":' | cut -d'"' -f4 | cut -c2-)
curl -sL "https://github.com/sebradloff/raw-k8s-tf-compiler/releases/download/v${latest_version}/raw-k8s-tf-compiler_${latest_version}_darwin_amd64.tar.gz" | tar xvz -f -

chmod +x rawk8stfc

mv rawk8stfc /usr/local/bin/rawk8stfc
```

#### On Linux

```
latest_version=$(curl -s "https://api.github.com/repos/sebradloff/raw-k8s-tf-compiler/releases/latest" | grep '"tag_name":' | cut -d'"' -f4 | cut -c2-)
curl -sL "https://github.com/sebradloff/raw-k8s-tf-compiler/releases/download/v${latest_version}/raw-k8s-tf-compiler_${latest_version}_linux_amd64.tar.gz" | tar xvz -f -

chmod +x rawk8stfc

mv rawk8stfc /usr/local/bin/rawk8stfc
```

### Go Get

Use go get to install the binary:

```
go get -u github.com/sebradloff/raw-k8s-tf-compiler
mv $GOBIN/raw-k8s-tf-compiler $GOBIN/rawk8stfc
```

## Usage

You provide the tool with a kubernetes manifest or direcetory of 
kubernetes manifests and it will generate a resource block for the 
[terraform-provider-k8s plugin](https://github.com/banzaicloud/terraform-provider-k8s).

Currently there are two options for generating the resource blocks; `inline` and `file-reference`.

The `inline` option generates resource blocks with each kubernetes 
resource injected as a heredoc in the content attribute of the 
k8s_manifest resource block.

`$ rawk8stfc inline -f deploy.yaml -o deploy.tf`

Example of generated file `deploy.tf`:
```hcl
resource "k8s_manifest" "default_apps-v1_Deployment_nginx-deployment" {
  content = <<EOT
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: nginx-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx
        ports:
        - containerPort: 80
EOT
}
```

The `file-reference` option generates resource blocks with each 
kubernetes manifest referenced with the terraform file function in the 
content attribute of the k8s_manifest resource block.

`$ rawk8stfc file-reference -f deploy.yaml -o deploy.tf`

Example of generated file `deploy.tf`:
```hcl
resource "k8s_manifest" "default_apps-v1_Deployment_nginx-deployment" {
  content = file("${path.module}/../testdata/k8s-files/one-obj.yaml")
}
```
